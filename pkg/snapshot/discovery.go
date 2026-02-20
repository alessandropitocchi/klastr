package snapshot

import (
	"fmt"
	"os/exec"
	"strings"
)

// execCommand is a package-level variable for creating exec.Cmd, replaceable in tests.
var execCommand = exec.Command

// APIResource represents a Kubernetes API resource type.
type APIResource struct {
	Name       string // e.g. "deployments"
	Group      string // e.g. "apps"
	Version    string // e.g. "v1"
	Kind       string // e.g. "Deployment"
	Namespaced bool
}

// GroupResource returns the fully qualified resource name (e.g. "deployments.apps").
func (r APIResource) GroupResource() string {
	if r.Group == "" {
		return r.Name
	}
	return r.Name + "." + r.Group
}

// excludedResources are resource names that should never be included in a snapshot.
var excludedResources = map[string]bool{
	"events":                          true,
	"events.events.k8s.io":           true,
	"endpoints":                       true,
	"endpointslices":                  true,
	"componentstatuses":               true,
	"nodes":                           true,
	"apiservices":                      true,
	"leases":                           true,
	"tokenreviews":                     true,
	"selfsubjectaccessreviews":         true,
	"selfsubjectrulesreviews":          true,
	"subjectaccessreviews":             true,
	"localsubjectaccessreviews":        true,
	"tokenrequests":                    true,
	"certificatesigningrequests":       true,
	"csidrivers":                       true,
	"csinodes":                         true,
	"csistoragecapacities":             true,
	"runtimeclasses":                   true,
	"prioritylevelconfigurations":      true,
	"flowschemas":                      true,
	"controllerrevisions":              true,
	"replicasets":                       true,
	"pods":                              true,
}

// DiscoverResources queries the cluster for available API resources.
func DiscoverResources(kubecontext string) ([]APIResource, error) {
	cmd := execCommand("kubectl", "--context", kubecontext, "api-resources", "-o", "wide", "--verbs=list,get")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to discover api-resources: %w", err)
	}
	return parseAPIResources(string(output))
}

// parseAPIResources parses the output of kubectl api-resources -o wide.
func parseAPIResources(output string) ([]APIResource, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("unexpected api-resources output: too few lines")
	}

	// Parse header to find column positions
	header := lines[0]
	nameIdx := 0
	apiVersionIdx := strings.Index(header, "APIVERSION")
	if apiVersionIdx < 0 {
		// Older kubectl uses APIGROUP
		apiVersionIdx = strings.Index(header, "APIGROUP")
	}
	namespacedIdx := strings.Index(header, "NAMESPACED")
	kindIdx := strings.Index(header, "KIND")

	if apiVersionIdx < 0 || namespacedIdx < 0 || kindIdx < 0 {
		return nil, fmt.Errorf("unexpected api-resources header format: %s", header)
	}

	var resources []APIResource
	for _, line := range lines[1:] {
		if len(line) < kindIdx {
			continue
		}

		name := strings.TrimSpace(line[nameIdx:min(apiVersionIdx, len(line))])
		apiVersion := strings.TrimSpace(line[apiVersionIdx:min(namespacedIdx, len(line))])
		namespacedStr := strings.TrimSpace(line[namespacedIdx:min(kindIdx, len(line))])
		kind := strings.Fields(strings.TrimSpace(line[kindIdx:]))[0]

		// Parse group and version from apiVersion field
		group := ""
		version := apiVersion
		if parts := strings.Split(apiVersion, "/"); len(parts) == 2 {
			group = parts[0]
			version = parts[1]
		}

		r := APIResource{
			Name:       name,
			Group:      group,
			Version:    version,
			Kind:       kind,
			Namespaced: strings.ToLower(namespacedStr) == "true",
		}

		// Skip excluded resources
		if excludedResources[r.Name] || excludedResources[r.GroupResource()] {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
