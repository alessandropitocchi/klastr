package certmanager

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/alepito/deploy-cluster/pkg/config"
	"github.com/alepito/deploy-cluster/pkg/logger"
	"github.com/alepito/deploy-cluster/pkg/retry"
)

// execCommand is a package-level variable for creating exec.Cmd, replaceable in tests.
var execCommand = exec.Command

const defaultVersion = "v1.16.3"

type Plugin struct {
	Log     *logger.Logger
	Timeout time.Duration
}

func New(log *logger.Logger, timeout time.Duration) *Plugin {
	return &Plugin{Log: log, Timeout: timeout}
}

func (p *Plugin) Name() string {
	return "cert-manager"
}

func (p *Plugin) manifestURL(version string) string {
	return fmt.Sprintf("https://github.com/cert-manager/cert-manager/releases/download/%s/cert-manager.yaml", version)
}

func (p *Plugin) Install(cfg *config.CertManagerConfig, kubecontext string) error {
	version := cfg.Version
	if version == "" {
		version = defaultVersion
	}

	p.Log.Info("Installing cert-manager %s...\n", version)

	url := p.manifestURL(version)
	err := retry.Run(3, 5*time.Second, p.Log.Warn, func() error {
		cmd := execCommand("kubectl", "--context", kubecontext, "apply", "-f", url)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
	if err != nil {
		return fmt.Errorf("failed to apply cert-manager manifest: %w", err)
	}

	// Wait for webhook to be ready (critical for cert-manager to work)
	p.Log.Info("Waiting for cert-manager-webhook to be ready...\n")
	waitCmd := execCommand("kubectl", "--context", kubecontext,
		"rollout", "status", "deployment/cert-manager-webhook",
		"-n", "cert-manager", "--timeout", p.Timeout.String())
	waitCmd.Stdout = os.Stdout
	waitCmd.Stderr = os.Stderr
	if err := waitCmd.Run(); err != nil {
		return fmt.Errorf("cert-manager-webhook not ready: %w", err)
	}

	// Also wait for the main controller
	p.Log.Info("Waiting for cert-manager controller to be ready...\n")
	waitCtrl := execCommand("kubectl", "--context", kubecontext,
		"rollout", "status", "deployment/cert-manager",
		"-n", "cert-manager", "--timeout", p.Timeout.String())
	waitCtrl.Stdout = os.Stdout
	waitCtrl.Stderr = os.Stderr
	if err := waitCtrl.Run(); err != nil {
		return fmt.Errorf("cert-manager controller not ready: %w", err)
	}

	p.Log.Success("cert-manager %s installed successfully\n", version)
	p.Log.Info("You can now create Issuer/ClusterIssuer and Certificate resources\n")
	return nil
}

func (p *Plugin) Uninstall(cfg *config.CertManagerConfig, kubecontext string) error {
	version := cfg.Version
	if version == "" {
		version = defaultVersion
	}

	p.Log.Info("Uninstalling cert-manager...\n")

	url := p.manifestURL(version)
	cmd := execCommand("kubectl", "--context", kubecontext, "delete", "-f", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete cert-manager: %w", err)
	}

	p.Log.Success("cert-manager uninstalled\n")
	return nil
}

func (p *Plugin) IsInstalled(kubecontext string) (bool, error) {
	cmd := execCommand("kubectl", "--context", kubecontext,
		"get", "deployment", "cert-manager", "-n", "cert-manager")
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}
