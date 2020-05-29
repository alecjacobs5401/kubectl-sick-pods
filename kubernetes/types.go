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

// PodCondition is a wrapper around a Kubernetes Pod Condition
type PodCondition struct {
	Type       string
	Successful bool
	Reason     string
	Message    string
}

// ContainerStatus is a wrapper around a Kubernetes Pod's Container Status
type ContainerStatus struct {
	Name  string
	Ready bool
}

// ConfigPathFromDirectory determines the kube config location from the HOME environment variable. If HOME is not defined, return empty.
func ConfigPathFromDirectory(d string) string {
	if d != "" {
		return filepath.Join(d, ".kube", "config")
	}
	return ""
}
