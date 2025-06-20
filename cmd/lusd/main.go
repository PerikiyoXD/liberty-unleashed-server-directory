package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Security constants
const (
	maxConfigFileSize = 1024 * 1024 * 32  // 32MB max config file size
	maxLogFileSize    = 100 * 1024 * 1024 // 100MB max log file size
	configFileMode    = 0600              // Owner read/write only
	logFileMode       = 0644              // Owner read/write, group/others read only
)

// Build information (set by build flags)
var (
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "unknown"
)

type Config struct {
	Port             int           `json:"port"`
	AllowedUserAgent string        `json:"allowedUserAgent"`
	StaleTimeout     time.Duration `json:"staleTimeout"`
	Blacklist        map[string]bool
	OfficialServers  []string
	LogFile          string
	LogEnabled       bool
}

// jsonConfig represents the structure of the config.json file
type jsonConfig struct {
	Port             int      `json:"port"`
	AllowedUserAgent string   `json:"allowedUserAgent"`
	StaleTimeout     string   `json:"staleTimeout"`
	Blacklist        []string `json:"blacklist"`
	OfficialServers  []string `json:"officialServers"`
	LogFile          string   `json:"logFile"`
	LogEnabled       bool     `json:"logEnabled"`
}

type ServerList struct {
	sync.Mutex
	Entries map[string]int64
	Config  Config
}

func NewServerList(cfg Config) *ServerList {
	s := &ServerList{
		Entries: make(map[string]int64),
		Config:  cfg,
	}
	go s.cleanupLoop()
	return s
}

func (s *ServerList) Report(ip string, port int) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	s.Lock()
	defer s.Unlock()
	s.Entries[addr] = time.Now().Unix()
}

func (s *ServerList) GetActive() []string {
	cutoff := time.Now().Add(-s.Config.StaleTimeout).Unix()
	s.Lock()
	defer s.Unlock()

	// Use a map to avoid duplicates
	activeMap := make(map[string]bool)

	// Add all non-stale servers from reported entries
	for addr, ts := range s.Entries {
		if ts >= cutoff {
			activeMap[addr] = true
		}
	}
	// Add all official servers
	for _, addr := range s.Config.OfficialServers {
		activeMap[addr] = true
	}

	// Convert to sorted slice
	var list []string
	for addr := range activeMap {
		list = append(list, addr)
	}
	sort.Strings(list)
	return list
}

func (s *ServerList) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-s.Config.StaleTimeout).Unix()
		s.Lock()
		for addr, ts := range s.Entries {
			if ts < cutoff {
				log.Printf("Removing stale server: %s (last seen at %d)", addr, ts)
				delete(s.Entries, addr)
			}
		}
		s.Unlock()
	}
}

