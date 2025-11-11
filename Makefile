.PHONY: build run test clean docker-build docker-run install dev

# Build the application
build:
	go build -o bin/grss cmd/grss/main.go

# Run the application
run: build
	./bin/grss

# Install dependencies
install:
	go mod download

# Run in development mode with hot reload (requires air)
dev:
	air

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out

# Build Docker image
docker-build:
	docker build -t grss:latest .

# Run Docker container
docker-run:
	docker run -p 1200:1200 grss:latest

# Run with docker-compose
docker-compose-up:
	docker-compose up -d

# Stop docker-compose
docker-compose-down:
	docker-compose down

# View docker-compose logs
docker-compose-logs:
	docker-compose logs -f

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Install development tools
install-tools:
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help
help:
	@echo "Available targets:"
	@echo "  build              - Build the application"
	@echo "  run                - Build and run the application"
	@echo "  install            - Install dependencies"
	@echo "  dev                - Run with hot reload (requires air)"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  clean              - Clean build artifacts"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run Docker container"
	@echo "  docker-compose-up  - Start with docker-compose"
	@echo "  docker-compose-down - Stop docker-compose"
	@echo "  fmt                - Format code"
	@echo "  lint               - Lint code"
	@echo "  install-tools      - Install development tools"
