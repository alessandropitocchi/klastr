package plugin

// Plugin defines the interface for cluster plugins (ArgoCD, storage, etc.)
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Install installs the plugin on the cluster
	// kubeconfig is the path to the kubeconfig file
	Install(kubeconfig string) error

	// Uninstall removes the plugin from the cluster
	Uninstall(kubeconfig string) error

	// IsInstalled checks if the plugin is already installed
	IsInstalled(kubeconfig string) (bool, error)
}
