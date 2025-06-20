# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Production-ready repository structure
- Docker support with multi-stage builds
- Docker Compose configuration
- Systemd service file for Linux deployment
- Build scripts for multiple platforms
- Makefile for easy building
- GitHub Actions CI/CD pipeline
- Health check endpoint (`/health`)
- Version endpoint (`/version`)
- Graceful shutdown handling
- Comprehensive README documentation
- Security scanning in CI pipeline
- Cross-platform build support

### Changed
- Improved error handling and logging
- Enhanced server structure with proper HTTP timeouts
- Better configuration management
- Updated .gitignore for production use

### Security
- Added security scanning with gosec and govulncheck
- Improved Docker security with non-root user
- Added dependency review in CI

## [1.0.0] - 2025-06-20

### Added
- Initial release of Liberty Unleashed Server Directory
- Dynamic server registration via HTTP POST
- Automatic cleanup of stale servers
- Official server support
- IP blacklisting functionality
- Configurable JSON settings
- Web interface for server listing
- RESTful API endpoints
- Logging support

### Features
- `/report.php` - Server registration endpoint
- `/servers.txt` - Active servers list
- `/official.txt` - Official servers list
- `/` - Web interface

### About Liberty Unleashed
Liberty Unleashed is a free online multiplayer modification for Grand Theft Auto 3, providing features such as custom vehicle placements, object placements, pickup placements, spawn points, and extensive scripting capabilities.
