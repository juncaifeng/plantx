package discovery

import "context"

// Instance represents a discovered service endpoint.
type Instance struct {
	ID      string
	Address string
	Port    int
	Meta    map[string]string
}

// Registry abstracts service discovery.
type Registry interface {
	Register(ctx context.Context, serviceName string, inst Instance) error
	Deregister(ctx context.Context, serviceName string, instID string) error
	Discover(ctx context.Context, serviceName string) ([]Instance, error)
	Close() error
}
