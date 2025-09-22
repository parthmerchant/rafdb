# RAFDB - Reliable and Fast Database
# Build, test, and deployment commands

# Default recipe - show available commands
default:
    @just --list

# Install dependencies
deps:
    go mod tidy
    go mod download

# Build the application
build:
    go build -o rafdb .

# Build for production (optimized)
build-prod:
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o rafdb .

# Run the application locally
run:
    go run .

# Run with live reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
    air

# Run tests
test:
    go test ./...

# Run tests with coverage
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Run benchmarks
bench:
    go test -bench=. ./...

# Format code
fmt:
    go fmt ./...

# Lint code (requires golangci-lint)
lint:
    golangci-lint run

# Clean build artifacts
clean:
    rm -f rafdb rafdb_data.json coverage.out coverage.html

# Docker commands

# Build Docker image
docker-build:
    docker build -t rafdb:latest .

# Run with Docker
docker-run: docker-build
    docker run -p 8080:8080 -v $(pwd)/data:/home/rafdb/data rafdb:latest

# Run with Docker Compose
docker-up:
    docker-compose up --build

# Stop Docker Compose
docker-down:
    docker-compose down

# Deploy RAFDB (build and run with Docker Compose)
deploy-rafdb: docker-up

# Database operations

# Test database with sample data
test-db:
    #!/usr/bin/env bash
    echo "Starting RAFDB for testing..."
    
    # Start the database in background
    go run . &
    DB_PID=$!
    
    # Wait for server to start
    sleep 3
    
    echo "Testing RAFDB API..."
    
    # Test health endpoint
    echo "1. Health check:"
    curl -s http://localhost:8080/api/v1/health | jq .
    
    # Create a collection
    echo "2. Creating 'users' collection:"
    curl -s -X POST http://localhost:8080/api/v1/collections \
        -H "Content-Type: application/json" \
        -d '{"name": "users"}' | jq .
    
    # Insert documents
    echo "3. Inserting user documents:"
    curl -s -X POST http://localhost:8080/api/v1/collections/users/documents \
        -H "Content-Type: application/json" \
        -d '{"id": "user1", "data": {"name": "Alice", "age": 30, "city": "New York"}}' | jq .
    
    curl -s -X POST http://localhost:8080/api/v1/collections/users/documents \
        -H "Content-Type: application/json" \
        -d '{"id": "user2", "data": {"name": "Bob", "age": 25, "city": "San Francisco"}}' | jq .
    
    # Get a document
    echo "4. Getting user1:"
    curl -s http://localhost:8080/api/v1/collections/users/documents/user1 | jq .
    
    # List all documents
    echo "5. Listing all users:"
    curl -s http://localhost:8080/api/v1/collections/users/documents | jq .
    
    # Query documents
    echo "6. Querying users by city (New York):"
    curl -s -X POST http://localhost:8080/api/v1/collections/users/query \
        -H "Content-Type: application/json" \
        -d '{"field": "city", "value": "New York"}' | jq .
    
    # Update a document
    echo "7. Updating user1:"
    curl -s -X PUT http://localhost:8080/api/v1/collections/users/documents/user1 \
        -H "Content-Type: application/json" \
        -d '{"data": {"name": "Alice Smith", "age": 31, "city": "New York", "status": "updated"}}' | jq .
    
    # Get stats
    echo "8. Database stats:"
    curl -s http://localhost:8080/api/v1/stats | jq .
    
    # List collections
    echo "9. List collections:"
    curl -s http://localhost:8080/api/v1/collections | jq .
    
    echo "Test completed! Stopping database..."
    kill $DB_PID
    
    echo "Data persisted to rafdb_data.json"

# Load test the database
load-test:
    #!/usr/bin/env bash
    echo "Starting load test..."
    
    # Start the database in background
    go run . &
    DB_PID=$!
    
    # Wait for server to start
    sleep 3
    
    # Create test collection
    curl -s -X POST http://localhost:8080/api/v1/collections \
        -H "Content-Type: application/json" \
        -d '{"name": "loadtest"}' > /dev/null
    
    echo "Inserting 1000 documents..."
    for i in {1..1000}; do
        curl -s -X POST http://localhost:8080/api/v1/collections/loadtest/documents \
            -H "Content-Type: application/json" \
            -d "{\"id\": \"doc$i\", \"data\": {\"index\": $i, \"value\": \"test-data-$i\", \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}}" > /dev/null
        
        if [ $((i % 100)) -eq 0 ]; then
            echo "Inserted $i documents..."
        fi
    done
    
    echo "Load test completed! Getting stats..."
    curl -s http://localhost:8080/api/v1/stats | jq .
    
    echo "Stopping database..."
    kill $DB_PID

# Benchmark the database
benchmark:
    #!/usr/bin/env bash
    echo "Running RAFDB benchmark..."
    
    # Start the database in background
    go run . &
    DB_PID=$!
    
    # Wait for server to start
    sleep 3
    
    # Create benchmark collection
    curl -s -X POST http://localhost:8080/api/v1/collections \
        -H "Content-Type: application/json" \
        -d '{"name": "benchmark"}' > /dev/null
    
    echo "Benchmarking write performance..."
    time for i in {1..100}; do
        curl -s -X POST http://localhost:8080/api/v1/collections/benchmark/documents \
            -H "Content-Type: application/json" \
            -d "{\"id\": \"bench$i\", \"data\": {\"value\": $i}}" > /dev/null
    done
    
    echo "Benchmarking read performance..."
    time for i in {1..100}; do
        curl -s http://localhost:8080/api/v1/collections/benchmark/documents/bench$i > /dev/null
    done
    
    echo "Benchmark completed!"
    kill $DB_PID

# Show help
help:
    @echo "RAFDB - Reliable and Fast Database"
    @echo ""
    @echo "Available commands:"
    @echo "  deps         - Install Go dependencies"
    @echo "  build        - Build the application"
    @echo "  run          - Run the application locally"
    @echo "  test         - Run tests"
    @echo "  test-db      - Test database with sample data"
    @echo "  load-test    - Run load test with 1000 documents"
    @echo "  benchmark    - Run performance benchmark"
    @echo "  docker-build - Build Docker image"
    @echo "  docker-run   - Run with Docker"
    @echo "  deploy-rafdb - Deploy with Docker Compose"
    @echo "  clean        - Clean build artifacts"
