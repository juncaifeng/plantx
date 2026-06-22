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

// Authorize asks OPA whether the action is allowed.
func (a *Authorizer) Authorize(ctx context.Context, req authz.Request) (authz.Decision, error) {
	if a.opts.URL == "" {
		// Fallback: allow if user has matching permission or role.
		allowed := hasPermission(req.User.Permissions, req.Action) || hasRole(req.User.Roles, req.Action)
		return authz.Decision{Allowed: allowed, Reason: "local RBAC fallback"}, nil
	}
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return authz.Decision{}, fmt.Errorf("opa returned %d", resp.StatusCode)
	}
	var d decisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return authz.Decision{}, fmt.Errorf("decode opa response: %w", err)
	}
	reason := "denied by policy"
	if d.Result {
		reason = "allowed by policy"
	}
	return authz.Decision{Allowed: d.Result, Reason: reason}, nil
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
