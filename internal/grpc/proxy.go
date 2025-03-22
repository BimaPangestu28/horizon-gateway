package grpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/types"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// ProxyHandler represents a gRPC proxy handler
type ProxyHandler struct {
	router       interfaces.Router
	logger       logging.Logger
	server       *grpc.Server
	descriptors  map[string]*descriptorMap
	connections  map[string]*grpc.ClientConn
	connMutex    sync.RWMutex
	interceptors []grpc.UnaryServerInterceptor
	metrics      *GRPCMetrics
	addr         string
	port         int
	tlsEnabled   bool
	certFile     string
	keyFile      string
	initialized  bool
	listeners    []net.Listener
}

// descriptorMap maps service names to their method descriptors
type descriptorMap struct {
	serviceDesc *grpc.ServiceDesc
	methods     map[string]*methodInfo
}

// methodInfo holds information about a gRPC method
type methodInfo struct {
	methodName     string
	serviceName    string
	isClientStream bool
	isServerStream bool
	handler        grpc.MethodDesc
}

// GRPCMetrics tracks gRPC metrics
type GRPCMetrics struct {
	RequestsTotal  int64
	ErrorsTotal    int64
	ErrorsByCode   map[codes.Code]int64
	RequestLatency []time.Duration
	ActiveRequests int64
	mu             sync.Mutex
}

// NewProxyHandler creates a new gRPC proxy handler
func NewProxyHandler(router interfaces.Router, logger logging.Logger, port int) *ProxyHandler {
	metrics := &GRPCMetrics{
		ErrorsByCode:   make(map[codes.Code]int64),
		RequestLatency: make([]time.Duration, 0, 100),
	}

	return &ProxyHandler{
		router:       router,
		logger:       logger,
		descriptors:  make(map[string]*descriptorMap),
		connections:  make(map[string]*grpc.ClientConn),
		interceptors: make([]grpc.UnaryServerInterceptor, 0),
		metrics:      metrics,
		port:         port,
		initialized:  false,
		listeners:    make([]net.Listener, 0),
	}
}

// Initialize initializes the gRPC server
func (h *ProxyHandler) Initialize() error {
	if h.initialized {
		return nil
	}

	// Create server options
	var opts []grpc.ServerOption

	// Add keepalive options
	opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Minute,
		MaxConnectionAge:      30 * time.Minute,
		MaxConnectionAgeGrace: 5 * time.Second,
		Time:                  5 * time.Minute,
		Timeout:               20 * time.Second,
	}))

	// Add interceptors
	opts = append(opts, grpc.ChainUnaryInterceptor(h.loggingInterceptor, h.metricInterceptor))

	// Add TLS if configured
	if h.tlsEnabled {
		creds, err := credentials.NewServerTLSFromFile(h.certFile, h.keyFile)
		if err != nil {
			return fmt.Errorf("failed to create TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Create server
	h.server = grpc.NewServer(opts...)

	// Register unknown service handler
	grpc.RegisterUnknownServiceHandler(h.server, h.handleUnknownService)

	// Enable reflection service
	reflection.Register(h.server)

	h.initialized = true
	return nil
}

// Start starts the gRPC server
func (h *ProxyHandler) Start() error {
	if !h.initialized {
		if err := h.Initialize(); err != nil {
			return err
		}
	}

	// Create listener
	addr := fmt.Sprintf(":%d", h.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create gRPC listener on %s: %w", addr, err)
	}

	h.listeners = append(h.listeners, listener)
	h.addr = addr

	h.logger.Info("Starting gRPC server", "addr", addr)

	// Serve in a goroutine
	go func() {
		if err := h.server.Serve(listener); err != nil {
			h.logger.Error("gRPC server error", "error", err)
		}
	}()

	return nil
}

// Stop stops the gRPC server
func (h *ProxyHandler) Stop() error {
	if h.server != nil {
		h.server.GracefulStop()
	}

	h.connMutex.Lock()
	defer h.connMutex.Unlock()

	for addr, conn := range h.connections {
		if err := conn.Close(); err != nil {
			h.logger.Error("Error closing gRPC connection", "addr", addr, "error", err)
		}
	}

	for _, listener := range h.listeners {
		if err := listener.Close(); err != nil {
			h.logger.Error("Error closing gRPC listener", "error", err)
		}
	}

	h.initialized = false
	return nil
}

// loggingInterceptor logs incoming gRPC requests
func (h *ProxyHandler) loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	h.logger.Info("gRPC request received",
		"method", info.FullMethod,
		"client", getClientAddress(ctx))

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	if err != nil {
		st, _ := status.FromError(err)
		h.logger.Error("gRPC request error",
			"method", info.FullMethod,
			"client", getClientAddress(ctx),
			"duration", duration,
			"code", st.Code(),
			"error", err)
	} else {
		h.logger.Info("gRPC request completed",
			"method", info.FullMethod,
			"client", getClientAddress(ctx),
			"duration", duration)
	}

	return resp, err
}

// metricInterceptor collects metrics for gRPC requests
func (h *ProxyHandler) metricInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	h.metrics.mu.Lock()
	h.metrics.RequestsTotal++
	h.metrics.ActiveRequests++
	h.metrics.mu.Unlock()

	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	h.metrics.mu.Lock()
	h.metrics.ActiveRequests--
	h.metrics.RequestLatency = append(h.metrics.RequestLatency, duration)
	if err != nil {
		h.metrics.ErrorsTotal++
		st, _ := status.FromError(err)
		h.metrics.ErrorsByCode[st.Code()]++
	}
	h.metrics.mu.Unlock()

	return resp, err
}

