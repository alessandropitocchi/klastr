package main

import (
	"fmt"

	"github.com/alepito/deploy-cluster/pkg/config"
	"github.com/alepito/deploy-cluster/pkg/plugin/argocd"
	"github.com/alepito/deploy-cluster/pkg/provider/kind"
	"github.com/spf13/cobra"
)

var (
	createConfigFile string
	createEnvFile    string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cluster from configuration",
	Long:  `Create a new Kubernetes cluster based on the provided configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load .env file
		if err := config.LoadEnvFile(createEnvFile); err != nil {
			return fmt.Errorf("failed to load env file: %w", err)
		}

		// Load config
		fmt.Printf("Loading configuration from %s...\n", createConfigFile)
		cfg, err := config.Load(createConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		fmt.Printf("\n")
		fmt.Printf("Cluster: %s\n", cfg.Name)
		fmt.Printf("Provider: %s\n", cfg.Provider.Type)
		fmt.Printf("Control planes: %d\n", cfg.Cluster.ControlPlanes)
		fmt.Printf("Workers: %d\n", cfg.Cluster.Workers)
		if cfg.Cluster.Version != "" {
			fmt.Printf("Kubernetes version: %s\n", cfg.Cluster.Version)
		}
		fmt.Printf("\n")

		// Get provider
		provider, err := getProvider(cfg.Provider.Type)
		if err != nil {
			return err
		}

		// Create cluster
		if err := provider.Create(cfg); err != nil {
			return err
		}

		// Determine kubecontext based on provider
		kubecontext := fmt.Sprintf("kind-%s", cfg.Name)

		// Install plugins
		if cfg.Plugins.ArgoCD != nil && cfg.Plugins.ArgoCD.Enabled {
			fmt.Println()
			argoPlugin := argocd.New()
			if err := argoPlugin.Install(cfg.Plugins.ArgoCD, kubecontext); err != nil {
				return fmt.Errorf("failed to install ArgoCD: %w", err)
			}
		}

		fmt.Printf("\nTo use the cluster:\n")
		fmt.Printf("  kubectl cluster-info --context %s\n", kubecontext)

		return nil
	},
}

func getProvider(providerType string) (interface {
	Create(*config.Config) error
}, error) {
	switch providerType {
	case "kind":
		return kind.New(), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerType)
	}
}

func init() {
	createCmd.Flags().StringVarP(&createConfigFile, "config", "c", "cluster.yaml", "cluster configuration file")
	createCmd.Flags().StringVarP(&createEnvFile, "env", "e", ".env", "environment file for secrets")
	rootCmd.AddCommand(createCmd)
}
