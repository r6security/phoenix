# Phoenix demo: Falco based on-demand application quarantine

This tutorial shows how to use Phoenix to put a specific pod into quarantine when a terminal is opened into it. For this Phoeinx relies on triggers (SecurityEvents) that are created by the Falco-Backend that translates Falco events towards Phoenix. 

In this tutorial you will learn how to:

- install Phoenix
- install Falco
- configure Falco
- install Falco-Backend to be able translate Falco notifications to SecurityEvents
- configure Phoenix

## Prerequisite:

Since quarantine is enforced via NetworkPolicy resources CNI plugin of the Kubernetes cluster must support it. In case of managed K8s cluster also make sure that network policies are enabled.

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
    kubectl -n falco-backend get pods

## Falco-Backend installation

    kubectl -n falco-backend apply -f deploy/manifests/deploy-falco-backend

## Phoenix configuration:

Before triggering the operator let's install a demo application and check that the network connection of the demo-page is working fine (this will be suspended later, when we define a rule that open a terminal results putting the application pod into quarantine):

### Deploy demo-page application

    kubectl apply -n demo-page -f deploy/manifests/demo-page/demo-page-deployment.yaml

Confirm that a terminal can be opened inside the pod:

    kubectl exec -it -n demo-page deployments/demo-page -c nginx -- sh

We can see that google.com is reachable from the pod:

    curl google.com

Let's exit:

    exit

Now we can activate the MTD configuration to take action in case of terminal opening events:

    kubectl apply -n demo-page -f deploy/manifests/falco-backend-quarantine-demo-amtd.yaml

### Trigger the operator

To trigger a Falco event let's open a terminal into a pod that is considered a security threat according to the current Falco configuration:

    kubectl exec -it -n demo-page deployments/demo-page -c nginx -- sh

Notice that in this case the pod does not terminate the connection, however there will be no access to any external resources, e.g. curl google.com, because of the quarantine. 

Try curl on google.com again (we should see no connection):

    curl google.com

Let's exit from the pod and check the list of the pod:

    exit