// handleUnknownService handles requests for services that aren't explicitly registered
func (h *ProxyHandler) handleUnknownService(srv interface{}, stream grpc.ServerStream) error {
	fullMethodName, ok := grpc.MethodFromServerStream(stream)
	if !ok {
		return status.Errorf(codes.Internal, "method name not found in stream context")
	}

	// Parse method and service name
	serviceName, methodName := parseMethodName(fullMethodName)
	if serviceName == "" || methodName == "" {
		return status.Errorf(codes.InvalidArgument, "invalid method name format: %s", fullMethodName)
	}

	h.logger.Info("Handling gRPC proxy request",
		"method", fullMethodName,
		"service", serviceName,
		"method", methodName)

	// Get metadata from context
	md, _ := metadata.FromIncomingContext(stream.Context())

	// Create a new context with incoming metadata
	outCtx := metadata.NewOutgoingContext(stream.Context(), md)

	// Create HTTP request route path for router
	routePath := fmt.Sprintf("/grpc/%s/%s", serviceName, methodName)

	// Find route for this service
	httpReq := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Path: routePath,
		},
		Header: http.Header{
			"Content-Type": []string{"application/grpc"},
		},
		Host: getClientAddress(stream.Context()),
	}

	// Pass headers from metadata to HTTP request
	for k, vs := range md {
		for _, v := range vs {
			httpReq.Header.Add(k, v)
		}
	}

	route, err := h.router.FindRoute(httpReq)
	if err != nil {
		return status.Errorf(codes.NotFound, "no route found for service %s", serviceName)
	}

	// Check if gRPC is enabled for this route
	if route.GRPC == nil || !route.GRPC.Enabled {
		return status.Errorf(codes.Unimplemented, "gRPC not enabled for route %s", route.Name)
	}

	// Get upstream connection
	conn, err := h.getUpstreamConnection(route, serviceName)
	if err != nil {
		h.logger.Error("Failed to get upstream connection",
			"route", route.Name,
			"service", serviceName,
			"error", err)
		return status.Errorf(codes.Unavailable, "failed to connect to upstream service")
	}

	// Create client stream
	clientStream, err := conn.NewStream(outCtx, &grpc.StreamDesc{
		ServerStreams: stream.FullMethod().isServerStream,
		ClientStreams: stream.FullMethod().isClientStream,
	}, fullMethodName)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create client stream: %v", err)
	}

	// Proxy data between client and upstream service
	return h.proxyStreams(stream, clientStream)
}

