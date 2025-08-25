# Neo4j Multi-tenant Proxy Makefile

.PHONY: help build test test-verbose coverage clean lint fmt vet run dev install-tools

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

# Build the application
build: ## Build the Neo4j proxy binary
	go build -o neo4j-proxy ./cmd/neo4j-proxy

# Run tests
test: ## Run all tests
	go run github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --randomize-suites --fail-on-pending --cover --trace --junit-report=junit.xml

# Run tests with verbose output
test-verbose: ## Run tests with verbose output
	go run github.com/onsi/ginkgo/v2/ginkgo -r -v --randomize-all --randomize-suites --fail-on-pending --cover --trace --junit-report=junit.xml

# Run tests with race detection
test-race: ## Run tests with race detection
	go run github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --randomize-suites --fail-on-pending --cover --trace --race --junit-report=junit.xml

# Generate coverage report
coverage: test ## Generate and view coverage report
	@echo "Coverage report:"
	@go tool cover -func=coverprofile.out | tail -1

# View coverage in browser
coverage-html: test ## Generate HTML coverage report
	go tool cover -html=coverprofile.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean: ## Clean build artifacts and test reports
	rm -f neo4j-proxy
	rm -f junit.xml
	rm -f *.out
	rm -f coverage.html

# Lint code
lint: ## Run linter
	golangci-lint run

# Format code
fmt: ## Format Go code
	go fmt ./...

# Vet code
vet: ## Run go vet
	go vet ./...

# Run the proxy
run: build ## Build and run the proxy
	./neo4j-proxy

# Development mode (with file watching would require additional tools)
dev: ## Run in development mode
	go run ./cmd/neo4j-proxy

# Install development tools
install-tools: ## Install required development tools
	go install github.com/onsi/ginkgo/v2/ginkgo@latest
	@echo "Development tools installed"

# Tidy dependencies
tidy: ## Tidy go modules
	go mod tidy

# Download dependencies
deps: ## Download dependencies
	go mod download

# Verify dependencies
verify: ## Verify dependencies
	go mod verify

# All quality checks
check: fmt vet lint test ## Run all quality checks

# Docker build (if you want to add Docker support later)
docker-build: ## Build Docker image
	docker build -t neo4j-proxy .

# Show Go environment
env: ## Show Go environment
	go env