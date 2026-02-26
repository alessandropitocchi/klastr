// Package env provides multi-environment support for klastr configurations.
// It implements an overlay system similar to Kustomize for managing
// environment-specific configurations (dev, staging, production).
package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alessandropitocchi/deploy-cluster/pkg/template"
	"gopkg.in/yaml.v3"
)

// Overlay represents an environment-specific configuration overlay
type Overlay struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Base        string            `yaml:"base,omitempty"` // Path to base config (default: ../)
	Patches     []Patch           `yaml:"patches,omitempty"`
	Values      map[string]string `yaml:"values,omitempty"` // Key-value overrides
}

// Patch represents a strategic merge patch for a specific path
type Patch struct {
	Target string      `yaml:"target"` // e.g., "plugins.monitoring.enabled"
	Value  interface{} `yaml:"value"`
}

// Manager handles environment operations
type Manager struct {
	basePath string
}

// NewManager creates a new environment manager
func NewManager(basePath string) *Manager {
	return &Manager{basePath: basePath}
}

// Load loads a configuration for a specific environment
func (m *Manager) Load(envName string) (*template.Template, error) {
	envPath := filepath.Join(m.basePath, "environments", envName)
	
	// Check if environment exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("environment %q not found at %s", envName, envPath)
	}

	// Load overlay configuration
	overlayFile := filepath.Join(envPath, "overlay.yaml")
	overlay := &Overlay{Name: envName}
	
	if data, err := os.ReadFile(overlayFile); err == nil {
		if err := yaml.Unmarshal(data, overlay); err != nil {
			return nil, fmt.Errorf("failed to parse overlay.yaml: %w", err)
		}
	}

	// Determine base path
	basePath := m.basePath
	if overlay.Base != "" {
		if filepath.IsAbs(overlay.Base) {
			basePath = overlay.Base
		} else {
			basePath = filepath.Join(envPath, overlay.Base)
		}
	}

	// Load base configuration
	loader := template.NewLoader()
	baseConfig, err := loader.Load(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	// Apply environment-specific patches
	if err := m.applyPatches(baseConfig, overlay.Patches); err != nil {
		return nil, fmt.Errorf("failed to apply patches: %w", err)
	}

	// Apply value overrides (templating)
	if err := m.applyValues(baseConfig, overlay.Values); err != nil {
		return nil, fmt.Errorf("failed to apply values: %w", err)
	}

	// Load additional environment-specific files
	envConfigPath := filepath.Join(envPath, "config.yaml")
	if _, err := os.Stat(envConfigPath); err == nil {
		envConfig, err := loader.LoadFromFile(envConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load env config: %w", err)
		}
		// Merge environment config with base
		loader.MergeTemplates(baseConfig, envConfig)
	}

	return baseConfig, nil
}

// applyPatches applies strategic merge patches to the template
func (m *Manager) applyPatches(cfg *template.Template, patches []Patch) error {
	for _, patch := range patches {
		if err := m.applyPatch(cfg, patch); err != nil {
			return fmt.Errorf("patch %q: %w", patch.Target, err)
		}
	}
	return nil
}

// applyPatch applies a single patch using path notation (e.g., "cluster.workers")
func (m *Manager) applyPatch(cfg *template.Template, patch Patch) error {
	parts := strings.Split(patch.Target, ".")
	if len(parts) == 0 {
		return fmt.Errorf("empty target path")
	}

	switch parts[0] {
	case "name":
		if s, ok := patch.Value.(string); ok {
			cfg.Name = s
		}
	case "cluster":
		if len(parts) < 2 {
			return fmt.Errorf("invalid cluster path")
		}
		return m.patchCluster(&cfg.Cluster, parts[1:], patch.Value)
	case "provider":
		if len(parts) < 2 {
			return fmt.Errorf("invalid provider path")
		}
		return m.patchProvider(&cfg.Provider, parts[1:], patch.Value)
	case "plugins":
		if len(parts) < 2 {
			return fmt.Errorf("invalid plugins path")
		}
		return m.patchPlugins(&cfg.Plugins, parts[1:], patch.Value)
	case "snapshot":
		if cfg.Snapshot == nil {
			cfg.Snapshot = &template.SnapshotConfig{}
		}
		return m.patchSnapshot(cfg.Snapshot, parts[1:], patch.Value)
	default:
		return fmt.Errorf("unknown target: %s", parts[0])
	}
	return nil
}

func (m *Manager) patchCluster(c *template.ClusterTemplate, path []string, value interface{}) error {
	if len(path) == 0 {
		return fmt.Errorf("empty cluster path")
	}

	switch path[0] {
	case "controlPlanes":
		if v, ok := value.(int); ok {
			c.ControlPlanes = v
		}
	case "workers":
		if v, ok := value.(int); ok {
			c.Workers = v
		}
	case "version":
		if v, ok := value.(string); ok {
			c.Version = v
		}
	default:
		return fmt.Errorf("unknown cluster field: %s", path[0])
	}
	return nil
}

