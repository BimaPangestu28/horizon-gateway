package grpc

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/bimapangestu28/horizon/internal/interfaces"
	"github.com/bimapangestu28/horizon/internal/types"
	"github.com/bimapangestu28/horizon/internal/utils/logging"
)

var (
	ErrGRPCNotEnabled          = errors.New("gRPC not enabled for this route")
	ErrGRPCServiceNotFound     = errors.New("gRPC service not found")
	ErrInvalidGRPCRequest      = errors.New("invalid gRPC request")
	ErrGRPCMethodNotFound      = errors.New("gRPC method not found")
	ErrProtoDescriptorNotFound = errors.New("proto descriptor not found")
)

type GRPCHandler struct {
	router        interfaces.Router
	logger        logging.Logger
	connections   map[string]*grpc.ClientConn
	connectionsMu sync.RWMutex
	descriptorsMu sync.RWMutex
	descriptors   map[string]*DescriptorInfo
}

type DescriptorInfo struct {
	Services   map[string]*ServiceInfo
	Methods    map[string]*MethodInfo
	Enums      map[string]*EnumInfo
	Messages   map[string]*MessageInfo
	ImportedAt time.Time
}

type ServiceInfo struct {
	Name        string
	FullName    string
	Methods     []*MethodInfo
	Description string
}

type MethodInfo struct {
	Name           string
	ServiceName    string
	InputType      string
	OutputType     string
	IsClientStream bool
	IsServerStream bool
	Description    string
}

type EnumInfo struct {
	Name        string
	FullName    string
	Values      map[string]int32
	Description string
}

type MessageInfo struct {
	Name        string
	FullName    string
	Fields      []*FieldInfo
	Description string
}

type FieldInfo struct {
	Name        string
	Number      int
	Type        string
	IsRepeated  bool
	Description string
}

type GRPCMetrics struct {
	RequestsTotal    int64
	RequestsActive   int64
	RequestDurations []time.Duration
	ErrorsTotal      int64
	ErrorsByCode     map[string]int64
	ErrorsByMethod   map[string]int64
	BytesReceived    int64
	BytesSent        int64
}

func NewGRPCHandler(router interfaces.Router, logger logging.Logger) *GRPCHandler {
	return &GRPCHandler{
		router:      router,
		logger:      logger,
		connections: make(map[string]*grpc.ClientConn),
		descriptors: make(map[string]*DescriptorInfo),
	}
}

func (h *GRPCHandler) HandleGRPC(c *fiber.Ctx) error {
	route, err := h.router.FindRoute(c.Path(), c.Method())
	if err != nil {
		h.logger.Error("gRPC route not found", "path", c.Path(), "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "gRPC route not found",
		})
	}

	if !h.isGRPCEnabled(route) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "gRPC not enabled for this route",
		})
	}

	clientConn, err := h.getConnection(route)
	if err != nil {
		h.logger.Error("Failed to get gRPC connection",
			"route", route.Name,
			"error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect to gRPC service",
		})
	}

	serviceName := route.GRPC.ServiceName
	if serviceName == "" {
		serviceName = extractServiceFromPath(c.Path())
	}

	methodName := extractMethodFromPath(c.Path())
	if methodName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid gRPC method name",
		})
	}

	fullMethodName := fmt.Sprintf("/%s/%s", serviceName, methodName)

	ctx := c.Context()
	md := metadata.New(nil)

	// Extract headers to forward
	for _, header := range route.GRPC.ForwardHeaders {
		if val := c.Get(header); val != "" {
			md.Set(header, val)
		}
	}

	// Add metadata from query parameters
	for key, values := range c.Queries() {
		md.Set(key, values...)
	}

	// Create context with metadata
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Set timeout
	timeout := 30 * time.Second // Default timeout
	if route.Timeout != nil && route.Timeout.RequestTimeout > 0 {
		timeout = route.Timeout.RequestTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Get request body
	reqBody := c.Body()

	// Make gRPC call
	respBody, respMD, err := h.invokeGRPC(ctx, clientConn, fullMethodName, reqBody)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(grpcStatusToHTTP(st.Code())).JSON(fiber.Map{
			"error": st.Message(),
			"code":  st.Code().String(),
		})
	}

	// Set response headers from metadata
	for key, values := range respMD {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	// Set content type
	c.Set("Content-Type", "application/grpc+json")

	// Return response
	return c.Send(respBody)
}

