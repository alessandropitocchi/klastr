package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "deploy-cluster",
	Short: "Deploy Kubernetes clusters with plugins",
	Long: `deploy-cluster is a CLI tool for deploying Kubernetes clusters
on various providers (kind, k3d) with configurable topology
and plugin support (ArgoCD, storage, ingress).`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
