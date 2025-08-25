# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go project implementing a multi-tenant Neo4j proxy server. The proxy allows multiple tenants to connect through a single endpoint while routing requests to different Neo4j backend instances based on tenant identification. This solves the limitation that open-source Neo4j doesn't support multi-tenancy (only available in enterprise versions).

## Project Structure

```
neo4j-proxy/
├── cmd/neo4j-proxy/          # Main application entry point
│   └── main.go
├── pkg/                      # Public packages
│   ├── bolt/                 # Bolt protocol implementation
│   ├── config/               # Configuration management
│   └── proxy/                # Proxy server core
├── internal/                 # Private packages
│   ├── auth/                 # Authentication and tenant extraction
│   └── router/               # Connection routing logic
├── test/                     # Ginkgo BDD tests
├── .github/workflows/        # GitHub Actions CI/CD
└── Makefile                  # Build and development commands
```

## Development Commands

### Using Make (Recommended)
- `make help` - Show all available make targets
- `make build` - Build the Neo4j proxy binary
- `make test` - Run all tests with coverage and JUnit reporting
- `make test-verbose` - Run tests with verbose output
- `make test-race` - Run tests with race detection
- `make coverage` - Generate and display coverage report
- `make coverage-html` - Generate HTML coverage report
- `make lint` - Run golangci-lint
- `make fmt` - Format Go code
- `make vet` - Run go vet
- `make run` - Build and run the proxy
- `make dev` - Run in development mode
- `make clean` - Clean build artifacts
- `make install-tools` - Install development tools
- `make check` - Run all quality checks (fmt, vet, lint, test)

### Direct Go Commands
- `go build ./cmd/neo4j-proxy` - Build the main binary
- `go run ./cmd/neo4j-proxy` - Run the proxy directly
- `go mod tidy` - Clean up module dependencies
- `go vet ./...` - Run Go vet for code analysis
- `go fmt ./...` - Format Go code

### Testing with Ginkgo
- `ginkgo -r` - Run all tests recursively
- `ginkgo -r --cover --junit-report=junit.xml` - Run tests with coverage and JUnit output
- `ginkgo -r -v` - Run tests with verbose output
- `ginkgo -r --race` - Run tests with race detection

## Architecture

### Multi-tenant Proxy Design
The proxy implements a layered architecture:

1. **Connection Handler** (`pkg/proxy/proxy.go`): Accepts client connections and manages the proxy lifecycle
2. **Bolt Protocol Parser** (`pkg/bolt/protocol.go`): Handles Neo4j's Bolt protocol handshake and message parsing
3. **Authentication & Tenant Extraction** (`internal/auth/auth.go`): Determines tenant routing based on username patterns or metadata
4. **Router** (`internal/router/router.go`): Routes connections to appropriate backend Neo4j instances
5. **Configuration** (`pkg/config/config.go`): Manages tenant-to-backend mappings

### Tenant Identification Strategies
- **Username-based**: Extract tenant from username format like `tenant1@user`
- **Database-based**: Extract tenant from database name in connection metadata
- **Metadata-based**: Extract tenant from connection metadata fields

### Configuration
Default configuration supports test environment:
- `tenant1` -> `yunhorn187:17687`
- `tenant2` -> `yunhorn187:27687`

Configuration can be loaded from JSON file via `CONFIG_FILE` environment variable.

## Key Features
- Bolt protocol compatibility (versions 1-4)
- Multiple tenant extraction strategies
- Thread-safe connection routing
- Graceful shutdown handling
- Comprehensive test coverage with Ginkgo BDD
- CI/CD pipeline with GitHub Actions
- JUnit test reporting
- Coverage reporting

## Testing Strategy
The project uses Ginkgo for BDD-style testing with comprehensive coverage:
- Unit tests for all major components
- Integration tests for proxy functionality
- Thread-safety tests for concurrent access
- Protocol compliance tests for Bolt implementation

## Development Notes
- All components are designed to be thread-safe
- Configuration can be updated at runtime
- Connections are proxied transparently after initial routing
- Failed backends are handled gracefully with proper error reporting