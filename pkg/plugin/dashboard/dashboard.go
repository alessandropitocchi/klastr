package dashboard

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/alepito/deploy-cluster/pkg/config"
	"github.com/alepito/deploy-cluster/pkg/logger"
	"github.com/alepito/deploy-cluster/pkg/retry"
)

const (
	defaultHeadlampChart   = "oci://ghcr.io/headlamp-k8s/charts/headlamp"
	defaultHeadlampVersion = "0.25.0"
	releaseName            = "headlamp"
	namespace              = "headlamp"
)

type Plugin struct {
	Log *logger.Logger
}

func New(log *logger.Logger) *Plugin {
	return &Plugin{Log: log}
}

func (p *Plugin) Name() string {
	return "dashboard"
}

func (p *Plugin) Install(cfg *config.DashboardConfig, kubecontext string) error {
	switch cfg.Type {
	case "headlamp":
		return p.installHeadlamp(cfg, kubecontext)
	default:
		return fmt.Errorf("unsupported dashboard type: %s (supported: headlamp)", cfg.Type)
	}
}

func (p *Plugin) Uninstall(cfg *config.DashboardConfig, kubecontext string) error {
	switch cfg.Type {
	case "headlamp":
		return p.uninstallHeadlamp(kubecontext)
	default:
		return fmt.Errorf("unsupported dashboard type: %s", cfg.Type)
	}
}

func (p *Plugin) IsInstalled(kubecontext string) (bool, error) {
	cmd := exec.Command("helm", "status", releaseName,
		"--namespace", namespace, "--kube-context", kubecontext)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (p *Plugin) chartVersion(cfg *config.DashboardConfig) string {
	if cfg.Version != "" {
		return cfg.Version
	}
	return defaultHeadlampVersion
}

func (p *Plugin) installHeadlamp(cfg *config.DashboardConfig, kubecontext string) error {
	version := p.chartVersion(cfg)
	p.Log.Info("Installing Headlamp %s via Helm...\n", version)

	args := []string{
		"upgrade", "--install", releaseName, defaultHeadlampChart,
		"--version", version,
		"--namespace", namespace,
		"--create-namespace",
		"--kube-context", kubecontext,
		"--wait",
		"--timeout", (5 * time.Minute).String(),
	}

	err := retry.Run(3, 5*time.Second, p.Log.Warn, func() error {
		cmd := exec.Command("helm", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
	if err != nil {
		return fmt.Errorf("failed to install Headlamp: %w", err)
	}

	// Create ClusterRoleBinding for the headlamp service account
	p.Log.Debug("Creating ClusterRoleBinding for Headlamp...\n")
	crbManifest := fmt.Sprintf(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: headlamp-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: headlamp
    namespace: %s`, namespace)

	crbCmd := exec.Command("kubectl", "--context", kubecontext, "apply", "-f", "-")
	crbCmd.Stdin = strings.NewReader(crbManifest)
	crbCmd.Stdout = os.Stdout
	crbCmd.Stderr = os.Stderr
	if err := crbCmd.Run(); err != nil {
		p.Log.Warn("Warning: failed to create ClusterRoleBinding: %v\n", err)
	}

	p.Log.Success("Headlamp installed successfully\n")

	// Configure ingress if enabled
	if cfg.Ingress != nil && cfg.Ingress.Enabled {
		if err := p.configureIngress(cfg.Ingress, kubecontext); err != nil {
			return fmt.Errorf("failed to configure Headlamp ingress: %w", err)
		}
	} else {
		p.Log.Info("\nTo access Headlamp:\n")
		p.Log.Info("  kubectl port-forward svc/headlamp -n %s 4466:80\n", namespace)
		p.Log.Info("  Open: http://localhost:4466\n")
	}

	return nil
}

func (p *Plugin) configureIngress(cfg *config.DashboardIngressConfig, kubecontext string) error {
	p.Log.Info("Configuring ingress for Headlamp...\n")

	manifest := fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: headlamp-ingress
  namespace: %s
  annotations:
    nginx.ingress.kubernetes.io/backend-protocol: "HTTP"
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
spec:
  ingressClassName: nginx
  rules:
    - host: %s
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: headlamp
                port:
                  number: 80`, namespace, cfg.Host)

	err := retry.Run(3, 5*time.Second, p.Log.Warn, func() error {
		cmd := exec.Command("kubectl", "--context", kubecontext, "apply", "-f", "-")
		cmd.Stdin = strings.NewReader(manifest)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
	if err != nil {
		return fmt.Errorf("failed to apply headlamp ingress: %w", err)
	}

	p.Log.Success("Headlamp available at: http://%s\n", cfg.Host)
	return nil
}

func (p *Plugin) uninstallHeadlamp(kubecontext string) error {
	p.Log.Info("Uninstalling Headlamp...\n")

	cmd := exec.Command("helm", "uninstall", releaseName,
		"--namespace", namespace,
		"--kube-context", kubecontext)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to uninstall Headlamp: %w", err)
	}

	// Clean up ClusterRoleBinding
	crbCmd := exec.Command("kubectl", "--context", kubecontext,
		"delete", "clusterrolebinding", "headlamp-admin", "--ignore-not-found=true")
	crbCmd.Stdout = os.Stdout
	crbCmd.Stderr = os.Stderr
	_ = crbCmd.Run()

	p.Log.Success("Headlamp uninstalled\n")
	return nil
}
