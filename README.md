# Horizon API Gateway

A high-performance, community-focused API Gateway built with Go.

![Horizon Logo](https://raw.githubusercontent.com/horizon-gateway/assets/main/logo.png)

> A high-performance, community-focused API Gateway built with Go.

[![Go Report Card](https://goreportcard.com/badge/github.com/horizon-gateway/horizon)](https://goreportcard.com/report/github.com/horizon-gateway/horizon)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## 📖 Overview

Horizon is an open-source, enterprise-grade API Gateway designed to be simple to use while offering powerful features for managing, securing, and optimizing API traffic. It's built from the ground up with performance in mind, making it suitable for both small projects and large-scale deployments.

## 🚀 Why Horizon?

- **Simplicity First**: Deploy and configure in minutes with sensible defaults
- **High Performance**: Built with Go for exceptional throughput and low latency
- **Community Driven**: 100% open source with a focus on community contributions
- **Enterprise Ready**: Includes features typically found only in premium solutions
- **Developer Friendly**: Comprehensive documentation and intuitive interfaces

## ✨ Features

### Core Features (Phase 1)
- ✅ HTTP/HTTPS reverse proxy with HTTP/2 support
- ✅ Dynamic routing based on path, headers, query params, and methods
- ✅ Advanced load balancing (round-robin, weighted, least connections)
- ✅ Circuit breaking and request timeouts
- ✅ Connection pooling and keep-alive management
- ✅ Docker and Kubernetes deployment

### Security & Performance (Phase 2)
- ✅ API key authentication
- ✅ JWT validation and generation
- ✅ Rate limiting and throttling with multiple algorithms
- ✅ IP filtering and geolocation rules
- ✅ Request validation
- ✅ CORS management
- ✅ Response caching with multiple strategies

### Management & Observability (Phase 3)
- ✅ Admin API for complete gateway management
- ✅ Admin UI dashboard with React and Tailwind CSS
- ✅ Structured logging (JSON format)
- ✅ Prometheus metrics with detailed monitoring
- ✅ Distributed tracing (OpenTelemetry)
- ✅ Configuration versioning and rollback
- ✅ Real-time analytics dashboard

### Coming Soon (Phase 4)
- ⬜ Plugins system
- ⬜ WebSocket/gRPC support
- ⬜ Service discovery
- ⬜ Advanced transformations

## 🛠️ Getting Started

### Prerequisites

- Go 1.20 or higher
- Docker (for containerized deployment)
- Node.js 16+ (for Admin UI)

### Installation

#### From Source

```bash
# Clone the repository
git clone https://github.com/horizon-gateway/horizon.git
cd horizon

# Build the binary
go build -o horizon cmd/horizon/main.go

# Build the Admin UI
cd ui
npm install
npm run build
cd ..

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

We welcome contributions of all kinds! See our [Contributing Guide](CONTRIBUTING.md) for details on how to get started.

## 📄 License

Horizon API Gateway is released under the [Apache 2.0 License](LICENSE).