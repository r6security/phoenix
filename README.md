# Phoenix: Automated Moving Target Defense for Kubernetes

Make your infrastructure unpredictable. Phoenix is an Automated Moving Target Defense (AMTD) framework designed to eliminate the static nature of modern cloud environments. By introducing managed entropy, Phoenix prevents attackers and "dwellers" from gaining a foothold, forcing them to hit a target that is constantly shifting.

## Why Phoenix?

Traditional security focuses on persistent defense (walls and patches). Phoenix assumes that given enough time, any static defense can be breached.

- Neutralize Dwell Time: Even if an attacker gains entry, their environment disappears and refreshes before they can pivot or exfiltrate data.

- Proactive vs. Reactive: Don't wait for an alert to patch a CVE; move the workload to a clean, hardened state automatically.

- Infrastructure Unpredictability: Turn your Kubernetes cluster into a "moving target" that is impossible to map or reliably exploit.

For more details please check the [documentation](docs/README.md).

> Warning: This project is in active development, consider this before deploying it in a production environment.  All APIs, SDKs, and packages are subject to change.

## Key Features
🔄 Dynamic Container & Node Rotation

Phoenix automatically cycles containers, nodes, and associated resources at randomized or policy-defined intervals. This disrupts attack patterns and clears unauthorized persistence without interrupting service availability.

📜 Real-Time Policy Adaptation

Leveraging deep integration with Prometheus telemetry, Phoenix dynamically scales its defensive posture. If suspicious activity is detected, rotation frequencies increase and security policies tighten automatically.

🛠️ Self-Healing and Managed Randomness

Phoenix doesn't just "fix" broken things; it proactively "heals" the environment back to a trusted state using Randomized Rotation.

- Chaos as a Shield: By constantly and randomly refreshing pods and nodes, Phoenix wipes out any "dwellers" (attackers hiding in the background).

- Autonomous Drift Correction: If any part of the infrastructure deviates from its cryptographically signed baseline, Phoenix treats it as a compromise and self-heals by instantly replacing the resource.

📉 Automated Rollbacks & State Recovery

Instantly restore environments to "known-good" configurations following a misconfiguration or detected breach, ensuring rapid recovery and minimal blast yields.

📈 Seamless Observability

Full out-of-the-box integration with Grafana provides actionable insights into AMTD activities, showing you exactly how and when your infrastructure is mutating to stay ahead of threats.

## How it Works

Phoenix sits alongside your existing DevOps workflow, acting as a "Mutation Engine" for your cluster. It coordinates with the K8s API to ensure that while resources are being refreshed, traffic is balanced and state is preserved, maintaining 100% uptime.

For more details please check the [documentation](docs/README.md).

### Installation

Prerequisites

    Kubernetes cluster (v1.22+)

    Helm v3+

    Prometheus (for adaptive policy features)

 [install guide](docs/INSTALL.md#deploy-with-helm).

## Maturity and Deployments

* Current Status: Production-Tested Beta

While the Phoenix open-source project is evolving its APIs, the core AMTD engine is currently deployed at 50+ organizations in production environments. It has been battle-tested against real-world threat vectors and high-scale traffic.

    Note: As we continue to move toward a 1.0 release, please monitor the [changelog](changelog.md) for updates to SDKs and CRD schemas.

## Community & Support

- Feedback: We welcome all ideas and bug reports via GitHub Issues.

- Enterprise: For advanced features and dedicated support, visit https://www.r6security.com.

## License

Copyright © 2022-2026 by [R6 Security](https://www.r6security.com), Inc. All rights reserved.

Distributed under the Server Side Public License (SSPL) - see [LICENSE](/LICENSE) for the full text.

Built to make your infrastructure a harder target
