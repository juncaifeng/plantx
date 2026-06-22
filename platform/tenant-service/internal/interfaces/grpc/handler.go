package grpc

import (
	"context"

	"github.com/plantx/platform/tenant-service/api"
	"github.com/plantx/platform/tenant-service/internal/app"
	"github.com/plantx/platform/tenant-service/internal/domain"
)

// Handler implements the TenantService gRPC server.
type Handler struct {
	api.UnimplementedTenantServiceServer
	app *app.TenantService
}

// NewHandler creates a new Handler.
func NewHandler(app *app.TenantService) *Handler {
	return &Handler{app: app}
}

// ListTenants handles listing tenants.
func (h *Handler) ListTenants(ctx context.Context, req *api.ListTenantsRequest) (*api.ListTenantsResponse, error) {
	tenants, err := h.app.ListTenants(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.Tenant, 0, len(tenants))
	for _, t := range tenants {
		out = append(out, toProto(t))
	}
	return &api.ListTenantsResponse{Tenants: out}, nil
}

// CreateTenant handles creating a tenant.
func (h *Handler) CreateTenant(ctx context.Context, req *api.CreateTenantRequest) (*api.Tenant, error) {
	t, err := h.app.CreateTenant(ctx, req.GetName())
	if err != nil {
		return nil, err
	}
	return toProto(t), nil
}

func toProto(t *domain.Tenant) *api.Tenant {
	return &api.Tenant{
		Id:        t.ID,
		Name:      t.Name,
		Status:    t.Status,
		CreatedAt: t.CreatedAt,
	}
}
