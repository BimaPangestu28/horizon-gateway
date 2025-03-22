package websocket

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var (
	ErrWebSocketNotEnabled     = errors.New("websocket not enabled for this route")
	ErrWebSocketUpgradeFailure = errors.New("failed to upgrade connection to websocket")
	ErrUpstreamConnectFailure  = errors.New("failed to connect to upstream websocket")
	ErrInvalidUpstreamURL      = errors.New("invalid upstream websocket URL")
)

type WebSocketProxy struct {
	router           interfaces.Router
	logger           logging.Logger
	connections      map[string]*ConnectionInfo
	connectionsMutex sync.RWMutex
	metrics          *WebSocketMetrics
	config           *WebSocketConfig
}

type ConnectionInfo struct {
	ID            string
	ClientConn    *websocket.Conn
	UpstreamConn  *websocket.Conn
	Route         string
	StartTime     time.Time
	BytesReceived int64
	BytesSent     int64
	MessagesIn    int64
	MessagesOut   int64
	LastActivity  time.Time
}

type WebSocketMetrics struct {
	ActiveConnections   int64
	TotalConnections    int64
	BytesReceived       int64
	BytesSent           int64
	MessagesReceived    int64
	MessagesSent        int64
	Errors              int64
	ConnectionDurations []time.Duration
	mutex               sync.Mutex
}

type WebSocketConfig struct {
	PingInterval      time.Duration
	PongWait          time.Duration
	WriteWait         time.Duration
	ReadBufferSize    int
	WriteBufferSize   int
	MaxMessageSize    int64
	EnableCompression bool
	AllowedOrigins    []string
	HeadersToForward  []string
}

func NewWebSocketProxy(router interfaces.Router, logger logging.Logger) *WebSocketProxy {
	metrics := &WebSocketMetrics{
		ConnectionDurations: make([]time.Duration, 0, 100),
	}

	config := &WebSocketConfig{
		PingInterval:      30 * time.Second,
		PongWait:          60 * time.Second,
		WriteWait:         10 * time.Second,
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		MaxMessageSize:    512 * 1024,
		EnableCompression: true,
		AllowedOrigins:    []string{"*"},
		HeadersToForward:  []string{"Authorization", "Cookie", "User-Agent"},
	}

	return &WebSocketProxy{
		router:      router,
		logger:      logger,
		connections: make(map[string]*ConnectionInfo),
		metrics:     metrics,
		config:      config,
	}
}

func (p *WebSocketProxy) HandleRequest(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "WebSocket upgrade required",
		})
	}

	route, err := p.findRoute(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	isEnabled, wsConfig := p.isWebSocketEnabled(route)
	if !isEnabled {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": ErrWebSocketNotEnabled.Error(),
		})
	}

	return websocket.New(func(clientConn *websocket.Conn) {
		p.handleConnection(clientConn, route, wsConfig)
	}, websocket.Config{
		ReadBufferSize:    p.config.ReadBufferSize,
		WriteBufferSize:   p.config.WriteBufferSize,
		EnableCompression: p.config.EnableCompression,
	})(c)
}

func (p *WebSocketProxy) findRoute(c *fiber.Ctx) (map[string]interface{}, error) {
	httpReq := &http.Request{
		Method:     c.Method(),
		URL:        createURL(c.Path(), c.Query()),
		Header:     make(http.Header),
		Host:       c.Hostname(),
		RemoteAddr: c.IP(),
	}

	c.Request().Header.VisitAll(func(key, value []byte) {
		httpReq.Header.Add(string(key), string(value))
	})

	routeInfo, err := p.router.FindRoute(httpReq)
	if err != nil {
		return nil, errors.New("route not found")
	}

	return routeInfo, nil
}

func (p *WebSocketProxy) isWebSocketEnabled(route map[string]interface{}) (bool, map[string]interface{}) {
	wsConfig, ok := route["websocket"].(map[string]interface{})
	if !ok {
		return false, nil
	}

	enabled, ok := wsConfig["enabled"].(bool)
	if !ok || !enabled {
		return false, nil
	}

	return true, wsConfig
}

