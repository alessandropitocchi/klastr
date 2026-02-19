package main

import (
	"fmt"
	"strings"

	"github.com/alepito/deploy-cluster/pkg/template"
	"github.com/alepito/deploy-cluster/pkg/plugin/argocd"
	"github.com/alepito/deploy-cluster/pkg/plugin/certmanager"
	"github.com/alepito/deploy-cluster/pkg/plugin/customapps"
	"github.com/alepito/deploy-cluster/pkg/plugin/dashboard"
	"github.com/alepito/deploy-cluster/pkg/plugin/ingress"
	"github.com/alepito/deploy-cluster/pkg/plugin/monitoring"
	"github.com/alepito/deploy-cluster/pkg/plugin/storage"
	"github.com/spf13/cobra"
)

var (
	createTemplateFile string
	createEnvFile    string
	createFailFast   bool
)

// pluginResult tracks the outcome of a plugin installation.
type pluginResult struct {
	Name string
	Err  error
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cluster from template",
	Long:  `Create a new Kubernetes cluster based on the provided template file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger("")

		// Load .env file
		if err := template.LoadEnvFile(createEnvFile); err != nil {
			return fmt.Errorf("failed to load env file: %w", err)
		}

		// Load config
		log.Info("Loading template from %s...\n", createTemplateFile)
		cfg, err := template.Load(createTemplateFile)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}

		log.Info("\n")
		log.Info("Cluster: %s\n", cfg.Name)
		log.Info("Provider: %s\n", cfg.Provider.Type)
		log.Info("Control planes: %d\n", cfg.Cluster.ControlPlanes)
		log.Info("Workers: %d\n", cfg.Cluster.Workers)
		if cfg.Cluster.Version != "" {
			log.Info("Kubernetes version: %s\n", cfg.Cluster.Version)
		}
		log.Info("\n")

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
		kubecontext := provider.KubeContext(cfg.Name)

		// Install plugins with result tracking
		var results []pluginResult

		if cfg.Plugins.Storage != nil && cfg.Plugins.Storage.Enabled {
			log.Info("\n")
			pluginLog := newLogger("[storage]")
			storagePlugin := storage.New(pluginLog, globalTimeout)
			err := storagePlugin.Install(cfg.Plugins.Storage, kubecontext)
			results = append(results, pluginResult{Name: "storage", Err: err})
			if err != nil && createFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to install, see summary above")
			}
		}

		if cfg.Plugins.Ingress != nil && cfg.Plugins.Ingress.Enabled {
			log.Info("\n")
			pluginLog := newLogger("[ingress]")
			ingressPlugin := ingress.New(pluginLog, globalTimeout)
			err := ingressPlugin.Install(cfg.Plugins.Ingress, kubecontext)
			results = append(results, pluginResult{Name: "ingress", Err: err})
			if err != nil && createFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to install, see summary above")
			}
		}

		if cfg.Plugins.CertManager != nil && cfg.Plugins.CertManager.Enabled {
			log.Info("\n")
			pluginLog := newLogger("[cert-manager]")
			cmPlugin := certmanager.New(pluginLog, globalTimeout)
			err := cmPlugin.Install(cfg.Plugins.CertManager, kubecontext)
			results = append(results, pluginResult{Name: "cert-manager", Err: err})
			if err != nil && createFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to install, see summary above")
			}
		}

		if cfg.Plugins.Monitoring != nil && cfg.Plugins.Monitoring.Enabled {
			log.Info("\n")
			pluginLog := newLogger("[monitoring]")
			monPlugin := monitoring.New(pluginLog, globalTimeout)
			err := monPlugin.Install(cfg.Plugins.Monitoring, kubecontext)
			results = append(results, pluginResult{Name: "monitoring", Err: err})
			if err != nil && createFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to install, see summary above")
			}
		}

		if cfg.Plugins.Dashboard != nil && cfg.Plugins.Dashboard.Enabled {
			log.Info("\n")
			pluginLog := newLogger("[dashboard]")
			dashPlugin := dashboard.New(pluginLog, globalTimeout)
			err := dashPlugin.Install(cfg.Plugins.Dashboard, kubecontext)
			results = append(results, pluginResult{Name: "dashboard", Err: err})
			if err != nil && createFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to install, see summary above")
			}
		}

		if len(cfg.Plugins.CustomApps) > 0 {
			log.Info("\n")
			pluginLog := newLogger("[customApps]")
			customPlugin := customapps.New(pluginLog, globalTimeout)
			err := customPlugin.InstallAll(cfg.Plugins.CustomApps, kubecontext)
			results = append(results, pluginResult{Name: "customApps", Err: err})
			if err != nil && createFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to install, see summary above")
			}
		}

		if cfg.Plugins.ArgoCD != nil && cfg.Plugins.ArgoCD.Enabled {
			log.Info("\n")
			pluginLog := newLogger("[argocd]")
			argoPlugin := argocd.New(pluginLog, globalTimeout)
			err := argoPlugin.Install(cfg.Plugins.ArgoCD, kubecontext)
			results = append(results, pluginResult{Name: "argocd", Err: err})
			if err != nil && createFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to install, see summary above")
			}
		}

		// Print summary and final info
		if len(results) > 0 {
			printSummary(results, log)
		}

		log.Info("\nTo use the cluster:\n")
		log.Info("  kubectl cluster-info --context %s\n", kubecontext)

		if hasErrors(results) {
			return fmt.Errorf("some plugins failed to install, see summary above")
		}

		return nil
	},
}

func printSummary(results []pluginResult, log interface{ Info(string, ...any) }) {
	fmt.Println()
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println("Plugin Installation Summary:")
	fmt.Println(strings.Repeat("-", 40))
	for _, r := range results {
		if r.Err == nil {
			fmt.Printf("  ✓ %s\n", r.Name)
		} else {
			fmt.Printf("  ✗ %s: %v\n", r.Name, r.Err)
		}
	}
	fmt.Println(strings.Repeat("-", 40))
}

func hasErrors(results []pluginResult) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}

func init() {
	createCmd.Flags().StringVarP(&createTemplateFile, "template", "t", "template.yaml", "cluster template file")
	createCmd.Flags().StringVarP(&createEnvFile, "env", "e", ".env", "environment file for secrets")
	createCmd.Flags().BoolVar(&createFailFast, "fail-fast", false, "stop at first plugin failure")
	rootCmd.AddCommand(createCmd)
}
