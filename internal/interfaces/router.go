package interfaces

import (
	"net/http"

	"github.com/bimapangestu28/horizon/internal/types"
)

// Router defines the interface for route finding functionality
type Router interface {
	FindRoute(req *http.Request) (*types.Route, error)
	GetAllRoutes() []*types.Route
}
