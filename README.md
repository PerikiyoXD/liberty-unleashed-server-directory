# Liberty Unleashed Server Directory

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](Dockerfile)

A lightweight, high-performance HTTP server directory service for Liberty Unleashed game servers. This application maintains a dynamic list of active game servers and provides RESTful endpoints for server registration and discovery.

## 🚀 Features

- **Dynamic Server Registration**: Liberty Unleashed servers can report their availability via HTTP POST requests
- **Automatic Cleanup**: Stale servers are automatically removed after a configurable timeout period
- **Official Server Support**: Maintains a separate list of official/trusted servers
- **IP Blacklisting**: Built-in protection against unwanted servers
- **Health Monitoring**: Built-in health check and metrics endpoints
- **Configurable Settings**: JSON-based configuration with sensible defaults
- **Production Ready**: Docker support, systemd service, graceful shutdown
- **Enhanced Security**: User-Agent validation, request rate limiting, input validation, and secure file operations
- **Cross-Platform**: Supports Windows, Linux, and macOS

## 🔒 Security Features

This application implements comprehensive security measures:

- **Input Validation**: All user inputs are validated and sanitized
- **Rate Limiting**: 60 requests per minute per IP to prevent abuse
- **Secure File Operations**: Path traversal protection and file size limits
- **Security Headers**: HTTP security headers to prevent common attacks
- **Error Handling**: Generic error messages to prevent information disclosure
- **Configuration Security**: Validation of all configuration parameters
- **Dependency Integrity**: Go modules with checksum verification

For detailed security information, see [SECURITY.md](docs/SECURITY.md) and [SECURITY_CHECKLIST.md](docs/SECURITY_CHECKLIST.md).