func (p *WebSocketProxy) handleConnection(clientConn *websocket.Conn, route map[string]interface{}, wsConfig map[string]interface{}) {
	connectionID := generateConnectionID()
	startTime := time.Now()

	p.incrementActiveConnections()
	defer p.decrementActiveConnections()

	routeName, _ := route["name"].(string)
	p.logger.Info("WebSocket connection established",
		"id", connectionID,
		"route", routeName,
		"client", clientConn.RemoteAddr().String())

	upstreamURL := getUpstreamURL(route, wsConfig)
	if upstreamURL == "" {
		p.logger.Error("Invalid upstream URL configuration", "id", connectionID)
		p.closeConnection(clientConn, websocket.CloseInternalServerErr, "Invalid upstream configuration")
		return
	}

	headers := extractHeadersToForward(clientConn, p.config.HeadersToForward)
	upstreamConn, _, err := websocket.DefaultDialer.Dial(upstreamURL, headers)
	if err != nil {
		p.logger.Error("Failed to connect to upstream",
			"id", connectionID,
			"upstream", upstreamURL,
			"error", err)
		p.closeConnection(clientConn, websocket.CloseInternalServerErr, "Failed to connect to service")
		return
	}
	defer upstreamConn.Close()

	connInfo := &ConnectionInfo{
		ID:           connectionID,
		ClientConn:   clientConn,
		UpstreamConn: upstreamConn,
		Route:        routeName,
		StartTime:    startTime,
		LastActivity: startTime,
	}

	p.registerConnection(connInfo)
	defer p.unregisterConnection(connectionID)

	done := make(chan struct{})
	defer close(done)

	go p.handleClientMessages(connInfo, done)
	go p.handleUpstreamMessages(connInfo, done)
	go p.pingClient(connInfo, done)

	<-done
	duration := time.Since(startTime)
	p.recordConnectionDuration(duration)

	p.logger.Info("WebSocket connection closed",
		"id", connectionID,
		"route", routeName,
		"duration", duration.String(),
		"bytes_in", connInfo.BytesReceived,
		"bytes_out", connInfo.BytesSent,
		"messages_in", connInfo.MessagesIn,
		"messages_out", connInfo.MessagesOut)
}

func (p *WebSocketProxy) handleClientMessages(conn *ConnectionInfo, done chan struct{}) {
	conn.ClientConn.SetReadLimit(p.config.MaxMessageSize)
	conn.ClientConn.SetReadDeadline(time.Now().Add(p.config.PongWait))
	conn.ClientConn.SetPongHandler(func(string) error {
		conn.ClientConn.SetReadDeadline(time.Now().Add(p.config.PongWait))
		return nil
	})

	for {
		messageType, message, err := conn.ClientConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				p.logger.Error("Client read error",
					"id", conn.ID,
					"error", err)
				p.metrics.Errors++
			}
			return
		}

		conn.LastActivity = time.Now()
		conn.BytesReceived += int64(len(message))
		conn.MessagesIn++
		p.metrics.MessagesReceived++
		p.metrics.BytesReceived += int64(len(message))

		err = conn.UpstreamConn.WriteMessage(messageType, message)
		if err != nil {
			p.logger.Error("Upstream write error",
				"id", conn.ID,
				"error", err)
			p.metrics.Errors++
			return
		}

		conn.MessagesSent++
		conn.BytesSent += int64(len(message))
		p.metrics.MessagesSent++
		p.metrics.BytesSent += int64(len(message))
	}
}

func (p *WebSocketProxy) handleUpstreamMessages(conn *ConnectionInfo, done chan struct{}) {
	for {
		messageType, message, err := conn.UpstreamConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				p.logger.Error("Upstream read error",
					"id", conn.ID,
					"error", err)
				p.metrics.Errors++
			}
			done <- struct{}{}
			return
		}

		conn.LastActivity = time.Now()
		conn.BytesReceived += int64(len(message))
		conn.MessagesIn++
		p.metrics.MessagesReceived++
		p.metrics.BytesReceived += int64(len(message))

		err = conn.ClientConn.WriteMessage(messageType, message)
		if err != nil {
			p.logger.Error("Client write error",
				"id", conn.ID,
				"error", err)
			p.metrics.Errors++
			done <- struct{}{}
			return
		}

		conn.MessagesOut++
		conn.BytesSent += int64(len(message))
		p.metrics.MessagesSent++
		p.metrics.BytesSent += int64(len(message))
	}
}

