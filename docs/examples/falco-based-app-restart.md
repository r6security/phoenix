# Phoenix demo: Falco based on-demand application restart

This tutorial shows how to use Phoenix to restart a specific pod when a terminal is opened into it. For this Phoenix relies on triggers (SecurityEvents) that are created by the Falco-integrator that translates Falco events towards Phoenix. 

In this tutorial you will learn how to:

- install Phoenix
- install Falco
- configure Falco
- install Falco-integrator to be able translate Falco notifications to SecurityEvents
- configure Phoenix

## Phoenix installation

    kubectl -n moving-target-defense apply -f deploy/manifests/deploy-phoenix

## Falco installation

    helm repo add falcosecurity https://falcosecurity.github.io/charts
    helm repo update
    helm install falco falcosecurity/falco --namespace falco --create-namespace

## Configure Falco

Load configuration to Falco that fits for this scenario:

    kubectl delete configmap -n falco falco
    kubectl create configmap -n falco falco --from-file deploy/manifests/config-falco/falco.yaml
    kubectl create configmap -n falco falco-rules --from-file deploy/manifests/config-falco/falco-rules.yaml
    kubectl patch -n falco daemonsets.apps falco --patch-file deploy/manifests/config-falco/falco-patch.yaml
    kubectl delete pods -n falco -l app.kubernetes.io/name=falco
    kubectl -n falco-integrator get pods

## Falco-integrator installation

    kubectl -n falco-integrator apply -f deploy/manifests/deploy-falco-integrator

## Phoenix configuration:

Before triggering the operator let's install a demo application and check that the we can open a terminal into the pod (this will be denied later, when we activite the configuration for the MTD operator to consider such thing a security threat and terminate the pod immediately):

### Deploy demo-page application

    kubectl apply -n demo-page -f deploy/manifests/demo-page/demo-page-deployment.yaml

Confirm that a terminal can be opened inside the pod:

    kubectl exec -it -n demo-page deployments/demo-page -c nginx -- sh

We can see that we are in the terminal, so let's exit from the pod:

    exit

Now we can activate the MTD configuration to take action in case of terminal opening events:

    kubectl apply -n demo-page -f deploy/manifests/falco-integrator-delete-demo-amtd.yaml

### Trigger the operator

To trigger a Falco event let's open a terminal into a pod that is considered a security threat according to the current Falco configuration:

    kubectl exec -it -n demo-page deployments/demo-page -c nginx -- sh

Watch pods to see that the pod where we opened the terminal was deleted (which restarted the application) and that is why the terminal was closed automatically:

    watch kubectl -n demo-page get pods 
