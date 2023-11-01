# Phoenix demo: KubeArmor based on-demand application restart

This tutorial shows how to use Phoenix to restart a specific pod when a terminal is opened into it. For this Phoenix relies on triggers (SecurityEvents) that are created by the KubeArmor-integrator that translates KubeArmor events towards Phoenix. 

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

    # create test application 
    kubectl -n demo-page apply -f deploy/manifests/demo-page/demo-page-deployment.yaml

    # create the policy
```
cat <<EOF | kubectl apply -f -
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

Before triggering Phoenix let's confirm that KubeArmor already blocks the package manager usage, however the pod is not deleted yet because Phoenix configuration is not active:

    kubectl exec -it -n demo-page deployments/demo-page -c nginx -- bash -c "apt update && apt install masscan"

    kubectl -n demo-page get pods

Let's activate the MTD configuration to delete pod (serving as an additional layer of security measure) after it gets notification about the block event of KubeArmor.

    kubectl -n demo-page apply -f deploy/manifests/falco-integrator-delete-demo-amtd.yaml 

### Trigger the Phoenix

To trigger a KubeArmor event let's execute our previous command again that tries to run the package manager:

    kubectl exec -it -n demo-page deployments/demo-page -c nginx -- bash -c "apt update && apt install masscan"

    kubectl -n demo-page get pods

Watch pods to see that the original pod was deleted (which restarted the application):

    watch kubectl -n demo-page get pods 
