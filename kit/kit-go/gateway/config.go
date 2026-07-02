package gateway

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/plantx/kit/kit-go/server"
	registryapi "github.com/plantx/platform/registry-service/api"
	"gopkg.in/yaml.v3"
)

// ServiceConfig describes the service itself.
type ServiceConfig struct {
	Name         string `yaml:"name"`
	GRPCHost     string `yaml:"grpc_host"`
	RESTPrefix   string `yaml:"rest_prefix"`
	RegistryAddr string `yaml:"registry_addr"`
	IAMAddr      string `yaml:"iam_addr"`
}

// ApplicationConfig describes the product application.
type ApplicationConfig struct {
	Key       string `yaml:"key"`
	Name      string `yaml:"name"`
	LabelKey  string `yaml:"label_key"`
	Icon      string `yaml:"icon"`
	SortOrder int32  `yaml:"sort_order"`
	Status    string `yaml:"status"`
}

// MicroAppConfig describes a qiankun micro-frontend manifest.
type MicroAppConfig struct {
	Name              string `yaml:"name"`
	Route             string `yaml:"route"`
	BundleURL         string `yaml:"bundle_url"`
	MenuLabelKey      string `yaml:"menu_label_key"`
	RequirePermission string `yaml:"require_permission"`
	Upstream          string `yaml:"upstream"`
}

// MenuConfig describes a portal menu entry.
type MenuConfig struct {
	LabelKey          string `yaml:"label_key"`
	Route             string `yaml:"route"`
	Icon              string `yaml:"icon"`
	ParentID          string `yaml:"parent_id"`
	ParentLabelKey    string `yaml:"parent_label_key"`
	SortOrder         int32  `yaml:"sort_order"`
	MicroAppName      string `yaml:"micro_app_name"`
	RequirePermission string `yaml:"require_permission"`
}

// PermissionConfig describes an RBAC permission exposed by the service.
type PermissionConfig struct {
	Name        string `yaml:"name"`
	Resource    string `yaml:"resource"`
	Operation   string `yaml:"operation"`
	Description string `yaml:"description"`
}

// Config is the top-level YAML configuration for gateway registration.
type Config struct {
	Service      ServiceConfig       `yaml:"service"`
	Application  ApplicationConfig   `yaml:"application"`
	MicroApps    []MicroAppConfig    `yaml:"micro_apps"`
	Menus        []MenuConfig        `yaml:"menus"`
	Permissions  []PermissionConfig  `yaml:"permissions"`
}

