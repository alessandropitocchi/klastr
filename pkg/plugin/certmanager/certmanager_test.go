package certmanager

import (
	"testing"

	"github.com/alepito/deploy-cluster/pkg/logger"
)

func testLogger() *logger.Logger {
	return logger.New("[cert-manager]", logger.LevelQuiet)
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
	if got := p.Name(); got != "cert-manager" {
		t.Errorf("Name() = %q, want %q", got, "cert-manager")
	}
}

func TestManifestURL_DefaultVersion(t *testing.T) {
	p := New(testLogger())
	url := p.manifestURL("v1.16.3")
	want := "https://github.com/cert-manager/cert-manager/releases/download/v1.16.3/cert-manager.yaml"
	if url != want {
		t.Errorf("manifestURL() = %q, want %q", url, want)
	}
}

func TestManifestURL_CustomVersion(t *testing.T) {
	p := New(testLogger())
	url := p.manifestURL("v1.14.0")
	want := "https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml"
	if url != want {
		t.Errorf("manifestURL() = %q, want %q", url, want)
	}
}