func (m *Manager) patchProvider(p *template.ProviderTemplate, path []string, value interface{}) error {
	if len(path) == 0 {
		return fmt.Errorf("empty provider path")
	}

	switch path[0] {
	case "type":
		if v, ok := value.(string); ok {
			p.Type = v
		}
	case "kubeconfig":
		if v, ok := value.(string); ok {
			p.Kubeconfig = v
		}
	case "context":
		if v, ok := value.(string); ok {
			p.Context = v
		}
	default:
		return fmt.Errorf("unknown provider field: %s", path[0])
	}
	return nil
}

func (m *Manager) patchPlugins(p *template.PluginsTemplate, path []string, value interface{}) error {
	if len(path) == 0 {
		return fmt.Errorf("empty plugins path")
	}

	// Handle plugin enablement: plugins.monitoring.enabled = true
	pluginName := path[0]
	
	switch pluginName {
	case "storage":
		if p.Storage == nil {
			p.Storage = &template.StorageTemplate{}
		}
		if len(path) > 1 && path[1] == "enabled" {
			if v, ok := value.(bool); ok {
				p.Storage.Enabled = v
			}
		}
	case "ingress":
		if p.Ingress == nil {
			p.Ingress = &template.IngressTemplate{}
		}
		if len(path) > 1 && path[1] == "enabled" {
			if v, ok := value.(bool); ok {
				p.Ingress.Enabled = v
			}
		}
	case "monitoring":
		if p.Monitoring == nil {
			p.Monitoring = &template.MonitoringTemplate{}
		}
		if len(path) > 1 && path[1] == "enabled" {
			if v, ok := value.(bool); ok {
				p.Monitoring.Enabled = v
			}
		}
	case "dashboard":
		if p.Dashboard == nil {
			p.Dashboard = &template.DashboardTemplate{}
		}
		if len(path) > 1 && path[1] == "enabled" {
			if v, ok := value.(bool); ok {
				p.Dashboard.Enabled = v
			}
		}
	case "certManager":
		if p.CertManager == nil {
			p.CertManager = &template.CertManagerTemplate{}
		}
		if len(path) > 1 && path[1] == "enabled" {
			if v, ok := value.(bool); ok {
				p.CertManager.Enabled = v
			}
		}
	// Add more plugins as needed
	default:
		return fmt.Errorf("patch for plugin %s not implemented", pluginName)
	}
	return nil
}

func (m *Manager) patchSnapshot(s *template.SnapshotConfig, path []string, value interface{}) error {
	if len(path) == 0 {
		return fmt.Errorf("empty snapshot path")
	}

	switch path[0] {
	case "enabled":
		if v, ok := value.(bool); ok {
			s.Enabled = v
		}
	case "bucket":
		if v, ok := value.(string); ok {
			s.Bucket = v
		}
	case "prefix":
		if v, ok := value.(string); ok {
			s.Prefix = v
		}
	case "region":
		if v, ok := value.(string); ok {
			s.Region = v
		}
	default:
		return fmt.Errorf("unknown snapshot field: %s", path[0])
	}
	return nil
}

// applyValues applies value overrides using simple templating
func (m *Manager) applyValues(cfg *template.Template, values map[string]string) error {
	// Values are applied during template processing, not here
	// This is handled by the templating system
	return nil
}

// List returns all available environments
func (m *Manager) List() ([]string, error) {
	envDir := filepath.Join(m.basePath, "environments")
	
	entries, err := os.ReadDir(envDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var envs []string
	for _, entry := range entries {
		if entry.IsDir() {
			envs = append(envs, entry.Name())
		}
	}
	return envs, nil
}

// Create creates a new environment structure
func (m *Manager) Create(name string, base string) error {
	envDir := filepath.Join(m.basePath, "environments", name)
	
	if err := os.MkdirAll(envDir, 0755); err != nil {
		return fmt.Errorf("failed to create env directory: %w", err)
	}

	overlay := &Overlay{
		Name:        name,
		Description: fmt.Sprintf("%s environment", name),
		Base:        base,
		Patches:     []Patch{},
		Values:      make(map[string]string),
	}

	data, err := yaml.Marshal(overlay)
	if err != nil {
		return fmt.Errorf("failed to marshal overlay: %w", err)
	}

	overlayFile := filepath.Join(envDir, "overlay.yaml")
	if err := os.WriteFile(overlayFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write overlay.yaml: %w", err)
	}

	return nil
}
