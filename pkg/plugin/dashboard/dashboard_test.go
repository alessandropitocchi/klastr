package dashboard

import (
	"testing"

	"github.com/alepito/deploy-cluster/pkg/config"
	"github.com/alepito/deploy-cluster/pkg/logger"
)

func testLogger() *logger.Logger {
	return logger.New("[dashboard]", logger.LevelQuiet)
}

func TestNew(t *testing.T) {
	p := New(testLogger())
	if p == nil {
		t.Fatal("New() should return non-nil plugin")
	}
	if p.Log == nil {
		t.Error("Log should not be nil")
	}
}

func TestName(t *testing.T) {
	p := New(testLogger())
	if got := p.Name(); got != "dashboard" {
		t.Errorf("Name() = %q, want %q", got, "dashboard")
	}
}

func TestInstall_UnsupportedType(t *testing.T) {
	p := New(testLogger())
	cfg := &config.DashboardConfig{Enabled: true, Type: "lens"}
	err := p.Install(cfg, "fake-context")
	if err == nil {
		t.Fatal("Install() should fail for unsupported type")
	}
	if got := err.Error(); got != "unsupported dashboard type: lens (supported: headlamp)" {
		t.Errorf("error = %q, want specific message", got)
	}
}

func TestUninstall_UnsupportedType(t *testing.T) {
	p := New(testLogger())
	cfg := &config.DashboardConfig{Enabled: true, Type: "lens"}
	err := p.Uninstall(cfg, "fake-context")
	if err == nil {
		t.Fatal("Uninstall() should fail for unsupported type")
	}
	if got := err.Error(); got != "unsupported dashboard type: lens" {
		t.Errorf("error = %q, want specific message", got)
	}
}

func TestChartVersion_Default(t *testing.T) {
	p := New(testLogger())
	cfg := &config.DashboardConfig{Enabled: true, Type: "headlamp"}
	if got := p.chartVersion(cfg); got != defaultHeadlampVersion {
		t.Errorf("chartVersion() = %q, want %q", got, defaultHeadlampVersion)
	}
}

func TestChartVersion_Custom(t *testing.T) {
	p := New(testLogger())
	cfg := &config.DashboardConfig{Enabled: true, Type: "headlamp", Version: "0.20.0"}
	if got := p.chartVersion(cfg); got != "0.20.0" {
		t.Errorf("chartVersion() = %q, want %q", got, "0.20.0")
	}
}
