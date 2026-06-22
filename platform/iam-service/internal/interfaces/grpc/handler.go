package grpc

import (
	"context"

	"github.com/plantx/platform/iam-service/api"
	"github.com/plantx/platform/iam-service/internal/app"
	"github.com/plantx/platform/iam-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Handler implements the IAMService gRPC server.
type Handler struct {
	api.UnimplementedIAMServiceServer
	app *app.IAMService
}

// NewHandler creates a new Handler.
func NewHandler(app *app.IAMService) *Handler {
	return &Handler{app: app}
}

// ListUsers handles listing users.
func (h *Handler) ListUsers(ctx context.Context, req *api.ListUsersRequest) (*api.ListUsersResponse, error) {
	users, err := h.app.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.User, 0, len(users))
	for _, u := range users {
		out = append(out, toProtoUser(u))
	}
	return &api.ListUsersResponse{Users: out}, nil
}

// CreateUser handles creating a user.
func (h *Handler) CreateUser(ctx context.Context, req *api.CreateUserRequest) (*api.User, error) {
	u, err := h.app.CreateUser(ctx, req.GetUsername(), req.GetTenantId(), req.GetRoleIds())
	if err != nil {
		return nil, err
	}
	return toProtoUser(u), nil
}

// ListRoles handles listing roles.
func (h *Handler) ListRoles(ctx context.Context, req *api.ListRolesRequest) (*api.ListRolesResponse, error) {
	roles, err := h.app.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.Role, 0, len(roles))
	for _, r := range roles {
		out = append(out, toProtoRole(r))
	}
	return &api.ListRolesResponse{Roles: out}, nil
}

// GetRole handles fetching a single role.
func (h *Handler) GetRole(ctx context.Context, req *api.GetRoleRequest) (*api.Role, error) {
	role, err := h.app.GetRole(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, status.Errorf(codes.NotFound, "role %q not found", req.GetId())
	}
	return toProtoRole(role), nil
}

// CreateRole handles role creation.
func (h *Handler) CreateRole(ctx context.Context, req *api.CreateRoleRequest) (*api.Role, error) {
	role, err := h.app.CreateRole(ctx, req.GetName(), req.GetDescription(), req.GetPermissions())
	if err != nil {
		return nil, err
	}
	return toProtoRole(role), nil
}

// UpdateRole handles role updates.
func (h *Handler) UpdateRole(ctx context.Context, req *api.UpdateRoleRequest) (*api.Role, error) {
	role, err := h.app.UpdateRole(ctx, req.GetId(), req.GetName(), req.GetDescription(), req.GetPermissions())
	if err != nil {
		return nil, err
	}
	return toProtoRole(role), nil
}

// DeleteRole handles role deletion.
func (h *Handler) DeleteRole(ctx context.Context, req *api.DeleteRoleRequest) (*emptypb.Empty, error) {
	if err := h.app.DeleteRole(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ListPermissions handles listing permissions.
func (h *Handler) ListPermissions(ctx context.Context, req *api.ListPermissionsRequest) (*api.ListPermissionsResponse, error) {
	perms, err := h.app.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.Permission, 0, len(perms))
	for _, p := range perms {
		out = append(out, toProtoPermission(p))
	}
	return &api.ListPermissionsResponse{Permissions: out}, nil
}

// CreatePermission handles permission creation.
func (h *Handler) CreatePermission(ctx context.Context, req *api.CreatePermissionRequest) (*api.Permission, error) {
	perm, err := h.app.CreatePermission(ctx, req.GetName(), req.GetResource(), req.GetOperation(), req.GetDescription())
	if err != nil {
		return nil, err
	}
	return toProtoPermission(perm), nil
}

// DeletePermission handles permission deletion.
func (h *Handler) DeletePermission(ctx context.Context, req *api.DeletePermissionRequest) (*emptypb.Empty, error) {
	if err := h.app.DeletePermission(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func toProtoUser(u *domain.User) *api.User {
	return &api.User{
		Id:       u.ID,
		Username: u.Username,
		TenantId: u.TenantID,
		RoleIds:  u.RoleIDs,
	}
}

func toProtoRole(r *domain.Role) *api.Role {
	return &api.Role{
		Id:          r.ID,
		Name:        r.Name,
		Permissions: r.Permissions,
		Description: r.Description,
	}
}

func toProtoPermission(p *domain.Permission) *api.Permission {
	return &api.Permission{
		Id:          p.ID,
		Name:        p.Name,
		Resource:    p.Resource,
		Operation:   p.Operation,
		Description: p.Description,
	}
}
