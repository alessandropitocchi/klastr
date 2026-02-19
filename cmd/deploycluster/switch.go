package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var switchExecCommand = exec.Command

var switchCmd = &cobra.Command{
	Use:   "switch [cluster-name]",
	Short: "Switch kubectl context between clusters",
	Long: `Switch the active kubectl context to a kind cluster.
Without arguments, lists all clusters and highlights the current context.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return listClusters()
		}
		return switchToCluster(args[0])
	},
}

func listClusters() error {
	// Get kind clusters
	cmd := switchExecCommand("kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	clusters := strings.TrimSpace(string(output))
	if clusters == "" {
		fmt.Println("No clusters found")
		return nil
	}

	// Get current context
	ctxCmd := switchExecCommand("kubectl", "config", "current-context")
	ctxOutput, _ := ctxCmd.Output()
	currentContext := strings.TrimSpace(string(ctxOutput))

	fmt.Println("KIND CLUSTERS")
	fmt.Println("─────────────")
	for _, cluster := range strings.Split(clusters, "\n") {
		cluster = strings.TrimSpace(cluster)
		if cluster == "" {
			continue
		}
		context := fmt.Sprintf("kind-%s", cluster)
		if context == currentContext {
			fmt.Printf("● %s (current)\n", cluster)
		} else {
			fmt.Printf("  %s\n", cluster)
		}
	}

	return nil
}

func switchToCluster(name string) error {
	// Verify cluster exists
	cmd := switchExecCommand("kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	found := false
	for _, cluster := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.TrimSpace(cluster) == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("cluster '%s' not found", name)
	}

	// Switch context
	context := fmt.Sprintf("kind-%s", name)
	switchCmd := switchExecCommand("kubectl", "config", "use-context", context)
	if err := switchCmd.Run(); err != nil {
		return fmt.Errorf("failed to switch context: %w", err)
	}

	fmt.Printf("Switched to context '%s'\n", context)
	return nil
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
