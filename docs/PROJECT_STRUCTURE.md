# Project Structure

This document describes the optimal project structure for the Liberty Unleashed Server Directory.

## Directory Layout

```
lusd/
├── .github/                  # GitHub workflows and templates
│   └── workflows/            # CI/CD pipeline definitions
├── cmd/                      # Main applications
│   └── lusd/                 # Liberty Unleashed Server Directory app
│       ├── main.go           # Main application entry point
│       └── main_test.go      # Application tests
├── configs/                  # Configuration files
│   ├── config.json           # Active configuration
│   └── config.example.json   # Example configuration template
├── docker/                   # Docker-related files
│   ├── Dockerfile            # Container build instructions
│   └── docker-compose.yml    # Multi-container orchestration
├── docs/                     # Documentation
│   ├── CHANGELOG.md          # Version history
│   └── ENVIRONMENT.md        # Environment variable docs
├── scripts/                  # Build and deployment scripts
│   ├── build.sh              # Unix build script
│   ├── build.bat             # Windows build script
│   └── deploy.sh             # Deployment automation
├── systemd/                  # Linux service files
│   └── lusd-server.service   # systemd service definition
├── .gitignore                # Git ignore rules
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
├── LICENSE                   # Project license
├── Makefile                  # Build automation
└── README.md                 # Project documentation
```

## Design Principles

### 1. **Standard Go Project Layout**
- `cmd/` contains main applications following Go conventions
- Tests are co-located with the code they test
- Configuration and documentation are properly organized

### 2. **Separation of Concerns**
- **Application Code**: `cmd/lusd/`
- **Configuration**: `configs/`
- **Documentation**: `docs/`
- **Infrastructure**: `docker/`, `systemd/`, `scripts/`

### 3. **Production Ready**
- Clear separation between development and deployment files
- Environment-specific configurations
- Automated build and deployment scripts
- Container support with proper file organization

### 4. **Maintainability**
- Logical grouping of related files
- Clear naming conventions
- Self-documenting structure
- Easy navigation for new contributors

## Benefits

1. **Professional Structure**: Follows industry standards and Go community conventions
2. **Scalability**: Easy to add new components or services
3. **Clear Ownership**: Each directory has a specific purpose
4. **Development Experience**: Developers can quickly understand the project layout
5. **CI/CD Friendly**: Build systems can easily locate required files
6. **Docker Optimization**: Clean separation enables efficient container builds

## Migration Notes

All file references have been updated in:
- Build scripts (`scripts/build.sh`, `scripts/build.bat`)
- Docker files (`docker/Dockerfile`, `docker/docker-compose.yml`)
- GitHub Actions (`.github/workflows/`)
- Documentation (`README.md`)
- Application code (static file serving)

The application maintains backward compatibility while providing a more professional and maintainable structure.
