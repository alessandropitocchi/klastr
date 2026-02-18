package main

import (
	"fmt"

	"github.com/alepito/deploy-cluster/pkg/config"
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
	upgradeConfigFile string
	upgradeEnvFile    string
	upgradeDryRun     bool
	upgradeFailFast   bool
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade an existing cluster applying configuration changes",
	Long: `Upgrade an existing Kubernetes cluster by applying only the differences
compared to the current state. The cluster is not recreated, but plugins
(ArgoCD repos/apps) are updated: additions, modifications and removals.

Use --dry-run to preview changes without applying them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger("")

		// Load .env file
		if err := config.LoadEnvFile(upgradeEnvFile); err != nil {
			return fmt.Errorf("failed to load env file: %w", err)
		}

		// Load config
		log.Info("Loading configuration from %s...\n", upgradeConfigFile)
		cfg, err := config.Load(upgradeConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
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

		var results []pluginResult

		// Upgrade storage plugin
		if cfg.Plugins.Storage != nil && cfg.Plugins.Storage.Enabled {
			pluginLog := newLogger("[storage]")
			storagePlugin := storage.New(pluginLog, globalTimeout)
			installed, checkErr := storagePlugin.IsInstalled(kubecontext)
			if checkErr != nil {
				results = append(results, pluginResult{Name: "storage", Err: fmt.Errorf("failed to check status: %w", checkErr)})
			} else {
				if !installed {
					pluginLog.Info("Storage not installed, performing installation...\n")
				} else {
					pluginLog.Info("Storage already installed, re-applying...\n")
				}
				err := storagePlugin.Install(cfg.Plugins.Storage, kubecontext)
				results = append(results, pluginResult{Name: "storage", Err: err})
			}
			if hasErrors(results) && upgradeFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to upgrade, see summary above")
			}
		}

		// Upgrade ingress plugin
		if cfg.Plugins.Ingress != nil && cfg.Plugins.Ingress.Enabled {
			pluginLog := newLogger("[ingress]")
			ingressPlugin := ingress.New(pluginLog, globalTimeout)
			installed, checkErr := ingressPlugin.IsInstalled(kubecontext)
			if checkErr != nil {
				results = append(results, pluginResult{Name: "ingress", Err: fmt.Errorf("failed to check status: %w", checkErr)})
			} else {
				if !installed {
					pluginLog.Info("Ingress not installed, performing installation...\n")
				} else {
					pluginLog.Info("Ingress already installed, re-applying...\n")
				}
				err := ingressPlugin.Install(cfg.Plugins.Ingress, kubecontext)
				results = append(results, pluginResult{Name: "ingress", Err: err})
			}
			if hasErrors(results) && upgradeFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to upgrade, see summary above")
			}
		}

		// Upgrade cert-manager plugin
		if cfg.Plugins.CertManager != nil && cfg.Plugins.CertManager.Enabled {
			pluginLog := newLogger("[cert-manager]")
			cmPlugin := certmanager.New(pluginLog, globalTimeout)
			installed, checkErr := cmPlugin.IsInstalled(kubecontext)
			if checkErr != nil {
				results = append(results, pluginResult{Name: "cert-manager", Err: fmt.Errorf("failed to check status: %w", checkErr)})
			} else {
				if !installed {
					pluginLog.Info("Not installed, performing installation...\n")
				} else {
					pluginLog.Info("Already installed, re-applying...\n")
				}
				err := cmPlugin.Install(cfg.Plugins.CertManager, kubecontext)
				results = append(results, pluginResult{Name: "cert-manager", Err: err})
			}
			if hasErrors(results) && upgradeFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to upgrade, see summary above")
			}
		}

		// Upgrade monitoring plugin
		if cfg.Plugins.Monitoring != nil && cfg.Plugins.Monitoring.Enabled {
			pluginLog := newLogger("[monitoring]")
			monPlugin := monitoring.New(pluginLog, globalTimeout)
			installed, checkErr := monPlugin.IsInstalled(kubecontext)
			if checkErr != nil {
				results = append(results, pluginResult{Name: "monitoring", Err: fmt.Errorf("failed to check status: %w", checkErr)})
			} else {
				if !installed {
					pluginLog.Info("Not installed, performing installation...\n")
				} else {
					pluginLog.Info("Already installed, re-applying...\n")
				}
				err := monPlugin.Install(cfg.Plugins.Monitoring, kubecontext)
				results = append(results, pluginResult{Name: "monitoring", Err: err})
			}
			if hasErrors(results) && upgradeFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to upgrade, see summary above")
			}
		}

		// Upgrade dashboard plugin
		if cfg.Plugins.Dashboard != nil && cfg.Plugins.Dashboard.Enabled {
			pluginLog := newLogger("[dashboard]")
			dashPlugin := dashboard.New(pluginLog, globalTimeout)
			installed, checkErr := dashPlugin.IsInstalled(kubecontext)
			if checkErr != nil {
				results = append(results, pluginResult{Name: "dashboard", Err: fmt.Errorf("failed to check status: %w", checkErr)})
			} else {
				if !installed {
					pluginLog.Info("Dashboard not installed, performing installation...\n")
				} else {
					pluginLog.Info("Dashboard already installed, re-applying...\n")
				}
				err := dashPlugin.Install(cfg.Plugins.Dashboard, kubecontext)
				results = append(results, pluginResult{Name: "dashboard", Err: err})
			}
			if hasErrors(results) && upgradeFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to upgrade, see summary above")
			}
		}

		// Upgrade custom apps
		if len(cfg.Plugins.CustomApps) > 0 {
			pluginLog := newLogger("[customApps]")
			pluginLog.Info("Upgrading custom apps...\n")
			customPlugin := customapps.New(pluginLog, globalTimeout)
			err := customPlugin.InstallAll(cfg.Plugins.CustomApps, kubecontext)
			results = append(results, pluginResult{Name: "customApps", Err: err})
			if err != nil && upgradeFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to upgrade, see summary above")
			}
		}

		// Upgrade ArgoCD plugin
		if cfg.Plugins.ArgoCD != nil && cfg.Plugins.ArgoCD.Enabled {
			pluginLog := newLogger("[argocd]")
			argoPlugin := argocd.New(pluginLog, globalTimeout)

			namespace := cfg.Plugins.ArgoCD.Namespace
			if namespace == "" {
				namespace = "argocd"
			}

			installed, checkErr := argoPlugin.IsInstalled(kubecontext, namespace)
			if checkErr != nil {
				results = append(results, pluginResult{Name: "argocd", Err: fmt.Errorf("failed to check status: %w", checkErr)})
			} else if !installed {
				pluginLog.Info("ArgoCD not installed, performing full installation...\n")
				err := argoPlugin.Install(cfg.Plugins.ArgoCD, kubecontext)
				results = append(results, pluginResult{Name: "argocd", Err: err})
			} else {
				err := argoPlugin.Upgrade(cfg.Plugins.ArgoCD, kubecontext)
				results = append(results, pluginResult{Name: "argocd", Err: err})
			}
			if hasErrors(results) && upgradeFailFast {
				printSummary(results, log)
				return fmt.Errorf("some plugins failed to upgrade, see summary above")
			}
		} else if cfg.Plugins.ArgoCD != nil && !cfg.Plugins.ArgoCD.Enabled {
			pluginLog := newLogger("[argocd]")
			argoPlugin := argocd.New(pluginLog, globalTimeout)
			namespace := cfg.Plugins.ArgoCD.Namespace
			if namespace == "" {
				namespace = "argocd"
			}
			installed, err := argoPlugin.IsInstalled(kubecontext, namespace)
			if err == nil && installed {
				log.Warn("[argocd] WARNING: ArgoCD is installed but disabled in config. It will NOT be automatically uninstalled.\n")
				log.Warn("[argocd] To uninstall manually: kubectl delete namespace %s --context %s\n", namespace, kubecontext)
			}
		}

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

func runUpgradeDryRun(cfg *config.Config, kubecontext string) error {
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
				fmt.Printf("  ~ %s (%s@%s) (update)\n", app.Name, app.Chart, version)
			} else {
				fmt.Printf("  + %s (%s@%s) (install)\n", app.Name, app.Chart, version)
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
	upgradeCmd.Flags().StringVarP(&upgradeConfigFile, "config", "c", "cluster.yaml", "cluster configuration file")
	upgradeCmd.Flags().StringVarP(&upgradeEnvFile, "env", "e", ".env", "environment file for secrets")
	upgradeCmd.Flags().BoolVar(&upgradeDryRun, "dry-run", false, "preview changes without applying them")
	upgradeCmd.Flags().BoolVar(&upgradeFailFast, "fail-fast", false, "stop at first plugin failure")
	rootCmd.AddCommand(upgradeCmd)
}
