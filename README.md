# Horizon API Gateway

![Horizon Logo](https://raw.githubusercontent.com/horizon-gateway/assets/main/logo.png)

> A high-performance, community-focused API Gateway built with Go.

[![Go Report Card](https://goreportcard.com/badge/github.com/BimaPangestu28/horizon-gateway)](https://goreportcard.com/report/github.com/BimaPangestu28/horizon-gateway)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/BimaPangestu28/horizon-gateway.svg)](https://github.com/BimaPangestu28/horizon-gateway/stargazers)
[![GitHub Issues](https://img.shields.io/github/issues/BimaPangestu28/horizon-gateway.svg)](https://github.com/BimaPangestu28/horizon-gateway/issues)

## 📖 Overview

Horizon is an open-source, enterprise-grade API Gateway designed to be simple to use while offering powerful features for managing, securing, and optimizing API traffic. It's built from the ground up with performance in mind, making it suitable for both small projects and large-scale deployments.

## 🚀 Why Horizon?

- **Simplicity First**: Deploy and configure in minutes with sensible defaults
- **High Performance**: Built with Go for exceptional throughput and low latency
- **Community Driven**: 100% open source with a focus on community contributions
- **Enterprise Ready**: Includes features typically found only in premium solutions
- **Developer Friendly**: Comprehensive documentation and intuitive interfaces

## ✨ Features

### Core Functionality

- ✅ HTTP/HTTPS reverse proxy with HTTP/2 support
- ✅ Dynamic routing based on path, headers, query params, and methods
- ✅ Advanced load balancing (round-robin, weighted, least connections)
- ✅ Circuit breaking and request timeouts
- ✅ Connection pooling and keep-alive management
- ✅ WebSocket, gRPC, and GraphQL support
- ✅ Request/response transformation
- ✅ Response caching with configurable strategies

### Security

- ✅ API key authentication
- ✅ JWT validation and generation
- ✅ OAuth2 server and client
- ✅ Rate limiting and throttling
- ✅ IP filtering and geolocation rules
- ✅ Request validation against schemas
- ✅ CORS management
- ✅ SSL/TLS termination with automated certificate management

### Observability & Management

- ✅ Structured logging (JSON format)
- ✅ Prometheus metrics
- ✅ Distributed tracing (OpenTelemetry)
- ✅ Health checks with configurable probes
- ✅ Real-time analytics dashboard
- ✅ Configuration API
- ✅ Admin UI
- ✅ Hot reload of configurations

### Service Discovery

- ✅ Static service discovery
- ✅ DNS-based service discovery
- ✅ Consul service discovery
- ✅ Kubernetes service discovery
- ✅ Etcd service discovery

### Developer Experience

- ✅ Playground/testing console
- ✅ Comprehensive documentation
- ✅ CLI tools for management
- ✅ Mock services for testing
- ✅ Debug mode with request/response inspection
- ✅ Request replaying

### Advanced Features

- ✅ Plugin system
- ✅ API aggregation
- ✅ GraphQL federation
- ✅ Custom validators
- ✅ WebAssembly plugins
- ✅ Request transformation scripting
- ✅ High availability clustering

## 🚧 Current Status

Horizon is currently under active development. We've completed the implementation of our core features according to our roadmap:

- **Phase 1 (Completed)**: Core proxy and routing functionality
  - HTTP/HTTPS proxy
  - Basic routing
  - Load balancing
  - Health checks
  - Docker deployment

- **Phase 2 (Completed)**: Security and performance
  - Authentication (API keys, JWT)
  - Rate limiting
  - Response caching
  - Circuit breaking

- **Phase 3 (Completed)**: Management and observability
  - Admin API
  - Admin UI
  - Logging and metrics
  - Tracing

- **Phase 4 (Completed)**: Advanced features
  - Plugins system
  - WebSocket/gRPC support
  - Service discovery
  - Advanced transformations

## 🛠️ Getting Started

### Prerequisites

- Go 1.20 or higher
- Docker (for containerized deployment)

### Installation

#### From Source

```bash
# Clone the repository
git clone https://github.com/BimaPangestu28/horizon-gateway.git
cd horizon-gateway

# Build the binary
go build -o horizon cmd/main.go

# Run with default configuration
./horizon
```

#### Using Docker

```bash
docker run -p 8080:8080 -p 8081:8081 \
  -v $(pwd)/config.yaml:/etc/horizon/config.yaml \
  horizongateway/horizon:latest
```

### Basic Configuration

Create a `config.yaml` file:

```yaml
server:
  port: 8080
  admin_port: 8081

routes:
  - name: example-api
    listen_path: /api/*
    upstream_url: http://api.example.com
    methods: ["GET", "POST"]
    
  - name: another-service
    listen_path: /service/*
    upstream_url: http://service.internal
    strip_path: true
    methods: ["*"]
```

See the [Configuration Guide](docs/configuration.md) for full details.

## 💻 Development

### Project Structure

```
horizon-gateway/
├── cmd/                # Command line applications
│   └── horizon/        # Main application entrypoint
├── internal/           # Private application code
│   ├── cache/          # Caching implementations
│   ├── config/         # Configuration management
│   ├── core/           # Core routing and proxy logic
│   ├── discovery/      # Service discovery
│   ├── errors/         # Error definitions
│   ├── graphql/        # GraphQL support
│   ├── grpc/           # gRPC support
│   ├── handlers/       # HTTP request handlers
│   ├── httphandlers/   # HTTP-specific handlers
│   ├── interfaces/     # Common interfaces
│   ├── middleware/     # Middleware components
│   ├── metrics/        # Metrics collection
│   ├── plugins/        # Plugin system
│   ├── proxy/          # Reverse proxy functionality
│   ├── resilience/     # Circuit breaking, retries
│   ├── security/       # Authentication and authorization
│   ├── server/         # HTTP server setup
│   ├── tracing/        # Distributed tracing
│   ├── transform/      # Request/response transformation
│   ├── types/          # Common type definitions
│   ├── utils/          # Utility functions
│   ├── validator/      # Request validation
│   └── websocket/      # WebSocket support
├── k8s/                # Kubernetes deployment files
├── monitoring/         # Monitoring setup files
├── plugins/            # Plugin implementations
│   ├── builtin/        # Built-in plugins
│   ├── custom/         # Custom plugin examples
│   ├── interfaces/     # Plugin interfaces
│   ├── loader/         # Plugin loading system
│   ├── registry/       # Plugin registry
│   └── wasm/           # WebAssembly plugins
├── ui/                 # Admin UI
└── docs/               # Documentation
```

### Build & Test

```bash
# Run tests
go test ./...

# Build for development
go build -tags dev -o horizon cmd/main.go

# Run with hot reloading (requires air)
air
```

### Development with Docker Compose

```bash
# Start development environment
docker-compose -f docker-compose.yaml -f docker-compose.dev.yaml up -d

# View logs
docker-compose logs -f

# Rebuild and restart
docker-compose -f docker-compose.yaml -f docker-compose.dev.yaml up -d --build
```

### Working with Plugins

```bash
# Build plugins
make build-plugins

# Create a new plugin
make setup-plugin
```

## 🖥️ Admin UI

Horizon includes a comprehensive Admin UI for managing all aspects of the API Gateway:

- **Dashboard**: Real-time metrics and gateway status
- **Routes Management**: Create, edit, and delete API routes
- **Authentication**: Manage API keys and JWT configurations
- **Rate Limiting**: Configure rate limits for your APIs
- **Circuit Breakers**: Control failure handling and graceful degradation
- **Caching**: Optimize performance with response caching
- **Analytics**: Visualize traffic patterns and error rates

To access the Admin UI, navigate to `http://localhost:8081` after starting Horizon.

## 📊 Observability

Horizon provides comprehensive observability features:

- **Metrics**: Prometheus-compatible metrics endpoint at `/metrics`
- **Logs**: Structured JSON logs for easy parsing and analysis
- **Tracing**: OpenTelemetry integration with support for Jaeger, Zipkin, and more
- **Health Checks**: Advanced health check endpoints for monitoring

## 🤝 Contributing

We welcome contributions of all kinds! Here's how you can contribute:

1. **Fork the Repository**: Create your own fork of the project
2. **Create a Feature Branch**: `git checkout -b feature/amazing-feature`
3. **Make Changes**: Implement your changes following the coding standards
4. **Run Tests**: Ensure all tests pass with `go test ./...`
5. **Commit Changes**: Commit with a descriptive message
6. **Push to Branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**: Submit your changes for review

Please see our [Contributing Guide](CONTRIBUTING.md) for detailed information.

### What We Need Help With

- Core functionality implementation
- Testing and bug fixes
- Documentation improvements
- Feature suggestions
- UI/UX enhancements

## 📄 License

Horizon API Gateway is released under the [Apache 2.0 License](LICENSE).

## 🙏 Acknowledgements

- Inspired by other great API gateways like Kong, Traefik, and KrakenD
- Built with amazing open-source technologies
- Made possible by our wonderful community contributors