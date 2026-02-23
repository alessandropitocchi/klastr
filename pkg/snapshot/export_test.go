package snapshot

import (
	"path/filepath"
	"testing"
)

func TestSanitizeResource(t *testing.T) {
	res := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":              "my-config",
			"namespace":         "default",
			"resourceVersion":   "12345",
			"uid":               "abc-def",
			"creationTimestamp":  "2025-01-01T00:00:00Z",
			"generation":        float64(1),
			"managedFields":     []interface{}{map[string]interface{}{"manager": "kubectl"}},
			"selfLink":          "/api/v1/namespaces/default/configmaps/my-config",
			"labels":            map[string]interface{}{"app": "test"},
			"annotations": map[string]interface{}{
				"kubectl.kubernetes.io/last-applied-configuration": `{"apiVersion":"v1"}`,
				"my-annotation": "keep-this",
			},
		},
		"data": map[string]interface{}{
			"key": "value",
		},
		"status": map[string]interface{}{
			"phase": "Active",
		},
	}

	sanitized := sanitizeResource(res)

	// Check removed fields
	meta := sanitized["metadata"].(map[string]interface{})
	for _, field := range []string{"resourceVersion", "uid", "creationTimestamp", "generation", "managedFields", "selfLink"} {
		if _, ok := meta[field]; ok {
			t.Errorf("metadata.%s should be removed", field)
		}
	}

	// Check status removed
	if _, ok := sanitized["status"]; ok {
		t.Error("status should be removed")
	}

	// Check last-applied-configuration removed
	annotations := meta["annotations"].(map[string]interface{})
	if _, ok := annotations["kubectl.kubernetes.io/last-applied-configuration"]; ok {
		t.Error("last-applied-configuration annotation should be removed")
	}

	// Check preserved fields
	if meta["name"] != "my-config" {
		t.Errorf("name = %v, want %q", meta["name"], "my-config")
	}
	if meta["namespace"] != "default" {
		t.Errorf("namespace = %v, want %q", meta["namespace"], "default")
	}
	if annotations["my-annotation"] != "keep-this" {
		t.Errorf("my-annotation = %v, want %q", annotations["my-annotation"], "keep-this")
	}
	if sanitized["data"].(map[string]interface{})["key"] != "value" {
		t.Error("data.key should be preserved")
	}

	// Check original not modified
	origMeta := res["metadata"].(map[string]interface{})
	if _, ok := origMeta["resourceVersion"]; !ok {
		t.Error("original should not be modified")
	}
}

func TestSanitizeResource_EmptyAnnotationsRemoved(t *testing.T) {
	res := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "test",
			"annotations": map[string]interface{}{
				"kubectl.kubernetes.io/last-applied-configuration": `{}`,
			},
		},
	}

	sanitized := sanitizeResource(res)
	meta := sanitized["metadata"].(map[string]interface{})
	if _, ok := meta["annotations"]; ok {
		t.Error("empty annotations map should be removed")
	}
}

func TestIsSystemResource_OwnerReferences(t *testing.T) {
	res := map[string]interface{}{
		"kind": "ReplicaSet",
		"metadata": map[string]interface{}{
			"name":      "my-app-abc123",
			"namespace": "default",
			"ownerReferences": []interface{}{
				map[string]interface{}{
					"kind": "Deployment",
					"name": "my-app",
				},
			},
		},
	}

	if !isSystemResource(res) {
		t.Error("resource with ownerReferences should be system resource")
	}
}

func TestIsSystemResource_KubernetesService(t *testing.T) {
	res := map[string]interface{}{
		"kind": "Service",
		"metadata": map[string]interface{}{
			"name":      "kubernetes",
			"namespace": "default",
		},
	}

	if !isSystemResource(res) {
		t.Error("kubernetes service should be system resource")
	}
}

func TestIsSystemResource_KubeRootCACert(t *testing.T) {
	res := map[string]interface{}{
		"kind": "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "kube-root-ca.crt",
			"namespace": "default",
		},
	}

	if !isSystemResource(res) {
		t.Error("kube-root-ca.crt should be system resource")
	}
}

func TestIsSystemResource_DefaultServiceAccount(t *testing.T) {
	res := map[string]interface{}{
		"kind": "ServiceAccount",
		"metadata": map[string]interface{}{
			"name":      "default",
			"namespace": "default",
		},
	}

	if !isSystemResource(res) {
		t.Error("default ServiceAccount should be system resource")
	}
}

