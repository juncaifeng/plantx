package activities

import (
	"context"
	"fmt"
	"log"

	registryapi "github.com/plantx/platform/registry-service/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RegistryActivities implements activities that interact with the registry-service gRPC API.
type RegistryActivities struct {
	RegistryAddr string
}

func (a *RegistryActivities) client(ctx context.Context) (registryapi.RegistryServiceClient, func(), error) {
	conn, err := grpc.DialContext(ctx, a.RegistryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("dial registry %q: %w", a.RegistryAddr, err)
	}
	cleanup := func() { conn.Close() }
	return registryapi.NewRegistryServiceClient(conn), cleanup, nil
}

// serviceMicroAppNames returns the names of the micro-apps registered for a service.
func (a *RegistryActivities) serviceMicroAppNames(ctx context.Context, serviceName string) (map[string]struct{}, error) {
	c, cleanup, err := a.client(ctx)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	resp, err := c.ListServices(ctx, &registryapi.ListServicesRequest{})
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}

	names := make(map[string]struct{})
	for _, svc := range resp.GetServices() {
		if svc.GetName() != serviceName {
			continue
		}
		for _, m := range svc.GetMicroApps() {
			names[m.GetName()] = struct{}{}
		}
		break
	}
	return names, nil
}

// PublishServiceMenus sets all menus belonging to a service to ONLINE.
func (a *RegistryActivities) PublishServiceMenus(ctx context.Context, serviceName string) error {
	return a.updateServiceMenus(ctx, serviceName, registryapi.ResourceStatus_RESOURCE_STATUS_ONLINE)
}

// UnpublishServiceMenus sets all menus belonging to a service to OFFLINE.
func (a *RegistryActivities) UnpublishServiceMenus(ctx context.Context, serviceName string) error {
	return a.updateServiceMenus(ctx, serviceName, registryapi.ResourceStatus_RESOURCE_STATUS_OFFLINE)
}

func (a *RegistryActivities) updateServiceMenus(ctx context.Context, serviceName string, status registryapi.ResourceStatus) error {
	names, err := a.serviceMicroAppNames(ctx, serviceName)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}

	c, cleanup, err := a.client(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	listResp, err := c.ListMenus(ctx, &registryapi.ListMenusRequest{})
	if err != nil {
		return fmt.Errorf("list menus: %w", err)
	}

	for _, menu := range listResp.GetMenus() {
		if _, ok := names[menu.GetMicroAppName()]; !ok {
			continue
		}
		if menu.GetStatus() == status {
			continue
		}
		_, err := c.UpdateMenu(ctx, &registryapi.UpdateMenuRequest{
			Id:                menu.GetId(),
			LabelKey:          menu.GetLabelKey(),
			Route:             menu.GetRoute(),
			Icon:              menu.GetIcon(),
			ParentId:          menu.GetParentId(),
			SortOrder:         menu.GetSortOrder(),
			MicroAppName:      menu.GetMicroAppName(),
			RequirePermission: menu.GetRequirePermission(),
			ApplicationId:     menu.GetApplicationId(),
			ApplicationKey:    menu.GetApplicationKey(),
			Status:            status,
		})
		if err != nil {
			return fmt.Errorf("update menu %s to %s: %w", menu.GetId(), status, err)
		}
	}
	return nil
}

// PublishServiceMicroApps sets all micro-apps belonging to a service to ONLINE.
func (a *RegistryActivities) PublishServiceMicroApps(ctx context.Context, serviceName string) error {
	return a.updateServiceMicroApps(ctx, serviceName, registryapi.ResourceStatus_RESOURCE_STATUS_ONLINE)
}

// UnpublishServiceMicroApps sets all micro-apps belonging to a service to OFFLINE.
func (a *RegistryActivities) UnpublishServiceMicroApps(ctx context.Context, serviceName string) error {
	return a.updateServiceMicroApps(ctx, serviceName, registryapi.ResourceStatus_RESOURCE_STATUS_OFFLINE)
}

func (a *RegistryActivities) updateServiceMicroApps(ctx context.Context, serviceName string, status registryapi.ResourceStatus) error {
	c, cleanup, err := a.client(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	resp, err := c.ListServices(ctx, &registryapi.ListServicesRequest{})
	if err != nil {
		return fmt.Errorf("list services: %w", err)
	}

	var target *registryapi.Service
	for _, svc := range resp.GetServices() {
		if svc.GetName() == serviceName {
			target = svc
			break
		}
	}
	if target == nil {
		return nil
	}

	for _, m := range target.GetMicroApps() {
		if m.GetStatus() == status {
			continue
		}
		_, err := c.UpdateMicroApp(ctx, &registryapi.UpdateMicroAppRequest{
			Name:              m.GetName(),
			Route:             m.GetRoute(),
			BundleUrl:         m.GetBundleUrl(),
			MenuLabelKey:      m.GetMenuLabelKey(),
			RequirePermission: m.GetRequirePermission(),
			Upstream:          m.GetUpstream(),
			Status:            status,
		})
		if err != nil {
			return fmt.Errorf("update micro-app %s to %s: %w", m.GetName(), status, err)
		}
	}
	return nil
}

// SetServiceStatus updates the lifecycle status of a service by name.
func (a *RegistryActivities) SetServiceStatus(ctx context.Context, serviceName string, status string) error {
	c, cleanup, err := a.client(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	s := registryapi.ResourceStatus_RESOURCE_STATUS_UNSPECIFIED
	switch status {
	case "ONLINE":
		s = registryapi.ResourceStatus_RESOURCE_STATUS_ONLINE
	case "OFFLINE":
		s = registryapi.ResourceStatus_RESOURCE_STATUS_OFFLINE
	case "PENDING":
		s = registryapi.ResourceStatus_RESOURCE_STATUS_PENDING
	case "DRAFT":
		s = registryapi.ResourceStatus_RESOURCE_STATUS_DRAFT
	case "UPDATING":
		s = registryapi.ResourceStatus_RESOURCE_STATUS_UPDATING
	}

	_, err = c.UpdateServiceStatus(ctx, &registryapi.UpdateServiceStatusRequest{
		Name:   serviceName,
		Status: s,
	})
	if err != nil {
		return fmt.Errorf("update service %s status to %s: %w", serviceName, status, err)
	}
	return nil
}

// WriteAuditLog writes an audit log entry for a service lifecycle event.
// Currently a local stdout log; can be replaced with audit-service gRPC later.
func (a *RegistryActivities) WriteAuditLog(ctx context.Context, serviceName string, event string) error {
	log.Printf("[audit] service=%s event=%s", serviceName, event)
	return nil
}
