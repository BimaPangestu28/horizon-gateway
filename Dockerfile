FROM node:23-alpine AS ui-builder

WORKDIR /app/ui
COPY ui/package*.json ./
RUN npm ci --silent
COPY ui/ ./
RUN npm run build

FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o horizon cmd/horizon/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata curl
WORKDIR /app
COPY --from=builder /app/horizon .
COPY --from=builder /app/config.yaml /etc/horizon/config.yaml
COPY --from=ui-builder /app/ui/dist /app/ui/dist

VOLUME ["/etc/horizon"]
EXPOSE 8080 8081 8443

ENV CONFIG_PATH=/etc/horizon/config.yaml \
    ADMIN_UI_PATH=/app/ui/dist

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["./horizon", "--config", "/etc/horizon/config.yaml"]