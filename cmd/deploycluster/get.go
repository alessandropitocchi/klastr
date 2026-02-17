package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get cluster information",
	Long:  `Display information about existing clusters.`,
}

var getClustersCmd = &cobra.Command{
	Use:   "clusters",
	Short: "List all clusters",
	Long:  `List all existing kind clusters.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get kind clusters
		kindCmd := exec.Command("kind", "get", "clusters")
		output, err := kindCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to list clusters: %w", err)
		}

		clusters := strings.TrimSpace(string(output))
		if clusters == "" {
			fmt.Println("No clusters found")
			return nil
		}

		fmt.Println("KIND CLUSTERS")
		fmt.Println("─────────────")
		for _, cluster := range strings.Split(clusters, "\n") {
			if cluster != "" {
				fmt.Printf("• %s\n", cluster)
			}
		}

		return nil
	},
}

var getNodesCmd = &cobra.Command{
	Use:   "nodes [cluster-name]",
	Short: "List nodes in a cluster",
	Long:  `List all nodes in a specific cluster.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterName := "kind"
		if len(args) > 0 {
			clusterName = args[0]
		}

		// Get nodes using docker
		dockerCmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("label=io.x-k8s.kind.cluster=%s", clusterName), "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}")
		output, err := dockerCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to list nodes: %w", err)
		}

		result := strings.TrimSpace(string(output))
		if result == "" || result == "NAMES\tSTATUS\tPORTS" {
			fmt.Printf("No nodes found for cluster '%s'\n", clusterName)
			return nil
		}

		fmt.Printf("NODES FOR CLUSTER '%s'\n", clusterName)
		fmt.Println("──────────────────────────────────────────────────────────────")
		fmt.Println(result)

		return nil
	},
}

var getKubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig [cluster-name]",
	Short: "Get kubeconfig for a cluster",
	Long:  `Output the kubeconfig for a specific cluster.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterName := "kind"
		if len(args) > 0 {
			clusterName = args[0]
		}

		kindCmd := exec.Command("kind", "get", "kubeconfig", "--name", clusterName)
		output, err := kindCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get kubeconfig: %w", err)
		}

		fmt.Print(string(output))
		return nil
	},
}

func init() {
	getCmd.AddCommand(getClustersCmd)
	getCmd.AddCommand(getNodesCmd)
	getCmd.AddCommand(getKubeconfigCmd)
	rootCmd.AddCommand(getCmd)
}
