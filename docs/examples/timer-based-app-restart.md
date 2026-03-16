# Phoenix demo: Timer based application restart

This tutorial shows how to use Phoenix to periodically restart specific pods relying on triggers (SecurityEvents) that are created with the Time-based Trigger.

In this tutorial you will learn how to:

- install Phoenix
- install Time-based Trigger
- configure Phoenix 
- configure Time-based Trigger

## Phoenix installation

    kubectl apply -n moving-target-defense -f deploy/manifests/deploy-phoenix

## Time-based Trigger installation

    kubectl apply -n time-based-trigger -f deploy/manifests/deploy-time-based-trigger

## Create a demo application to have something to illustrate the scenario with

    kubectl apply -n demo-page -f deploy/manifests/demo-page/demo-page-deployment.yaml

## Configure Phoenix

    kubectl apply -n demo-page -f deploy/manifests/time-based-trigger-demo-amtd.yaml

## Configure Time-based Trigger

The Time-based Trigger needs **both** annotations on the pod: `time-based-trigger.amtd.r6security.com/enabled` and `time-based-trigger.amtd.r6security.com/schedule`. Set both in one patch (under `spec.template.metadata.annotations`):

```bash
kubectl patch -n demo-page deployments.apps demo-page -p '{"spec":{"template":{"metadata":{"annotations":{"time-based-trigger.amtd.r6security.com/enabled":"true","time-based-trigger.amtd.r6security.com/schedule":"30s"}}}}}'
```

If the trigger doesn't run, verify the pod has both annotations:
`kubectl get pod -n demo-page -l app=demo-page -o jsonpath='{.items[0].metadata.annotations}'`

Watch pods to see the restarts every 30 seconds:

	watch kubectl -n demo-page get pods

## Clean up

    kubectl -n moving-target-defense delete -f deploy/manifests/deploy-phoenix
    kubectl -n time-based-trigger delete -f deploy/manifests/deploy-time-based-trigger
    kubectl delete namespace demo-page
