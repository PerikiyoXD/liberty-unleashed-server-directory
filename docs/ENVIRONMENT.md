# Environment Variables

You can override configuration settings using environment variables:

## Available Environment Variables

| Environment Variable | Config Field | Default | Description |
|---------------------|-------------|---------|-------------|
| `LUSD_PORT` | port | 80 | Port to listen on |
| `LUSD_USER_AGENT` | allowedUserAgent | "LU-Server/0.1" | Required User-Agent for registration |
| `LUSD_STALE_TIMEOUT` | staleTimeout | "10m" | Server stale timeout |
| `LUSD_LOG_FILE` | logFile | "lusd_server.log" | Log file path |
| `LUSD_LOG_ENABLED` | logEnabled | true | Enable/disable logging |

## Example Usage

### Docker
```bash
docker run -d \
  -p 80:80 \
  -e LUSD_PORT=8080 \
  -e LUSD_LOG_ENABLED=true \
  lusd:latest
```

### Docker Compose
```yaml
version: '3.8'
services:
  lusd-server:
    image: lusd:latest
    environment:
      - LUSD_PORT=8080
      - LUSD_LOG_ENABLED=true
    ports:
      - "8080:8080"
```

### Command Line
```bash
# Linux/macOS
export LUSD_PORT=8080
export LUSD_LOG_ENABLED=true
./lusd

# Windows
set LUSD_PORT=8080
set LUSD_LOG_ENABLED=true
lusd.exe
```

### Systemd Service
```ini
[Service]
Environment=LUSD_PORT=8080
Environment=LUSD_LOG_ENABLED=true
ExecStart=/opt/lusd/lusd
```

## Priority Order

Configuration values are applied in the following order (later values override earlier ones):

1. Default values (hardcoded)
2. Configuration file (`config.json`)
3. Environment variables
4. Command line flags (if implemented)

## Note

Environment variables are checked when the application starts. Changes to environment variables after startup require a restart to take effect.
