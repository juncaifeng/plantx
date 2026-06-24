package gateway

import (
	"fmt"
	"os"
	"regexp"
	"strings"

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
	SortOrder         int32  `yaml:"sort_order"`
	MicroAppName      string `yaml:"micro_app_name"`
	RequirePermission string `yaml:"require_permission"`
}

// Config is the top-level YAML configuration for gateway registration.
type Config struct {
	Service      ServiceConfig       `yaml:"service"`
	Application  ApplicationConfig   `yaml:"application"`
	MicroApps    []MicroAppConfig    `yaml:"micro_apps"`
	Menus        []MenuConfig        `yaml:"menus"`
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
		cfg.Menus[i].MicroAppName = expandEnv(cfg.Menus[i].MicroAppName)
		cfg.Menus[i].RequirePermission = expandEnv(cfg.Menus[i].RequirePermission)
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
			SortOrder:         m.SortOrder,
			MicroAppName:      m.MicroAppName,
			RequirePermission: m.RequirePermission,
		}))
	}

	return AutoRegister(cfg.Service.Name, opts...), nil
}
