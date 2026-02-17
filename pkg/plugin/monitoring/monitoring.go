package monitoring

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/alepito/deploy-cluster/pkg/config"
)

const (
	// kube-prometheus-stack setup/teardown manifests
	kubePrometheusVersion = "v0.14.0"
	kubePrometheusBaseURL = "https://raw.githubusercontent.com/prometheus-operator/kube-prometheus/" + kubePrometheusVersion + "/manifests"
)

type Plugin struct {
	Verbose bool
}

func New() *Plugin {
	return &Plugin{Verbose: true}
}

func (p *Plugin) Name() string {
	return "monitoring"
}

func (p *Plugin) log(format string, args ...any) {
	if p.Verbose {
		fmt.Printf(format, args...)
	}
}

func (p *Plugin) Install(cfg *config.MonitoringConfig, kubecontext string) error {
	switch cfg.Type {
	case "prometheus":
		return p.installPrometheus(kubecontext)
	default:
		return fmt.Errorf("unsupported monitoring type: %s (supported: prometheus)", cfg.Type)
	}
}

func (p *Plugin) Uninstall(cfg *config.MonitoringConfig, kubecontext string) error {
	switch cfg.Type {
	case "prometheus":
		return p.uninstallPrometheus(kubecontext)
	default:
		return fmt.Errorf("unsupported monitoring type: %s", cfg.Type)
	}
}

func (p *Plugin) IsInstalled(kubecontext string) (bool, error) {
	cmd := exec.Command("kubectl", "--context", kubecontext,
		"get", "deployment", "prometheus-operator", "-n", "monitoring")
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (p *Plugin) installPrometheus(kubecontext string) error {
	p.log("[monitoring] Installing kube-prometheus-stack %s...\n", kubePrometheusVersion)

	// Step 1: Apply CRDs and namespace setup first
	p.log("[monitoring] Applying setup manifests (CRDs, namespace)...\n")
	setupURL := kubePrometheusBaseURL + "/setup"
	cmd := exec.Command("kubectl", "--context", kubecontext,
		"apply", "--server-side", "-f", setupURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply setup manifests: %w", err)
	}

	// Step 2: Wait for CRDs to be established
	p.log("[monitoring] Waiting for CRDs to be established...\n")
	crdCmd := exec.Command("kubectl", "--context", kubecontext,
		"wait", "--for", "condition=Established",
		"--all", "CustomResourceDefinition",
		"--namespace=monitoring", "--timeout=60s")
	crdCmd.Stdout = os.Stdout
	crdCmd.Stderr = os.Stderr
	if err := crdCmd.Run(); err != nil {
		// Non-fatal: CRDs might already be established
		p.log("[monitoring] Warning: CRD wait returned: %v (continuing)\n", err)
	}

	// Step 3: Apply main manifests
	p.log("[monitoring] Applying main manifests...\n")
	mainCmd := exec.Command("kubectl", "--context", kubecontext,
		"apply", "-f", kubePrometheusBaseURL)
	mainCmd.Stdout = os.Stdout
	mainCmd.Stderr = os.Stderr
	if err := mainCmd.Run(); err != nil {
		return fmt.Errorf("failed to apply monitoring manifests: %w", err)
	}

	// Step 4: Wait for key deployments
	deployments := []string{"prometheus-operator", "grafana", "kube-state-metrics"}
	for _, dep := range deployments {
		p.log("[monitoring] Waiting for %s to be ready...\n", dep)
		waitCmd := exec.Command("kubectl", "--context", kubecontext,
			"rollout", "status", "deployment/"+dep,
			"-n", "monitoring", "--timeout", (3 * time.Minute).String())
		waitCmd.Stdout = os.Stdout
		waitCmd.Stderr = os.Stderr
		if err := waitCmd.Run(); err != nil {
			p.log("[monitoring] Warning: %s not ready: %v\n", dep, err)
		}
	}

	p.log("[monitoring] ✓ kube-prometheus-stack installed successfully\n")
	p.log("\n[monitoring] To access Grafana:\n")
	p.log("  kubectl port-forward svc/grafana -n monitoring 3000:3000\n")
	p.log("  Open: http://localhost:3000 (admin/admin)\n")
	p.log("\n[monitoring] To access Prometheus:\n")
	p.log("  kubectl port-forward svc/prometheus-k8s -n monitoring 9090:9090\n")
	p.log("  Open: http://localhost:9090\n")
	return nil
}

func (p *Plugin) uninstallPrometheus(kubecontext string) error {
	p.log("[monitoring] Uninstalling kube-prometheus-stack...\n")

	// Delete main manifests first, then setup (reverse order)
	mainCmd := exec.Command("kubectl", "--context", kubecontext,
		"delete", "--ignore-not-found=true", "-f", kubePrometheusBaseURL)
	mainCmd.Stdout = os.Stdout
	mainCmd.Stderr = os.Stderr
	if err := mainCmd.Run(); err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("failed to delete monitoring manifests: %w", err)
		}
	}

	setupCmd := exec.Command("kubectl", "--context", kubecontext,
		"delete", "--ignore-not-found=true", "-f", kubePrometheusBaseURL+"/setup")
	setupCmd.Stdout = os.Stdout
	setupCmd.Stderr = os.Stderr
	if err := setupCmd.Run(); err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("failed to delete setup manifests: %w", err)
		}
	}

	p.log("[monitoring] ✓ kube-prometheus-stack uninstalled\n")
	return nil
}
