package event

import "context"

// Event is a domain event that can be published through the bus.
type Event interface {
	EventName() string
}

// Handler processes a single event.
type Handler func(ctx context.Context, payload []byte, metadata Metadata) error

// Metadata carries cross-cutting context for an event.
type Metadata struct {
	TraceID  string
	TenantID string
	UserID   string
}

// Bus abstracts the event broker implementation.
type Bus interface {
	Publish(ctx context.Context, e Event) error
	Subscribe(subject string, h Handler) error
	Close() error
}
