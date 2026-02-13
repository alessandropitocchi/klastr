package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name     string          `yaml:"name"`
	Provider ProviderConfig  `yaml:"provider"`
	Cluster  ClusterConfig   `yaml:"cluster"`
	Plugins  PluginsConfig   `yaml:"plugins,omitempty"`
}

type ProviderConfig struct {
	Type string `yaml:"type"` // kind, k3d, etc.
}

type ClusterConfig struct {
	ControlPlanes int    `yaml:"controlPlanes"`
	Workers       int    `yaml:"workers"`
	Version       string `yaml:"version,omitempty"` // Kubernetes version
}

type PluginsConfig struct {
	ArgoCD  *ArgoCDConfig  `yaml:"argocd,omitempty"`
	Storage *StorageConfig `yaml:"storage,omitempty"`
	Ingress *IngressConfig `yaml:"ingress,omitempty"`
}

type ArgoCDConfig struct {
	Enabled   bool              `yaml:"enabled"`
	Namespace string            `yaml:"namespace,omitempty"` // ArgoCD installation namespace
	Version   string            `yaml:"version,omitempty"`   // ArgoCD version
	Repos     []ArgoCDRepoConfig `yaml:"repos,omitempty"`    // Repositories to add
}

type ArgoCDRepoConfig struct {
	Name        string `yaml:"name,omitempty"`        // Repository name (optional)
	URL         string `yaml:"url"`                   // Repository URL
	Type        string `yaml:"type,omitempty"`        // git or helm (default: git)
	Insecure    *bool  `yaml:"insecure,omitempty"`    // Skip TLS verification (auto-detected for non-HTTPS)
	Username    string `yaml:"username,omitempty"`    // For private repos (HTTPS)
	Password    string `yaml:"password,omitempty"`    // For private repos (HTTPS)
	SSHKeyEnv   string `yaml:"sshKeyEnv,omitempty"`   // Env var containing SSH private key
	SSHKeyFile  string `yaml:"sshKeyFile,omitempty"`  // Path to SSH private key file
}

type StorageConfig struct {
	Enabled bool   `yaml:"enabled"`
	Type    string `yaml:"type"` // local-path, openebs
}

type IngressConfig struct {
	Enabled bool   `yaml:"enabled"`
	Type    string `yaml:"type"` // nginx, traefik
}

// DefaultConfig returns a starter configuration
func DefaultConfig() *Config {
	return &Config{
		Name: "my-cluster",
		Provider: ProviderConfig{
			Type: "kind",
		},
		Cluster: ClusterConfig{
			ControlPlanes: 1,
			Workers:       2,
			Version:       "v1.31.0",
		},
		Plugins: PluginsConfig{
			ArgoCD: &ArgoCDConfig{
				Enabled:   false,
				Namespace: "argocd",
				Version:   "stable",
				Repos: []ArgoCDRepoConfig{
					{
						Name: "my-gitops-repo",
						URL:  "https://github.com/user/gitops-repo.git",
						Type: "git",
					},
				},
			},
		},
	}
}

// Save writes the config to a YAML file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads a config from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
