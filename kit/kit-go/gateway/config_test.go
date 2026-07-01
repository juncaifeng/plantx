package gateway

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "service.yaml")
	content := `
service:
  name: demo-service
  grpc_host: ${DEMO_GRPC_HOST:-demo-service:8080}
  rest_prefix: /api/demo/v1
  registry_addr: registry-service:8080

application:
  key: demo
  name: Demo
  label_key: nav.demo
  sort_order: 100

micro_apps:
  - name: demo-ui
    route: /demo
    bundle_url: /apps/demo-ui/demo-ui.js
    upstream: demo-ui:80

menus:
  - label_key: nav.demo.home
    route: /demo
    icon: HomeOutlined
    sort_order: 10
    micro_app_name: demo-ui
    require_permission: item:list

permissions:
  - name: item:list
    resource: item
    operation: list
    description: List demo items

  - name: setting:admin
    resource: setting
    operation: admin
    description: Admin settings
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	os.Setenv("DEMO_GRPC_HOST", "192.168.1.11:8080")
	defer os.Unsetenv("DEMO_GRPC_HOST")

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Service.Name != "demo-service" {
		t.Errorf("service name = %q, want demo-service", cfg.Service.Name)
	}
	if cfg.Service.GRPCHost != "192.168.1.11:8080" {
		t.Errorf("grpc host = %q, want 192.168.1.11:8080", cfg.Service.GRPCHost)
	}
	if cfg.Application.Key != "demo" {
		t.Errorf("application key = %q, want demo", cfg.Application.Key)
	}
	if len(cfg.MicroApps) != 1 || cfg.MicroApps[0].Name != "demo-ui" {
		t.Errorf("micro apps = %+v, want 1 demo-ui", cfg.MicroApps)
	}
	if len(cfg.Menus) != 1 || cfg.Menus[0].LabelKey != "nav.demo.home" {
		t.Errorf("menus = %+v, want 1 nav.demo.home", cfg.Menus)
	}
	if len(cfg.Permissions) != 2 {
		t.Errorf("permissions = %+v, want 2", cfg.Permissions)
	}
	if cfg.Permissions[0].Name != "item:list" {
		t.Errorf("permission[0].name = %q, want item:list", cfg.Permissions[0].Name)
	}
}
