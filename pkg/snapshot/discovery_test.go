package snapshot

import (
	"testing"
)

func TestParseAPIResources(t *testing.T) {
	// Format WITHOUT SHORTNAMES (some kubectl versions)
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

func TestParseAPIResources_WithShortnames(t *testing.T) {
	// Real kubectl output format with SHORTNAMES and CATEGORIES columns
	output := `NAME                                SHORTNAMES         APIVERSION                        NAMESPACED   KIND                               VERBS                                                        CATEGORIES
configmaps                          cm                 v1                                true         ConfigMap                          create,delete,deletecollection,get,list,patch,update,watch
endpoints                           ep                 v1                                true         Endpoints                          create,delete,deletecollection,get,list,patch,update,watch
events                              ev                 v1                                true         Event                              create,delete,deletecollection,get,list,patch,update,watch
namespaces                          ns                 v1                                false        Namespace                          create,delete,get,list,patch,update,watch
nodes                               no                 v1                                false        Node                               create,delete,deletecollection,get,list,patch,update,watch
pods                                po                 v1                                true         Pod                                create,delete,deletecollection,get,list,patch,update,watch   all
secrets                                                v1                                true         Secret                             create,delete,deletecollection,get,list,patch,update,watch
services                            svc                v1                                true         Service                            create,delete,deletecollection,get,list,patch,update,watch   all
deployments                         deploy             apps/v1                           true         Deployment                         create,delete,deletecollection,get,list,patch,update,watch   all
clusterroles                                           rbac.authorization.k8s.io/v1      false        ClusterRole                        create,delete,deletecollection,get,list,patch,update,watch
ingresses                           ing                networking.k8s.io/v1              true         Ingress                            create,delete,deletecollection,get,list,patch,update,watch`

	resources, err := parseAPIResources(output)
	if err != nil {
		t.Fatalf("parseAPIResources() error = %v", err)
	}

	found := make(map[string]APIResource)
	for _, r := range resources {
		found[r.Name] = r
	}

	// Deployments should be parsed correctly despite SHORTNAMES column
	dep, ok := found["deployments"]
	if !ok {
		t.Fatal("deployments not found in results")
	}
	if dep.Group != "apps" {
		t.Errorf("deployments group = %q, want %q", dep.Group, "apps")
	}
	if dep.Version != "v1" {
		t.Errorf("deployments version = %q, want %q", dep.Version, "v1")
	}
	if dep.Kind != "Deployment" {
		t.Errorf("deployments kind = %q, want %q", dep.Kind, "Deployment")
	}
	if !dep.Namespaced {
		t.Error("deployments should be namespaced")
	}

	// Secrets (no shortname) should work
	sec, ok := found["secrets"]
	if !ok {
		t.Fatal("secrets not found in results")
	}
	if sec.Group != "" {
		t.Errorf("secrets group = %q, want empty", sec.Group)
	}

	// Ingresses should be parsed correctly
	ing, ok := found["ingresses"]
	if !ok {
		t.Fatal("ingresses not found in results")
	}
	if ing.Group != "networking.k8s.io" {
		t.Errorf("ingresses group = %q, want %q", ing.Group, "networking.k8s.io")
	}

	// ClusterRoles (no shortname, grouped) should work
	cr, ok := found["clusterroles"]
	if !ok {
		t.Fatal("clusterroles not found in results")
	}
	if cr.Group != "rbac.authorization.k8s.io" {
		t.Errorf("clusterroles group = %q, want %q", cr.Group, "rbac.authorization.k8s.io")
	}
	if cr.Namespaced {
		t.Error("clusterroles should not be namespaced")
	}

	// Excluded resources should be filtered
	for _, name := range []string{"events", "endpoints", "nodes", "pods"} {
		if _, ok := found[name]; ok {
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
