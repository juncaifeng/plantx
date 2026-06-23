// Package repo provides audit-service repository implementations.
package repo

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/plantx/platform/audit-service/internal/domain"
)

// InMemoryRepo is a temporary in-memory repository for audit logs.
type InMemoryRepo struct {
	mu   sync.RWMutex
	logs []*domain.Log
}

// NewInMemoryRepo creates a new InMemoryRepo pre-populated with sample logs.
func NewInMemoryRepo() *InMemoryRepo {
	now := time.Now()
	r := &InMemoryRepo{
		logs: []*domain.Log{
			{ID: uuid.NewString(), TenantID: "tenant-1", UserID: "user-1", Action: "create", Resource: "order", Timestamp: now.Add(-time.Hour), Detail: "created order ORD-001"},
			{ID: uuid.NewString(), TenantID: "tenant-1", UserID: "user-2", Action: "list", Resource: "order", Timestamp: now.Add(-30 * time.Minute), Detail: "listed orders"},
			{ID: uuid.NewString(), TenantID: "tenant-2", UserID: "user-3", Action: "create", Resource: "tenant", Timestamp: now.Add(-15 * time.Minute), Detail: "created tenant tenant-2"},
		},
	}
	return r
}

// Query returns audit logs matching the provided filter.
func (r *InMemoryRepo) Query(_ context.Context, filter domain.Filter) ([]*domain.Log, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.Log
	for _, l := range r.logs {
		if filter.TenantID != "" && l.TenantID != filter.TenantID {
			continue
		}
		if filter.StartTime != 0 && l.Timestamp.Unix() < filter.StartTime {
			continue
		}
		if filter.EndTime != 0 && l.Timestamp.Unix() > filter.EndTime {
			continue
		}
		result = append(result, l)
		if filter.Limit > 0 && len(result) >= int(filter.Limit) {
			break
		}
	}
	return result, nil
}

// Save stores an audit log entry.
func (r *InMemoryRepo) Save(_ context.Context, log *domain.Log) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if log.ID == "" {
		log.ID = uuid.NewString()
	}
	r.logs = append(r.logs, log)
	return nil
}
