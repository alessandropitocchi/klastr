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
// Handles variable columns: NAME, SHORTNAMES (optional), APIVERSION/APIGROUP, NAMESPACED, KIND, VERBS, CATEGORIES (optional).
func parseAPIResources(output string) ([]APIResource, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("unexpected api-resources output: too few lines")
	}

	// Parse header to find column positions
	header := lines[0]

	// Build ordered list of column starts from the header
	type column struct {
		name  string
		start int
	}
	knownColumns := []string{"NAME", "SHORTNAMES", "APIVERSION", "APIGROUP", "NAMESPACED", "KIND", "VERBS", "CATEGORIES"}
	var cols []column
	for _, name := range knownColumns {
		idx := strings.Index(header, name)
		if idx >= 0 {
			cols = append(cols, column{name: name, start: idx})
		}
	}

	// Find required columns
	colStart := func(name string) int {
		for _, c := range cols {
			if c.name == name {
				return c.start
			}
		}
		return -1
	}
	// colEnd returns the start of the next column after the named one, or -1 if it's the last.
	colEnd := func(name string) int {
		for i, c := range cols {
			if c.name == name {
				if i+1 < len(cols) {
					return cols[i+1].start
				}
				return -1 // last column
			}
		}
		return -1
	}

	nameStart := colStart("NAME")
	apiVersionStart := colStart("APIVERSION")
	if apiVersionStart < 0 {
		apiVersionStart = colStart("APIGROUP")
	}
	namespacedStart := colStart("NAMESPACED")
	kindStart := colStart("KIND")

	if nameStart < 0 || apiVersionStart < 0 || namespacedStart < 0 || kindStart < 0 {
		return nil, fmt.Errorf("unexpected api-resources header format: %s", header)
	}

	nameEnd := colEnd("NAME")

	var resources []APIResource
	for _, line := range lines[1:] {
		if len(line) < kindStart {
			continue
		}

		// Extract NAME column only (not SHORTNAMES)
		end := nameEnd
		if end < 0 || end > len(line) {
			end = len(line)
		}
		name := strings.TrimSpace(line[nameStart:end])
		// NAME could still contain shortnames if SHORTNAMES column is missing;
		// take only the first field to be safe
		if fields := strings.Fields(name); len(fields) > 0 {
			name = fields[0]
		}

		apiVersion := strings.TrimSpace(line[apiVersionStart:min(namespacedStart, len(line))])
		namespacedStr := strings.TrimSpace(line[namespacedStart:min(kindStart, len(line))])

		kindFields := strings.Fields(strings.TrimSpace(line[kindStart:]))
		if len(kindFields) == 0 {
			continue
		}
		kind := kindFields[0]

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
