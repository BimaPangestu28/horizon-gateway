# Horizon API Gateway

![Horizon Logo](https://raw.githubusercontent.com/horizon-gateway/assets/main/logo.png)

> A high-performance, community-focused API Gateway built with Go.

[![Go Report Card](https://goreportcard.com/badge/github.com/horizon-gateway/horizon)](https://goreportcard.com/report/github.com/horizon-gateway/horizon)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/horizon-gateway/horizon.svg)](https://github.com/horizon-gateway/horizon/stargazers)
[![GitHub Issues](https://img.shields.io/github/issues/horizon-gateway/horizon.svg)](https://github.com/horizon-gateway/horizon/issues)

## 📖 Overview

Horizon is an open-source, enterprise-grade API Gateway designed to be simple to use while offering powerful features for managing, securing, and optimizing API traffic. It's built from the ground up with performance in mind, making it suitable for both small projects and large-scale deployments.

## 🚀 Why Horizon?

- **Simplicity First**: Deploy and configure in minutes with sensible defaults
- **High Performance**: Built with Go for exceptional throughput and low latency
- **Community Driven**: 100% open source with a focus on community contributions
- **Enterprise Ready**: Includes features typically found only in premium solutions
- **Developer Friendly**: Comprehensive documentation and intuitive interfaces

## 🔧 Tech Stack

- **Language**: Go (Golang) 1.20+
- **HTTP Framework**: [Fiber](https://github.com/gofiber/fiber)
- **Configuration**: YAML/JSON with hot reload capability
- **Storage**: [Badger DB](https://github.com/dgraph-io/badger) (embedded), with optional PostgreSQL for clustering
- **Authentication**: Native JWT, API Keys, OAuth2 support
- **Monitoring**: Prometheus metrics, structured logging (JSON)
- **UI**: Admin dashboard built with React and Tailwind CSS
- **Deployment**: Docker, Kubernetes-ready

## ✨ Features

### Core

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

### Developer Experience

- ✅ Playground/testing console
- ✅ Comprehensive documentation
- ✅ CLI tools for management
- ✅ Mock services for testing
- ✅ Debug mode with request/response inspection
- ✅ Request replaying

## 🚧 Current Status

Horizon is currently under active development. We're working on implementing the core features first, followed by security, observability, and management components.

### Roadmap

- **Phase 1 (Current)**: Core proxy and routing functionality
  - HTTP/HTTPS proxy
  - Basic routing
  - Load balancing
  - Health checks
  - Docker deployment

- **Phase 2**: Security and performance
  - Authentication (API keys, JWT)
  - Rate limiting
  - Response caching
  - Circuit breaking

- **Phase 3**: Management and observability
  - Admin API
  - Admin UI
  - Logging and metrics
  - Tracing

- **Phase 4**: Advanced features
  - Plugins system
  - WebSocket/gRPC support
  - Service discovery
  - Advanced transformations

## 🛠️ Getting Started

### Prerequisites

- Go 1.20 or higher
- Node.js 23 or higher
- Docker (for containerized deployment)

### Installation

#### From Source

```bash
# Clone the repository
git clone https://github.com/horizon-gateway/horizon.git
cd horizon

# Build the backend and UI
make all

# Run with default configuration
make run
```

This will start:
- API Gateway on port 8080
- Admin API and UI on port 8081

#### Using Docker

```bash
docker run -p 8080:8080 -p 8081:8081 \
  -v $(pwd)/config.yaml:/etc/horizon/config.yaml \
  horizongateway/horizon:latest
```

### Development Workflow

```bash
# Run backend with hot reload
make dev-hot

# Build UI and run backend with hot reload
make dev-full

# Run UI in development mode
make ui-dev
```

### Accessing the Application

After starting the application:

- Access the API Gateway at http://localhost:8080
- Access the Admin API and UI at http://localhost:8081/admin

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
horizon/
├── cmd/                # Command line applications
├── config/             # Configuration management
├── core/               # Core gateway functionality
├── handlers/           # HTTP handlers
├── middleware/         # Middleware components
├── plugins/            # Plugin system
├── security/           # Authentication and authorization
├── storage/            # Data storage implementations
├── ui/                 # Admin UI
└── utils/              # Utility functions
```

### Build & Test

```bash
# Run tests
make test

# Build for development
make build

# Run with hot reloading
make dev-hot
```

## 📖 Documentation

Comprehensive documentation is available at [docs.horizongateway.io](https://docs.horizongateway.io)

- [Getting Started](docs/getting-started.md)
- [Configuration Reference](docs/configuration.md)
- [API Documentation](docs/api.md)
- [Deployment Guide](docs/deployment.md)
- [Contribution Guidelines](CONTRIBUTING.md)

## 🤝 How to Become a Contributor

We welcome contributions from developers of all experience levels! Here's how you can get involved:

### 1. Set Up Your Development Environment

```bash
# Clone the repository
git clone https://github.com/horizon-gateway/horizon.git
cd horizon

# Set up the development environment
make setup-dev
```

### 2. Find Something to Work On

- Check our [GitHub Issues](https://github.com/horizon-gateway/horizon/issues) for open tasks
- Look for issues tagged with `good-first-issue` if you're new to the project
- Join our community chat to discuss ideas (Discord/Slack link)

### 3. Create a Branch and Make Your Changes

```bash
# Create a branch with a descriptive name
git checkout -b feature/your-feature-name

# Make your changes and commit them
git commit -m "Add your meaningful commit message"
```

### 4. Submit a Pull Request

- Push your branch to your fork of the repository
- Open a pull request with a clear description of your changes
- Wait for code review and address any feedback

### 5. Contribution Guidelines

- Follow the Go style guide and coding conventions
- Write tests for your code
- Document new features
- Keep pull requests focused on a single change
- Sign off your commits using `git commit -s`

### 6. Code of Conduct

All contributors are expected to adhere to our Code of Conduct, which promotes a respectful and inclusive environment for everyone. Harassment or disrespectful behavior will not be tolerated.

### 7. Recognition

Contributors are recognized in the following ways:
- Listed in our CONTRIBUTORS.md file
- Acknowledged in release notes
- Potential to join the core team based on consistent contributions

See our [Contribution Guidelines](CONTRIBUTING.md) for more detailed information.

## 📄 License

Horizon API Gateway is released under the [Apache 2.0 License](LICENSE).

## 🙏 Acknowledgements

- Inspired by other great API gateways like Kong, Traefik, and KrakenD
- Built with amazing open-source technologies
- Made possible by our wonderful community contributors