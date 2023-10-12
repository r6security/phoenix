## Prerequisites

In itself Phoenix does not have any specific dependency regarding its core installation, [AMTD Operator](/docs/concepts/...) and [Time-based Trigger](/docs/concepts/...), however if you want to integrate it with 3rd party tools, e.g. [Falco](https://falco.org/), [KubeArmor](https://kubearmor.io/), etc. you need to have these tools installed and configured as well.
For example, to provide timely based pod restarts with Phonenix, no 3rd party tool is necessray. However, to provide on-demand pod restarts in reaction to specific security threats that such 3rd party tools can detect - like noticing that someone opened a terminal and modified a file - you need:
* the specific 3rd party tool installed in your environment that is configured to communicate the threats towards a specific Phoenix integration backend (this is basically just setting up a webhook)
* the specific Phoenix integration backend that receives the threats information, translates it to a unified format for Phoenix

By design a specific backend exist for each 3rd party tool.

For more details see the following pages:

* [Falco integration](/docs/concepts/integration/falco)
* [KubeArmor integration (coming soon)](/docs/concepts/integration/kubearmor)

## Installation

You can deploy Phoenix by executing the following commands:

### Deploy Time-based Trigger 
```
 curl "https://github.com/..." | kubectl apply -f -
```

### Deploy Phoenix
```
 curl "https://github.com/..." | kubectl apply -f -
 curl "https://github.com/..." | kubectl apply -f -
```

### Check that all pods are in a running state
```
kubectl -n timer-backend get pods
kubectl -n moving-target-defense get pods
```

## Setup Scheduled restart with Time-based Trigger ("Hello World" example)

Let's start a demo-page application that is restarted by Phoenix on a scheduled basis as a security measure. To do this we use the Time-based Trigger that creates security events that are handled by Phoenix. This setup consists the following steps:

1. Deploying the demo-page application:

```
curl "https://github.com/..." | kubectl apply -f -
kubectl -n demo-page wait --for=condition=ready pod --all
```

2. Deploy an MTD configuration:

```
curl "https://github.com/..." | kubectl apply -f -
kubectl -n moving-target-defense get AdaptiveMovingTargetDefense
```

3. Enable time backend for the demo-page deployment and schdedule the restart in every 30s:

```
kubectl -n demo-page patch ...
kubectl -n demo-page patch ...
```

Watch pods to see the restarts in every 30 seconds:

watch kubectl -n demo-page get pods

## Try it out!

You can try Phoenix out at Killercoda in a self paced tutorial where you can try out the following scenarios:
* Scheduled restart with Time-based Trigger
* On-demand restart with Falco Backend
* On-Demand quarantine with Falco Backend
