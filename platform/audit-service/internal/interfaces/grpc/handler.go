// Package grpc implements the audit-service gRPC handlers.
package grpc

import (
	"context"

	"github.com/plantx/platform/audit-service/api"
	"github.com/plantx/platform/audit-service/internal/app"
	"github.com/plantx/platform/audit-service/internal/domain"
)

// Handler implements the AuditService gRPC server.
type Handler struct {
	api.UnimplementedAuditServiceServer
	app *app.AuditService
}

// NewHandler creates a new Handler.
func NewHandler(app *app.AuditService) *Handler {
	return &Handler{app: app}
}

// ListAuditLogs handles querying audit logs.
func (h *Handler) ListAuditLogs(ctx context.Context, req *api.ListAuditLogsRequest) (*api.ListAuditLogsResponse, error) {
	logs, err := h.app.ListLogs(ctx, req.GetTenantId(), req.GetStartTime(), req.GetEndTime(), req.GetLimit())
	if err != nil {
		return nil, err
	}
	out := make([]*api.AuditLog, 0, len(logs))
	for _, l := range logs {
		out = append(out, toProto(l))
	}
	return &api.ListAuditLogsResponse{Logs: out}, nil
}

func toProto(l *domain.Log) *api.AuditLog {
	return &api.AuditLog{
		Id:        l.ID,
		TenantId:  l.TenantID,
		UserId:    l.UserID,
		Action:    l.Action,
		Resource:  l.Resource,
		Timestamp: l.Timestamp.Unix(),
		Detail:    l.Detail,
	}
}
