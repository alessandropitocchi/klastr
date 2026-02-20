package snapshot

import (
	"testing"
)

func TestParseAPIResources(t *testing.T) {
	output := `NAME                              APIVERSION                               NAMESPACED   KIND                             VERBS
configmaps                        v1                                       true         ConfigMap                        [create delete deletecollection get list patch update watch]
namespaces                        v1                                       false        Namespace                        [create delete get list patch update watch]
nodes                             v1                                       false        Node                             [create delete deletecollection get list patch update watch]
pods                              v1                                       true         Pod                              [create delete deletecollection get list patch update watch]
secrets                           v1                                       true         Secret                           [create delete deletecollection get list patch update watch]
services                          v1                                       true         Service                          [create delete deletecollection get list patch update watch]
deployments                       apps/v1                                  true         Deployment                       [create delete deletecollection get list patch update watch]
events                            v1                                       true         Event                            [create delete deletecollection get list patch update watch]
endpoints                         v1                                       true         Endpoints                        [create delete deletecollection get list patch update watch]
clusterroles                      rbac.authorization.k8s.io/v1             false        ClusterRole                      [create delete deletecollection get list patch update watch]`

	resources, err := parseAPIResources(output)
	if err != nil {
		t.Fatalf("parseAPIResources() error = %v", err)
	}

	// Check that excluded resources are filtered
	for _, r := range resources {
		if excludedResources[r.Name] {
			t.Errorf("excluded resource %q should not be in results", r.Name)
		}
	}

	// Check we got expected resources
	found := make(map[string]bool)
	for _, r := range resources {
		found[r.Name] = true
	}

	expectedResources := []string{"configmaps", "namespaces", "secrets", "services", "deployments", "clusterroles"}
	for _, name := range expectedResources {
		if !found[name] {
			t.Errorf("expected resource %q not found in results", name)
		}
	}

	// Verify excluded
	excludedExpected := []string{"events", "endpoints", "nodes", "pods"}
	for _, name := range excludedExpected {
		if found[name] {
			t.Errorf("excluded resource %q should not be in results", name)
		}
	}
}

func TestParseAPIResources_GroupParsing(t *testing.T) {
	output := `NAME                              APIVERSION                               NAMESPACED   KIND                             VERBS
deployments                       apps/v1                                  true         Deployment                       [create delete get list]
configmaps                        v1                                       true         ConfigMap                        [create delete get list]`

	resources, err := parseAPIResources(output)
	if err != nil {
		t.Fatalf("parseAPIResources() error = %v", err)
	}

	for _, r := range resources {
		switch r.Name {
		case "deployments":
			if r.Group != "apps" {
				t.Errorf("deployments group = %q, want %q", r.Group, "apps")
			}
			if r.Version != "v1" {
				t.Errorf("deployments version = %q, want %q", r.Version, "v1")
			}
			if r.Kind != "Deployment" {
				t.Errorf("deployments kind = %q, want %q", r.Kind, "Deployment")
			}
			if !r.Namespaced {
				t.Error("deployments should be namespaced")
			}
		case "configmaps":
			if r.Group != "" {
				t.Errorf("configmaps group = %q, want empty", r.Group)
			}
			if r.Version != "v1" {
				t.Errorf("configmaps version = %q, want %q", r.Version, "v1")
			}
		}
	}
}

func TestParseAPIResources_TooFewLines(t *testing.T) {
	_, err := parseAPIResources("HEADER ONLY")
	if err == nil {
		t.Fatal("expected error for too few lines")
	}
}

func TestParseAPIResources_EmptyOutput(t *testing.T) {
	_, err := parseAPIResources("")
	if err == nil {
		t.Fatal("expected error for empty output")
	}
}

func TestAPIResource_GroupResource(t *testing.T) {
	tests := []struct {
		name     string
		resource APIResource
		want     string
	}{
		{
			name:     "core resource",
			resource: APIResource{Name: "configmaps", Group: ""},
			want:     "configmaps",
		},
		{
			name:     "grouped resource",
			resource: APIResource{Name: "deployments", Group: "apps"},
			want:     "deployments.apps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource.GroupResource()
			if got != tt.want {
				t.Errorf("GroupResource() = %q, want %q", got, tt.want)
			}
		})
	}
}
