package certmanager

import (
	"testing"
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
	if got := p.Name(); got != "cert-manager" {
		t.Errorf("Name() = %q, want %q", got, "cert-manager")
	}
}

func TestManifestURL_DefaultVersion(t *testing.T) {
	p := New()
	url := p.manifestURL("v1.16.3")
	want := "https://github.com/cert-manager/cert-manager/releases/download/v1.16.3/cert-manager.yaml"
	if url != want {
		t.Errorf("manifestURL() = %q, want %q", url, want)
	}
}

func TestManifestURL_CustomVersion(t *testing.T) {
	p := New()
	url := p.manifestURL("v1.14.0")
	want := "https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml"
	if url != want {
		t.Errorf("manifestURL() = %q, want %q", url, want)
	}
}

func TestLog_Silent(t *testing.T) {
	p := New()
	p.Verbose = false
	p.log("test %s\n", "message")
}
