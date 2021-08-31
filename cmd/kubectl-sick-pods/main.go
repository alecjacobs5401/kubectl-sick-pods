package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/alecjacobs5401/kube-client-wrapper/pkg/api"
	"github.com/alecjacobs5401/kube-client-wrapper/pkg/types"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

const version = "0.3.0-rc1"

func main() {
	args := os.Args[1:]

	app := kingpin.New("kubectl sick-pods", "A kubectl plugin used to find Pods that are in a 'NotReady' state and display debugging information about them")
	app.UsageWriter(os.Stdout)
	app.Version(version)
	app.HelpFlag.Short('h')

	podSelectors := types.PodSelectors{}
	clientConfig := types.ClientConfig{}

	app.Arg("pod", "A pod name to target (Accepts multiple pod names)").StringsVar(&podSelectors.Names)
	app.Flag("selector", "Pod Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)").
		Short('l').StringVar(&podSelectors.Label)
	app.Flag("field-selector", "Pod Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.").
		StringVar(&podSelectors.Field)

	app.Flag("kubeconfig", "The path to a pre-existing kubeconfig file that you want to have the new cluster authorization merged into. If not provided, the KUBECONFIG environment variable will be checked for a file. If that is empty, $HOME/.kube/config will be used.").
		Envar("KUBECONFIG").StringVar(&clientConfig.ConfigFile)
	app.Flag("namespace", "The Kubernetes namespace to use. Default to the current namespace in your kube config file.").
		Short('n').StringVar(&clientConfig.Namespace)
	app.Flag("context", "The Kubernetes context to use. Defaults to the current context in your kube config file.").
		StringVar(&clientConfig.Context)
	app.Flag("all-namespaces", "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace").
		Short('A').BoolVar(&clientConfig.AllNamespaces)

	_, err := app.Parse(args)
	if err != nil {
		appContext, _ := app.ParseContext(args)
		app.FatalUsageContext(appContext, err.Error())
	}

	if clientConfig.ConfigFile == "" {
		clientConfig.ConfigFile = api.ConfigPathFromDirectory(os.Getenv("HOME"))
	}

	if err := diagnose(podSelectors, clientConfig); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func diagnose(podSelectors types.PodSelectors, clientConfig types.ClientConfig) error {
	client, err := api.NewClient(clientConfig)
	if err != nil {
		return errors.Wrap(err, "building client")
	}

	// duplicate podSelectors
	selectors := podSelectors
	if selectors.Field != "" {
		selectors.Field = fmt.Sprintf("%s,%s", "status.phase!=Succeeded", selectors.Field)
	} else {
		selectors.Field = "status.phase!=Succeeded"
	}
	pods, err := client.Pods(selectors)
	if err != nil {
		return errors.Wrap(err, "getting pods")
	}

	for _, pod := range pods {
		if !api.PodIsReady(pod) {
			reason := pod.Status.Reason
			if reason == "" {
				reason = "None"
			}
			podDisplayName := pod.Name
			if clientConfig.AllNamespaces {
				podDisplayName = fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
			}
			fmt.Printf("'%s' is not ready! Reason Provided: %s\n", podDisplayName, reason)
			failedPodConditions := api.FailedPodConditions(pod)
			if len(failedPodConditions) != 0 {
				fmt.Println("\tFailed Pod Conditions:")
				// minwidth, tabwidth, padding, padchar, flags
				w := tabwriter.NewWriter(os.Stdout, 8, 8, 1, '\t', 0)

				format := "\t\t%s\t%s\t%s\n"
				fmt.Fprintf(w, format, "CONDITION", "REASON", "MESSAGE")
				for _, condition := range failedPodConditions {
					fmt.Fprintf(w, format, condition.Type, condition.Reason, condition.Message)
				}

				w.Flush()
			}
			fmt.Println("\n\tPod Events:")

			events, err := client.Events(types.EventSelectors{Field: fmt.Sprintf("involvedObject.name=%s", pod.Name)})
			if err != nil {
				return errors.Wrapf(err, "getting events for pod %s", podDisplayName)
			}

			// minwidth, tabwidth, padding, padchar, flags
			w := tabwriter.NewWriter(os.Stdout, 8, 8, 1, '\t', 0)

			format := "\t\t%s\t%s\t%s\t%s\n"
			fmt.Fprintf(w, format, "LAST SEEN", "TYPE", "REASON", "MESSAGE")
			for _, event := range events {
				fmt.Fprintf(w, format, event.LastTimestamp, event.Type, event.Reason, event.Message)
			}
			w.Flush()
			fmt.Println()

			notReadyContainers := api.NotReadyPodContainerStatus(pod)
			if len(notReadyContainers) != 0 {
				for _, container := range notReadyContainers {
					fmt.Printf("\tContainer '%s' is not ready!\n", container.Name)
					fmt.Println("\t\tContainer Logs:")
					logs, err := client.PodLogs(pod, container.Name)
					if err == nil {
						for _, log := range strings.Split(logs, "\n") {
							fmt.Printf("\t\t\t%s\n", log)
						}
					} else {
						fmt.Printf("\t\t\tErrored Getting Logs: %s\n", err.Error())
					}
				}
			}
		}
	}

	return nil
}
