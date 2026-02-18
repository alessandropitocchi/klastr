package monitoring

import (
	"testing"
	"time"

	"github.com/alepito/deploy-cluster/pkg/config"
	"github.com/alepito/deploy-cluster/pkg/logger"
)

func testLogger() *logger.Logger {
	return logger.New("[monitoring]", logger.LevelQuiet)
}

func TestNew(t *testing.T) {
	p := New(testLogger(), 5*time.Minute)
	if p == nil {
		t.Fatal("New() should return non-nil plugin")
	}
	if p.Log == nil {
		t.Error("Log should not be nil")
	}
}

func TestName(t *testing.T) {
	p := New(testLogger(), 5*time.Minute)
	if got := p.Name(); got != "monitoring" {
		t.Errorf("Name() = %q, want %q", got, "monitoring")
	}
}

func TestInstall_UnsupportedType(t *testing.T) {
	p := New(testLogger(), 5*time.Minute)
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
	p := New(testLogger(), 5*time.Minute)
	cfg := &config.MonitoringConfig{Enabled: true, Type: "datadog"}
	err := p.Uninstall(cfg, "fake-context")
	if err == nil {
		t.Fatal("Uninstall() should fail for unsupported type")
	}
	if got := err.Error(); got != "unsupported monitoring type: datadog" {
		t.Errorf("error = %q, want specific message", got)
	}
}

func TestChartVersion_Default(t *testing.T) {
	p := New(testLogger(), 5*time.Minute)
	cfg := &config.MonitoringConfig{Enabled: true, Type: "prometheus"}
	if got := p.chartVersion(cfg); got != defaultChartVersion {
		t.Errorf("chartVersion() = %q, want %q", got, defaultChartVersion)
	}
}

func TestChartVersion_Custom(t *testing.T) {
	p := New(testLogger(), 5*time.Minute)
	cfg := &config.MonitoringConfig{Enabled: true, Type: "prometheus", Version: "70.0.0"}
	if got := p.chartVersion(cfg); got != "70.0.0" {
		t.Errorf("chartVersion() = %q, want %q", got, "70.0.0")
	}
}
