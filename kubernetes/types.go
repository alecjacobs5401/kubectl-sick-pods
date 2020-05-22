package kubernetes

import (
	"path/filepath"

	k8s "k8s.io/client-go/kubernetes"
)

// PodSelectors represents what selectors to use when listing pods
type PodSelectors struct {
	Label string
	Field string
	Names []string
}

// EventSelectors represents what selectors to use when listing pods
type EventSelectors struct {
	Label string
	Field string
}

// ClientConfig represents configuration for the Kubernetes Client
type ClientConfig struct {
	ConfigFile string
	Namespace  string
	Context    string
}

// Client is a wrapper around a Kubernetes Interface
type Client struct {
	client    k8s.Interface
	namespace string
}

// ConfigPathFromDirectory determines the kube config location from the HOME environment variable. If HOME is not defined, return empty.
func ConfigPathFromDirectory(d string) string {
	if d != "" {
		return filepath.Join(d, ".kube", "config")
	}
	return ""
}
