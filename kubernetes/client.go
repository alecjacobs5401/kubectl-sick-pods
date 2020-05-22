package kubernetes

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	// required by client-go library to auth via oidc
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// NewClient takes in a ClientConfig and generates a new Client to interface with the
// Kubernetes Cluster as outlined in the ClientConfig
func NewClient(config ClientConfig) (*Client, error) {
	c := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: config.ConfigFile},
		&clientcmd.ConfigOverrides{CurrentContext: config.Context, Context: clientcmdapi.Context{Namespace: config.Namespace}},
	)
	cc, err := c.ClientConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "building config from kube config located at %s", config.ConfigFile)
	}
	namespace, _, err := c.Namespace()
	if err != nil {
		return nil, errors.Wrap(err, "getting namepace for client")
	}

	client, err := k8s.NewForConfig(cc)
	if err != nil {
		return nil, errors.Wrap(err, "building kubernetes client")
	}

	return &Client{client: client, namespace: namespace}, nil
}

// Pods takes a PodSelectors and returns an array of Kubernetes Pods based on those selectors
func (c *Client) Pods(selectors PodSelectors) ([]corev1.Pod, error) {
	podsAPI := c.client.CoreV1().Pods(c.namespace)

	pods := []corev1.Pod{}
	if len(selectors.Names) == 0 {
		podList, err := podsAPI.List(metav1.ListOptions{LabelSelector: selectors.Label, FieldSelector: selectors.Field})
		if err != nil {
			return nil, errors.Wrapf(err, "listing pods with LabelSelector: %s and FieldSelector: %s", selectors.Label, selectors.Field)
		}
		pods = podList.Items
	} else {
		for _, name := range selectors.Names {
			pod, err := podsAPI.Get(name, metav1.GetOptions{})
			if err != nil && !k8sErrors.IsNotFound(err) {
				return nil, errors.Wrapf(err, "Pod %s failed to delete!", name)
			}

			if err == nil || !k8sErrors.IsNotFound(err) {
				pods = append(pods, *pod)
			}
		}
	}

	return pods, nil
}

// Events takes a EventSelectors and returns an array of Kubernetes Events based on those selectors
func (c *Client) Events(selectors EventSelectors) ([]corev1.Event, error) {
	eventList, err := c.client.CoreV1().Events(c.namespace).List(metav1.ListOptions{LabelSelector: selectors.Label, FieldSelector: selectors.Field})
	if err != nil {
		return nil, errors.Wrapf(err, "listing events with LabelSelector: %s and FieldSelector: %s", selectors.Label, selectors.Field)
	}

	return eventList.Items, nil
}
