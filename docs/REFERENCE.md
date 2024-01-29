## Custom Resources 

A custom resource is an extension of the Kubernetes API that is not necessarily available in a default Kubernetes installation, however, it can be added at any time by deploying a CustomResourceDefinition. 
The CustomResourceDefinition API resource allows you to define custom resources. Defining a CRD object creates a new custom resource with a name and schema that you specify. 

Here we describe the custome resrouces that Phoenix uses and for which the CRDs are installed during its deployment.

### AdaptiveMovingTargetDefense

Each `AdaptiveMovingTargetDefense` contains a `podSelector` and a `strategy`. Phoenix for that `AdaptiveMovingTargetDefense` continuously scans for Pods that match its selector, and in case of security threats that refer to a watched `Pod`, it checks the `strategy` section to determine how to react. 
For this `strategy` contains a set of `rule`-`action` pairs, where `rule` consists labels that are matched against labels in a `SecurityEvent` that describes the threat. Phoenix executes the `action` for the best matching rule. 

```
  apiVersion: amtd.r6security.com/v1beta1
  kind: AdaptiveMovingTargetDefense
  metadata:
    name: amtd-sample
  spec:
    podSelector:
      app: booking-frontend
    strategy:
      - rule: 
          type: default
        action:
          disable: {}
      - rule: 
          type: test
          threatLevel: warning
          source: TimerBackend
        action:
          delete: {}
      - rule: 
          type: network-attack
        action:
          quarantine: {}
      - rule: 
          type: filesystem-corruption
          threatLevel: medium
          source: falco
        action:
          quarantine: {}
      - rule: 
          threatLevel: medium
        action:
          quarantine: {}
```

Applying this manifest creates a new `AdaptiveMovingTargetDefense` named "amtd-sample" that makes Phoenix to watch pods that have the label "app: booking-frontend". The `strategy` contains five `rule`-`action` pairs. The `rule` with "type: default" is  mandatory, since every `AdaptiveMovingTargetDefense` should have a default action that is executed in case there is no better matching rule. The `action` is configurable for the default rule too, which in the example is "disable".

Below is a quick reference of the most important fields of the AdaptiveMovingTargetDefense spec.

| Field | Type | Description | Required |
| :--- | :---: | :--- | :---: |
| `strategy` | `list` | List of `rule`-`action` pairs containing at least with one item where `rule.type = default`. | Yes |
| `strategy.[*].rule` | `object` | Object with the following keys: `type`, `threatLevel`, `source`, where at least one key is mandatory. | Yes |
| `strategy.[*].action` | `string` | Defines the type of action that is executed in case of matching rule. | Yes |

### SecurityEvent

Each `SecurityEvent` represents a threat for pods that are listed in `targets` field. The threat is characterized by multpile labels under the `rule` field. The `description` field is for providing information for human operators.

A `SecurityEvent` is intended to be created by Integration Backends and not by human operators. However, for testing purposes manual creation can be an option if proper RBAC is configured for that.

```
apiVersion: amtd.r6security.com/v1beta1
kind: SecurityEvent
metadata:
   name: se-sample
spec:
  targets:
    - default/booking-frontend-789f54744c-qsjqb
  rule:
    type: filesystem-corruption
    threatLevel: medium
    source: falco
  description: "Falco: I saw a non-authorazied edit in /etc/shadow."
```

This manifest creates a `SecurityEvent` with the name "se-sample" for the pod "booking-frontend-789f54744c-qsjqb" in the "default" namespace. The threat is characterized as a "filesystem-corruption", that is considered at "medium" level by Falco. The `description` summarizes it as a non-authorazied edit in the file "/etc/shadow".

Below is a quick reference of the most important fields of the SecurityEvent spec.

| Field | Type | Description | Required |
| :--- | :---: | :--- | :---: |
| `targets` | `list` | List of pods in `<namespace>/<pod-name>`. | Yes |
| `rule` | `object` | Object with the following keys: `type`, `threatLevel`, `source`, where at least one key is mandatory. | Yes |
| `description` | `string` |  Description helps describe a SecurityEvent with more details | Yes |

