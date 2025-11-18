.PHONY: help build test run clean docker-build docker-run install dev demo

# Default target
help:
	@echo "Available targets:"
	@echo "  make build        - Build the demo application"
	@echo "  make test         - Run all tests"
	@echo "  make test-cover   - Run tests with coverage report"
	@echo "  make run          - Run the demo application"
	@echo "  make dev          - Run in development mode with auto-reload"
	@echo "  make demo         - Build and run demo application"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run application in Docker"
	@echo "  make install      - Install dependencies"
	@echo "  make lint         - Run linters"
	@echo "  make fmt          - Format code"

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Build the demo application
build:
	@echo "Building demo application..."
	go build -o bin/demo ./cmd/demo

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run the demo application
run: build
	@echo "Starting demo application..."
	./bin/demo

# Development mode (requires air or similar tool)
dev:
	@echo "Starting in development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Install 'air' for auto-reload: go install github.com/cosmtrek/air@latest"; \
		$(MAKE) run; \
	fi

# Build and run demo
demo: build
	@echo "Starting demo application..."
	@echo "--------------------------------------"
	@echo "GUI: http://localhost:8080"
	@echo "API: http://localhost:8080/api"
	@echo "--------------------------------------"
	./bin/demo

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f demo.db
	rm -f *.log

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

# Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "Install golangci-lint: https://golangci-lint.run/usage/install/"; \
		go vet ./...; \
	fi

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t go-configutil:latest .

# Docker run
docker-run:
	@echo "Running in Docker..."
	docker run -p 8080:8080 --env-file .env go-configutil:latest

# Security scan
security:
	@echo "Running security scan..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "Install gosec: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# Benchmark tests
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Check for vulnerabilities
vuln-check:
	@echo "Checking for vulnerabilities..."
	@if command -v govulncheck > /dev/null; then \
		govulncheck ./...; \
	else \
		echo "Install govulncheck: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi
