package app

import (
	"context"

	"github.com/plantx/services/test-service/internal/domain"
)

// Service implements the TestService business logic.
type Service struct{}

// New creates a new Service.
func New() *Service {
	return &Service{}
}

// Ping returns a simple pong message.
func (s *Service) Ping(ctx context.Context) string {
	return "pong"
}

// Echo returns the requested message along with caller identity metadata.
func (s *Service) Echo(ctx context.Context, msg, userID, tenantID string) *domain.EchoResult {
	return &domain.EchoResult{
		Message:  msg,
		UserID:   userID,
		TenantID: tenantID,
	}
}
