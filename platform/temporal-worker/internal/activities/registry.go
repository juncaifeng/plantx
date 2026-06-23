package activities

import "context"

// UpdateMenuStatus is a placeholder activity that updates the status of a menu.
func UpdateMenuStatus(ctx context.Context, menuID string, status string) error {
	// TODO: call registry-service gRPC to update menu status.
	_ = ctx
	_ = menuID
	_ = status
	return nil
}

// UpdateMicroAppStatus is a placeholder activity that updates the status of a micro-app.
func UpdateMicroAppStatus(ctx context.Context, microAppID string, status string) error {
	// TODO: call registry-service gRPC to update micro-app status.
	_ = ctx
	_ = microAppID
	_ = status
	return nil
}

// AuditLog is a placeholder activity that writes an audit log entry.
func AuditLog(ctx context.Context, resource string, action string) error {
	// TODO: call audit-service gRPC to write audit log.
	_ = ctx
	_ = resource
	_ = action
	return nil
}
