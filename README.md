# Horizon API Gateway

A high-performance, community-focused API Gateway built with Go.

## Features

### Core Features
- HTTP/HTTPS proxy with HTTP/2 support
- Advanced routing based on path, method, headers, query parameters, and host
- Load balancing (round-robin, weighted, least connections)
- Connection draining for graceful service removal
- Health checking with circuit breaking
- Configuration hot reload with versioning and migration
- Docker and Kubernetes deployment

## Getting Started

- Go 1.20 or higher
- Docker and Docker Compose
- Git

## Getting Started

### Clone the Repository

```bash
git clone https://github.com/horizon-gateway/horizon.git
cd horizon
```

### Building the Project

Build the project locally:

```bash
go build -o horizon cmd/horizon/main.go
```

### Running the Gateway

Run the gateway directly:

```bash
./horizon
```

The gateway will start with default configuration:
- Main port: 8080
- Admin port: 8081

### Using Docker

Build and run using Docker:

```bash
docker build -t horizon .
docker run -p 8080:8080 -p 8081:8081 -v $(pwd)/config.yaml:/etc/horizon/config.yaml horizon
```

### Using Docker Compose

```bash
docker-compose up -d
```

This will start the gateway and a mock service for testing.

## Development

### Project Structure

```
horizon/
├── cmd/                # Command line applications
│   └── horizon/        # Main application entrypoint
│       └── main.go
├── internal/           # Private application code
│   ├── config/         # Configuration management
│   ├── core/           # Core gateway functionality
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # Middleware components
│   └── utils/          # Utility functions
├── config.yaml         # Default configuration
├── Dockerfile          # Docker build instructions
├── docker-compose.yml  # Docker Compose setup
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
└── .air.toml           # Configuration for hot reload
```

### Development with Hot Reload

For development with automatic rebuilding and restarting when code changes:

1. Install Air:
```bash
go install github.com/air-verse/air@latest
```

2. Run the setup-dev make target to set up the development environment:
```bash
make setup-dev
```

3. Run the gateway with hot reload:
```bash
make dev-hot
```

Air will monitor your source code files and automatically rebuild and restart the gateway when changes are detected.

### Testing the Gateway

Once running, you can test the gateway with:

```bash
# Test the health endpoint
curl http://localhost:8080/health

# Test proxying through the gateway
curl http://localhost:8080/api/get

# Test the admin API
curl http://localhost:8081/admin/routes
```

## Configuration

The gateway is configured via a YAML file. By default, it looks for `config.yaml` in the current directory, but you can specify a different path:

```bash
./horizon /path/to/config.yaml
```

See the example config file for available options.

## Contributing

Please follow the code standards and git branching strategy as defined in the project documentation when contributing to this project.