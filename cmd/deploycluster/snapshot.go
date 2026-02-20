package main

import (
	"fmt"
	"strings"

	"github.com/alepito/deploy-cluster/pkg/snapshot"
	"github.com/alepito/deploy-cluster/pkg/template"
	"github.com/spf13/cobra"
)

var (
	snapshotTemplateFile string
	snapshotNamespaces   string
	snapshotDryRun       bool
	snapshotEnvFile      string
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage cluster snapshots",
	Long:  `Save, restore, list, and delete Kubernetes cluster snapshots.`,
}

var snapshotSaveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Save a snapshot of cluster resources",
	Long: `Export all Kubernetes resources from the cluster to a local snapshot.
The snapshot can later be restored to the same or a different cluster.

Note: Snapshots may contain Kubernetes Secrets in plain text.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger("[snapshot]")
		name := args[0]

		// Load .env file
		if err := template.LoadEnvFile(snapshotEnvFile); err != nil {
			return fmt.Errorf("failed to load env file: %w", err)
		}

		// Load template
		cfg, err := template.Load(snapshotTemplateFile)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}

		// Get provider
		provider, err := getProvider(cfg.Provider.Type)
		if err != nil {
			return err
		}

		// Check cluster exists
		exists, err := provider.Exists(cfg.Name)
		if err != nil {
			return fmt.Errorf("failed to check cluster: %w", err)
		}
		if !exists {
			return fmt.Errorf("cluster %q does not exist", cfg.Name)
		}

		kubecontext := provider.KubeContext(cfg.Name)

		// Parse namespaces flag
		var namespaces []string
		if snapshotNamespaces != "" {
			namespaces = strings.Split(snapshotNamespaces, ",")
		}

		return snapshot.Save(name, kubecontext, cfg.Name, cfg.Provider.Type, snapshotTemplateFile, namespaces, log)
	},
}

var snapshotRestoreCmd = &cobra.Command{
	Use:   "restore <name>",
	Short: "Restore a snapshot to the cluster",
	Long: `Apply resources from a saved snapshot to the cluster.
Use --dry-run to preview what would be applied without making changes.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		log := newLogger("[snapshot]")
		name := args[0]

		// Load .env file
		if err := template.LoadEnvFile(snapshotEnvFile); err != nil {
			return fmt.Errorf("failed to load env file: %w", err)
		}

		// Load template
		cfg, err := template.Load(snapshotTemplateFile)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}

		// Get provider
		provider, err := getProvider(cfg.Provider.Type)
		if err != nil {
			return err
		}

		// Check cluster exists
		exists, err := provider.Exists(cfg.Name)
		if err != nil {
			return fmt.Errorf("failed to check cluster: %w", err)
		}
		if !exists {
			return fmt.Errorf("cluster %q does not exist", cfg.Name)
		}

		kubecontext := provider.KubeContext(cfg.Name)

		return snapshot.Restore(name, kubecontext, snapshotDryRun, log)
	},
}

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all snapshots",
	Long:  `Display all saved snapshots with their metadata.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshots, err := snapshot.List()
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}

		if len(snapshots) == 0 {
			fmt.Println("No snapshots found")
			return nil
		}

		fmt.Println("SNAPSHOTS")
		fmt.Println("─────────────────────────────────────────────────────────────")
		for _, s := range snapshots {
			fmt.Printf("• %s\n", s.Name)
			fmt.Printf("  Cluster: %s (%s)\n", s.ClusterName, s.Provider)
			fmt.Printf("  Resources: %d\n", s.ResourceCount)
			fmt.Printf("  Created: %s\n", s.CreatedAt.Format("2006-01-02 15:04:05"))
			if len(s.Namespaces) > 0 {
				fmt.Printf("  Namespaces: %s\n", strings.Join(s.Namespaces, ", "))
			}
			fmt.Println()
		}

		return nil
	},
}

var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a snapshot",
	Long:  `Permanently delete a saved snapshot from disk.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		log := newLogger("[snapshot]")

		if err := snapshot.Delete(name); err != nil {
			return fmt.Errorf("failed to delete snapshot: %w", err)
		}

		log.Success("Snapshot %q deleted\n", name)
		return nil
	},
}

func init() {
	snapshotSaveCmd.Flags().StringVarP(&snapshotTemplateFile, "template", "t", "template.yaml", "cluster template file")
	snapshotSaveCmd.Flags().StringVar(&snapshotNamespaces, "namespace", "", "comma-separated list of namespaces to snapshot (default: all non-system)")
	snapshotSaveCmd.Flags().StringVarP(&snapshotEnvFile, "env", "e", ".env", "environment file for secrets")

	snapshotRestoreCmd.Flags().StringVarP(&snapshotTemplateFile, "template", "t", "template.yaml", "cluster template file")
	snapshotRestoreCmd.Flags().BoolVar(&snapshotDryRun, "dry-run", false, "preview restore without applying")
	snapshotRestoreCmd.Flags().StringVarP(&snapshotEnvFile, "env", "e", ".env", "environment file for secrets")

	snapshotCmd.AddCommand(snapshotSaveCmd)
	snapshotCmd.AddCommand(snapshotRestoreCmd)
	snapshotCmd.AddCommand(snapshotListCmd)
	snapshotCmd.AddCommand(snapshotDeleteCmd)
	rootCmd.AddCommand(snapshotCmd)
}
