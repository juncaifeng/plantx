package worker

import (
	"context"
	"fmt"

	"github.com/plantx/platform/temporal-worker/internal/activities"
	"github.com/plantx/platform/temporal-worker/internal/workflows"
	"go.temporal.io/sdk/client"
	temporalworker "go.temporal.io/sdk/worker"
)

func Start(ctx context.Context, temporalHost string) error {
	c, err := client.Dial(client.Options{
		HostPort: temporalHost,
	})
	if err != nil {
		return fmt.Errorf("dial temporal: %w", err)
	}
	defer c.Close()

	w := temporalworker.New(c, "plantx-platform", temporalworker.Options{})
	w.RegisterWorkflow(workflows.ServiceLifecycleWorkflow)
	w.RegisterActivity(activities.UpdateMenuStatus)
	w.RegisterActivity(activities.UpdateMicroAppStatus)
	w.RegisterActivity(activities.AuditLog)

	if err := w.Run(temporalworker.InterruptCh()); err != nil {
		return fmt.Errorf("run worker: %w", err)
	}
	return nil
}
