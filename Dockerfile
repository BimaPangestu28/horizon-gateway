# Build stage
FROM golang:1.20-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o horizon cmd/horizon/main.go

# Final stage
FROM alpine:3.16

# Add ca certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/horizon .

# Copy default configuration
COPY config.yaml /etc/horizon/config.yaml

# Create volume for configuration
VOLUME ["/etc/horizon"]

# Expose ports
EXPOSE 8080 8081

# Set environment variables
ENV CONFIG_PATH=/etc/horizon/config.yaml

# Run the application
CMD ["./horizon", "/etc/horizon/config.yaml"]