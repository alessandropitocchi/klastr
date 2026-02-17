package ingress

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/alepito/deploy-cluster/pkg/config"
)

const (
	nginxManifestURL = "https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.0/deploy/static/provider/kind/deploy.yaml"
)

type Plugin struct {
	Verbose bool
}

func New() *Plugin {
	return &Plugin{Verbose: true}
}

func (p *Plugin) Name() string {
	return "ingress"
}

func (p *Plugin) log(format string, args ...any) {
	if p.Verbose {
		fmt.Printf(format, args...)
	}
}

func (p *Plugin) Install(cfg *config.IngressConfig, kubecontext string) error {
	switch cfg.Type {
	case "nginx":
		return p.installNginx(kubecontext)
	default:
		return fmt.Errorf("unsupported ingress type: %s (supported: nginx)", cfg.Type)
	}
}

func (p *Plugin) Uninstall(cfg *config.IngressConfig, kubecontext string) error {
	switch cfg.Type {
	case "nginx":
		return p.uninstallNginx(kubecontext)
	default:
		return fmt.Errorf("unsupported ingress type: %s", cfg.Type)
	}
}

func (p *Plugin) IsInstalled(kubecontext string) (bool, error) {
	cmd := exec.Command("kubectl", "--context", kubecontext,
		"get", "deployment", "ingress-nginx-controller", "-n", "ingress-nginx")
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (p *Plugin) installNginx(kubecontext string) error {
	p.log("[ingress] Installing nginx ingress controller...\n")

	cmd := exec.Command("kubectl", "--context", kubecontext,
		"apply", "-f", nginxManifestURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply nginx ingress manifest: %w", err)
	}

	p.log("[ingress] Waiting for nginx ingress controller to be ready...\n")
	waitCmd := exec.Command("kubectl", "--context", kubecontext,
		"rollout", "status", "deployment/ingress-nginx-controller",
		"-n", "ingress-nginx", "--timeout", (3 * time.Minute).String())
	waitCmd.Stdout = os.Stdout
	waitCmd.Stderr = os.Stderr
	if err := waitCmd.Run(); err != nil {
		return fmt.Errorf("nginx ingress controller not ready: %w", err)
	}

	p.log("[ingress] ✓ nginx ingress controller installed successfully\n")
	p.log("[ingress] Ingress class: nginx\n")
	return nil
}

func (p *Plugin) uninstallNginx(kubecontext string) error {
	p.log("[ingress] Uninstalling nginx ingress controller...\n")

	cmd := exec.Command("kubectl", "--context", kubecontext,
		"delete", "-f", nginxManifestURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete nginx ingress: %w", err)
	}

	p.log("[ingress] ✓ nginx ingress controller uninstalled\n")
	return nil
}
