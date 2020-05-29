# kubectl-diagnose
[![GitHub Release](https://img.shields.io/github/release/alecjacobs5401/kubectl-diagnose.svg?logo=github&style=flat-square)](https://github.com/alecjacobs5401/kubectl-diagnose/releases/latest)

[kubectl plugins](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) for diagnosing Pods.

The idea for these plugins has been shamelessly stolen from [Melanie Cebula's QCOn London Talk](https://www.infoq.com/presentations/airbnb-kubernetes-services/) and enhanced to be _slightly_ less opinionated.

The `kubectl-diagnose` plugin helps find and debug Kubernetes Pods that are "Not Ready" (that have failing Pod Conditions or Containers)

The `kubectl-podevents` plugin displays all pod events pods in your currently configured namespace.

Both plugins take standard Pod selection arguments as well as one or multiple pod names to explicitly diagnose/grab events for.

## Installation

Installation steps will install both `kubectl-diagnose` and `kubectl-podevents`

### Mac / Linux
```
curl -sL https://raw.githubusercontent.com/alecjacobs5401/kubectl-diagnose/master/install.sh | sh -s
```

### Windows

Download Windows binaries from [releases](https://github.com/alecjacobs5401/kubectl-diagnose/releases)

## Upgrading
See installation.

## Usage

### `kubectl diagnose`
```
usage: kubectl diagnose [<flags>] [<pod>...]

A kubectl plugin used to find Pods that are in a 'NotReady' state and display debugging information about them

Flags:
  -h, --help                   Show context-sensitive help (also try --help-long and --help-man).
      --version                Show application version.
  -l, --selector=SELECTOR      Pod Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
      --field-selector=FIELD-SELECTOR
                               Pod Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.
      --kubeconfig=KUBECONFIG  The path to a pre-existing kubeconfig file that you want to have the new cluster authorization merged into. If not provided, the KUBECONFIG environment variable will be checked for a file. If that is empty, $HOME/.kube/config will be used.
  -n, --namespace=NAMESPACE    The Kubernetes namespace to use. Default to the current namespace in your kube config file.
      --context=CONTEXT        The Kubernetes context to use. Defaults to the current context in your kube config file.

Args:
  [<pod>]  A pod name to target (Accepts multiple pod names)
```

### `kubectl podevents`
```
usage: kubectl podevents [<flags>] [<pod>...]

A kubectl plugin used to view pod events for the given pod selector

Flags:
  -h, --help                   Show context-sensitive help (also try --help-long and --help-man).
      --version                Show application version.
  -l, --selector=SELECTOR      Pod Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
      --field-selector=FIELD-SELECTOR
                               Pod Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.
      --event-selector=EVENT-SELECTOR
                               Pod Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
      --event-field-selector=EVENT-FIELD-SELECTOR
                               Event Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.
      --kubeconfig=KUBECONFIG  The path to a pre-existing kubeconfig file that you want to have the new cluster authorization merged into. If not provided, the KUBECONFIG environment variable will be checked for a file. If that is empty, $HOME/.kube/config will be used.
  -n, --namespace=NAMESPACE    The Kubernetes namespace to use. Default to the current namespace in your kube config file.
      --context=CONTEXT        The Kubernetes context to use. Defaults to the current context in your kube config file.

Args:
  [<pod>]  A pod name to target (Accepts multiple pod names)
```

## Example Usages
### Without Arguments / Flags
```
$ kubectl diagnose
'bad-pod-764ccf854d-kbsq2' is not ready! Reason Provided: None
	Failed Pod Conditions:
		CONDITION	    REASON			    MESSAGE
		Ready		    ContainersNotReady	containers with unready status: [bad-container]
		ContainersReady	ContainersNotReady	containers with unready status: [bad-container]

	Pod Events:
		LAST SEEN			            TYPE	REASON			    MESSAGE
		2020-05-29 14:15:32 -0700 PDT	Warning	DNSConfigForming	Search Line limits were exceeded, some search paths have been omitted, the applied search line is: ajacobs-playground.svc.cluster.local svc.cluster.local cluster.local ec2.internal us-east-1.invocadev.com invocadev.com

	Container 'bad-container' is not ready!
		Container Logs:
			-e:1:in `<main>': ah (RuntimeError)
			yo
```
```
$ kubectl podevents
Events for bad-pod-764ccf854d-kbsq2:
LAST SEEN			            TYPE	REASON		    	MESSAGE
2020-05-29 14:15:32 -0700 PDT	Warning	DNSConfigForming	Search Line limits were exceeded, some search paths have been omitted, the applied search line is: ajacobs-playground.svc.cluster.local svc.cluster.local cluster.local ec2.internal us-east-1.invocadev.com invocadev.com

Events for redis-59b66855fc-94sn9:
LAST SEEN			TYPE	REASON		    	MESSAGE
2020-05-29 14:17:33 -0700 PDT	Warning	DNSConfigForming	Search Line limits were exceeded, some search paths have been omitted, the applied search line is: ajacobs-playground.svc.cluster.local svc.cluster.local cluster.local ec2.internal us-east-1.invocadev.com invocadev.com
```


### Diagnosing based on Pod Names
```
kubectl diagnose pod-abc-123 pod-def-456
kubectl podevents pod-abc-123 pod-def-456
```

### With event-field-selector
```
# filters out events with a reason of DNSConfigForming
kubectl podevents --event-field-selector reason!=DNSConfigForming
```
