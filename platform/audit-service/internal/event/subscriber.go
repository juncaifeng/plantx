// Package event contains audit-service event bus subscribers.
package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/plantx/kit/kit-go/event"
	"github.com/plantx/kit/kit-go/server"
	"github.com/plantx/platform/audit-service/internal/app"
	"github.com/plantx/platform/audit-service/internal/domain"
)

// Subscriber consumes audit events from the event bus.
type Subscriber struct {
	app *app.AuditService
}

// NewSubscriber creates a new audit event subscriber.
func NewSubscriber(app *app.AuditService) *Subscriber {
	return &Subscriber{app: app}
}

// Handle processes a raw audit event payload and persists it.
func (s *Subscriber) Handle(ctx context.Context, payload []byte, _ event.Metadata) error {
	var evt server.AuditEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return err
	}
	ts := time.Now()
	if evt.Timestamp > 0 {
		ts = time.UnixMilli(evt.Timestamp)
	}
	log := &domain.Log{
		UserID:    evt.UserID,
		TenantID:  evt.TenantID,
		Action:    evt.Action,
		Resource:  evt.Resource,
		Timestamp: ts,
		Detail:    evt.Method,
	}
	return s.app.SaveLog(ctx, log)
}
