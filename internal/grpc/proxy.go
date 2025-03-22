package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	ErrGRPCNotEnabled  = errors.New("gRPC is not enabled for this route")
	ErrServiceNotFound = errors.New("service not found")
	ErrMethodNotFound  = errors.New("method not found")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrUpstreamFailure = errors.New("upstream service failure")
)

type GRPCProxy struct {
	router        interfaces.Router
	logger        logging.Logger
	server        *grpc.Server
	connections   map[string]*grpc.ClientConn
	connMutex     sync.RWMutex
	metrics       *GRPCMetrics
	port          int
	serverStarted bool
	listeners     []net.Listener
}

type GRPCMetrics struct {
	RequestsTotal    int64
	RequestsActive   int64
	RequestLatencies []time.Duration
	ErrorsTotal      int64
	ErrorsByCode     map[codes.Code]int64
	ServiceCalls     map[string]int64
	mutex            sync.Mutex
}

func NewGRPCProxy(router interfaces.Router, logger logging.Logger, port int) *GRPCProxy {
	metrics := &GRPCMetrics{
		RequestLatencies: make([]time.Duration, 0, 100),
		ErrorsByCode:     make(map[codes.Code]int64),
		ServiceCalls:     make(map[string]int64),
	}

	return &GRPCProxy{
		router:        router,
		logger:        logger,
		connections:   make(map[string]*grpc.ClientConn),
		metrics:       metrics,
		port:          port,
		serverStarted: false,
		listeners:     make([]net.Listener, 0),
	}
}

func (p *GRPCProxy) Initialize() error {
	serverOptions := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute,
			MaxConnectionAge:      30 * time.Minute,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Minute,
			Timeout:               20 * time.Second,
		}),
		grpc.ChainUnaryInterceptor(
			p.loggingInterceptor,
			p.metricsInterceptor,
		),
	}

	p.server = grpc.NewServer(serverOptions...)

	grpc.UnknownServiceHandler(p.server, p.handleUnknownService)
	reflection.Register(p.server)

	return nil
}

func (p *GRPCProxy) Start() error {
	if p.serverStarted {
		return nil
	}

	if p.server == nil {
		if err := p.Initialize(); err != nil {
			return err
		}
	}

	address := fmt.Sprintf(":%d", p.port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	p.listeners = append(p.listeners, listener)
	p.serverStarted = true

	p.logger.Info("Starting gRPC server", "address", address)

	go func() {
		if err := p.server.Serve(listener); err != nil {
			p.logger.Error("gRPC server error", "error", err)
		}
	}()

	return nil
}

func (p *GRPCProxy) Stop() error {
	if !p.serverStarted {
		return nil
	}

	p.server.GracefulStop()
	p.serverStarted = false

	p.connMutex.Lock()
	defer p.connMutex.Unlock()

	for address, conn := range p.connections {
		if err := conn.Close(); err != nil {
			p.logger.Error("Error closing gRPC connection", "address", address, "error", err)
		}
	}

	return nil
}

func (p *GRPCProxy) loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	startTime := time.Now()

	p.logger.Info("gRPC request received",
		"method", info.FullMethod,
		"client_addr", getClientAddress(ctx))

	resp, err := handler(ctx, req)

	duration := time.Since(startTime)
	if err != nil {
		st, _ := status.FromError(err)
		p.logger.Error("gRPC request failed",
			"method", info.FullMethod,
			"code", st.Code().String(),
			"duration_ms", duration.Milliseconds(),
			"error", err)
	} else {
		p.logger.Info("gRPC request completed",
			"method", info.FullMethod,
			"duration_ms", duration.Milliseconds())
	}

	return resp, err
}

func (p *GRPCProxy) metricsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	p.metrics.mutex.Lock()
	p.metrics.RequestsTotal++
	p.metrics.RequestsActive++
	serviceName := extractServiceFromMethod(info.FullMethod)
	p.metrics.ServiceCalls[serviceName]++
	p.metrics.mutex.Unlock()

	startTime := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(startTime)

	p.metrics.mutex.Lock()
	p.metrics.RequestsActive--
	p.metrics.RequestLatencies = append(p.metrics.RequestLatencies, duration)

	if len(p.metrics.RequestLatencies) > 1000 {
		p.metrics.RequestLatencies = p.metrics.RequestLatencies[len(p.metrics.RequestLatencies)-1000:]
	}

	if err != nil {
		p.metrics.ErrorsTotal++
		st, _ := status.FromError(err)
		p.metrics.ErrorsByCode[st.Code()]++
	}
	p.metrics.mutex.Unlock()

	return resp, err
}

func (p *GRPCProxy) handleUnknownService(srv interface{}, stream grpc.ServerStream) error {
	fullMethod, ok := grpc.MethodFromServerStream(stream)
	if !ok {
		return status.Error(codes.Internal, "failed to get method from stream")
	}

	serviceName, methodName := splitMethodName(fullMethod)
	if serviceName == "" || methodName == "" {
		return status.Errorf(codes.InvalidArgument, "invalid method name format: %s", fullMethod)
	}

	p.logger.Info("Handling gRPC request",
		"full_method", fullMethod,
		"service", serviceName,
		"method", methodName)

	md, _ := metadata.FromIncomingContext(stream.Context())
	outCtx := metadata.NewOutgoingContext(stream.Context(), md)

	route, err := p.findRoute(serviceName, methodName)
	if err != nil {
		return status.Errorf(codes.NotFound, "service route not found: %s", err.Error())
	}

	grpcConfig, ok := route["grpc"].(map[string]interface{})
	if !ok || !isGRPCEnabled(grpcConfig) {
		return status.Errorf(codes.Unimplemented, "gRPC not enabled for route")
	}

	upstreamConn, err := p.getUpstreamConnection(route)
	if err != nil {
		p.logger.Error("Failed to connect to upstream",
			"service", serviceName,
			"error", err)
		return status.Errorf(codes.Unavailable, "upstream connection failed: %v", err)
	}

	isClientStream, isServerStream := checkStreamingType(stream)

	clientStream, err := createClientStream(upstreamConn, outCtx, fullMethod, isClientStream, isServerStream)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create upstream stream: %v", err)
	}

	return proxyStream(stream, clientStream)
}

