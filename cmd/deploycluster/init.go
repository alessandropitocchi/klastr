package main

import (
	"fmt"
	"os"

	"github.com/alepito/deploy-cluster/pkg/config"
	"github.com/spf13/cobra"
)

var (
	initOutput string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate an initial cluster configuration file",
	Long: `Generate a starter cluster.yaml configuration file
with default values that you can customize.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if file already exists
		if _, err := os.Stat(initOutput); err == nil {
			return fmt.Errorf("file %s already exists, use --output to specify a different name", initOutput)
		}

		cfg := config.DefaultConfig()
		if err := cfg.Save(initOutput); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		fmt.Printf("Created %s\n", initOutput)
		fmt.Println("\nEdit the file to customize your cluster, then run:")
		fmt.Printf("  deploy-cluster create --config %s\n", initOutput)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&initOutput, "output", "o", "cluster.yaml", "output file path")
	rootCmd.AddCommand(initCmd)
}
