package websocket

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	fibws "github.com/gofiber/websocket/v2"

	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/types"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

type ProxyHandler struct {
	router            interfaces.Router
	logger            logging.Logger
	metrics           *WebSocketMetrics
	config            *WebSocketProxyConfig
	activeConnections sync.Map
}

type WebSocketProxyConfig struct {
	PingInterval        time.Duration
	PongWait            time.Duration
	WriteWait           time.Duration
	ReadBufferSize      int
	WriteBufferSize     int
	MaxMessageSize      int64
	MessageBufferSize   int
	EnableCompression   bool
	ForwardHeaders      []string
	AllowedOrigins      []string
	EnableProxyProtocol bool
}

type ConnectionPair struct {
	ClientConn   *fibws.Conn
	UpstreamConn *websocket.Conn
	Started      time.Time
	BytesIn      int64
	BytesOut     int64
	MessageCount int64
}

func NewWebSocketProxyHandler(router interfaces.Router, logger logging.Logger) *ProxyHandler {
	return &ProxyHandler{
		router: router,
		logger: logger,
		metrics: &WebSocketMetrics{
			ConnectionsActive: 0,
			ConnectionsTotal:  0,
			MessagesReceived:  0,
			MessagesSent:      0,
			BytesReceived:     0,
			BytesSent:         0,
			ConnectionErrors:  0,
			MessageErrors:     0,
		},
		config: &WebSocketProxyConfig{
			PingInterval:        30 * time.Second,
			PongWait:            60 * time.Second,
			WriteWait:           10 * time.Second,
			ReadBufferSize:      4096,
			WriteBufferSize:     4096,
			MaxMessageSize:      512 * 1024, // 512 KB
			MessageBufferSize:   256,
			EnableCompression:   true,
			ForwardHeaders:      []string{"Authorization", "X-API-Key"},
			AllowedOrigins:      []string{"*"},
			EnableProxyProtocol: false,
		},
		activeConnections: sync.Map{},
	}
}

func (h *ProxyHandler) HandleRequest(c *fiber.Ctx) error {
	// Check if it's a WebSocket upgrade request
	if !websocket.FastHTTPIsWebSocketUpgrade(c.Context()) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "WebSocket upgrade required",
		})
	}

	// Find route
	httpReq := &http.Request{
		Method: c.Method(),
		URL:    c.Request().URI().QueryArgs().QueryString(),
		Header: make(http.Header),
		Host:   string(c.Request().Host()),
	}

	// Copy headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		httpReq.Header.Add(string(key), string(value))
	})

	route, err := h.router.FindRoute(httpReq)
	if err != nil {
		h.logger.Error("WebSocket route not found", "path", c.Path(), "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found for WebSocket connection",
		})
	}

	// Check if WebSocket is enabled for this route
	if !h.isWebSocketEnabled(route) {
		h.logger.Error("WebSocket not enabled for route", "path", c.Path(), "route", route.Name)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "WebSocket not enabled for this route",
		})
	}

	// Handle WebSocket connection
	return fibws.New(func(clientConn *fibws.Conn) {
		h.handleWebSocketConnection(clientConn, route)
	}, fibws.Config{
		ReadBufferSize:    h.config.ReadBufferSize,
		WriteBufferSize:   h.config.WriteBufferSize,
		EnableCompression: h.config.EnableCompression,
		HandshakeTimeout:  h.config.PongWait,
	})(c)
}