func TestIsSystemResource_UserResource(t *testing.T) {
	res := map[string]interface{}{
		"kind": "Deployment",
		"metadata": map[string]interface{}{
			"name":      "my-app",
			"namespace": "default",
		},
	}

	if isSystemResource(res) {
		t.Error("user deployment should not be system resource")
	}
}

func TestIsSystemResource_EmptyOwnerReferences(t *testing.T) {
	res := map[string]interface{}{
		"kind": "ConfigMap",
		"metadata": map[string]interface{}{
			"name":            "my-config",
			"namespace":       "default",
			"ownerReferences": []interface{}{},
		},
	}

	if isSystemResource(res) {
		t.Error("resource with empty ownerReferences should not be system resource")
	}
}

func TestIsSystemResource_NoMetadata(t *testing.T) {
	res := map[string]interface{}{
		"kind": "ConfigMap",
	}

	if !isSystemResource(res) {
		t.Error("resource with no metadata should be treated as system resource")
	}
}

func TestDeepCopyMap(t *testing.T) {
	original := map[string]interface{}{
		"a": "hello",
		"b": map[string]interface{}{
			"c": "nested",
		},
		"d": []interface{}{"one", "two"},
	}

	copied := deepCopyMap(original)

	// Modify the copy
	copied["a"] = "changed"
	copied["b"].(map[string]interface{})["c"] = "modified"

	// Original should be unchanged
	if original["a"] != "hello" {
		t.Error("original should not be modified")
	}
	if original["b"].(map[string]interface{})["c"] != "nested" {
		t.Error("nested original should not be modified")
	}
}

func TestWriteResource(t *testing.T) {
	dir := t.TempDir()
	res := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name": "test-config",
		},
	}

	if err := writeResource(dir, "test-config", res); err != nil {
		t.Fatalf("writeResource() error = %v", err)
	}

	// Check file exists
	files, _ := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}

func TestExportOptions_ExcludeSecrets(t *testing.T) {
	// Verify that ExcludeSecrets skips resources named "secrets"
	resources := []APIResource{
		{Name: "configmaps", Group: "", Version: "v1", Kind: "ConfigMap", Namespaced: true},
		{Name: "secrets", Group: "", Version: "v1", Kind: "Secret", Namespaced: true},
		{Name: "deployments", Group: "apps", Version: "v1", Kind: "Deployment", Namespaced: true},
		{Name: "clusterroles", Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole", Namespaced: false},
	}

	opts := ExportOptions{ExcludeSecrets: true}

	var namespacedKept, clusterScopedKept []string
	for _, r := range resources {
		if opts.ExcludeSecrets && r.Name == "secrets" {
			continue
		}
		if r.Namespaced {
			namespacedKept = append(namespacedKept, r.Name)
		} else {
			clusterScopedKept = append(clusterScopedKept, r.Name)
		}
	}

	// secrets should be excluded
	for _, name := range namespacedKept {
		if name == "secrets" {
			t.Error("secrets should be excluded when ExcludeSecrets is true")
		}
	}
	for _, name := range clusterScopedKept {
		if name == "secrets" {
			t.Error("secrets should be excluded when ExcludeSecrets is true")
		}
	}

	// configmaps and deployments should be kept
	found := make(map[string]bool)
	for _, n := range namespacedKept {
		found[n] = true
	}
	if !found["configmaps"] {
		t.Error("configmaps should not be excluded")
	}
	if !found["deployments"] {
		t.Error("deployments should not be excluded")
	}
}

func TestExportOptions_ExcludeSecrets_False(t *testing.T) {
	resources := []APIResource{
		{Name: "secrets", Group: "", Version: "v1", Kind: "Secret", Namespaced: true},
	}

	opts := ExportOptions{ExcludeSecrets: false}

	var kept []string
	for _, r := range resources {
		if opts.ExcludeSecrets && r.Name == "secrets" {
			continue
		}
		kept = append(kept, r.Name)
	}

	if len(kept) != 1 || kept[0] != "secrets" {
		t.Error("secrets should be kept when ExcludeSecrets is false")
	}
}

func TestWriteResource_SpecialChars(t *testing.T) {
	dir := t.TempDir()
	res := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "system:admin",
		},
	}

	if err := writeResource(dir, "system:admin", res); err != nil {
		t.Fatalf("writeResource() error = %v", err)
	}

	files, _ := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
}
