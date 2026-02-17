package monitoring

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
	if got := p.Name(); got != "monitoring" {
		t.Errorf("Name() = %q, want %q", got, "monitoring")
	}
}

func TestInstall_UnsupportedType(t *testing.T) {
	p := New()
	p.Verbose = false
	cfg := &config.MonitoringConfig{Enabled: true, Type: "datadog"}
	err := p.Install(cfg, "fake-context")
	if err == nil {
		t.Fatal("Install() should fail for unsupported type")
	}
	if got := err.Error(); got != "unsupported monitoring type: datadog (supported: prometheus)" {
		t.Errorf("error = %q, want specific message", got)
	}
}

func TestUninstall_UnsupportedType(t *testing.T) {
	p := New()
	p.Verbose = false
	cfg := &config.MonitoringConfig{Enabled: true, Type: "datadog"}
	err := p.Uninstall(cfg, "fake-context")
	if err == nil {
		t.Fatal("Uninstall() should fail for unsupported type")
	}
	if got := err.Error(); got != "unsupported monitoring type: datadog" {
		t.Errorf("error = %q, want specific message", got)
	}
}

func TestLog_Silent(t *testing.T) {
	p := New()
	p.Verbose = false
	p.log("test %s\n", "message")
}