func (h *ProxyHandler) handleWebSocketConnection(clientConn *fibws.Conn, route *types.Route) {
	h.metrics.ConnectionsActive++
	h.metrics.ConnectionsTotal++

	connStartTime := time.Now()
	connectionID := fmt.Sprintf("%s-%d", clientConn.RemoteAddr().String(), connStartTime.UnixNano())

	h.logger.Info("WebSocket connection established",
		"id", connectionID,
		"route", route.Name,
		"client", clientConn.RemoteAddr().String())

	defer func() {
		h.metrics.ConnectionsActive--
		duration := time.Since(connStartTime)

		h.logger.Info("WebSocket connection closed",
			"id", connectionID,
			"route", route.Name,
			"client", clientConn.RemoteAddr().String(),
			"duration", duration.String())

		h.activeConnections.Delete(connectionID)
	}()

	// Get upstream URL
	upstreamURL := route.WebSocket.UpstreamURL
	if upstreamURL == "" {
		upstreamURL = route.UpstreamURL
	}

	// Make sure it's a WebSocket URL
	if !strings.HasPrefix(upstreamURL, "ws://") && !strings.HasPrefix(upstreamURL, "wss://") {
		if strings.HasPrefix(upstreamURL, "http://") {
			upstreamURL = "ws://" + upstreamURL[7:]
		} else if strings.HasPrefix(upstreamURL, "https://") {
			upstreamURL = "wss://" + upstreamURL[8:]
		} else {
			upstreamURL = "ws://" + upstreamURL
		}
	}

	// Prepare headers for upstream connection
	headers := http.Header{}
	for _, headerName := range h.config.ForwardHeaders {
		if value := clientConn.Headers(headerName); len(value) > 0 {
			headers.Set(headerName, value)
		}
	}

	// Connect to upstream WebSocket
	dialer := &websocket.Dialer{
		ReadBufferSize:    h.config.ReadBufferSize,
		WriteBufferSize:   h.config.WriteBufferSize,
		HandshakeTimeout:  h.config.PongWait,
		EnableCompression: h.config.EnableCompression,
	}

	upstreamConn, resp, err := dialer.Dial(upstreamURL, headers)
	if err != nil {
		h.metrics.ConnectionErrors++
		errMsg := err.Error()
		if resp != nil {
			errMsg = fmt.Sprintf("Upstream server returned HTTP %d: %s", resp.StatusCode, err.Error())
		}

		h.logger.Error("Failed to connect to upstream WebSocket",
			"id", connectionID,
			"route", route.Name,
			"upstream", upstreamURL,
			"error", errMsg)

		// Send close message to client
		clientConn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to connect to upstream service"),
			time.Now().Add(time.Second),
		)
		return
	}
	defer upstreamConn.Close()

	// Store connection pair
	connPair := &ConnectionPair{
		ClientConn:   clientConn,
		UpstreamConn: upstreamConn,
		Started:      connStartTime,
		BytesIn:      0,
		BytesOut:     0,
		MessageCount: 0,
	}
	h.activeConnections.Store(connectionID, connPair)

	// Create channels for message passing and error handling
	doneCh := make(chan struct{})
	errorCh := make(chan error, 2)

	// Start goroutines for bi-directional message forwarding
	go h.readFromClientAndWriteToUpstream(clientConn, upstreamConn, connPair, errorCh, doneCh)
	go h.readFromUpstreamAndWriteToClient(upstreamConn, clientConn, connPair, errorCh, doneCh)

	// Set up ping/pong handling
	pingTicker := time.NewTicker(h.config.PingInterval)
	defer pingTicker.Stop()

	// Wait for completion or error
	select {
	case <-doneCh:
		// Normal completion
		return
	case err := <-errorCh:
		h.logger.Error("WebSocket error occurred",
			"id", connectionID,
			"route", route.Name,
			"error", err.Error())
		return
	case <-pingTicker.C:
		// Send ping to client
		if err := clientConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(h.config.WriteWait)); err != nil {
			h.logger.Error("Failed to write ping message to client",
				"id", connectionID,
				"error", err.Error())
			return
		}
	}
}

func (h *ProxyHandler) readFromClientAndWriteToUpstream(
	clientConn *fibws.Conn,
	upstreamConn *websocket.Conn,
	connPair *ConnectionPair,
	errorCh chan<- error,
	doneCh chan<- struct{},
) {
	defer func() {
		close(doneCh)
	}()

	// Set read deadline and message size limit
	clientConn.SetReadLimit(h.config.MaxMessageSize)
	clientConn.SetReadDeadline(time.Now().Add(h.config.PongWait))

	// Set up pong handler to reset read deadline
	clientConn.SetPongHandler(func(string) error {
		clientConn.SetReadDeadline(time.Now().Add(h.config.PongWait))
		return nil
	})

	for {
		msgType, msg, err := clientConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
				websocket.CloseNoStatusReceived) {
				errorCh <- fmt.Errorf("unexpected close error from client: %w", err)
			}
			break
		}

		// Update metrics
		h.metrics.MessagesReceived++
		msgSize := int64(len(msg))
		h.metrics.BytesReceived += msgSize
		connPair.BytesIn += msgSize
		connPair.MessageCount++

		// Write to upstream
		err = upstreamConn.WriteMessage(msgType, msg)
		if err != nil {
			h.metrics.MessageErrors++
			errorCh <- fmt.Errorf("error writing message to upstream: %w", err)
			break
		}

		// Update metrics
		h.metrics.MessagesSent++
		h.metrics.BytesSent += msgSize
		connPair.BytesOut += msgSize
	}
}

