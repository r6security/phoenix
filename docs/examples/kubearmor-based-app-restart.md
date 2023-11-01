# Phoenix demo: KubeArmor based on-demand application restart

This tutorial shows how to use Phoenix to restart a specific pod when a corresponding KubeArmor alert is created. Since KubeArmor is capable of doing security enforcement on its own, this scenario would like to show an example of how Phoenix can be used to serve as another layer of security measures. In this simple example, KubeArmor will be configured to block any usage of package manager inside the pod, and in response to the alerts that KuberArmor generates, Phoenix immediately restarts the pod (the specific action is configurable). For this, Phoenix relies on triggers (SecurityEvents) that are created by the KubeArmor-integrator which translates KubeArmor alerts towards Phoenix.

In this tutorial you will learn how to:

- install Phoenix
- install KubeArmor
- create an example application and create a KubeArmor policy for it
- install KubeArmor-integrator to be able translate KubeArmor notifications to SecurityEvents
- configure Phoenix

## Phoenix installation

    kubectl -n moving-target-defense apply -f deploy/manifests/deploy-phoenix

## KubeArmor installation

    helm repo add kubearmor https://kubearmor.github.io/charts
    helm repo update kubearmor
    helm upgrade --install kubearmor-operator kubearmor/kubearmor-operator -n kubearmor --create-namespace
    kubectl apply -f https://raw.githubusercontent.com/kubearmor/KubeArmor/main/pkg/KubeArmorOperator/config/samples/sample-config.yml

### Deploy demo-page application and configure KubeArmor to deny execution of package management tools (apt/apt-get) by creating a policy

 Create the test application:

    kubectl -n demo-page apply -f deploy/manifests/demo-page/demo-page-deployment.yaml

Create the policy:

```
cat <<EOF | kubectl -n demo-page apply -f -
apiVersion: security.kubearmor.com/v1
kind: KubeArmorPolicy
metadata:
  name: block-pkg-mgmt-tools-exec
spec:
  selector:
    matchLabels:
      app: demo-page
  process:
    matchPaths:
    - path: /usr/bin/apt
    - path: /usr/bin/apt-get
  action:
    Block
EOF
```

## KubeArmor-integrator installation

    kubectl -n kubearmor-integrator apply -f deploy/manifests/deploy-kubearmor-integrator

## Phoenix configuration:

Before configuring Phoenix let's confirm that KubeArmor already blocks the package manager usage, however, the pod is not deleted yet because Phoenix configuration is not active:

    kubectl exec -it -n demo-page deployments/demo-page -c nginx -- bash -c "apt update && apt install masscan"

It will be denied permission enforced by KubeArmor:

    bash: line 1: /usr/bin/apt: Permission denied
    command terminated with exit code 126

However, no pod restart was carried out, since Phoenix has not been configured yet. 

    kubectl -n demo-page get pods

Let's activate the MTD configuration to delete pod (serving as an additional layer of security measure) after it gets notification about the block event of KubeArmor.

    kubectl -n demo-page apply -f deploy/manifests/kubearmor-integrator-delete-demo-amtd.yaml 

### Trigger Phoenix

To trigger a KubeArmor event let's execute our previous command again that tries to run the package manager:

    kubectl exec -it -n demo-page deployments/demo-page -c nginx -- bash -c "apt update && apt install masscan"

It will be denied permission enforced by KubeArmor:

    bash: line 1: /usr/bin/apt: Permission denied
    command terminated with exit code 126

Also, this time the pod was deleted (which restarted the application) by Phoenix: 

    kubectl -n demo-page get pods
