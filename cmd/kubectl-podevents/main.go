package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/alecjacobs5401/kubectl-diagnose/kubernetes"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

const version = "0.1.0"

func main() {
	args := os.Args[1:]

	app := kingpin.New("kubectl podevents", "A kubectl plugin used to view pod events for the given pod selector")
	app.UsageWriter(os.Stdout)
	app.Version(version)
	app.HelpFlag.Short('h')

	podSelectors := kubernetes.PodSelectors{}
	eventSelectors := kubernetes.EventSelectors{}
	clientConfig := kubernetes.ClientConfig{}

	app.Arg("pod", "A pod name to target (Accepts multiple pod names)").StringsVar(&podSelectors.Names)
	app.Flag("selector", "Pod Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)").
		Short('l').StringVar(&podSelectors.Label)
	app.Flag("field-selector", "Pod Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.").
		StringVar(&podSelectors.Field)
	app.Flag("event-selector", "Pod Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)").
		StringVar(&eventSelectors.Label)
	app.Flag("event-field-selector", "Event Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.").
		StringVar(&eventSelectors.Field)

	app.Flag("kubeconfig", "The path to a pre-existing kubeconfig file that you want to have the new cluster authorization merged into. If not provided, the KUBECONFIG environment variable will be checked for a file. If that is empty, $HOME/.kube/config will be used.").
		Envar("KUBECONFIG").StringVar(&clientConfig.ConfigFile)
	app.Flag("namespace", "The Kubernetes namespace to use. Default to the current namespace in your kube config file.").
		Short('n').StringVar(&clientConfig.Namespace)
	app.Flag("context", "The Kubernetes context to use. Defaults to the current context in your kube config file.").
		StringVar(&clientConfig.Context)

	_, err := app.Parse(args)
	if err != nil {
		appContext, _ := app.ParseContext(args)
		app.FatalUsageContext(appContext, err.Error())
	}

	if clientConfig.ConfigFile == "" {
		clientConfig.ConfigFile = kubernetes.ConfigPathFromDirectory(os.Getenv("HOME"))
	}

	if err := podevents(podSelectors, eventSelectors, clientConfig); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func podevents(podSelectors kubernetes.PodSelectors, eventSelectors kubernetes.EventSelectors, clientConfig kubernetes.ClientConfig) error {
	client, err := kubernetes.NewClient(clientConfig)
	if err != nil {
		return errors.Wrap(err, "building client")
	}

	podNames := podSelectors.Names
	if len(podNames) == 0 {
		pods, err := client.Pods(podSelectors)
		if err != nil {
			return errors.Wrap(err, "getting pods")
		}
		for _, pod := range pods {
			podNames = append(podNames, pod.Name)
		}
	}

	for index, pod := range podNames {
		if index != 0 {
			fmt.Println()
		}

		// minwidth, tabwidth, padding, padchar, flags
		w := tabwriter.NewWriter(os.Stdout, 8, 8, 1, '\t', 0)

		format := "%s\t%s\t%s\t%s\n"
		fmt.Fprintf(w, format, "LAST SEEN", "TYPE", "REASON", "MESSAGE")

		dupEventSelectors := eventSelectors
		fieldSelectors := fmt.Sprintf("involvedObject.name=%s", pod)
		if dupEventSelectors.Field != "" {
			fieldSelectors = fmt.Sprintf("%s,%s", fieldSelectors, dupEventSelectors.Field)
		}
		dupEventSelectors.Field = fieldSelectors
		events, err := client.Events(dupEventSelectors)
		if err != nil {
			return errors.Wrapf(err, "getting events for pod '%s'", pod)
		}

		fmt.Printf("Events for %s:\n", pod)
		for _, event := range events {
			fmt.Fprintf(w, format, event.LastTimestamp, event.Type, event.Reason, event.Message)
		}
		w.Flush()
	}

	return nil
}
