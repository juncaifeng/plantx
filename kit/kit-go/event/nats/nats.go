// Package nats provides a NATS JetStream event bus implementation.
package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	kitctx "github.com/plantx/kit/kit-go/context"
	"github.com/plantx/kit/kit-go/event"
)

// Options configures the NATS event bus.
type Options struct {
	URL         string
	StreamName  string
	DurableName string
}

// New creates a NATS JetStream-backed event bus.
// If URL is empty, a local in-memory stub bus is returned for development/tests.
func New(opts Options) (event.Bus, error) {
	if opts.URL == "" {
		return NewInMemory(), nil
	}
	nc, err := nats.Connect(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("nats jetstream: %w", err)
	}
	if opts.StreamName == "" {
		opts.StreamName = "PLANTX_EVENTS"
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     opts.StreamName,
		Subjects: []string{"plantx.>"},
	})
	if err != nil && !errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
		nc.Close()
		return nil, fmt.Errorf("nats add stream: %w", err)
	}
	return &bus{nc: nc, js: js, opts: opts}, nil
}

type bus struct {
	nc   *nats.Conn
	js   nats.JetStreamContext
	opts Options
}

func (b *bus) Publish(ctx context.Context, e event.Event) error {
	payload, err := json.Marshal(e)
	if err != nil {
		return err
	}
	trace := kitctx.GetTrace(ctx)
	user := kitctx.GetUser(ctx)
	tenant := kitctx.GetTenant(ctx)
	meta := event.Metadata{
		TraceID:  trace.TraceID,
		TenantID: tenant.ID,
	}
	if user != nil {
		meta.UserID = user.ID
		if meta.TenantID == "" {
			meta.TenantID = user.TenantID
		}
	}
	envelope := struct {
		Name       string          `json:"event_name"`
		Payload    json.RawMessage `json:"payload"`
		Metadata   event.Metadata  `json:"metadata"`
		OccurredAt int64           `json:"occurred_at"`
	}{
		Name:       e.EventName(),
		Payload:    payload,
		Metadata:   meta,
		OccurredAt: time.Now().UnixMilli(),
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	subject := "plantx." + e.EventName()
	_, err = b.js.Publish(subject, data)
	return err
}

func (b *bus) Subscribe(subject string, h event.Handler) error {
	durable := b.opts.DurableName
	if durable == "" {
		durable = "plantx-consumer"
	}
	// The bus prefixes event names with "plantx." when publishing; mirror that
	// here so subscribers receive the same subject.
	if !strings.HasPrefix(subject, "plantx.") {
		subject = "plantx." + subject
	}
	_, err := b.js.Subscribe(subject, func(msg *nats.Msg) {
		_ = b.handle(msg, h)
	}, nats.Durable(durable))
	return err
}

func (b *bus) handle(msg *nats.Msg, h event.Handler) error {
	type envelope struct {
		Payload  json.RawMessage `json:"payload"`
		Metadata event.Metadata  `json:"metadata"`
	}
	var e envelope
	if err := json.Unmarshal(msg.Data, &e); err != nil {
		return err
	}
	ctx := context.Background()
	if e.Metadata.TraceID != "" {
		ctx = kitctx.WithTrace(ctx, kitctx.TraceContext{TraceID: e.Metadata.TraceID})
	}
	if err := h(ctx, e.Payload, e.Metadata); err != nil {
		return err
	}
	return msg.Ack()
}

func (b *bus) Close() error {
	b.nc.Close()
	return nil
}

// InMemory is a local stub bus for tests and development without NATS.
type InMemory struct {
	mu        sync.RWMutex
	handlers  map[string][]event.Handler
	published []any
}

// NewInMemory creates an in-memory event bus.
func NewInMemory() *InMemory {
	return &InMemory{handlers: make(map[string][]event.Handler)}
}

// Publish records the event locally and dispatches it to matching in-memory
// subscribers asynchronously.
func (m *InMemory) Publish(ctx context.Context, e event.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	payload, _ := json.Marshal(e)
	trace := kitctx.GetTrace(ctx)
	user := kitctx.GetUser(ctx)
	tenant := kitctx.GetTenant(ctx)
	meta := event.Metadata{
		TraceID:  trace.TraceID,
		TenantID: tenant.ID,
	}
	if user != nil {
		meta.UserID = user.ID
		if meta.TenantID == "" {
			meta.TenantID = user.TenantID
		}
	}
	m.published = append(m.published, map[string]any{
		"name":    e.EventName(),
		"payload": payload,
		"meta":    meta,
	})
	for _, h := range m.handlers[e.EventName()] {
		go func() { _ = h(ctx, payload, meta) }()
	}
	return nil
}

// Subscribe registers an in-memory handler for the given subject.
func (m *InMemory) Subscribe(subject string, h event.Handler) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[subject] = append(m.handlers[subject], h)
	return nil
}

// Close is a no-op for the in-memory bus.
func (m *InMemory) Close() error { return nil }
