package domain

import (
	"context"
	"time"
)

// Item is a simple tenant-scoped entity.
type Item struct {
	ID        string
	TenantID  string
	Title     string
	CreatedAt time.Time
}

// SettingScope controls whether a setting is global or tenant-scoped.
type SettingScope int32

const (
	SettingScopeUnspecified SettingScope = 0
	SettingScopeGlobal      SettingScope = 1
	SettingScopeTenant      SettingScope = 2
)

// Setting is a configuration entry supporting RBAC + ABAC demos.
type Setting struct {
	ID        string
	TenantID  string
	Key       string
	Value     string
	Scope     SettingScope
	UpdatedBy string
	UpdatedAt time.Time
}

// Repository defines storage for demo entities.
type Repository interface {
	CreateItem(ctx context.Context, tenantID, title string) (*Item, error)
	ListItems(ctx context.Context, tenantID string) ([]*Item, error)

	CreateSetting(ctx context.Context, setting *Setting) (*Setting, error)
	GetSetting(ctx context.Context, id string) (*Setting, error)
	ListSettings(ctx context.Context, tenantID string) ([]*Setting, error)
	UpdateSetting(ctx context.Context, id, value string) (*Setting, error)
	DeleteSetting(ctx context.Context, id string) (*Setting, error)
}
