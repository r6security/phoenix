<p align="center">
  <img alt="Phoenix", src="docs/img/phoenix-logo.png" width="30%" height="30%"></br>
</p>

# Phoenix: Automated Moving Target Defense for Kubernetes

Protect your Kubernetes environments with ease.
Phoenix leverages Automated Moving Target Defense (AMTD) to deliver dynamic, scalable, and intelligent security. It integrates seamlessly with your existing DevOps workflows, providing robust protection against evolving threats without slowing down your pipelines.
> Warning: This project is in active development, consider this before deploying it in a production environment.  All APIs, SDKs, and packages are subject to change.

## Features
🔄 Dynamic Container Refresh

    Automatically rotates containers, nodes, and resources.
    Disrupts attack patterns while ensuring 100% uptime.

📜 Real-Time Policy Adaptation

    Adjusts security policies dynamically using Prometheus telemetry.
    Reduces false positives and eliminates manual configurations.

🔄 Automated Rollbacks

    Restores environments to known-good states after misconfigurations or breaches.

🛠️ Self-Healing Infrastructure

    Node Infrastructure Modules (NIMs) autonomously detect and recover from anomalies.

📈 Seamless Observability

    Full integration with Prometheus and Grafana for actionable insights into AMTD activities.
    
## Documentation

For more details please check the [documentation](docs/README.md).

## Caveats

* The project is in an early stage where the current focus is to be able to provide a proof-of-concept implementation that a wider range of potential users can try out. We are welcome all feedbacks and ideas as we continuously improve the project and introduc new features.

## Help

Phoenix development is coordinated in Discord, feel free to [join](https://discord.gg/zMt663CG).

## License

Copyright 2021-2024 by [R6 Security](https://www.r6security.com), Inc. Some rights reserved.

Server Side Public License - see [LICENSE](/LICENSE) for full text.
