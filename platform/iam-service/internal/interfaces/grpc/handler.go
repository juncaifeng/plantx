package grpc

import (
	"context"
	"fmt"
	"strings"

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

// ListAttributes handles listing ABAC attributes.
func (h *Handler) ListAttributes(ctx context.Context, req *api.ListAttributesRequest) (*api.ListAttributesResponse, error) {
	attrs, err := h.app.ListAttributes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.Attribute, 0, len(attrs))
	for _, a := range attrs {
		out = append(out, toProtoAttribute(a))
	}
	return &api.ListAttributesResponse{Attributes: out}, nil
}

// CreateAttribute handles creating an ABAC attribute.
func (h *Handler) CreateAttribute(ctx context.Context, req *api.CreateAttributeRequest) (*api.Attribute, error) {
	a, err := h.app.CreateAttribute(ctx, req.GetKey(), req.GetValueType(), req.GetDescription())
	if err != nil {
		return nil, err
	}
	return toProtoAttribute(a), nil
}

// UpdateAttribute handles updating an ABAC attribute.
func (h *Handler) UpdateAttribute(ctx context.Context, req *api.UpdateAttributeRequest) (*api.Attribute, error) {
	a, err := h.app.UpdateAttribute(ctx, req.GetId(), req.GetKey(), req.GetValueType(), req.GetDescription())
	if err != nil {
		return nil, err
	}
	return toProtoAttribute(a), nil
}

// DeleteAttribute handles deleting an ABAC attribute.
func (h *Handler) DeleteAttribute(ctx context.Context, req *api.DeleteAttributeRequest) (*emptypb.Empty, error) {
	if err := h.app.DeleteAttribute(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ListConditions handles listing ABAC conditions.
func (h *Handler) ListConditions(ctx context.Context, req *api.ListConditionsRequest) (*api.ListConditionsResponse, error) {
	conds, err := h.app.ListConditions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.Condition, 0, len(conds))
	for _, c := range conds {
		out = append(out, toProtoCondition(c))
	}
	return &api.ListConditionsResponse{Conditions: out}, nil
}

// CreateCondition handles creating an ABAC condition.
func (h *Handler) CreateCondition(ctx context.Context, req *api.CreateConditionRequest) (*api.Condition, error) {
	c, err := h.app.CreateCondition(ctx, req.GetName(), req.GetAttributeKey(), req.GetOperator(), req.GetValue(), req.GetDescription())
	if err != nil {
		return nil, err
	}
	return toProtoCondition(c), nil
}

// UpdateCondition handles updating an ABAC condition.
func (h *Handler) UpdateCondition(ctx context.Context, req *api.UpdateConditionRequest) (*api.Condition, error) {
	c, err := h.app.UpdateCondition(ctx, req.GetId(), req.GetName(), req.GetAttributeKey(), req.GetOperator(), req.GetValue(), req.GetDescription())
	if err != nil {
		return nil, err
	}
	return toProtoCondition(c), nil
}

// DeleteCondition handles deleting an ABAC condition.
func (h *Handler) DeleteCondition(ctx context.Context, req *api.DeleteConditionRequest) (*emptypb.Empty, error) {
	if err := h.app.DeleteCondition(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ListPolicies handles listing ABAC policies.
func (h *Handler) ListPolicies(ctx context.Context, req *api.ListPoliciesRequest) (*api.ListPoliciesResponse, error) {
	policies, err := h.app.ListPolicies(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*api.Policy, 0, len(policies))
	for _, p := range policies {
		out = append(out, toProtoPolicy(p))
	}
	return &api.ListPoliciesResponse{Policies: out}, nil
}

// CreatePolicy handles creating an ABAC policy.
func (h *Handler) CreatePolicy(ctx context.Context, req *api.CreatePolicyRequest) (*api.Policy, error) {
	p, err := h.app.CreatePolicy(ctx, req.GetName(), req.GetDescription(), req.GetEffect(), req.GetPriority(), req.GetPermissions(), req.GetConditionIds())
	if err != nil {
		return nil, err
	}
	return toProtoPolicy(p), nil
}

// UpdatePolicy handles updating an ABAC policy.
func (h *Handler) UpdatePolicy(ctx context.Context, req *api.UpdatePolicyRequest) (*api.Policy, error) {
	p, err := h.app.UpdatePolicy(ctx, req.GetId(), req.GetName(), req.GetDescription(), req.GetEffect(), req.GetPriority(), req.GetPermissions(), req.GetConditionIds())
	if err != nil {
		return nil, err
	}
	return toProtoPolicy(p), nil
}

// DeletePolicy handles deleting an ABAC policy.
func (h *Handler) DeletePolicy(ctx context.Context, req *api.DeletePolicyRequest) (*emptypb.Empty, error) {
	if err := h.app.DeletePolicy(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// EvaluatePolicy handles evaluating ABAC policies for a permission.
func (h *Handler) EvaluatePolicy(ctx context.Context, req *api.EvaluatePolicyRequest) (*api.EvaluatePolicyResponse, error) {
	policies, err := h.app.ListPolicies(ctx)
	if err != nil {
		return nil, err
	}
	conds, err := h.app.ListConditions(ctx)
	if err != nil {
		return nil, err
	}
	allowed, reason := evaluateABAC(req.GetPermission(), req.GetUserAttributes(), req.GetResourceAttributes(), policies, conds)
	return &api.EvaluatePolicyResponse{Allowed: allowed, Reason: reason}, nil
}

func evaluateABAC(permission string, userAttrs, resourceAttrs map[string]string, policies []*domain.Policy, conditions []*domain.Condition) (bool, string) {
	condByID := make(map[string]*domain.Condition)
	for _, c := range conditions {
		condByID[c.ID] = c
	}
	for _, p := range policies {
		if !contains(p.Permissions, permission) {
			continue
		}
		allMatch := true
		for _, cid := range p.ConditionIDs {
			c := condByID[cid]
			if c == nil {
				allMatch = false
				break
			}
			left := userAttrs[c.AttributeKey]
			if left == "" {
				left = resourceAttrs[c.AttributeKey]
			}
			if !matchCondition(left, c.Operator, c.Value) {
				allMatch = false
				break
			}
		}
		if allMatch {
			if p.Effect == "deny" {
				return false, fmt.Sprintf("denied by policy %q", p.Name)
			}
			return true, fmt.Sprintf("allowed by policy %q", p.Name)
		}
	}
	return false, "no matching ABAC policy"
}

func matchCondition(left, operator, right string) bool {
	switch operator {
	case "eq":
		return left == right
	case "ne":
		return left != right
	case "in":
		return contains(splitValues(right), left)
	case "not_in":
		return !contains(splitValues(right), left)
	default:
		return false
	}
}

func splitValues(s string) []string {
	parts := make([]string, 0)
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func contains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
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

func toProtoAttribute(a *domain.Attribute) *api.Attribute {
	return &api.Attribute{
		Id:          a.ID,
		Key:         a.Key,
		ValueType:   a.ValueType,
		Description: a.Description,
	}
}

func toProtoCondition(c *domain.Condition) *api.Condition {
	return &api.Condition{
		Id:           c.ID,
		Name:         c.Name,
		AttributeKey: c.AttributeKey,
		Operator:     c.Operator,
		Value:        c.Value,
		Description:  c.Description,
	}
}

func toProtoPolicy(p *domain.Policy) *api.Policy {
	return &api.Policy{
		Id:           p.ID,
		Name:         p.Name,
		Description:  p.Description,
		Permissions:  p.Permissions,
		ConditionIds: p.ConditionIDs,
		Effect:       p.Effect,
		Priority:     p.Priority,
	}
}
