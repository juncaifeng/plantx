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

// ServiceLifecycleWorkflow orchestrates lifecycle changes for a registered service.
func ServiceLifecycleWorkflow(ctx workflow.Context, serviceName string, event LifecycleEvent) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	switch event {
	case ServiceRegistered, ServiceHealthy:
		if err := workflow.ExecuteActivity(ctx, "SetServiceStatus", serviceName, "ONLINE").Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "PublishServiceMenus", serviceName).Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "PublishServiceMicroApps", serviceName).Get(ctx, nil); err != nil {
			return err
		}
	case ServiceDeregistered, ServiceUnhealthy:
		if err := workflow.ExecuteActivity(ctx, "UnpublishServiceMenus", serviceName).Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "UnpublishServiceMicroApps", serviceName).Get(ctx, nil); err != nil {
			return err
		}
		if err := workflow.ExecuteActivity(ctx, "SetServiceStatus", serviceName, "OFFLINE").Get(ctx, nil); err != nil {
			return err
		}
	}

	return workflow.ExecuteActivity(ctx, "WriteAuditLog", serviceName, string(event)).Get(ctx, nil)
}