// loadConfig attempts to load configuration from a JSON file.
// Falls back to default configuration if file not found or invalid.
func loadConfig(configPath string) Config { // Default configuration
	defaultCfg := Config{
		Port:             80,
		AllowedUserAgent: "LU-Server/0.1",
		StaleTimeout:     10 * time.Minute,
		Blacklist:        map[string]bool{},
		OfficialServers:  []string{},
		LogFile:          "lusd_server.log",
		LogEnabled:       true,
	}

	// Validate config path
	configPath = filepath.Clean(configPath)
	if strings.Contains(configPath, "..") {
		log.Printf("Invalid config path detected, using defaults")
		return defaultCfg
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Config file not found, creating default config")

		// Create default JSON config
		defaultJsonCfg := jsonConfig{
			Port:             defaultCfg.Port,
			AllowedUserAgent: defaultCfg.AllowedUserAgent,
			StaleTimeout:     "10m",
			LogFile:          defaultCfg.LogFile,
			LogEnabled:       defaultCfg.LogEnabled,
		}

		// Convert blacklist map to slice
		for ip := range defaultCfg.Blacklist {
			defaultJsonCfg.Blacklist = append(defaultJsonCfg.Blacklist, ip)
		}

		// Convert to JSON with indentation
		jsonData, err := json.MarshalIndent(defaultJsonCfg, "", "  ")
		if err != nil {
			log.Printf("Error creating default config, using defaults")
			return defaultCfg
		}

		// Write to file with secure permissions
		if err := secureWriteFile(configPath, jsonData, configFileMode); err != nil {
			log.Printf("Error writing default config, using defaults")
		} else {
			log.Printf("Created default config")
		}

		return defaultCfg
	}

	// Read the config file with security checks
	data, err := secureReadFile(configPath, maxConfigFileSize)
	if err != nil {
		log.Printf("Error reading config file, using defaults")
		return defaultCfg
	}
	// Parse the JSON
	var jsonCfg jsonConfig
	if err := json.Unmarshal(data, &jsonCfg); err != nil {
		log.Printf("Error parsing config file, using defaults")
		return defaultCfg
	}
	// Convert JSON config to internal config
	cfg := Config{
		Port:             jsonCfg.Port,
		AllowedUserAgent: jsonCfg.AllowedUserAgent,
		Blacklist:        make(map[string]bool),
		OfficialServers:  jsonCfg.OfficialServers,
		LogFile:          jsonCfg.LogFile,
		LogEnabled:       jsonCfg.LogEnabled,
	}

	// Parse stale timeout
	if duration, err := time.ParseDuration(jsonCfg.StaleTimeout); err != nil {
		log.Printf("Invalid staleTimeout format, using default")
		cfg.StaleTimeout = defaultCfg.StaleTimeout
	} else {
		cfg.StaleTimeout = duration
	}

	// Parse blacklist with validation
	for _, ip := range jsonCfg.Blacklist {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		// Validate IP address format
		if net.ParseIP(ip) == nil {
			log.Printf("Skipping invalid IP in blacklist: %s", ip)
			continue
		}
		cfg.Blacklist[ip] = true
	}

	// Clean up official servers list - remove empty entries and validate IPs
	var validOfficialServers []string
	for _, addr := range jsonCfg.OfficialServers {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}

		// Check if it's a valid IP address
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr // If not in host:port format, use the whole string
		}

		if net.ParseIP(host) == nil {
			log.Printf("Skipping official server: not a valid IP")
			continue
		}

		validOfficialServers = append(validOfficialServers, addr)
	}
	cfg.OfficialServers = validOfficialServers

	// Validate port
	if cfg.Port < 1 || cfg.Port > 65535 {
		log.Printf("Invalid port number, using default")
		cfg.Port = defaultCfg.Port
	}

	// Validate allowed user agent
	if cfg.AllowedUserAgent == "" {
		log.Printf("Empty allowedUserAgent, using default")
		cfg.AllowedUserAgent = defaultCfg.AllowedUserAgent
	}

	// Validate log file
	if cfg.LogFile == "" {
		log.Printf("Empty logFile, using default")
		cfg.LogFile = defaultCfg.LogFile
	}
	log.Printf("Successfully loaded config")
	// Override with environment variables if present (with validation)
	if port := os.Getenv("LUSD_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 && p <= 65535 {
			cfg.Port = p
			log.Printf("Port overridden by environment variable")
		} else {
			log.Printf("Invalid LUSD_PORT environment variable, ignoring")
		}
	}

	if userAgent := os.Getenv("LUSD_USER_AGENT"); userAgent != "" {
		// Validate user agent string (basic validation)
		if len(userAgent) > 0 && len(userAgent) <= 100 {
			cfg.AllowedUserAgent = userAgent
			log.Printf("User agent overridden by environment variable")
		} else {
			log.Printf("Invalid LUSD_USER_AGENT environment variable, ignoring")
		}
	}

	if timeout := os.Getenv("LUSD_STALE_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil && duration > 0 {
			cfg.StaleTimeout = duration
			log.Printf("Stale timeout overridden by environment variable")
		} else {
			log.Printf("Invalid LUSD_STALE_TIMEOUT environment variable, ignoring")
		}
	}

	if logFile := os.Getenv("LUSD_LOG_FILE"); logFile != "" {
		// Validate log file path
		if len(logFile) > 0 && len(logFile) <= 255 && !strings.Contains(logFile, "..") {
			cfg.LogFile = logFile
			log.Printf("Log file overridden by environment variable")
		} else {
			log.Printf("Invalid LUSD_LOG_FILE environment variable, ignoring")
		}
	}

	if logEnabled := os.Getenv("LUSD_LOG_ENABLED"); logEnabled != "" {
		if enabled, err := strconv.ParseBool(logEnabled); err == nil {
			cfg.LogEnabled = enabled
			log.Printf("Log enabled overridden by environment variable")
		} else {
			log.Printf("Invalid LUSD_LOG_ENABLED environment variable, ignoring")
		}
	}

	return cfg
}

