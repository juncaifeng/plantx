package app

import (
	"context"
	"fmt"
	"time"

	"github.com/plantx/kit/kit-go/auth"
	kitctx "github.com/plantx/kit/kit-go/context"
	"github.com/plantx/demo_app/backend/internal/domain"
)

// DemoService implements the application use-cases.
type DemoService struct {
	repo domain.Repository
}

// NewDemoService creates a new DemoService.
func NewDemoService(repo domain.Repository) *DemoService {
	return &DemoService{repo: repo}
}

// CreateItem creates an item in the current tenant.
func (s *DemoService) CreateItem(ctx context.Context, title string) (*domain.Item, error) {
	tenant := kitctx.GetTenant(ctx)
	return s.repo.CreateItem(ctx, tenant.ID, title)
}

// ListItems lists items in the current tenant.
func (s *DemoService) ListItems(ctx context.Context) ([]*domain.Item, error) {
	tenant := kitctx.GetTenant(ctx)
	return s.repo.ListItems(ctx, tenant.ID)
}

// CreateSetting creates a setting with RBAC + ABAC enforcement.
// RBAC: the caller already has "setting:create" (enforced by kit-go authz).
// ABAC: scope-level rules based on user attributes.
func (s *DemoService) CreateSetting(ctx context.Context, key, value string, scope domain.SettingScope) (*domain.Setting, error) {
	user := kitctx.GetUser(ctx)
	tenant := kitctx.GetTenant(ctx)

	if scope == domain.SettingScopeGlobal {
		// ABAC: only platform admins may create global settings.
		if !isPlatformAdmin(user) {
			return nil, fmt.Errorf("global settings require platform admin role")
		}
	}

	setting := &domain.Setting{
		ID:        newSettingID(),
		TenantID:  tenant.ID,
		Key:       key,
		Value:     value,
		Scope:     scope,
		UpdatedBy: user.ID,
		UpdatedAt: time.Now(),
	}
	if scope == domain.SettingScopeGlobal {
		setting.TenantID = ""
	}
	return s.repo.CreateSetting(ctx, setting)
}

// GetSetting returns a setting if visible to the current tenant.
func (s *DemoService) GetSetting(ctx context.Context, id string) (*domain.Setting, error) {
	setting, err := s.repo.GetSetting(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canAccessSetting(ctx, setting) {
		return nil, fmt.Errorf("setting not accessible")
	}
	return setting, nil
}

// ListSettings returns global settings plus the current tenant's settings.
func (s *DemoService) ListSettings(ctx context.Context) ([]*domain.Setting, error) {
	tenant := kitctx.GetTenant(ctx)
	return s.repo.ListSettings(ctx, tenant.ID)
}

// UpdateSetting updates a setting with ABAC enforcement.
func (s *DemoService) UpdateSetting(ctx context.Context, id, value string) (*domain.Setting, error) {
	user := kitctx.GetUser(ctx)
	setting, err := s.GetSetting(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canModifySetting(ctx, setting) {
		return nil, fmt.Errorf("insufficient scope to update setting")
	}
	setting.Value = value
	setting.UpdatedBy = user.ID
	setting.UpdatedAt = time.Now()
	return s.repo.UpdateSetting(ctx, id, value)
}

// DeleteSetting deletes a setting with ABAC enforcement.
func (s *DemoService) DeleteSetting(ctx context.Context, id string) (*domain.Setting, error) {
	setting, err := s.GetSetting(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canModifySetting(ctx, setting) {
		return nil, fmt.Errorf("insufficient scope to delete setting")
	}
	return s.repo.DeleteSetting(ctx, id)
}

func isPlatformAdmin(user *auth.UserInfo) bool {
	if user == nil {
		return false
	}
	for _, r := range user.Roles {
		if r == "admin" || r == "platform_admin" {
			return true
		}
	}
	return false
}

func canAccessSetting(ctx context.Context, setting *domain.Setting) bool {
	if setting.Scope == domain.SettingScopeGlobal {
		return true
	}
	tenant := kitctx.GetTenant(ctx)
	return setting.TenantID == tenant.ID
}

func canModifySetting(ctx context.Context, setting *domain.Setting) bool {
	user := kitctx.GetUser(ctx)
	if setting.Scope == domain.SettingScopeGlobal {
		return isPlatformAdmin(user)
	}
	tenant := kitctx.GetTenant(ctx)
	return setting.TenantID == tenant.ID
}

func newSettingID() string {
	return fmt.Sprintf("st-%d", time.Now().UnixNano())
}
