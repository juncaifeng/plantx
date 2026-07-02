package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Permission describes an RBAC permission exposed by a service.
type Permission struct {
	Name        string
	Resource    string
	Operation   string
	Description string
}

// iamPermission is the JSON shape returned by iam-service.
type iamPermission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Resource    string `json:"resource"`
	Operation   string `json:"operation"`
	Description string `json:"description"`
}

type iamPermissionsResponse struct {
	Permissions []iamPermission `json:"permissions"`
}

type iamCreatePermissionRequest struct {
	Name        string `json:"name"`
	Resource    string `json:"resource"`
	Operation   string `json:"operation"`
	Description string `json:"description"`
}

// IAMClient is a tiny HTTP client for the iam-service permission API.
type IAMClient struct {
	baseURL string
	client  *http.Client
}

// NewIAMClient creates an IAM client. addr may be a host:port or a full URL.
func NewIAMClient(addr string) *IAMClient {
	base := addr
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + base
	}
	return &IAMClient{
		baseURL: strings.TrimRight(base, "/"),
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// ListPermissions returns the existing permission catalog.
func (c *IAMClient) ListPermissions(ctx context.Context) (map[string]iamPermission, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/iam/v1/permissions", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list permissions returned %d: %s", resp.StatusCode, string(body))
	}

	var data iamPermissionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode permissions: %w", err)
	}

	out := make(map[string]iamPermission, len(data.Permissions))
	for _, p := range data.Permissions {
		out[p.Name] = p
	}
	return out, nil
}

// CreatePermission registers a permission in the catalog.
func (c *IAMClient) CreatePermission(ctx context.Context, p Permission) error {
	body, err := json.Marshal(iamCreatePermissionRequest(p))
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/iam/v1/permissions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("create permission %s: %w", p.Name, err)
	}
	defer resp.Body.Close()

	// 409 means the permission already exists; treat as success.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create permission %s returned %d: %s", p.Name, resp.StatusCode, string(body))
	}
	return nil
}

// SyncPermissions ensures all declared permissions exist in IAM.
func (c *IAMClient) SyncPermissions(ctx context.Context, permissions []Permission) error {
	existing, err := c.ListPermissions(ctx)
	if err != nil {
		return err
	}

	for _, p := range permissions {
		if _, ok := existing[p.Name]; ok {
			continue
		}
		if err := c.CreatePermission(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

// ValidatePermissions returns an error if any required permission is missing
// from the declared list and cannot be found in IAM.
func (c *IAMClient) ValidatePermissions(ctx context.Context, declared []Permission, required []string) error {
	declaredSet := make(map[string]struct{}, len(declared))
	for _, p := range declared {
		declaredSet[p.Name] = struct{}{}
	}

	existing, err := c.ListPermissions(ctx)
	if err != nil {
		return fmt.Errorf("cannot validate permissions: %w", err)
	}

	var missing []string
	for _, name := range required {
		if name == "" {
			continue
		}
		if _, ok := declaredSet[name]; ok {
			continue
		}
		if _, ok := existing[name]; ok {
			continue
		}
		missing = append(missing, name)
	}

	if len(missing) > 0 {
		return fmt.Errorf("required permissions not declared or registered: %s", strings.Join(missing, ", "))
	}
	return nil
}