// secureReadFile safely reads a file with size limits and path validation
func secureReadFile(path string, maxSize int64) ([]byte, error) {
	// Validate and clean the path
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return nil, fmt.Errorf("invalid file path")
	}

	// Check if path is absolute and within expected boundaries
	if filepath.IsAbs(cleanPath) {
		// For absolute paths, ensure they don't escape to system directories
		if strings.Contains(cleanPath, "..") {
			return nil, fmt.Errorf("path traversal detected")
		}
	}

	// Check file info first
	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("file access error")
	}

	// Check file size
	if info.Size() > maxSize {
		return nil, fmt.Errorf("file too large")
	}

	// Read file
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("file read error")
	}

	return data, nil
}

// secureWriteFile safely writes a file with proper permissions and path validation
func secureWriteFile(path string, data []byte, perm os.FileMode) error {
	// Validate and clean the path
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return fmt.Errorf("invalid file path")
	}

	// Check for path traversal
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal detected")
	}

	// Ensure parent directory exists
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory")
	}

	// Write file with secure permissions
	if err := os.WriteFile(cleanPath, data, perm); err != nil {
		return fmt.Errorf("file write error")
	}

	return nil
}

// secureOpenFile safely opens a file for logging with proper permissions
func secureOpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	// Validate and clean the path
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return nil, fmt.Errorf("invalid file path")
	}

	// Check for path traversal
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("path traversal detected")
	}

	// If file exists, check its size for log rotation
	if info, err := os.Stat(cleanPath); err == nil {
		if info.Size() > maxLogFileSize {
			// Rotate log file
			backupPath := cleanPath + ".old"
			if err := os.Rename(cleanPath, backupPath); err != nil {
				log.Printf("Warning: Could not rotate log file: %v", err)
			}
		}
	}

	// Open file with secure permissions
	file, err := os.OpenFile(cleanPath, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("file open error")
	}

	return file, nil
}

// validateConfigPath ensures config file path is safe
func validateConfigPath(execPath string) string {
	// Always place config file next to executable for security
	execDir := filepath.Dir(execPath)
	configPath := filepath.Join(execDir, "config.json")

	// Ensure the path is clean and doesn't contain traversal attempts
	return filepath.Clean(configPath)
}

// validateLogPath ensures log file path is safe
func validateLogPath(logFile, execPath string) (string, error) {
	if logFile == "" {
		return "", fmt.Errorf("empty log file path")
	}

	var logPath string
	if filepath.IsAbs(logFile) {
		// For absolute paths, ensure they're within reasonable bounds
		logPath = filepath.Clean(logFile)

		// Prevent writing to system directories
		systemDirs := []string{"/etc", "/bin", "/sbin", "/usr/bin", "/usr/sbin", "C:\\Windows", "C:\\Program Files"}
		for _, sysDir := range systemDirs {
			if strings.HasPrefix(strings.ToLower(logPath), strings.ToLower(sysDir)) {
				return "", fmt.Errorf("cannot write logs to system directory")
			}
		}
	} else {
		// For relative paths, place next to executable
		execDir := filepath.Dir(execPath)
		logPath = filepath.Join(execDir, logFile)
	}

	// Final validation
	logPath = filepath.Clean(logPath)
	if strings.Contains(logPath, "..") {
		return "", fmt.Errorf("path traversal detected in log path")
	}

	return logPath, nil
}

