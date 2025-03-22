package websocket

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/types"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

var (
	ErrWebSocketUpgradeRequired = errors.New("websocket upgrade required")
	ErrUpstreamNotAvailable     = errors.New("upstream websocket service not available")
	ErrInvalidWebSocketRequest  = errors.New("invalid websocket request")
)

type WebSocketHandler struct {
	router  interfaces.Router
	logger  logging.Logger
	metrics *WebSocketMetrics
	config  *WebSocketConfig
}

type WebSocketConfig struct {
	PingInterval       time.Duration
	PongWait           time.Duration
	WriteWait          time.Duration
	ReadBufferSize     int
	WriteBufferSize    int
	MaxMessageSize     int64
	MessageBufferSize  int
	EnableCompression  bool
	UpgradeCheckOrigin func(r *http.Request) bool
}

type WebSocketMetrics struct {
	ConnectionsActive   int64
	ConnectionsTotal    int64
	MessagesReceived    int64
	MessagesSent        int64
	ConnectionErrors    int64
	MessageErrors       int64
	BytesReceived       int64
	BytesSent           int64
	ConnectionLifetimes []time.Duration
}

func NewWebSocketHandler(router interfaces.Router, logger logging.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		router: router,
		logger: logger,
		metrics: &WebSocketMetrics{
			ConnectionLifetimes: make([]time.Duration, 0),
		},
		config: &WebSocketConfig{
			PingInterval:      30 * time.Second,
			PongWait:          60 * time.Second,
			WriteWait:         10 * time.Second,
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			MaxMessageSize:    512 * 1024,
			MessageBufferSize: 256,
			EnableCompression: true,
			UpgradeCheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.NewError(fiber.StatusBadRequest, "WebSocket upgrade required")
	}

	route, err := h.router.FindRoute(c.Path(), c.Method())
	if err != nil {
		h.logger.Error("WebSocket route not found", "path", c.Path(), "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "WebSocket route not found",
		})
	}

	if !h.isWebSocketEnabled(route) {
		h.logger.Error("WebSocket not enabled for route", "path", c.Path(), "route", route.Name)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "WebSocket not enabled for this route",
		})
	}

	return websocket.New(func(conn *websocket.Conn) {
		h.handleConnection(conn, route)
	}, websocket.Config{
		ReadBufferSize:    h.config.ReadBufferSize,
		WriteBufferSize:   h.config.WriteBufferSize,
		EnableCompression: h.config.EnableCompression,
		HandshakeTimeout:  h.config.PongWait,
		Origins:           []string{"*"},
	})(c)
}

func (h *WebSocketHandler) handleConnection(clientConn *websocket.Conn, route *types.Route) {
	h.metrics.ConnectionsActive++
	h.metrics.ConnectionsTotal++
	h.logger.Info("WebSocket connection established", "route", route.Name, "client", clientConn.RemoteAddr().String())

	connectionStart := time.Now()
	defer func() {
		h.metrics.ConnectionsActive--
		connectionDuration := time.Since(connectionStart)
		h.metrics.ConnectionLifetimes = append(h.metrics.ConnectionLifetimes, connectionDuration)
		h.logger.Info("WebSocket connection closed",
			"route", route.Name,
			"client", clientConn.RemoteAddr().String(),
			"duration", connectionDuration.String())
	}()

	upstreamURL := route.UpstreamURL
	header := http.Header{}
	for k, v := range clientConn.Headers() {
		header.Set(k, v)
	}

	upstreamConn, _, err := ws.DefaultDialer.Dial(upstreamURL, header)
	if err != nil {
		h.metrics.ConnectionErrors++
		h.logger.Error("Failed to connect to upstream WebSocket",
			"route", route.Name,
			"upstream", upstreamURL,
			"error", err)
		closeMessage := websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Failed to connect to upstream service")
		clientConn.WriteMessage(websocket.CloseMessage, closeMessage)
		return
	}
	defer upstreamConn.Close()

	done := make(chan struct{})
	errorChan := make(chan error, 2)
	messageChan := make(chan []byte, h.config.MessageBufferSize)

	go h.readFromClient(clientConn, messageChan, errorChan, done)
	go h.writeToUpstream(upstreamConn, messageChan, errorChan, done)
	go h.readFromUpstream(upstreamConn, clientConn, errorChan, done)

	ticker := time.NewTicker(h.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case err := <-errorChan:
			h.logger.Error("WebSocket error",
				"route", route.Name,
				"client", clientConn.RemoteAddr().String(),
				"error", err)
			return
		case <-ticker.C:
			if err := clientConn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(h.config.WriteWait)); err != nil {
				h.logger.Error("Failed to write ping message", "error", err)
				return
			}
		}
	}
}

func (h *WebSocketHandler) readFromClient(clientConn *websocket.Conn, messageChan chan<- []byte, errorChan chan<- error, done chan<- struct{}) {
	defer func() {
		close(done)
	}()

	clientConn.SetReadLimit(h.config.MaxMessageSize)
	clientConn.SetReadDeadline(time.Now().Add(h.config.PongWait))
	clientConn.SetPongHandler(func(string) error {
		clientConn.SetReadDeadline(time.Now().Add(h.config.PongWait))
		return nil
	})

	for {
		_, message, err := clientConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNoStatusReceived) {
				errorChan <- fmt.Errorf("unexpected close: %w", err)
			}
			break
		}

		h.metrics.MessagesReceived++
		h.metrics.BytesReceived += int64(len(message))
		messageChan <- message
	}
}

func (h *WebSocketHandler) writeToUpstream(upstreamConn net.Conn, messageChan <-chan []byte, errorChan chan<- error, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		case message := <-messageChan:
			err := wsutil.WriteClientMessage(upstreamConn, ws.OpBinary, message)
			if err != nil {
				h.metrics.MessageErrors++
				errorChan <- fmt.Errorf("failed to write to upstream: %w", err)
				return
			}
			h.metrics.MessagesSent++
			h.metrics.BytesSent += int64(len(message))
		}
	}
}

func (h *WebSocketHandler) readFromUpstream(upstreamConn net.Conn, clientConn *websocket.Conn, errorChan chan<- error, done chan struct{}) {
	defer func() {
		close(done)
	}()

	for {
		select {
		case <-done:
			return
		default:
			msg, op, err := wsutil.ReadServerData(upstreamConn)
			if err != nil {
				h.metrics.MessageErrors++
				errorChan <- fmt.Errorf("failed to read from upstream: %w", err)
				return
			}

			h.metrics.MessagesReceived++
			h.metrics.BytesReceived += int64(len(msg))

			clientConn.SetWriteDeadline(time.Now().Add(h.config.WriteWait))
			if err := clientConn.WriteMessage(int(op), msg); err != nil {
				h.metrics.MessageErrors++
				errorChan <- fmt.Errorf("failed to write to client: %w", err)
				return
			}
			h.metrics.MessagesSent++
			h.metrics.BytesSent += int64(len(msg))
		}
	}
}

func (h *WebSocketHandler) isWebSocketEnabled(route *types.Route) bool {
	return route.WebSocket != nil && route.WebSocket.Enabled
}