var envVarPattern = regexp.MustCompile(`\$\{([^}:]+)(?::-([^}]*))?\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

// expandEnv replaces ${VAR}, ${VAR:-default}, $VAR with environment variable values.
// If the variable is not set and no default is provided, the placeholder is left as-is.
func expandEnv(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		groups := envVarPattern.FindStringSubmatch(match)
		if groups == nil {
			return match
		}
		key := groups[1]
		defaultValue := groups[2]
		if key == "" {
			key = groups[3]
		}
		if v, ok := os.LookupEnv(key); ok {
			return v
		}
		if defaultValue != "" {
			return defaultValue
		}
		return match
	})
}

// expandConfig recursively expands environment variables in the config.
func expandConfig(cfg *Config) {
	cfg.Service.Name = expandEnv(cfg.Service.Name)
	cfg.Service.GRPCHost = expandEnv(cfg.Service.GRPCHost)
	cfg.Service.RESTPrefix = expandEnv(cfg.Service.RESTPrefix)
	cfg.Service.RegistryAddr = expandEnv(cfg.Service.RegistryAddr)
	cfg.Service.IAMAddr = expandEnv(cfg.Service.IAMAddr)

	cfg.Application.Key = expandEnv(cfg.Application.Key)
	cfg.Application.Name = expandEnv(cfg.Application.Name)
	cfg.Application.LabelKey = expandEnv(cfg.Application.LabelKey)
	cfg.Application.Icon = expandEnv(cfg.Application.Icon)
	cfg.Application.Status = expandEnv(cfg.Application.Status)

	for i := range cfg.MicroApps {
		cfg.MicroApps[i].Name = expandEnv(cfg.MicroApps[i].Name)
		cfg.MicroApps[i].Route = expandEnv(cfg.MicroApps[i].Route)
		cfg.MicroApps[i].BundleURL = expandEnv(cfg.MicroApps[i].BundleURL)
		cfg.MicroApps[i].MenuLabelKey = expandEnv(cfg.MicroApps[i].MenuLabelKey)
		cfg.MicroApps[i].RequirePermission = expandEnv(cfg.MicroApps[i].RequirePermission)
		cfg.MicroApps[i].Upstream = expandEnv(cfg.MicroApps[i].Upstream)
	}

	for i := range cfg.Menus {
		cfg.Menus[i].LabelKey = expandEnv(cfg.Menus[i].LabelKey)
		cfg.Menus[i].Route = expandEnv(cfg.Menus[i].Route)
		cfg.Menus[i].Icon = expandEnv(cfg.Menus[i].Icon)
		cfg.Menus[i].ParentID = expandEnv(cfg.Menus[i].ParentID)
		cfg.Menus[i].ParentLabelKey = expandEnv(cfg.Menus[i].ParentLabelKey)
		cfg.Menus[i].MicroAppName = expandEnv(cfg.Menus[i].MicroAppName)
		cfg.Menus[i].RequirePermission = expandEnv(cfg.Menus[i].RequirePermission)
	}

	for i := range cfg.Permissions {
		cfg.Permissions[i].Name = expandEnv(cfg.Permissions[i].Name)
		cfg.Permissions[i].Resource = expandEnv(cfg.Permissions[i].Resource)
		cfg.Permissions[i].Operation = expandEnv(cfg.Permissions[i].Operation)
		cfg.Permissions[i].Description = expandEnv(cfg.Permissions[i].Description)
	}
}

// LoadConfig reads a YAML file and returns a Config with environment variables expanded.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}

	expandConfig(&cfg)
	return &cfg, nil
}

// toApplicationStatus maps a string to registry application status.
func toApplicationStatus(s string) registryapi.ApplicationStatus {
	if s == "" {
		return registryapi.ApplicationStatus_APPLICATION_STATUS_ACTIVE
	}
	switch strings.ToUpper(s) {
	case "ACTIVE":
		return registryapi.ApplicationStatus_APPLICATION_STATUS_ACTIVE
	case "OFFLINE":
		return registryapi.ApplicationStatus_APPLICATION_STATUS_OFFLINE
	default:
		return registryapi.ApplicationStatus_APPLICATION_STATUS_UNSPECIFIED
	}
}

// AutoRegisterFromConfig builds a GatewayRegistrar from a YAML config file.
// It is a convenience wrapper around LoadConfig + AutoRegister for services
// that prefer declarative registration over code-based options.
// It also syncs declared RBAC permissions to iam-service when an iam_addr is
// configured.
func AutoRegisterFromConfig(path string) (server.GatewayRegistrar, error) {
	cfg, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}

	opts := make([]Option, 0)

	if cfg.Service.RegistryAddr != "" {
		opts = append(opts, WithRegistryAddr(cfg.Service.RegistryAddr))
	}
	if cfg.Service.GRPCHost != "" {
		opts = append(opts, WithGRPCHost(cfg.Service.GRPCHost))
	}
	if cfg.Service.RESTPrefix != "" {
		opts = append(opts, WithRESTPrefix(cfg.Service.RESTPrefix))
	}

	if cfg.Application.Key != "" {
		opts = append(opts, WithApplication(Application{
			Key:       cfg.Application.Key,
			Name:      cfg.Application.Name,
			LabelKey:  cfg.Application.LabelKey,
			Icon:      cfg.Application.Icon,
			SortOrder: cfg.Application.SortOrder,
			Status:    toApplicationStatus(cfg.Application.Status),
		}))
	}

	for _, m := range cfg.MicroApps {
		opts = append(opts, WithMicroApp(MicroApp{
			Name:              m.Name,
			Route:             m.Route,
			BundleURL:         m.BundleURL,
			MenuLabelKey:      m.MenuLabelKey,
			RequirePermission: m.RequirePermission,
			Upstream:          m.Upstream,
		}))
	}

	for _, m := range cfg.Menus {
		opts = append(opts, WithMenu(Menu{
			LabelKey:          m.LabelKey,
			Route:             m.Route,
			Icon:              m.Icon,
			ParentID:          m.ParentID,
			ParentLabelKey:    m.ParentLabelKey,
			SortOrder:         m.SortOrder,
			MicroAppName:      m.MicroAppName,
			RequirePermission: m.RequirePermission,
		}))
	}

	registrar := AutoRegister(cfg.Service.Name, opts...)

	// Sync permissions to IAM when iam_addr is configured.
	if len(cfg.Permissions) > 0 {
		iamAddr := cfg.Service.IAMAddr
		if iamAddr == "" {
			iamAddr = defaultIAMAddr()
		}
		if iamAddr != "" {
			client := NewIAMClient(iamAddr)
			permissions := make([]Permission, len(cfg.Permissions))
			for i, p := range cfg.Permissions {
				permissions[i] = Permission{
					Name:        p.Name,
					Resource:    p.Resource,
					Operation:   p.Operation,
					Description: p.Description,
				}
			}
			required := collectRequiredPermissions(cfg)
			if err := syncAndValidatePermissions(client, permissions, required); err != nil {
				return nil, err
			}
		}
	}

	return registrar, nil
}

func defaultIAMAddr() string {
	if v := getEnvAny("IAM_SERVICE_ADDR", "IAM_ADDR"); v != "" {
		return v
	}
	return "iam-service:8081"
}

func getEnvAny(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func collectRequiredPermissions(cfg *Config) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, m := range cfg.MicroApps {
		if m.RequirePermission == "" {
			continue
		}
		if _, ok := seen[m.RequirePermission]; ok {
			continue
		}
		seen[m.RequirePermission] = struct{}{}
		out = append(out, m.RequirePermission)
	}
	for _, m := range cfg.Menus {
		if m.RequirePermission == "" {
			continue
		}
		if _, ok := seen[m.RequirePermission]; ok {
			continue
		}
		seen[m.RequirePermission] = struct{}{}
		out = append(out, m.RequirePermission)
	}
	return out
}

func syncAndValidatePermissions(client *IAMClient, declared []Permission, required []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.SyncPermissions(ctx, declared); err != nil {
		return fmt.Errorf("sync permissions to IAM: %w", err)
	}
	if err := client.ValidatePermissions(ctx, declared, required); err != nil {
		return fmt.Errorf("validate permissions: %w", err)
	}
	return nil
}