func (p *GRPCProxy) findRoute(serviceName, methodName string) (map[string]interface{}, error) {
	httpPath := fmt.Sprintf("/grpc/%s/%s", serviceName, methodName)

	httpReq := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Path: httpPath,
		},
		Header: http.Header{
			"Content-Type": []string{"application/grpc"},
		},
	}

	route, err := p.router.FindRoute(httpReq)
	if err != nil {
		return nil, err
	}

	return route, nil
}

func (p *GRPCProxy) getUpstreamConnection(route map[string]interface{}) (*grpc.ClientConn, error) {
	upstreamURL, err := getUpstreamURL(route)
	if err != nil {
		return nil, err
	}

	p.connMutex.RLock()
	conn, exists := p.connections[upstreamURL]
	p.connMutex.RUnlock()

	if exists {
		return conn, nil
	}

	p.connMutex.Lock()
	defer p.connMutex.Unlock()

	conn, exists = p.connections[upstreamURL]
	if exists {
		return conn, nil
	}

	grpcConfig, _ := route["grpc"].(map[string]interface{})
	options := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTimeout(10 * time.Second),
	}

	if strings.HasPrefix(upstreamURL, "https://") || strings.HasPrefix(upstreamURL, "grpcs://") {
		options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	maxSize := getMaxMessageSize(grpcConfig)
	if maxSize > 0 {
		options = append(options,
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(maxSize),
				grpc.MaxCallSendMsgSize(maxSize),
			),
		)
	}

	cleanURL := cleanupGRPCURL(upstreamURL)
	conn, err = grpc.Dial(cleanURL, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	p.connections[upstreamURL] = conn
	return conn, nil
}

func (p *GRPCProxy) GetMetrics() *GRPCMetrics {
	return p.metrics
}

func getClientAddress(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "unknown"
	}
	return p.Addr.String()
}

func splitMethodName(fullMethod string) (serviceName, methodName string) {
	if !strings.HasPrefix(fullMethod, "/") {
		return "", ""
	}

	parts := strings.Split(fullMethod[1:], "/")
	if len(parts) != 2 {
		return "", ""
	}

	return parts[0], parts[1]
}

func extractServiceFromMethod(fullMethod string) string {
	service, _ := splitMethodName(fullMethod)
	return service
}

func isGRPCEnabled(grpcConfig map[string]interface{}) bool {
	enabled, ok := grpcConfig["enabled"].(bool)
	return ok && enabled
}

func getUpstreamURL(route map[string]interface{}) (string, error) {
	grpcConfig, ok := route["grpc"].(map[string]interface{})
	if ok {
		if upstreamURL, ok := grpcConfig["upstream_url"].(string); ok && upstreamURL != "" {
			return upstreamURL, nil
		}
	}

	if upstreamURL, ok := route["upstream_url"].(string); ok && upstreamURL != "" {
		return upstreamURL, nil
	}

	return "", errors.New("no upstream URL configured")
}

func getMaxMessageSize(grpcConfig map[string]interface{}) int {
	if grpcConfig == nil {
		return 0
	}

	if maxSize, ok := grpcConfig["max_message_size"].(int); ok {
		return maxSize
	}

	if maxSizeStr, ok := grpcConfig["max_message_size"].(string); ok {
		if maxSize, err := strconv.Atoi(maxSizeStr); err == nil {
			return maxSize
		}
	}

	return 0
}

func cleanupGRPCURL(url string) string {
	url = strings.TrimPrefix(url, "grpc://")
	url = strings.TrimPrefix(url, "grpcs://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	return url
}

func checkStreamingType(stream grpc.ServerStream) (isClientStream, isServerStream bool) {
	streamDesc := stream.StreamDesc()
	if streamDesc == nil {
		return false, false
	}
	return streamDesc.ClientStreams, streamDesc.ServerStreams
}

func createClientStream(conn *grpc.ClientConn, ctx context.Context, fullMethod string, isClientStream, isServerStream bool) (grpc.ClientStream, error) {
	desc := &grpc.StreamDesc{
		ServerStreams: isServerStream,
		ClientStreams: isClientStream,
	}
	return conn.NewStream(ctx, desc, fullMethod)
}

func proxyStream(serverStream grpc.ServerStream, clientStream grpc.ClientStream) error {
	egress := func() error {
		for {
			message := make([]byte, 0)
			err := serverStream.RecvMsg(&message)
			if err == io.EOF {
				clientStream.CloseSend()
				return nil
			}
			if err != nil {
				return err
			}

			err = clientStream.SendMsg(message)
			if err != nil {
				return err
			}
		}
	}

	ingress := func() error {
		for {
			message := make([]byte, 0)
			err := clientStream.RecvMsg(&message)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}

			err = serverStream.SendMsg(message)
			if err != nil {
				return err
			}
		}
	}

	done := make(chan error, 2)
	go func() {
		done <- egress()
	}()
	go func() {
		done <- ingress()
	}()

	err := <-done
	if err != nil && err != io.EOF {
		return err
	}

	return <-done
}
