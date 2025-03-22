package types

import (
	"time"
)

type WebSocketConfig struct {
	Enabled            bool          `yaml:"enabled" json:"enabled"`
	SubProtocols       []string      `yaml:"sub_protocols" json:"sub_protocols"`
	UpstreamURL        string        `yaml:"upstream_url" json:"upstream_url"`
	PingInterval       time.Duration `yaml:"ping_interval" json:"ping_interval"`
	PongWait           time.Duration `yaml:"pong_wait" json:"pong_wait"`
	WriteWait          time.Duration `yaml:"write_wait" json:"write_wait"`
	ReadBufferSize     int           `yaml:"read_buffer_size" json:"read_buffer_size"`
	WriteBufferSize    int           `yaml:"write_buffer_size" json:"write_buffer_size"`
	MaxMessageSize     int64         `yaml:"max_message_size" json:"max_message_size"`
	MessageBufferSize  int           `yaml:"message_buffer_size" json:"message_buffer_size"`
	EnableCompression  bool          `yaml:"enable_compression" json:"enable_compression"`
	AllowedOrigins     []string      `yaml:"allowed_origins" json:"allowed_origins"`
	ForwardHeaders     []string      `yaml:"forward_headers" json:"forward_headers"`
	ProxyProtocolLevel int           `yaml:"proxy_protocol_level" json:"proxy_protocol_level"`
}

type WebSocketStats struct {
	ConnectionsActive  int64   `json:"connections_active"`
	ConnectionsTotal   int64   `json:"connections_total"`
	MessagesReceived   int64   `json:"messages_received"`
	MessagesSent       int64   `json:"messages_sent"`
	BytesReceived      int64   `json:"bytes_received"`
	BytesSent          int64   `json:"bytes_sent"`
	ConnectionErrors   int64   `json:"connection_errors"`
	MessageErrors      int64   `json:"message_errors"`
	AvgConnectionTime  float64 `json:"avg_connection_time"`
	MaxConnectionTime  float64 `json:"max_connection_time"`
	AvgMessageSize     float64 `json:"avg_message_size"`
	LastConnectionTime string  `json:"last_connection_time"`
}