func (h *GRPCHandler) invokeGRPC(ctx context.Context, conn *grpc.ClientConn, fullMethodName string, reqBody []byte) ([]byte, metadata.MD, error) {
	var respMD metadata.MD

	// Create a gRPC call
	callOpts := []grpc.CallOption{
		grpc.Header(&respMD),
	}

	// Invoke unary method
	resp, err := conn.Invoke(ctx, fullMethodName, reqBody, callOpts...)
	if err != nil {
		return nil, nil, err
	}

	// Extract response body
	respBody, ok := resp.([]byte)
	if !ok {
		return nil, respMD, fmt.Errorf("invalid response type: %T", resp)
	}

	return respBody, respMD, nil
}

func (h *GRPCHandler) getConnection(route *types.Route) (*grpc.ClientConn, error) {
	upstreamURL := route.GRPC.UpstreamURL
	if upstreamURL == "" {
		upstreamURL = route.UpstreamURL
	}

	h.connectionsMu.RLock()
	conn, exists := h.connections[upstreamURL]
	h.connectionsMu.RUnlock()

	if exists {
		return conn, nil
	}

	h.connectionsMu.Lock()
	defer h.connectionsMu.Unlock()

	// Double-check if another goroutine created the connection
	conn, exists = h.connections[upstreamURL]
	if exists {
		return conn, nil
	}

	// Create dial options
	dialOpts := []grpc.DialOption{
		grpc.WithBlock(),
	}

	// Check if TLS is enabled
	if strings.HasPrefix(upstreamURL, "https://") {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	// Add keepalive options
	if route.GRPC.KeepAlive > 0 {
		dialOpts = append(dialOpts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                route.GRPC.KeepAlive,
			Timeout:             route.GRPC.KeepAliveTimeout,
			PermitWithoutStream: true,
		}))
	}

	// Create connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, upstreamURL, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	// Store connection
	h.connections[upstreamURL] = conn
	return conn, nil
}

func (h *GRPCHandler) isGRPCEnabled(route *types.Route) bool {
	return route.GRPC != nil && route.GRPC.Enabled
}

func extractServiceFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}

func extractMethodFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func grpcStatusToHTTP(code int32) int {
	switch code {
	case 0: // OK
		return fiber.StatusOK
	case 1: // Canceled
		return fiber.StatusRequestTimeout
	case 2: // Unknown
		return fiber.StatusInternalServerError
	case 3: // InvalidArgument
		return fiber.StatusBadRequest
	case 4: // DeadlineExceeded
		return fiber.StatusGatewayTimeout
	case 5: // NotFound
		return fiber.StatusNotFound
	case 6: // AlreadyExists
		return fiber.StatusConflict
	case 7: // PermissionDenied
		return fiber.StatusForbidden
	case 8: // ResourceExhausted
		return fiber.StatusTooManyRequests
	case 9: // FailedPrecondition
		return fiber.StatusPreconditionFailed
	case 10: // Aborted
		return fiber.StatusConflict
	case 11: // OutOfRange
		return fiber.StatusBadRequest
	case 12: // Unimplemented
		return fiber.StatusNotImplemented
	case 13: // Internal
		return fiber.StatusInternalServerError
	case 14: // Unavailable
		return fiber.StatusServiceUnavailable
	case 15: // DataLoss
		return fiber.StatusInternalServerError
	case 16: // Unauthenticated
		return fiber.StatusUnauthorized
	default:
		return fiber.StatusInternalServerError
	}
}

func (h *GRPCHandler) Shutdown() error {
	h.connectionsMu.Lock()
	defer h.connectionsMu.Unlock()

	var errs []error
	for url, conn := range h.connections {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error closing gRPC connection to %s: %w", url, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors shutting down gRPC handler: %v", errs)
	}

	return nil
}
