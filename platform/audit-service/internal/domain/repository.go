package domain

import "context"

// Filter defines optional filters for querying audit logs.
type Filter struct {
	TenantID  string
	StartTime int64
	EndTime   int64
	Limit     int32
}

// Repository defines persistence operations for audit logs.
type Repository interface {
	Query(ctx context.Context, filter Filter) ([]*Log, error)
	Save(ctx context.Context, log *Log) error
}