// getUpstreamConnection gets or creates a connection to an upstream gRPC service
func (h *ProxyHandler) getUpstreamConnection(route *types.Route, serviceName string) (*grpc.ClientConn, error) {
	// Get upstream URL from route
	upstreamURL := route.GRPC.UpstreamURL
	if upstreamURL == "" {
		upstreamURL = route.UpstreamURL
	}

	h.connMutex.RLock()
	conn, exists := h.connections[upstreamURL]
	h.connMutex.RUnlock()

	if exists {
		return conn, nil
	}

	h.connMutex.Lock()
	defer h.connMutex.Unlock()

	// Check again after acquiring the write lock
	conn, exists = h.connections[upstreamURL]
	if exists {
		return conn, nil
	}

	// Create connection options
	opts := []grpc.DialOption{
		grpc.WithBlock(),
	}

	// Add keepalive options
	if route.GRPC.KeepAlive > 0 {
		opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                route.GRPC.KeepAlive,
			Timeout:             route.GRPC.KeepAliveTimeout,
			PermitWithoutStream: true,
		}))
	}

	// Add TLS if needed
	if strings.HasPrefix(upstreamURL, "https://") {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// Create max message size option if configured
	if route.GRPC.MaxMessageSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(route.GRPC.MaxMessageSize),
			grpc.MaxCallSendMsgSize(route.GRPC.MaxMessageSize),
		))
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to upstream service
	conn, err := grpc.DialContext(ctx, upstreamURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to upstream gRPC service: %w", err)
	}

	// Store connection for reuse
	h.connections[upstreamURL] = conn
	return conn, nil
}

// proxyStreams proxies data between client and upstream service streams
func (h *ProxyHandler) proxyStreams(clientStream grpc.ServerStream, upstreamStream grpc.ClientStream) error {
	// Create error channel
	errChan := make(chan error, 2)

	// Start forwarding in both directions
	go func() {
		err := h.forwardClientToUpstream(clientStream, upstreamStream)
		errChan <- err
	}()

	go func() {
		err := h.forwardUpstreamToClient(upstreamStream, clientStream)
		errChan <- err
	}()

	// Wait for one direction to complete or error
	err := <-errChan
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

// forwardClientToUpstream forwards messages from client to upstream service
func (h *ProxyHandler) forwardClientToUpstream(src grpc.ServerStream, dst grpc.ClientStream) error {
	for {
		// Receive message from client
		message := make([]byte, 0)
		err := src.RecvMsg(&message)
		if err == io.EOF {
			// Close the client stream
			dst.CloseSend()
			return nil
		}
		if err != nil {
			return err
		}

		// Send message to upstream
		if err := dst.SendMsg(message); err != nil {
			return err
		}
	}
}

// forwardUpstreamToClient forwards messages from upstream service to client
func (h *ProxyHandler) forwardUpstreamToClient(src grpc.ClientStream, dst grpc.ServerStream) error {
	for {
		// Receive message from upstream
		message := make([]byte, 0)
		err := src.RecvMsg(&message)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// Send message to client
		if err := dst.SendMsg(message); err != nil {
			return err
		}
	}
}

// parseMethodName parses a gRPC full method name into service and method names
func parseMethodName(fullMethod string) (service, method string) {
	if len(fullMethod) == 0 || fullMethod[0] != '/' {
		return "", ""
	}

	parts := strings.Split(fullMethod[1:], "/")
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}

// getClientAddress gets the client address from the context
func getClientAddress(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "unknown"
	}
	return p.Addr.String()
}

// GetMetrics returns gRPC metrics
func (h *ProxyHandler) GetMetrics() *GRPCMetrics {
	return h.metrics
}

// RegisterInterceptor registers a gRPC interceptor
func (h *ProxyHandler) RegisterInterceptor(interceptor grpc.UnaryServerInterceptor) {
	h.interceptors = append(h.interceptors, interceptor)
}
