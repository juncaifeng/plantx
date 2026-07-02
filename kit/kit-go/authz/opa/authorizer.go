// Package opa provides an OPA-based policy authorizer for Kit services.
package opa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/plantx/kit/kit-go/authz"
)

// Options configures the OPA authorizer.
type Options struct {
	URL    string
	Policy string
	// DecisionPath is the OPA decision path, e.g. "/v1/data/plantx/authz/allow".
	DecisionPath string
	// IAMURL is the base URL of the IAM service HTTP gateway. When OPA is not
	// configured, the authorizer can evaluate ABAC policies via IAM.
	IAMURL string
}

// Authorizer requests policy decisions from OPA.
type Authorizer struct {
	opts   Options
	client *http.Client
}

// New creates an OPA authorizer.
func New(opts Options) *Authorizer {
	if opts.DecisionPath == "" {
		opts.DecisionPath = "/v1/data/plantx/authz/allow"
	}
	opts.URL = strings.TrimRight(opts.URL, "/")
	opts.IAMURL = strings.TrimRight(opts.IAMURL, "/")
	if !strings.HasPrefix(opts.DecisionPath, "/") {
		opts.DecisionPath = "/" + opts.DecisionPath
	}
	return &Authorizer{
		opts:   opts,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

type decisionRequest struct {
	Input map[string]any `json:"input"`
}

type decisionResponse struct {
	Result bool `json:"result"`
}

type abacEvaluateRequest struct {
	Permission         string            `json:"permission"`
	UserAttributes     map[string]string `json:"user_attributes"`
	ResourceAttributes map[string]string `json:"resource_attributes"`
}

type abacEvaluateResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

// Authorize asks OPA whether the action is allowed.
func (a *Authorizer) Authorize(ctx context.Context, req authz.Request) (authz.Decision, error) {
	var rbacAllowed bool
	var rbacReason string
	if a.opts.URL == "" {
		rbacAllowed = hasPermission(req.User.Permissions, req.Action) || hasRole(req.User.Roles, req.Action)
		rbacReason = "local RBAC fallback"
	} else {
		d, err := a.evaluateOPA(ctx, req)
		if err != nil {
			return authz.Decision{}, err
		}
		rbacAllowed = d.Allowed
		rbacReason = d.Reason
	}
	if rbacAllowed {
		return authz.Decision{Allowed: true, Reason: rbacReason}, nil
	}
	// RBAC denied; try ABAC via IAM if configured.
	if a.opts.IAMURL != "" {
		abacDecision, err := a.evaluateABAC(ctx, req)
		if err != nil {
			return authz.Decision{}, err
		}
		if abacDecision.Allowed {
			return abacDecision, nil
		}
	}
	return authz.Decision{Allowed: false, Reason: rbacReason}, nil
}

func (a *Authorizer) evaluateOPA(ctx context.Context, req authz.Request) (authz.Decision, error) {
	input := map[string]any{
		"user": map[string]any{
			"id":          req.User.ID,
			"tenant_id":   req.User.TenantID,
			"roles":       req.User.Roles,
			"permissions": req.User.Permissions,
		},
		"action": map[string]any{
			"service":   req.Action.Service,
			"resource":  req.Action.Resource,
			"operation": req.Action.Operation,
		},
		"resource": req.ResourceContext,
	}
	body, err := json.Marshal(decisionRequest{Input: input})
	if err != nil {
		return authz.Decision{}, fmt.Errorf("marshal decision input: %w", err)
	}
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.opts.URL+a.opts.DecisionPath, bytes.NewReader(body))
	if err != nil {
		return authz.Decision{}, err
	}
	hreq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(hreq)
	if err != nil {
		return authz.Decision{}, fmt.Errorf("opa request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return authz.Decision{}, fmt.Errorf("opa returned %d", resp.StatusCode)
	}
	var d decisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return authz.Decision{}, fmt.Errorf("decode opa response: %w", err)
	}
	reason := "denied by OPA policy"
	if d.Result {
		reason = "allowed by OPA policy"
	}
	return authz.Decision{Allowed: d.Result, Reason: reason}, nil
}

func (a *Authorizer) evaluateABAC(ctx context.Context, req authz.Request) (authz.Decision, error) {
	permission := fmt.Sprintf("%s:%s", req.Action.Resource, req.Action.Operation)
	body, err := json.Marshal(abacEvaluateRequest{
		Permission:         permission,
		UserAttributes:     req.User.Claims,
		ResourceAttributes: toStringMap(req.ResourceContext),
	})
	if err != nil {
		return authz.Decision{}, fmt.Errorf("marshal abac request: %w", err)
	}
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.opts.IAMURL+"/api/iam/v1/policies/evaluate", bytes.NewReader(body))
	if err != nil {
		return authz.Decision{}, err
	}
	hreq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(hreq)
	if err != nil {
		return authz.Decision{}, fmt.Errorf("abac request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return authz.Decision{}, fmt.Errorf("abac evaluate returned %d", resp.StatusCode)
	}
	var d abacEvaluateResponse
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return authz.Decision{}, fmt.Errorf("decode abac response: %w", err)
	}
	return authz.Decision{Allowed: d.Allowed, Reason: d.Reason}, nil
}

func toStringMap(m map[string]any) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}

func hasPermission(perms []string, action authz.Action) bool {
	want := fmt.Sprintf("%s:%s", action.Resource, action.Operation)
	for _, p := range perms {
		if p == want || p == "*:*" {
			return true
		}
	}
	return false
}

func hasRole(roles []string, _ authz.Action) bool {
	// Simple RBAC: admin role grants everything.
	for _, r := range roles {
		if r == "admin" || r == "platform_admin" {
			return true
		}
	}
	return false
}