func main() {
	// Record start time for uptime calculation
	startTime := time.Now()

	// Determine config file path - use secure path validation
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("Warning: Could not determine executable path, using current directory")
		execPath = "."
	}
	configPath := validateConfigPath(execPath)

	// Load configuration
	cfg := loadConfig(configPath)

	// Setup logging to file if enabled with security checks
	if cfg.LogEnabled && cfg.LogFile != "" {
		logFilePath, err := validateLogPath(cfg.LogFile, execPath)
		if err != nil {
			log.Printf("Error validating log file path: %v, continuing with console logging only", err)
		} else {
			logFile, err := secureOpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFileMode)
			if err != nil {
				log.Printf("Error opening log file: %v, continuing with console logging only", err)
			} else {
				log.Printf("Logging to file enabled")
				multiWriter := io.MultiWriter(os.Stdout, logFile)
				log.SetOutput(multiWriter)
			}
		}
	}

	// Add security headers and input validation to HTTP handlers
	servers := NewServerList(cfg)

	// Rate limiting map (simple in-memory rate limiting)
	var rateLimitMutex sync.Mutex
	rateLimitMap := make(map[string][]int64)
	const maxRequestsPerMinute = 60

	// Helper function to check rate limits
	checkRateLimit := func(ip string) bool {
		rateLimitMutex.Lock()
		defer rateLimitMutex.Unlock()

		now := time.Now().Unix()
		minute := now / 60

		if times, exists := rateLimitMap[ip]; exists {
			// Remove old entries
			var newTimes []int64
			for _, t := range times {
				if t >= minute-1 { // Keep last 2 minutes
					newTimes = append(newTimes, t)
				}
			}
			rateLimitMap[ip] = newTimes

			// Count requests in current minute
			count := 0
			for _, t := range newTimes {
				if t == minute {
					count++
				}
			}

			if count >= maxRequestsPerMinute {
				return false
			}
		}

		// Add current request
		rateLimitMap[ip] = append(rateLimitMap[ip], minute)
		return true
	}

	// Security middleware
	securityMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Add security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Get client IP
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}

			// Check rate limit
			if !checkRateLimit(ip) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next(w, r)
		}
	}

	http.HandleFunc("/report.php", securityMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.UserAgent() != cfg.AllowedUserAgent {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Parse form with size limit
		r.Body = http.MaxBytesReader(w, r.Body, 1024) // 1KB limit
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		portStr := r.FormValue("port")
		if portStr == "" {
			http.Error(w, "Missing port parameter", http.StatusBadRequest)
			return
		}

		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1024 || port > 65535 {
			http.Error(w, "Invalid port", http.StatusBadRequest)
			return
		}

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil || cfg.Blacklist[ip] {
			// Silent drop for blacklisted IPs
			w.WriteHeader(http.StatusOK)
			return
		}

		// Validate IP address
		if net.ParseIP(ip) == nil {
			http.Error(w, "Invalid IP address", http.StatusBadRequest)
			return
		}
		log.Printf("Received report from %s:%d", ip, port)

		servers.Report(ip, port)
		w.WriteHeader(http.StatusOK)
	}))

	http.HandleFunc("/servers.txt", securityMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		active := servers.GetActive()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		_, _ = w.Write([]byte(strings.Join(active, "\n")))
	}))

	http.HandleFunc("/official.txt", securityMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		_, _ = w.Write([]byte(strings.Join(servers.Config.OfficialServers, "\n")))
	}))

	// Health check endpoint
	http.HandleFunc("/health", securityMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		health := map[string]interface{}{
			"status":        "ok",
			"version":       Version,
			"timestamp":     time.Now().Unix(),
			"uptime":        time.Since(startTime).Seconds(),
			"activeServers": len(servers.GetActive()),
		}
		json.NewEncoder(w).Encode(health)
	}))

	// Version endpoint
	http.HandleFunc("/version", securityMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		version := map[string]string{
			"version": Version,
		}
		json.NewEncoder(w).Encode(version)
	}))
	// Create HTTP server with security timeouts and limits
	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:           addr,
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %d...", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
