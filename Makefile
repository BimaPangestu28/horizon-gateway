.PHONY: build run test lint clean docker docker-compose dev-hot

# Build the gateway
build:
	go build -o horizon cmd/horizon/main.go

# Run the gateway
run: build
	./horizon

# Run tests
test:
	go test ./...

# Run linting
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f horizon
	rm -rf tmp

# Build the Docker image
docker:
	docker build -t horizon .

# Start the gateway with Docker Compose
docker-compose:
	docker-compose up -d

# Stop the Docker Compose environment
docker-compose-down:
	docker-compose down

# Setup development environment
setup-dev:
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest

# Run the gateway in development mode
dev:
	go run cmd/horizon/main.go

# Run the gateway in development mode with hot reload
dev-hot:
	air

# Build for all supported platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o horizon-linux-amd64 cmd/horizon/main.go
	GOOS=darwin GOARCH=amd64 go build -o horizon-darwin-amd64 cmd/horizon/main.go
	GOOS=windows GOARCH=amd64 go build -o horizon-windows-amd64.exe cmd/horizon/main.go

# Generate test configuration for development
gen-test-config:
	cp config.yaml config.test.yaml