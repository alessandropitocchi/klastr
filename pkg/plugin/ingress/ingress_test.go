package ingress

import (
	"testing"

	"github.com/alepito/deploy-cluster/pkg/config"
)

func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() should return non-nil plugin")
	}
	if !p.Verbose {
		t.Error("Verbose should default to true")
	}
}

func TestName(t *testing.T) {
	p := New()
	if got := p.Name(); got != "ingress" {
		t.Errorf("Name() = %q, want %q", got, "ingress")
	}
}

func TestInstall_UnsupportedType(t *testing.T) {
	p := New()
	p.Verbose = false
	cfg := &config.IngressConfig{Enabled: true, Type: "traefik"}
	err := p.Install(cfg, "fake-context")
	if err == nil {
		t.Fatal("Install() should fail for unsupported type")
	}
	if got := err.Error(); got != "unsupported ingress type: traefik (supported: nginx)" {
		t.Errorf("error = %q, want specific message", got)
	}
}

func TestUninstall_UnsupportedType(t *testing.T) {
	p := New()
	p.Verbose = false
	cfg := &config.IngressConfig{Enabled: true, Type: "traefik"}
	err := p.Uninstall(cfg, "fake-context")
	if err == nil {
		t.Fatal("Uninstall() should fail for unsupported type")
	}
	if got := err.Error(); got != "unsupported ingress type: traefik" {
		t.Errorf("error = %q, want specific message", got)
	}
}

func TestLog_Verbose(t *testing.T) {
	p := New()
	p.log("test %s\n", "message")
}

func TestLog_Silent(t *testing.T) {
	p := New()
	p.Verbose = false
	p.log("test %s\n", "message")
}
