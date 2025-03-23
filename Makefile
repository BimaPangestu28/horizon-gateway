# Declare phony targets (targets that don't create files)
.PHONY: build run test lint clean docker docker-compose dev ui-build ui-dev all setup

# Build the Go backend
build:
	go build -o horizon cmd/horizon/main.go

# Build the UI (React frontend)
ui-build:
	cd ui && npm install && npm run build

# Build both backend and UI
all: build ui-build

# Run the backend only
run: build
	./horizon

# Build both backend and UI, then run them together
run-with-ui: all
	@echo "Starting backend with UI on port 8081"
	@echo "Access the API and UI at http://localhost:8081/admin"
	./horizon

# Run all tests
test:
	go test ./...

# Run code linting
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f horizon
	rm -rf tmp
	rm -rf ui/build ui/dist

# Build Docker image
docker:
	docker build -t horizon .

# Start with Docker Compose
docker-compose:
	docker-compose up -d

# Stop Docker Compose services
docker-compose-down:
	docker-compose down

# Setup development environment
setup-dev:
	go mod download
	go install github.com/golangci/golint/cmd/golint@latest
	go install github.com/air-verse/air@latest
	cd ui && npm install

# Run the backend in development mode
dev:
	go run cmd/horizon/main.go

# Run the backend with hot reload
dev-hot:
	air

# Run the UI in development mode with HMR
ui-dev:
	cd ui && npm run dev

# Full development environment with hot reload
dev-full: ui-build dev-hot

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o horizon-linux-amd64 cmd/horizon/main.go
	GOOS=darwin GOARCH=amd64 go build -o horizon-darwin-amd64 cmd/horizon/main.go
	GOOS=windows GOARCH=amd64 go build -o horizon-windows-amd64.exe cmd/horizon/main.go

# Generate test configuration
gen-test-config:
	cp config.yaml config.test.yaml

# Start metrics collection
start-metrics:
	docker-compose -f monitoring/prometheus/docker-compose.yaml up -d

# Start distributed tracing
start-tracing:
	docker-compose -f monitoring/jaeger/docker-compose.yaml up -d

# Start full monitoring stack
start-monitoring:
	docker-compose -f monitoring/docker-compose.yaml up -d

# Stop monitoring stack
stop-monitoring:
	docker-compose -f monitoring/docker-compose.yaml down

# Create configuration backup
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