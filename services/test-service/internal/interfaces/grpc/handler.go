package grpc

import (
	"context"

	kitctx "github.com/plantx/kit/kit-go/context"
	"github.com/plantx/services/test-service/api"
	"github.com/plantx/services/test-service/internal/app"
)

// Handler implements the TestService gRPC server.
type Handler struct {
	api.UnimplementedTestServiceServer
	app *app.Service
}

// NewHandler creates a new Handler.
func NewHandler(app *app.Service) *Handler {
	return &Handler{app: app}
}

// Ping handles health-check pings.
func (h *Handler) Ping(ctx context.Context, req *api.PingRequest) (*api.PongResponse, error) {
	return &api.PongResponse{Message: h.app.Ping(ctx)}, nil
}

// Echo handles echo requests.
func (h *Handler) Echo(ctx context.Context, req *api.EchoRequest) (*api.EchoResponse, error) {
	userID := ""
	if u := kitctx.GetUser(ctx); u != nil {
		userID = u.ID
	}
	tenantID := ""
	if t := kitctx.GetTenant(ctx); t.ID != "" {
		tenantID = t.ID
	}
	res := h.app.Echo(ctx, req.Message, userID, tenantID)
	return &api.EchoResponse{
		Message:  res.Message,
		UserId:   res.UserID,
		TenantId: res.TenantID,
	}, nil
}
