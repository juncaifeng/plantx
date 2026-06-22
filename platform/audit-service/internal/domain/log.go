package domain

import "time"

// Log represents a single audit log entry.
type Log struct {
	ID        string
	TenantID  string
	UserID    string
	Action    string
	Resource  string
	Timestamp time.Time
	Detail    string
}
