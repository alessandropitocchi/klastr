package certmanager

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/alepito/deploy-cluster/pkg/config"
)

const defaultVersion = "v1.16.3"

type Plugin struct {
	Verbose bool
}

func New() *Plugin {
	return &Plugin{Verbose: true}
}

func (p *Plugin) Name() string {
	return "cert-manager"
}

func (p *Plugin) log(format string, args ...any) {
	if p.Verbose {
		fmt.Printf(format, args...)
	}
}

func (p *Plugin) manifestURL(version string) string {
	return fmt.Sprintf("https://github.com/cert-manager/cert-manager/releases/download/%s/cert-manager.yaml", version)
}

func (p *Plugin) Install(cfg *config.CertManagerConfig, kubecontext string) error {
	version := cfg.Version
	if version == "" {
		version = defaultVersion
	}

	p.log("[cert-manager] Installing cert-manager %s...\n", version)

	url := p.manifestURL(version)
	cmd := exec.Command("kubectl", "--context", kubecontext, "apply", "-f", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply cert-manager manifest: %w", err)
	}

	// Wait for webhook to be ready (critical for cert-manager to work)
	p.log("[cert-manager] Waiting for cert-manager-webhook to be ready...\n")
	waitCmd := exec.Command("kubectl", "--context", kubecontext,
		"rollout", "status", "deployment/cert-manager-webhook",
		"-n", "cert-manager", "--timeout", (3 * time.Minute).String())
	waitCmd.Stdout = os.Stdout
	waitCmd.Stderr = os.Stderr
	if err := waitCmd.Run(); err != nil {
		return fmt.Errorf("cert-manager-webhook not ready: %w", err)
	}

	// Also wait for the main controller
	p.log("[cert-manager] Waiting for cert-manager controller to be ready...\n")
	waitCtrl := exec.Command("kubectl", "--context", kubecontext,
		"rollout", "status", "deployment/cert-manager",
		"-n", "cert-manager", "--timeout", (2 * time.Minute).String())
	waitCtrl.Stdout = os.Stdout
	waitCtrl.Stderr = os.Stderr
	if err := waitCtrl.Run(); err != nil {
		return fmt.Errorf("cert-manager controller not ready: %w", err)
	}

	p.log("[cert-manager] ✓ cert-manager %s installed successfully\n", version)
	p.log("[cert-manager] You can now create Issuer/ClusterIssuer and Certificate resources\n")
	return nil
}

func (p *Plugin) Uninstall(cfg *config.CertManagerConfig, kubecontext string) error {
	version := cfg.Version
	if version == "" {
		version = defaultVersion
	}

	p.log("[cert-manager] Uninstalling cert-manager...\n")

	url := p.manifestURL(version)
	cmd := exec.Command("kubectl", "--context", kubecontext, "delete", "-f", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete cert-manager: %w", err)
	}

	p.log("[cert-manager] ✓ cert-manager uninstalled\n")
	return nil
}

func (p *Plugin) IsInstalled(kubecontext string) (bool, error) {
	cmd := exec.Command("kubectl", "--context", kubecontext,
		"get", "deployment", "cert-manager", "-n", "cert-manager")
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}
