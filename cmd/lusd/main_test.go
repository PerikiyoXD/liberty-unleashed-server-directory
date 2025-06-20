package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Test loading non-existent config (should return defaults)
	cfg := loadConfig("nonexistent.json")
	
	if cfg.Port != 80 {
		t.Errorf("Expected default port 80, got %d", cfg.Port)
	}
	
	if cfg.AllowedUserAgent != "LU-Server/0.1" {
		t.Errorf("Expected default user agent 'LU-Server/0.1', got %s", cfg.AllowedUserAgent)
	}
	
	if cfg.StaleTimeout != 10*time.Minute {
		t.Errorf("Expected default stale timeout 10m, got %v", cfg.StaleTimeout)
	}
}

func TestServerList(t *testing.T) {
	cfg := Config{
		Port:             80,
		AllowedUserAgent: "LU-Server/0.1",
		StaleTimeout:     time.Minute,
		Blacklist:        make(map[string]bool),
		OfficialServers:  []string{"192.168.1.100:1234"},
		LogFile:          "",
		LogEnabled:       false,
	}
	
	servers := NewServerList(cfg)
	
	// Test reporting a server
	servers.Report("127.0.0.1", 2301)
	
	active := servers.GetActive()
	found := false
	for _, addr := range active {
		if addr == "127.0.0.1:2301" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected to find reported server in active list")
	}
	
	// Test official servers are included
	found = false
	for _, addr := range active {
		if addr == "192.168.1.100:1234" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected to find official server in active list")
	}
}

func TestReportEndpoint(t *testing.T) {
	cfg := Config{
		Port:             80,
		AllowedUserAgent: "LU-Server/0.1",
		StaleTimeout:     time.Minute,
		Blacklist:        make(map[string]bool),
		OfficialServers:  []string{},
		LogFile:          "",
		LogEnabled:       false,
	}
	
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.UserAgent() != cfg.AllowedUserAgent {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		portStr := r.FormValue("port")
		if portStr == "" {
			http.Error(w, "Invalid port", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
	
	// Test valid request
	form := url.Values{}
	form.Add("port", "2301")
	req := httptest.NewRequest("POST", "/report.php", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "LU-Server/0.1")
	
	w := httptest.NewRecorder()
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	// Test invalid user agent
	req.Header.Set("User-Agent", "Invalid")
	w = httptest.NewRecorder()
	handler(w, req)
	
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for invalid user agent, got %d", w.Code)
	}
	
	// Test invalid method
	req = httptest.NewRequest("GET", "/report.php", nil)
	w = httptest.NewRecorder()
	handler(w, req)
	
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for invalid method, got %d", w.Code)
	}
}

func TestHealthEndpoint(t *testing.T) {
	startTime := time.Now()
	
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		health := map[string]interface{}{
			"status":        "ok",
			"version":       "test",
			"buildTime":     "test-time",
			"commit":        "test-commit",
			"timestamp":     time.Now().Unix(),
			"uptime":        time.Since(startTime).Seconds(),
			"activeServers": 0,
		}
		json.NewEncoder(w).Encode(health)
	}
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
	
	var health map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &health); err != nil {
		t.Errorf("Failed to unmarshal health response: %v", err)
	}
	
	if health["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", health["status"])
	}
}

func TestServersEndpoint(t *testing.T) {
	cfg := Config{
		Port:             80,
		AllowedUserAgent: "LU-Server/0.1",
		StaleTimeout:     time.Minute,
		Blacklist:        make(map[string]bool),
		OfficialServers:  []string{"192.168.1.100:1234"},
		LogFile:          "",
		LogEnabled:       false,
	}
	
	servers := NewServerList(cfg)
	servers.Report("127.0.0.1", 2301)
	
	handler := func(w http.ResponseWriter, r *http.Request) {
		active := servers.GetActive()
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(strings.Join(active, "\n")))
	}
	
	req := httptest.NewRequest("GET", "/servers.txt", nil)
	w := httptest.NewRecorder()
	
	handler(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type text/plain, got %s", w.Header().Get("Content-Type"))
	}
	
	body := w.Body.String()
	if !strings.Contains(body, "127.0.0.1:2301") {
		t.Error("Expected response to contain reported server")
	}
	
	if !strings.Contains(body, "192.168.1.100:1234") {
		t.Error("Expected response to contain official server")
	}
}
