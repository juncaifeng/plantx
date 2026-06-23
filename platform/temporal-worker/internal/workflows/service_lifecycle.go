package workflows

import "go.temporal.io/sdk/workflow"

// ServiceLifecycleWorkflow orchestrates lifecycle changes for a registered service.
// It is intentionally a placeholder in the scaffold and will be expanded to drive
// menu/micro-app/permission status transitions through activities.
func ServiceLifecycleWorkflow(ctx workflow.Context, serviceName string, event string) error {
	// TODO: implement lifecycle orchestration (Task 9).
	_ = ctx
	_ = serviceName
	_ = event
	return nil
}
