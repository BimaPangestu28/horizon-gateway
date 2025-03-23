FROM node:23-alpine AS ui-builder

WORKDIR /app/ui

COPY ui/package*.json ./

RUN npm install

COPY ui/ ./

RUN npm run build

FROM golang:1.20-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o horizon cmd/horizon/main.go

FROM alpine:3.16

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/horizon .

COPY config.yaml /etc/horizon/config.yaml

COPY --from=ui-builder /app/ui/dist /app/ui/dist

VOLUME ["/etc/horizon"]

EXPOSE 8080 8081

ENV CONFIG_PATH=/etc/horizon/config.yaml
ENV ADMIN_UI_PATH=/app/ui/dist

CMD ["./horizon", "/etc/horizon/config.yaml"]