func (p *WebSocketProxy) pingClient(conn *ConnectionInfo, done chan struct{}) {
	ticker := time.NewTicker(p.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			conn.ClientConn.SetWriteDeadline(time.Now().Add(p.config.WriteWait))
			if err := conn.ClientConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				p.logger.Error("Ping failed", "id", conn.ID, "error", err)
				done <- struct{}{}
				return
			}
		case <-done:
			return
		}
	}
}

func (p *WebSocketProxy) closeConnection(conn *websocket.Conn, code int, message string) {
	closeMessage := websocket.FormatCloseMessage(code, message)
	conn.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(time.Second))
	conn.Close()
}

func (p *WebSocketProxy) incrementActiveConnections() {
	p.metrics.mutex.Lock()
	p.metrics.ActiveConnections++
	p.metrics.TotalConnections++
	p.metrics.mutex.Unlock()
}

func (p *WebSocketProxy) decrementActiveConnections() {
	p.metrics.mutex.Lock()
	p.metrics.ActiveConnections--
	p.metrics.mutex.Unlock()
}

func (p *WebSocketProxy) recordConnectionDuration(duration time.Duration) {
	p.metrics.mutex.Lock()
	p.metrics.ConnectionDurations = append(p.metrics.ConnectionDurations, duration)
	if len(p.metrics.ConnectionDurations) > 1000 {
		p.metrics.ConnectionDurations = p.metrics.ConnectionDurations[len(p.metrics.ConnectionDurations)-1000:]
	}
	p.metrics.mutex.Unlock()
}

func (p *WebSocketProxy) registerConnection(conn *ConnectionInfo) {
	p.connectionsMutex.Lock()
	defer p.connectionsMutex.Unlock()
	p.connections[conn.ID] = conn
}

func (p *WebSocketProxy) unregisterConnection(id string) {
	p.connectionsMutex.Lock()
	defer p.connectionsMutex.Unlock()
	delete(p.connections, id)
}

func (p *WebSocketProxy) GetActiveConnections() int {
	p.connectionsMutex.RLock()
	defer p.connectionsMutex.RUnlock()
	return len(p.connections)
}

func (p *WebSocketProxy) SetConfig(config *WebSocketConfig) {
	p.config = config
}

func (p *WebSocketProxy) GetMetrics() *WebSocketMetrics {
	return p.metrics
}

func (p *WebSocketProxy) CloseAllConnections() {
	p.connectionsMutex.Lock()
	connections := make([]*ConnectionInfo, 0, len(p.connections))
	for _, conn := range p.connections {
		connections = append(connections, conn)
	}
	p.connectionsMutex.Unlock()

	for _, conn := range connections {
		p.closeConnection(conn.ClientConn, websocket.CloseNormalClosure, "Server shutdown")
	}
}

func createURL(path string, queryString string) *url.URL {
	u := &url.URL{
		Path: path,
	}
	if queryString != "" {
		u.RawQuery = queryString
	}
	return u
}

func getUpstreamURL(route map[string]interface{}, wsConfig map[string]interface{}) string {
	wsUpstream, ok := wsConfig["upstream_url"].(string)
	if ok && wsUpstream != "" {
		return ensureWebSocketProtocol(wsUpstream)
	}

	upstreamURL, ok := route["upstream_url"].(string)
	if !ok || upstreamURL == "" {
		return ""
	}

	return ensureWebSocketProtocol(upstreamURL)
}

func ensureWebSocketProtocol(url string) string {
	if strings.HasPrefix(url, "ws://") || strings.HasPrefix(url, "wss://") {
		return url
	}

	if strings.HasPrefix(url, "http://") {
		return "ws://" + url[7:]
	}

	if strings.HasPrefix(url, "https://") {
		return "wss://" + url[8:]
	}

	return "ws://" + url
}

func extractHeadersToForward(conn *websocket.Conn, headersToForward []string) http.Header {
	headers := http.Header{}

	for _, headerName := range headersToForward {
		if value := conn.Headers(headerName); value != "" {
			headers.Set(headerName, value)
		}
	}

	return headers
}

func generateConnectionID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
