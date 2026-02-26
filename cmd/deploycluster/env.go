package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/alessandropitocchi/deploy-cluster/pkg/env"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	envBasePath string
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage multi-environment configurations",
	Long: `Manage multi-environment configurations for klastr.

This command helps you organize configurations for different environments
(dev, staging, production) using an overlay system similar to Kustomize.

Directory structure:
  my-cluster/
  ├── klastr.yaml              # Base configuration
  ├── plugins/
  └── environments/
      ├── dev/
      │   ├── overlay.yaml     # Environment-specific patches
      │   └── config.yaml      # Additional config (optional)
      ├── staging/
      │   └── overlay.yaml
      └── production/
          └── overlay.yaml

Example overlay.yaml:
  name: production
  description: Production environment
  base: ../../                  # Relative path to base
  patches:
    - target: cluster.workers
      value: 5
    - target: plugins.monitoring.enabled
      value: true
  values:
    DOMAIN: prod.example.com
    REPLICAS: "3"
`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available environments",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := env.NewManager(envBasePath)
		envs, err := manager.List()
		if err != nil {
			return fmt.Errorf("failed to list environments: %w", err)
		}

		if len(envs) == 0 {
			fmt.Println("No environments found.")
			fmt.Printf("Create one with: klastr env create <name>\n")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ENVIRONMENT\tDESCRIPTION")
		fmt.Fprintln(w, "-----------\t-----------")

		for _, name := range envs {
			// Try to load overlay for description
			overlayFile := envBasePath + "/environments/" + name + "/overlay.yaml"
			if data, err := os.ReadFile(overlayFile); err == nil {
				// Simple parsing to get description
				var desc struct {
					Description string `yaml:"description"`
				}
				// Ignore errors, just show name
				_ = yaml.Unmarshal(data, &desc)
				fmt.Fprintf(w, "%s\t%s\n", name, desc.Description)
			} else {
				fmt.Fprintf(w, "%s\t\n", name)
			}
		}
		w.Flush()
		return nil
	},
}

var envCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new environment",
	Args:  cobra.ExactArgs(1),
	Example: `  # Create dev environment
  klastr env create dev

  # Create production with custom base
  klastr env create production --base ./base`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		base, _ := cmd.Flags().GetString("base")

		manager := env.NewManager(envBasePath)
		if err := manager.Create(name, base); err != nil {
			return fmt.Errorf("failed to create environment: %w", err)
		}

		fmt.Printf("✓ Created environment %q\n", name)
		fmt.Printf("  Location: %s/environments/%s/\n", envBasePath, name)
		fmt.Printf("  Edit %s/environments/%s/overlay.yaml to customize\n", envBasePath, name)
		return nil
	},
}

var envShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show environment configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		manager := env.NewManager(envBasePath)
		cfg, err := manager.Load(name)
		if err != nil {
			return fmt.Errorf("failed to load environment: %w", err)
		}

		fmt.Printf("Environment: %s\n", name)
		fmt.Printf("Cluster: %s\n", cfg.Name)
		fmt.Printf("Provider: %s\n", cfg.Provider.Type)
		fmt.Printf("Control Planes: %d\n", cfg.Cluster.ControlPlanes)
		fmt.Printf("Workers: %d\n", cfg.Cluster.Workers)
		fmt.Printf("Version: %s\n", cfg.Cluster.Version)

		return nil
	},
}

func init() {
	envCmd.PersistentFlags().StringVar(&envBasePath, "base-path", ".", "Base path for environments")

	envCreateCmd.Flags().String("base", "../../", "Base configuration path")

	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envCreateCmd)
	envCmd.AddCommand(envShowCmd)
	rootCmd.AddCommand(envCmd)
}