## 📋 Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Docker Deployment](#docker-deployment)
- [Building from Source](#building-from-source)
- [Production Deployment](#production-deployment)
- [Contributing](#contributing)
- [License](#license)

## 🏗️ Installation

### Quick Start with Docker

```bash
# Clone the repository
git clone https://github.com/PerikiyoXD/liberty-unleashed-server-directory
cd lusd

# Run with Docker Compose
docker-compose up -d
```

### Binary Installation

Download the latest release for your platform from the [releases page](https://github.com/PerikiyoXD/liberty-unleashed-server-directory/releases).

#### Linux/macOS
```bash
# Clone the repository
git clone https://github.com/PerikiyoXD/liberty-unleashed-server-directory
cd lusd

# Download and install
wget https://github.com/PerikiyoXD/liberty-unleashed-server-directory/releases/latest/download/lusd-linux-amd64
chmod +x lusd-linux-amd64
sudo mv lusd-linux-amd64 /usr/local/bin/lusd

# Create config directory and copy example config
sudo mkdir -p /etc/lusd
sudo cp configs/config.example.json /etc/lusd/config.json

# Run
lusd
```

#### Windows
1. Download `lusd-windows-amd64.exe`
2. Copy `configs/config.example.json` to `config.json`
3. Run `lusd-windows-amd64.exe`

## ⚙️ Configuration

Create a `configs/config.json` file based on `configs/config.example.json`:

```json
{
  "port": 80,
  "allowedUserAgent": "LU-Server/0.1",
  "staleTimeout": "2m40s",
  "blacklist": [
    "31.220.49.160",
    "138.68.97.15"
  ],
  "officialServers": [
    "12.141.44.231:8001",
    "12.141.44.231:9000"
  ],
  "logFile": "lusd_server.log",
  "logEnabled": true
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `port` | int | 80 | Port to listen on |
| `allowedUserAgent` | string | "LU-Server/0.1" | Required User-Agent for server registration |
| `staleTimeout` | string | "10m" | Time after which servers are considered stale |
| `blacklist` | array | [] | List of blocked IP addresses |
| `officialServers` | array | [] | List of official servers (always shown) |
| `logFile` | string | "lusd_server.log" | Log file path |
| `logEnabled` | bool | true | Enable/disable file logging |

## 🚀 Usage

### Starting the Server

```bash
# Using binary
./lusd

# Using Docker
docker run -p 80:80 -v $(pwd)/config.json:/app/config.json lusd:latest

# Using Docker Compose
docker-compose -f docker/docker-compose.yml up -d
```

### Server Registration

Liberty Unleashed servers can register themselves by sending a POST request:

```bash
curl -X POST http://your-directory-server/report.php \
  -H "User-Agent: LU-Server/0.1" \
  -d "port=2301"
```

## 📡 API Endpoints

### Core Endpoints

| Endpoint | Method | Description |
|----------|---------|-------------|
| `/servers.txt` | GET | List of active servers (plain text) |
| `/official.txt` | GET | List of official servers (plain text) |
| `/report.php` | POST | Server registration endpoint |

### Monitoring Endpoints

| Endpoint | Method | Description |
|----------|---------|-------------|
| `/health` | GET | Health check with system information |
| `/version` | GET | Version and build information |

### Health Check Response

```json
{
  "status": "ok",
  "version": "v1.0.0",
  "buildTime": "2025-06-20_12:00:00",
  "commit": "abc1234",
  "timestamp": 1718881200,
  "uptime": 3661.5,
  "activeServers": 5
}
```

## 🐳 Docker Deployment

### Using Docker Compose (Recommended)

```yaml
version: '3.8'

services:
  lusd-server:
    build: .
    ports:
      - "80:80"
    volumes:
      - ./config.json:/app/config.json:ro
      - ./logs:/app/logs
    restart: unless-stopped
```

# Using Docker CLI

```bash
# Build image
docker build -f docker/Dockerfile -t lusd:latest .

# Run container
docker run -d \
  --name lusd-server \
  -p 80:80 \
  -v $(pwd)/configs/config.json:/app/config.json:ro \
  --restart unless-stopped \
  lusd:latest
```

## 🔨 Building from Source

### Prerequisites

- Go 1.23 or later
- Make (optional, for using Makefile)

### Build Commands

```bash
# Clone repository
git clone https://github.com/PerikiyoXD/liberty-unleashed-server-directory
cd lusd

# Build for current platform
make build
# or
go build -o lusd main.go

# Build for all platforms
make build-all

# Build with custom version
VERSION=v1.0.0 make build
```

### Available Make Targets

```bash
make help              # Show available targets
make build             # Build for current platform
make build-all         # Build for all platforms
make clean             # Clean build artifacts
make test              # Run tests
make lint              # Run linter
make docker-build      # Build Docker image
make docker-run        # Run Docker container
make install           # Install locally (Linux/macOS)
```

## 🏭 Production Deployment

### Linux (systemd)

1. **Install the binary**:
   ```bash
   sudo cp lusd /opt/lusd/
   sudo chown root:root /opt/lusd/lusd
   sudo chmod +x /opt/lusd/lusd
   ```

2. **Install systemd service**:
   ```bash
   sudo cp systemd/lusd-server.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable lusd-server
   sudo systemctl start lusd-server
   ```

3. **Check status**:
   ```bash
   sudo systemctl status lusd-server
   ```

### Using the Deployment Script

```bash
# Deploy to remote server
./scripts/deploy.sh your-server.com lusd-user

# Deploy locally
./scripts/deploy.sh localhost $USER
```

### Nginx Reverse Proxy

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## 📊 Monitoring and Logging

### Health Monitoring

The server exposes a `/health` endpoint for monitoring:

```bash
# Check health
curl http://localhost/health

# Monitor with curl
watch -n 30 'curl -s http://localhost/health | jq .'
```

### Log Management

Logs are written to both console and file (if enabled):

```bash
# View logs (systemd)
sudo journalctl -u lusd-server -f

# View log file
tail -f lusd_server.log

# Rotate logs (add to crontab)
0 0 * * * /usr/sbin/logrotate /etc/logrotate.d/lusd-server
```

## 🛡️ Security Considerations

1. **Firewall**: Only expose necessary ports
2. **User Agent Validation**: Configure a secure User-Agent string
3. **Blacklisting**: Regularly update the IP blacklist
4. **Rate Limiting**: Consider adding a reverse proxy with rate limiting
5. **HTTPS**: Use HTTPS in production (via reverse proxy)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/lusd.git
cd lusd

# Install dependencies
go mod download

# Run tests
make test

# Run linter
make lint

# Build and test
make build
./lusd
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Liberty Unleashed](https://www.liberty-unleashed.co.uk/) - The GTA3 multiplayer modification
- Go community for excellent tooling and libraries

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/PerikiyoXD/liberty-unleashed-server-directory/issues)
- **Wiki**: [GitHub Wiki](https://github.com/PerikiyoXD/liberty-unleashed-server-directory/wiki)
- **Discord**: [Liberty Unleashed Discord](https://discord.gg/liberty-unleashed)

---

**About Liberty Unleashed**

Liberty Unleashed is a free online multiplayer modification for Grand Theft Auto 3. Its main goals are to provide players with a great multiplayer experience by offering features such as:

- Custom Vehicle Placements
- Custom Object Placements  
- Custom Pickup Placements
- Custom Spawn Points
- Game modification options (time, weather, gravity)
- Both client-side and server-side scripting capabilities

*Note: Liberty Unleashed requires GTA3 v1.1*
