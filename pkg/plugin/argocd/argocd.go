package argocd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/alepito/deploy-cluster/pkg/config"
)

const repoSecretTemplate = `apiVersion: v1
kind: Secret
metadata:
  name: repo-{{ .Name }}
  namespace: {{ .Namespace }}
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  type: {{ .Type }}
  url: {{ .URL }}
{{- if .Username }}
  username: {{ .Username }}
{{- end }}
{{- if .Password }}
  password: {{ .Password }}
{{- end }}
{{- if .Insecure }}
  insecure: "true"
{{- end }}
{{- if .SSHPrivateKey }}
  sshPrivateKey: |
{{ .SSHPrivateKey | indent 4 }}
{{- end }}
`

type Plugin struct {
	Verbose bool
}

func New() *Plugin {
	return &Plugin{Verbose: true}
}

func (p *Plugin) Name() string {
	return "argocd"
}

func (p *Plugin) log(format string, args ...interface{}) {
	if p.Verbose {
		fmt.Printf(format, args...)
	}
}

func (p *Plugin) Install(cfg *config.ArgoCDConfig, kubecontext string) error {
	// Set defaults
	namespace := cfg.Namespace
	if namespace == "" {
		namespace = "argocd"
	}

	version := cfg.Version
	if version == "" {
		version = "stable"
	}

	p.log("[argocd] Installing ArgoCD...\n")
	p.log("[argocd] Namespace: %s\n", namespace)
	p.log("[argocd] Version: %s\n", version)

	// Create namespace
	p.log("[argocd] Creating namespace '%s'...\n", namespace)
	if err := p.runKubectl(kubecontext, "create", "namespace", namespace); err != nil {
		// Ignore if already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create namespace: %w", err)
		}
		p.log("[argocd] Namespace already exists, continuing...\n")
	}

	// Install ArgoCD
	manifestURL := fmt.Sprintf("https://raw.githubusercontent.com/argoproj/argo-cd/%s/manifests/install.yaml", version)
	p.log("[argocd] Applying manifest from %s...\n", manifestURL)

	if err := p.runKubectlApply(kubecontext, namespace, manifestURL); err != nil {
		return fmt.Errorf("failed to install ArgoCD: %w", err)
	}

	// Wait for ArgoCD to be ready
	p.log("[argocd] Waiting for ArgoCD server to be ready...\n")
	if err := p.waitForDeployment(kubecontext, namespace, "argocd-server", 5*time.Minute); err != nil {
		return fmt.Errorf("ArgoCD server not ready: %w", err)
	}
	p.log("[argocd] ✓ ArgoCD server is ready\n")

	// Add repositories
	if len(cfg.Repos) > 0 {
		p.log("[argocd] Adding repositories...\n")
		for _, repo := range cfg.Repos {
			if err := p.addRepository(repo, kubecontext, namespace); err != nil {
				return fmt.Errorf("failed to add repository %s: %w", repo.URL, err)
			}
		}
		p.log("[argocd] ✓ Repositories added\n")
	}

	// Print access info
	p.log("\n[argocd] ✓ ArgoCD installed successfully!\n")
	p.log("\n[argocd] To access ArgoCD UI:\n")
	p.log("  kubectl port-forward svc/argocd-server -n %s 8080:443\n", namespace)
	p.log("  Open: https://localhost:8080\n")
	p.log("\n[argocd] Get admin password:\n")
	p.log("  kubectl -n %s get secret argocd-initial-admin-secret -o jsonpath=\"{.data.password}\" | base64 -d\n", namespace)

	if len(cfg.Repos) > 0 {
		p.log("\n[argocd] Configured repositories:\n")
		for _, repo := range cfg.Repos {
			p.log("  • %s (%s)\n", repo.URL, repo.Name)
		}
	}

	return nil
}

