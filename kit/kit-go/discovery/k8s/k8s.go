// Package k8s provides a Kubernetes DNS-based service discovery registry.
package k8s

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/plantx/kit/kit-go/discovery"
)

// Options configures the K8s DNS registry.
type Options struct {
	// Namespace is the Kubernetes namespace to resolve services in.
	Namespace string
	// Port is the default gRPC port if SRV lookup does not return one.
	Port int
}

// New creates a K8s DNS discovery registry.
func New(opts Options) *Registry {
	if opts.Namespace == "" {
		opts.Namespace = "default"
	}
	if opts.Port == 0 {
		opts.Port = 8080
	}
	return &Registry{opts: opts}
}

// Registry discovers services via Kubernetes DNS.
type Registry struct {
	opts Options
}

// Register is a no-op for the K8s DNS registry.
func (r *Registry) Register(_ context.Context, _ string, _ discovery.Instance) error {
	return nil
}

// Deregister is a no-op for the K8s DNS registry.
func (r *Registry) Deregister(_ context.Context, _ string, _ string) error {
	return nil
}

// Discover resolves service instances via Kubernetes DNS.
func (r *Registry) Discover(ctx context.Context, serviceName string) ([]discovery.Instance, error) {
	// Headless service DNS: <service>.<namespace>.svc.cluster.local
	host := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, r.opts.Namespace)
	addrs, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("lookup %s: %w", host, err)
	}
	_, srvs, err := net.DefaultResolver.LookupSRV(ctx, "", "", host)
	if err != nil {
		srvs = nil
	}
	instances := make([]discovery.Instance, 0, len(addrs))
	for i, addr := range addrs {
		port := r.opts.Port
		for _, srv := range srvs {
			if srv.Target == addr || srv.Target == host {
				port = int(srv.Port)
				break
			}
		}
		instances = append(instances, discovery.Instance{
			ID:      fmt.Sprintf("%s-%d", host, i),
			Address: addr,
			Port:    port,
			Meta:    map[string]string{"source": "k8s-dns"},
		})
	}
	return instances, nil
}

// Close is a no-op for the K8s DNS registry.
func (r *Registry) Close() error { return nil }

// Atoi parses a port string.
func Atoi(s string) int {
	if s == "" {
		return 0
	}
	v, _ := strconv.Atoi(s)
	return v
}