func (h *ProxyHandler) readFromUpstreamAndWriteToClient(
	upstreamConn *websocket.Conn,
	clientConn *fibws.Conn,
	connPair *ConnectionPair,
	errorCh chan<- error,
	doneCh chan<- struct{},
) {
	// Set read deadline and message size limit
	upstreamConn.SetReadLimit(h.config.MaxMessageSize)
	upstreamConn.SetReadDeadline(time.Now().Add(h.config.PongWait))

	// Set up pong handler to reset read deadline
	upstreamConn.SetPongHandler(func(string) error {
		upstreamConn.SetReadDeadline(time.Now().Add(h.config.PongWait))
		return nil
	})

	for {
		msgType, msg, err := upstreamConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
				websocket.CloseNoStatusReceived) {
				errorCh <- fmt.Errorf("unexpected close error from upstream: %w", err)
			}
			return
		}

		// Update metrics
		h.metrics.MessagesReceived++
		msgSize := int64(len(msg))
		h.metrics.BytesReceived += msgSize
		connPair.BytesIn += msgSize
		connPair.MessageCount++

		// Write to client
		clientConn.SetWriteDeadline(time.Now().Add(h.config.WriteWait))
		err = clientConn.WriteMessage(msgType, msg)
		if err != nil {
			h.metrics.MessageErrors++
			errorCh <- fmt.Errorf("error writing message to client: %w", err)
			return
		}

		// Update metrics
		h.metrics.MessagesSent++
		h.metrics.BytesSent += msgSize
		connPair.BytesOut += msgSize
	}
}

func (h *ProxyHandler) isWebSocketEnabled(route *types.Route) bool {
	return route.WebSocket != nil && route.WebSocket.Enabled
}

func (h *ProxyHandler) GetStats() WebSocketStats {
	var stats WebSocketStats

	stats.ConnectionsActive = h.metrics.ConnectionsActive
	stats.ConnectionsTotal = h.metrics.ConnectionsTotal
	stats.MessagesReceived = h.metrics.MessagesReceived
	stats.MessagesSent = h.metrics.MessagesSent
	stats.BytesReceived = h.metrics.BytesReceived
	stats.BytesSent = h.metrics.BytesSent
	stats.ConnectionErrors = h.metrics.ConnectionErrors
	stats.MessageErrors = h.metrics.MessageErrors

	// Calculate average connection time
	var totalDuration float64
	var maxDuration float64
	var totalMsgSize int64
	var count int

	h.activeConnections.Range(func(_, value interface{}) bool {
		connPair, ok := value.(*ConnectionPair)
		if !ok {
			return true
		}

		duration := time.Since(connPair.Started).Seconds()
		totalDuration += duration

		if duration > maxDuration {
			maxDuration = duration
		}

		totalMsgSize += connPair.BytesIn + connPair.BytesOut
		count++

		return true
	})

	if count > 0 {
		stats.AvgConnectionTime = totalDuration / float64(count)
		stats.MaxConnectionTime = maxDuration

		if h.metrics.MessagesReceived+h.metrics.MessagesSent > 0 {
			stats.AvgMessageSize = float64(totalMsgSize) / float64(h.metrics.MessagesReceived+h.metrics.MessagesSent)
		}
	}

	if count > 0 {
		stats.LastConnectionTime = time.Now().Format(time.RFC3339)
	}

	return stats
}

func (h *ProxyHandler) GetActiveConnections() int {
	var count int
	h.activeConnections.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func (h *ProxyHandler) CloseAllConnections() {
	h.activeConnections.Range(func(key, value interface{}) bool {
		connPair, ok := value.(*ConnectionPair)
		if !ok {
			return true
		}

		// Close connections
		connPair.ClientConn.Close()
		connPair.UpstreamConn.Close()

		// Remove from map
		h.activeConnections.Delete(key)

		return true
	})
}

func (h *ProxyHandler) SetConfig(config *WebSocketProxyConfig) {
	h.config = config
}
