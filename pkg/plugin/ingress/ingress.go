package ingress

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/alepito/deploy-cluster/pkg/template"
	"github.com/alepito/deploy-cluster/pkg/logger"
	"github.com/alepito/deploy-cluster/pkg/retry"
)

// execCommand is a package-level variable for creating exec.Cmd, replaceable in tests.
var execCommand = exec.Command

const (
	nginxManifestURL = "https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.0/deploy/static/provider/kind/deploy.yaml"
)

type Plugin struct {
	Log     *logger.Logger
	Timeout time.Duration
}

func New(log *logger.Logger, timeout time.Duration) *Plugin {
	return &Plugin{Log: log, Timeout: timeout}
}

func (p *Plugin) Name() string {
	return "ingress"
}

func (p *Plugin) Install(cfg *template.IngressTemplate, kubecontext string) error {
	switch cfg.Type {
	case "nginx":
		return p.installNginx(kubecontext)
	default:
		return fmt.Errorf("unsupported ingress type: %s (supported: nginx)", cfg.Type)
	}
}

func (p *Plugin) Uninstall(cfg *template.IngressTemplate, kubecontext string) error {
	switch cfg.Type {
	case "nginx":
		return p.uninstallNginx(kubecontext)
	default:
		return fmt.Errorf("unsupported ingress type: %s", cfg.Type)
	}
}

func (p *Plugin) IsInstalled(kubecontext string) (bool, error) {
	cmd := execCommand("kubectl", "--context", kubecontext,
		"get", "deployment", "ingress-nginx-controller", "-n", "ingress-nginx")
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (p *Plugin) installNginx(kubecontext string) error {
	p.Log.Info("Installing nginx ingress controller...\n")

	err := retry.Run(3, 5*time.Second, p.Log.Warn, func() error {
		cmd := execCommand("kubectl", "--context", kubecontext,
			"apply", "-f", nginxManifestURL)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
	if err != nil {
		return fmt.Errorf("failed to apply nginx ingress manifest: %w", err)
	}

	p.Log.Info("Waiting for nginx ingress controller to be ready...\n")
	waitCmd := execCommand("kubectl", "--context", kubecontext,
		"rollout", "status", "deployment/ingress-nginx-controller",
		"-n", "ingress-nginx", "--timeout", p.Timeout.String())
	waitCmd.Stdout = os.Stdout
	waitCmd.Stderr = os.Stderr
	if err := waitCmd.Run(); err != nil {
		return fmt.Errorf("nginx ingress controller not ready: %w", err)
	}

	p.Log.Success("nginx ingress controller installed successfully\n")
	p.Log.Info("Ingress class: nginx\n")
	return nil
}

func (p *Plugin) uninstallNginx(kubecontext string) error {
	p.Log.Info("Uninstalling nginx ingress controller...\n")

	cmd := execCommand("kubectl", "--context", kubecontext,
		"delete", "-f", nginxManifestURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete nginx ingress: %w", err)
	}

	p.Log.Success("nginx ingress controller uninstalled\n")
	return nil
}
