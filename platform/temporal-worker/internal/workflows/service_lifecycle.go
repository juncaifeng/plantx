package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// LifecycleEvent represents a lifecycle change for a service.
type LifecycleEvent string

const (
	// ServiceRegistered is emitted when a service is registered.
	ServiceRegistered LifecycleEvent = "REGISTERED"
	// ServiceDeregistered is emitted when a service is deregistered.
	ServiceDeregistered LifecycleEvent = "DEREGISTERED"
	// ServiceUnhealthy is emitted when a service becomes unhealthy.
	ServiceUnhealthy LifecycleEvent = "UNHEALTHY"
	// ServiceHealthy is emitted when a service recovers.
	ServiceHealthy LifecycleEvent = "HEALTHY"
)

// ServiceLifecycleInput is the workflow input for ServiceLifecycleWorkflow.
// It is intentionally JSON-tagged so callers (e.g. registry-service) can pass
// a matching anonymous struct without importing this package.
type ServiceLifecycleInput struct {
	ServiceName   string         `json:"serviceName"`
	Event         LifecycleEvent `json:"event"`
	MicroAppNames []string       `json:"microAppNames"`
}

// ServiceLifecycleWorkflow orchestrates lifecycle changes for a registered service.
func ServiceLifecycleWorkflow(ctx workflow.Context, input ServiceLifecycleInput) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	switch input.Event {
	case ServiceRegistered, ServiceHealthy:
		if err := workflow.ExecuteActivity(ctx, "SetServiceStatus", input.ServiceName, "ONLINE").Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "PublishServiceMenus", input.ServiceName, input.MicroAppNames).Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "PublishServiceMicroApps", input.ServiceName, input.MicroAppNames).Get(ctx, nil); err != nil {
			return err
		}
	case ServiceUnhealthy:
		if err := workflow.ExecuteActivity(ctx, "SetServiceStatus", input.ServiceName, "OFFLINE").Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "UnpublishServiceMenus", input.ServiceName, input.MicroAppNames).Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "UnpublishServiceMicroApps", input.ServiceName, input.MicroAppNames).Get(ctx, nil); err != nil {
			return err
		}
	case ServiceDeregistered:
		// The service row is deleted by the registry-service immediately after
		// triggering this workflow, so we only cascade the offline status to
		// related menus and micro-apps. We intentionally skip SetServiceStatus.
		if err := workflow.ExecuteActivity(ctx, "UnpublishServiceMenus", input.ServiceName, input.MicroAppNames).Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "UnpublishServiceMicroApps", input.ServiceName, input.MicroAppNames).Get(ctx, nil); err != nil {
			return err
		}
	}

	return workflow.ExecuteActivity(ctx, "WriteAuditLog", input.ServiceName, string(input.Event)).Get(ctx, nil)
}
