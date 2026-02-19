package main

import (
	"fmt"

	"github.com/alepito/deploy-cluster/pkg/plugin/argocd"
	"github.com/alepito/deploy-cluster/pkg/plugin/certmanager"
	"github.com/alepito/deploy-cluster/pkg/plugin/customapps"
	"github.com/alepito/deploy-cluster/pkg/plugin/dashboard"
	"github.com/alepito/deploy-cluster/pkg/plugin/ingress"
	"github.com/alepito/deploy-cluster/pkg/plugin/monitoring"
	"github.com/alepito/deploy-cluster/pkg/plugin/storage"
	"github.com/alepito/deploy-cluster/pkg/template"
	"github.com/spf13/cobra"
)

var (
	upgradeTemplateFile string
	upgradeEnvFile      string
	upgradeDryRun       bool
	upgradeFailFast     bool
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade an existing cluster applying template changes",
	Long: `Upgrade an existing Kubernetes cluster by applying only the differences
compared to the current state. The cluster is not recreated, but plugins
(ArgoCD repos/apps) are updated: additions, modifications and removals.

Use --dry-run to preview changes without applying them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger("")

		// Load .env file
		if err := template.LoadEnvFile(upgradeEnvFile); err != nil {
			return fmt.Errorf("failed to load env file: %w", err)
		}

		// Load config
		log.Info("Loading template from %s...\n", upgradeTemplateFile)
		cfg, err := template.Load(upgradeTemplateFile)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}

		// Get provider and check cluster exists
		provider, err := getProvider(cfg.Provider.Type)
		if err != nil {
			return err
		}

		exists, err := provider.Exists(cfg.Name)
		if err != nil {
			return fmt.Errorf("failed to check cluster existence: %w", err)
		}
		if !exists {
			return fmt.Errorf("cluster '%s' does not exist. Use 'create' to create it first", cfg.Name)
		}

		kubecontext := provider.KubeContext(cfg.Name)

		if upgradeDryRun {
			return runUpgradeDryRun(cfg, kubecontext)
		}

		log.Info("Upgrading cluster '%s'...\n\n", cfg.Name)

		results := upgradePlugins(cfg, kubecontext, upgradeFailFast)

		// Print summary
		if len(results) > 0 {
			printSummary(results, log)
		}

		if hasErrors(results) {
			return fmt.Errorf("some plugins failed to upgrade, see summary above")
		}

		log.Success("\nUpgrade completed for cluster '%s'.\n", cfg.Name)
		return nil
	},
}

func runUpgradeDryRun(cfg *template.Template, kubecontext string) error {
	fmt.Printf("Dry-run for cluster '%s':\n", cfg.Name)

	// Storage
	if cfg.Plugins.Storage != nil && cfg.Plugins.Storage.Enabled {
		pluginLog := newLogger("[storage]")
		storagePlugin := storage.New(pluginLog, globalTimeout)
		installed, err := storagePlugin.IsInstalled(kubecontext)
		if err != nil {
			return fmt.Errorf("failed to check storage status: %w", err)
		}
		if installed {
			fmt.Printf("\n[storage] %s: installed (re-apply)\n", cfg.Plugins.Storage.Type)
		} else {
			fmt.Printf("\n[storage] %s: not installed (will install)\n", cfg.Plugins.Storage.Type)
		}
	}

	// Ingress
	if cfg.Plugins.Ingress != nil && cfg.Plugins.Ingress.Enabled {
		pluginLog := newLogger("[ingress]")
		ingressPlugin := ingress.New(pluginLog, globalTimeout)
		installed, err := ingressPlugin.IsInstalled(kubecontext)
		if err != nil {
			return fmt.Errorf("failed to check ingress status: %w", err)
		}
		if installed {
			fmt.Printf("\n[ingress] %s: installed (re-apply)\n", cfg.Plugins.Ingress.Type)
		} else {
			fmt.Printf("\n[ingress] %s: not installed (will install)\n", cfg.Plugins.Ingress.Type)
		}
	}

	// Cert-manager
	if cfg.Plugins.CertManager != nil && cfg.Plugins.CertManager.Enabled {
		pluginLog := newLogger("[cert-manager]")
		cmPlugin := certmanager.New(pluginLog, globalTimeout)
		installed, err := cmPlugin.IsInstalled(kubecontext)
		if err != nil {
			return fmt.Errorf("failed to check cert-manager status: %w", err)
		}
		version := cfg.Plugins.CertManager.Version
		if version == "" {
			version = "v1.16.3"
		}
		if installed {
			fmt.Printf("\n[cert-manager] %s: installed (re-apply)\n", version)
		} else {
			fmt.Printf("\n[cert-manager] %s: not installed (will install)\n", version)
		}
	}

	// Monitoring
	if cfg.Plugins.Monitoring != nil && cfg.Plugins.Monitoring.Enabled {
		pluginLog := newLogger("[monitoring]")
		monPlugin := monitoring.New(pluginLog, globalTimeout)
		installed, err := monPlugin.IsInstalled(kubecontext)
		if err != nil {
			return fmt.Errorf("failed to check monitoring status: %w", err)
		}
		if installed {
			fmt.Printf("\n[monitoring] %s: installed (re-apply)\n", cfg.Plugins.Monitoring.Type)
		} else {
			fmt.Printf("\n[monitoring] %s: not installed (will install)\n", cfg.Plugins.Monitoring.Type)
		}
	}

	// Dashboard
	if cfg.Plugins.Dashboard != nil && cfg.Plugins.Dashboard.Enabled {
		pluginLog := newLogger("[dashboard]")
		dashPlugin := dashboard.New(pluginLog, globalTimeout)
		installed, err := dashPlugin.IsInstalled(kubecontext)
		if err != nil {
			return fmt.Errorf("failed to check dashboard status: %w", err)
		}
		if installed {
			fmt.Printf("\n[dashboard] %s: installed (re-apply)\n", cfg.Plugins.Dashboard.Type)
		} else {
			fmt.Printf("\n[dashboard] %s: not installed (will install)\n", cfg.Plugins.Dashboard.Type)
		}
	}

	// Custom Apps
	if len(cfg.Plugins.CustomApps) > 0 {
		pluginLog := newLogger("[customApps]")
		customPlugin := customapps.New(pluginLog, globalTimeout)
		fmt.Println("\n[customApps] Custom apps:")
		for _, app := range cfg.Plugins.CustomApps {
			ns := app.Namespace
			if ns == "" {
				ns = app.Name
			}
			installed, err := customPlugin.IsInstalled(app.Name, ns, kubecontext)
			if err != nil {
				return fmt.Errorf("failed to check custom app %s status: %w", app.Name, err)
			}
			version := app.Version
			if version == "" {
				version = "latest"
			}
			if installed {
				fmt.Printf("  ~ %s (%s@%s) (update)\n", app.Name, app.ChartName, version)
			} else {
				fmt.Printf("  + %s (%s@%s) (install)\n", app.Name, app.ChartName, version)
			}
		}
	}

	// ArgoCD
	if cfg.Plugins.ArgoCD != nil && cfg.Plugins.ArgoCD.Enabled {
		pluginLog := newLogger("[argocd]")
		argoPlugin := argocd.New(pluginLog, globalTimeout)

		namespace := cfg.Plugins.ArgoCD.Namespace
		if namespace == "" {
			namespace = "argocd"
		}

		installed, err := argoPlugin.IsInstalled(kubecontext, namespace)
		if err != nil {
			return fmt.Errorf("failed to check ArgoCD status: %w", err)
		}

		if !installed {
			fmt.Printf("\n[argocd] Not installed (will perform full install)\n")
		} else {
			if err := argoPlugin.DryRun(cfg.Plugins.ArgoCD, kubecontext); err != nil {
				return fmt.Errorf("failed to dry-run ArgoCD: %w", err)
			}
		}
	}

	fmt.Println("\nNo changes applied (dry-run).")
	return nil
}

func init() {
	upgradeCmd.Flags().StringVarP(&upgradeTemplateFile, "template", "t", "template.yaml", "cluster template file")
	upgradeCmd.Flags().StringVarP(&upgradeEnvFile, "env", "e", ".env", "environment file for secrets")
	upgradeCmd.Flags().BoolVar(&upgradeDryRun, "dry-run", false, "preview changes without applying them")
	upgradeCmd.Flags().BoolVar(&upgradeFailFast, "fail-fast", false, "stop at first plugin failure")
	rootCmd.AddCommand(upgradeCmd)
}
