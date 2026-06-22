// Package static provides a static service discovery registry.
package static

import (
	"context"
	"sync"

	"github.com/plantx/kit/kit-go/discovery"
)

// New creates a static registry seeded with the given instances.
func New(instances map[string][]discovery.Instance) *Registry {
	if instances == nil {
		instances = make(map[string][]discovery.Instance)
	}
	return &Registry{instances: instances}
}

// Registry is a memory-backed discovery registry.
type Registry struct {
	mu        sync.RWMutex
	instances map[string][]discovery.Instance
}

// Register adds an instance to the static registry.
func (r *Registry) Register(ctx context.Context, serviceName string, inst discovery.Instance) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.instances[serviceName] = append(r.instances[serviceName], inst)
	return nil
}

// Deregister removes an instance from the static registry by ID.
func (r *Registry) Deregister(ctx context.Context, serviceName string, instID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var filtered []discovery.Instance
	for _, inst := range r.instances[serviceName] {
		if inst.ID != instID {
			filtered = append(filtered, inst)
		}
	}
	r.instances[serviceName] = filtered
	return nil
}

// Discover returns all registered instances for a service.
func (r *Registry) Discover(ctx context.Context, serviceName string) ([]discovery.Instance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.instances[serviceName], nil
}

// Close releases any resources held by the registry.
func (r *Registry) Close() error { return nil }
