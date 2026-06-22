package app

import (
	"context"

	"github.com/plantx/platform/audit-service/internal/domain"
)

// AuditService implements audit log use cases.
type AuditService struct {
	repo domain.Repository
}

// NewAuditService creates a new AuditService.
func NewAuditService(repo domain.Repository) *AuditService {
	return &AuditService{repo: repo}
}

// ListLogs queries audit logs using the provided filter.
func (s *AuditService) ListLogs(ctx context.Context, tenantID string, startTime, endTime int64, limit int32) ([]*domain.Log, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.repo.Query(ctx, domain.Filter{
		TenantID:  tenantID,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     limit,
	})
}

// SaveLog persists a single audit log entry.
func (s *AuditService) SaveLog(ctx context.Context, log *domain.Log) error {
	return s.repo.Save(ctx, log)
}
