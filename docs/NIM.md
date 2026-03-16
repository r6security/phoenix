# NIM (Node Infrastructure Module) add-on

Phoenix includes an optional **NIM controller** that performs NIM-enhanced pod restarts when pods have **time-based** SecurityEvents applied (e.g. from a timer or the [Time-based Trigger](https://github.com/r6security/time-based-trigger)). The controller updates NIM annotations on the pod and then deletes the pod with a configurable grace period so the workload is rescheduled.

## When NIM acts

- The pod must have the annotation `amtd.r6security.com/applied-sec-events` set (JSON array of SecurityEvent objects), typically by the Phoenix operator or the Time-based Trigger.
- NIM only reacts when at least one of those events has `rule.type=timed` and `rule.source=TimeBasedTrigger`.
- Other event types are ignored.

## Annotation contract

| Annotation | Set by | Description |
|------------|--------|-------------|
| `amtd.r6security.com/applied-sec-events` | Phoenix / Time-based Trigger | JSON array of SecurityEvent objects applied to the pod. |
| `nim.r6security.com/pod-startup-state` | NIM controller | One of `pending`, `starting`, `running`, `failed`. |
| `nim.r6security.com/triggers-stopped` | NIM controller | Set to `true` when NIM has applied a restart. |
| `nim.r6security.com/action-active` | NIM controller | Set to `true` when a NIM action is in progress. |
| `nim.r6security.com/last-action-timestamp` | NIM controller | Unix timestamp of the last NIM action. |

## Default behavior

- Grace period for pod termination: 30 seconds.
- Requeue interval: 10 seconds.
- Trigger: `type=timed`, `source=TimeBasedTrigger`.

These can be made configurable in a future release.