func (p *Plugin) addRepository(repo config.ArgoCDRepoConfig, kubecontext string, namespace string) error {
	// Set defaults
	name := repo.Name
	if name == "" {
		// Generate name from URL
		name = strings.ReplaceAll(repo.URL, "https://", "")
		name = strings.ReplaceAll(name, "http://", "")
		name = strings.ReplaceAll(name, "git@", "")
		name = strings.ReplaceAll(name, ":", "-")
		name = strings.ReplaceAll(name, "/", "-")
		name = strings.ReplaceAll(name, ".", "-")
		name = strings.TrimSuffix(name, "-git")
	}

	repoType := repo.Type
	if repoType == "" {
		repoType = "git"
	}

	// Get SSH key if configured
	var sshPrivateKey string
	if repo.SSHKeyEnv != "" {
		sshPrivateKey = os.Getenv(repo.SSHKeyEnv)
		if sshPrivateKey == "" {
			return fmt.Errorf("environment variable %s is not set", repo.SSHKeyEnv)
		}
		p.log("[argocd] Adding repository: %s (using SSH key from $%s)\n", repo.URL, repo.SSHKeyEnv)
	} else if repo.SSHKeyFile != "" {
		keyPath := repo.SSHKeyFile
		if strings.HasPrefix(keyPath, "~/") {
			home, _ := os.UserHomeDir()
			keyPath = home + keyPath[1:]
		}
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("failed to read SSH key file %s: %w", repo.SSHKeyFile, err)
		}
		sshPrivateKey = string(keyData)
		p.log("[argocd] Adding repository: %s (using SSH key from %s)\n", repo.URL, repo.SSHKeyFile)
	} else {
		p.log("[argocd] Adding repository: %s\n", repo.URL)
	}

	// Template functions
	funcMap := template.FuncMap{
		"indent": func(spaces int, s string) string {
			pad := strings.Repeat(" ", spaces)
			lines := strings.Split(s, "\n")
			for i, line := range lines {
				if line != "" {
					lines[i] = pad + line
				}
			}
			return strings.Join(lines, "\n")
		},
	}

	// Generate Secret manifest
	tmpl, err := template.New("repo").Funcs(funcMap).Parse(repoSecretTemplate)
	if err != nil {
		return err
	}

	// Determine insecure flag: explicit config or auto-detect for non-HTTPS
	insecure := false
	if repo.Insecure != nil {
		insecure = *repo.Insecure
	} else if !strings.HasPrefix(repo.URL, "https://") {
		insecure = true
	}
	if insecure {
		p.log("[argocd] Using insecure mode (skip TLS verification)\n")
	}

	data := struct {
		Name          string
		Namespace     string
		Type          string
		URL           string
		Insecure      bool
		Username      string
		Password      string
		SSHPrivateKey string
	}{
		Name:          name,
		Namespace:     namespace,
		Type:          repoType,
		URL:           repo.URL,
		Insecure:      insecure,
		Username:      repo.Username,
		Password:      repo.Password,
		SSHPrivateKey: sshPrivateKey,
	}

	var manifest bytes.Buffer
	if err := tmpl.Execute(&manifest, data); err != nil {
		return err
	}

	// Apply manifest
	cmd := exec.Command("kubectl", "--context", kubecontext, "apply", "-f", "-")
	cmd.Stdin = &manifest
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *Plugin) Uninstall(kubecontext string, namespace string) error {
	if namespace == "" {
		namespace = "argocd"
	}

	p.log("[argocd] Uninstalling ArgoCD from namespace '%s'...\n", namespace)

	if err := p.runKubectl(kubecontext, "delete", "namespace", namespace); err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	p.log("[argocd] ✓ ArgoCD uninstalled\n")
	return nil
}

func (p *Plugin) IsInstalled(kubecontext string, namespace string) (bool, error) {
	if namespace == "" {
		namespace = "argocd"
	}

	cmd := exec.Command("kubectl", "--context", kubecontext, "get", "deployment", "argocd-server", "-n", namespace)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

func (p *Plugin) runKubectl(kubecontext string, args ...string) error {
	fullArgs := append([]string{"--context", kubecontext}, args...)
	cmd := exec.Command("kubectl", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *Plugin) runKubectlApply(kubecontext string, namespace string, url string) error {
	cmd := exec.Command("kubectl", "--context", kubecontext, "apply", "-n", namespace, "-f", url, "--server-side", "--force-conflicts")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *Plugin) waitForDeployment(kubecontext string, namespace string, name string, timeout time.Duration) error {
	cmd := exec.Command("kubectl", "--context", kubecontext, "rollout", "status", "deployment/"+name, "-n", namespace, "--timeout", timeout.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
