package grpc

import (
	"context"

	demoapi "github.com/plantx/demo_app/backend/api"
	"github.com/plantx/demo_app/backend/internal/app"
	"github.com/plantx/demo_app/backend/internal/domain"
)

// Handler implements the DemoService gRPC server.
type Handler struct {
	demoapi.UnimplementedDemoServiceServer
	app *app.DemoService
}

// NewHandler creates a new Handler.
func NewHandler(application *app.DemoService) *Handler {
	return &Handler{app: application}
}

// CreateItem creates a new item.
func (h *Handler) CreateItem(ctx context.Context, req *demoapi.CreateItemRequest) (*demoapi.Item, error) {
	item, err := h.app.CreateItem(ctx, req.GetTitle())
	if err != nil {
		return nil, err
	}
	return itemToPB(item), nil
}

// ListItems lists items for the current tenant.
func (h *Handler) ListItems(ctx context.Context, _ *demoapi.ListItemsRequest) (*demoapi.ItemList, error) {
	items, err := h.app.ListItems(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*demoapi.Item, 0, len(items))
	for _, item := range items {
		out = append(out, itemToPB(item))
	}
	return &demoapi.ItemList{Items: out}, nil
}

// CreateSetting creates a new setting.
func (h *Handler) CreateSetting(ctx context.Context, req *demoapi.CreateSettingRequest) (*demoapi.Setting, error) {
	setting, err := h.app.CreateSetting(ctx, req.GetKey(), req.GetValue(), domainSettingScope(req.GetScope()))
	if err != nil {
		return nil, err
	}
	return settingToPB(setting), nil
}

// GetSetting returns a setting by id.
func (h *Handler) GetSetting(ctx context.Context, req *demoapi.GetSettingRequest) (*demoapi.Setting, error) {
	setting, err := h.app.GetSetting(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return settingToPB(setting), nil
}

// ListSettings lists settings visible to the current tenant.
func (h *Handler) ListSettings(ctx context.Context, _ *demoapi.ListSettingsRequest) (*demoapi.SettingList, error) {
	settings, err := h.app.ListSettings(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*demoapi.Setting, 0, len(settings))
	for _, s := range settings {
		out = append(out, settingToPB(s))
	}
	return &demoapi.SettingList{Settings: out}, nil
}

// UpdateSetting updates a setting value.
func (h *Handler) UpdateSetting(ctx context.Context, req *demoapi.UpdateSettingRequest) (*demoapi.Setting, error) {
	setting, err := h.app.UpdateSetting(ctx, req.GetId(), req.GetValue())
	if err != nil {
		return nil, err
	}
	return settingToPB(setting), nil
}

// DeleteSetting deletes a setting.
func (h *Handler) DeleteSetting(ctx context.Context, req *demoapi.DeleteSettingRequest) (*demoapi.Setting, error) {
	setting, err := h.app.DeleteSetting(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return settingToPB(setting), nil
}

func itemToPB(item *domain.Item) *demoapi.Item {
	return &demoapi.Item{
		Id:        item.ID,
		TenantId:  item.TenantID,
		Title:     item.Title,
		CreatedAt: item.CreatedAt.Unix(),
	}
}

func settingToPB(s *domain.Setting) *demoapi.Setting {
	return &demoapi.Setting{
		Id:        s.ID,
		TenantId:  s.TenantID,
		Key:       s.Key,
		Value:     s.Value,
		Scope:     pbSettingScope(s.Scope),
		UpdatedBy: s.UpdatedBy,
		UpdatedAt: s.UpdatedAt.Unix(),
	}
}

func domainSettingScope(s demoapi.SettingScope) domain.SettingScope {
	switch s {
	case demoapi.SettingScope_SETTING_SCOPE_GLOBAL:
		return domain.SettingScopeGlobal
	case demoapi.SettingScope_SETTING_SCOPE_TENANT:
		return domain.SettingScopeTenant
	default:
		return domain.SettingScopeUnspecified
	}
}

func pbSettingScope(s domain.SettingScope) demoapi.SettingScope {
	switch s {
	case domain.SettingScopeGlobal:
		return demoapi.SettingScope_SETTING_SCOPE_GLOBAL
	case domain.SettingScopeTenant:
		return demoapi.SettingScope_SETTING_SCOPE_TENANT
	default:
		return demoapi.SettingScope_SETTING_SCOPE_UNSPECIFIED
	}
}
