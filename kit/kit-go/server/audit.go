package server

// AuditEvent is published by the kit server logging interceptor for every
// gRPC/HTTP request so that platform audit-service can collect it.
type AuditEvent struct {
	Method    string `json:"method"`
	Action    string `json:"action"`
	Resource  string `json:"resource,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	TenantID  string `json:"tenant_id,omitempty"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

// EventName returns the subject used to publish/subscribe audit events.
func (e *AuditEvent) EventName() string { return "audit.events" }
