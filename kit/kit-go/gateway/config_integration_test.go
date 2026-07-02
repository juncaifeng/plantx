package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/plantx/platform/registry-service/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestAutoRegisterFromConfigSyncsPermissions(t *testing.T) {
	registryAddr, cleanupRegistry := startTestRegistry(t)
	defer cleanupRegistry()

	var createdPermissions []Permission
	iamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/iam/v1/permissions" && r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(iamPermissionsResponse{Permissions: nil})
			return
		}
		if r.URL.Path == "/api/iam/v1/permissions" && r.Method == http.MethodPost {
			var req iamCreatePermissionRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			createdPermissions = append(createdPermissions, Permission(req))
			w.WriteHeader(http.StatusCreated)
			return
		}
		http.NotFound(w, r)
	}))
	defer iamServer.Close()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "service.yaml")
	content := `
service:
  name: demo-service
  grpc_host: demo-service:8080
  registry_addr: ` + registryAddr + `
  iam_addr: ` + iamServer.URL + `

application:
  key: demo
  name: Demo
  label_key: nav.demo
  sort_order: 100

permissions:
  - name: item:list
    resource: item
    operation: list

  - name: setting:admin
    resource: setting
    operation: admin

menus:
  - label_key: nav.demo.home
    route: /demo
    micro_app_name: demo-ui
    require_permission: item:list
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	registrar, err := AutoRegisterFromConfig(configPath)
	if err != nil {
		t.Fatalf("auto register from config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := registrar.Register(ctx); err != nil {
		t.Fatalf("register: %v", err)
	}
	defer func() { _ = registrar.Deregister(ctx) }()

	// Verify registry service was registered.
	conn, err := grpc.NewClient(registryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial registry: %v", err)
	}
	defer func() { _ = conn.Close() }()
	registryClient := api.NewRegistryServiceClient(conn)

	svcs, err := registryClient.ListServices(ctx, &api.ListServicesRequest{})
	if err != nil {
		t.Fatalf("list services: %v", err)
	}
	found := false
	for _, s := range svcs.GetServices() {
		if s.GetName() == "demo-service" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("demo-service not registered")
	}

	// Verify permissions were synced to IAM.
	if len(createdPermissions) != 2 {
		t.Fatalf("created permissions = %+v, want 2", createdPermissions)
	}
	names := make(map[string]bool)
	for _, p := range createdPermissions {
		names[p.Name] = true
	}
	if !names["item:list"] || !names["setting:admin"] {
		t.Fatalf("unexpected permission names: %+v", createdPermissions)
	}
}
