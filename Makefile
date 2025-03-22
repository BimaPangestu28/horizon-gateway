.PHONY: build run test lint clean docker docker-compose dev ui-build ui-dev all setup

# Build the gateway
build:
	go build -o horizon cmd/horizon/main.go

# Build UI
ui-build:
	cd ui && npm install && npm run build

# Build everything
all: build ui-build

# Run the gateway
run: build
	./horizon

# Run with Admin UI
run-with-ui: build ui-build
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
	rm -rf ui/build ui/dist

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
	go install github.com/golangci/golint/cmd/golint@latest
	go install github.com/air-verse/air@latest
	cd ui && npm install

# Run the gateway in development mode
dev:
	go run cmd/horizon/main.go

# Run the gateway in development mode with hot reload
dev-hot:
	air

# Run UI in development mode
ui-dev:
	cd ui && npm run dev

# Build for all supported platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o horizon-linux-amd64 cmd/horizon/main.go
	GOOS=darwin GOARCH=amd64 go build -o horizon-darwin-amd64 cmd/horizon/main.go
	GOOS=windows GOARCH=amd64 go build -o horizon-windows-amd64.exe cmd/horizon/main.go

# Generate test configuration for development
gen-test-config:
	cp config.yaml config.test.yaml

# Start metrics collection (Prometheus)
start-metrics:
	docker-compose -f monitoring/prometheus/docker-compose.yaml up -d

# Start tracing (Jaeger)
start-tracing:
	docker-compose -f monitoring/jaeger/docker-compose.yaml up -d

# Start monitoring stack (Prometheus, Grafana, Jaeger)
start-monitoring:
	docker-compose -f monitoring/docker-compose.yaml up -d

# Stop monitoring stack
stop-monitoring:
	docker-compose -f monitoring/docker-compose.yaml down

# Create a config backup
backup-config:
	mkdir -p config-backups
	cp config.yaml config-backups/config-$$(date +%Y%m%d%H%M%S).yaml

# Setup UI project from scratch
setup-ui:
	mkdir -p ui
	cd ui && npm init -y
	cd ui && npm install react react-dom react-router-dom recharts
	cd ui && npm install react-feather
	cd ui && npm install tailwindcss postcss autoprefixer
	cd ui && npm install @headlessui/react
	cd ui && npm install --save-dev vite @vitejs/plugin-react
	cd ui && npx tailwindcss init -p
	mkdir -p ui/src/components ui/src/pages ui/src/services