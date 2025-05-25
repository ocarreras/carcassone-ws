.PHONY: build run test clean docker-build docker-run docker-compose-up docker-compose-down help

# Variables
BINARY_NAME=carcassonne-ws
DOCKER_IMAGE=carcassonne-ws
PORT=8080

# Default target
help:
	@echo "Available targets:"
	@echo "  build           - Build the Go binary"
	@echo "  run             - Run the server locally"
	@echo "  test            - Run tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  docker-compose-up   - Start with docker-compose"
	@echo "  docker-compose-down - Stop docker-compose"
	@echo "  deps            - Download dependencies"

# Build the Go binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) ./cmd/server

# Run the server locally
run:
	@echo "Starting server on port $(PORT)..."
	go run ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

# Run Docker container
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p $(PORT):$(PORT) --rm $(DOCKER_IMAGE)

# Start with docker-compose
docker-compose-up:
	@echo "Starting with docker-compose..."
	docker-compose up --build

# Stop docker-compose
docker-compose-down:
	@echo "Stopping docker-compose..."
	docker-compose down

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@echo "Run 'make run' to start the server"
	@echo "Open http://localhost:$(PORT) in your browser"

# Production build
prod-build:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o $(BINARY_NAME) ./cmd/server
