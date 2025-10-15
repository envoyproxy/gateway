---
date: 2025-10-15
title: From Evaluation to Integration: SAP’s Journey with Envoy Gateway
linkTitle: SAP
author: Guy Daich (SAP)
---

## Why Envoy Gateway

Our journey with Envoy Gateway (EG) began in late 2023, when multiple SAP product teams undertook an initiative to modernize and streamline their existing gateway infrastructure, which consists of various 3rd party, open source, and managed solutions running on diverse infrastructure. We decided on Envoy as the data plane and conducted a rigorous control-plane evaluation. Envoy Gateway, then [pre-GA](https://sched.co/1Rj4s) (General Availability), emerged as the strategic choice. Selecting Envoy Gateway was not a trivial decision given the number of mature control planes in the market.

However, the following factors made Envoy Gateway the right strategic choice for us:
* Strong alignment with the Envoy Proxy organization and maintainer community.
* An [open-source-first roadmap](https://gateway.envoyproxy.io/contributions/roadmap/) that matched our requirements and balanced community and vendor needs.
* A leading [implementation of Gateway API](https://gateway-api.sigs.k8s.io/implementations/#envoy-gateway), providing portability and long-term viability.
* Multi-vendor backing committed to a [common layer for Envoy-based gateways](https://blog.envoyproxy.io/introducing-envoy-gateway-ad385cc59532) and announced plans to migrate to this common layer.
* Strong [extensibility](https://gateway.envoyproxy.io/docs/tasks/extensibility/) that accelerates delivery without waiting on upstream changes.

During our evaluation, we also identified capability and readiness gaps. We concluded that successful adoption required us to take a proactive role as significant contributors, committing to a sustained engineering effort to close those gaps in collaboration with the community. This was essential to meeting enterprise timelines and helping the broader ecosystem reach GA with confidence.

## Driving Production Readiness

Our contributions were initially guided by the project’s [GA roadmap](https://github.com/envoyproxy/gateway/issues/2249) and focused on three areas to help the project reach this critical milestone for both our adoption and the community at large:
* Features: Extended client and backend traffic policies to cover [timeouts](https://gateway.envoyproxy.io/docs/api/extension_types/#timeout), [HTTP](https://gateway.envoyproxy.io/docs/api/extension_types/#http1settings) and [TLS](https://gateway.envoyproxy.io/docs/api/extension_types/#clienttlssettings) options, [circuit breakers](https://gateway.envoyproxy.io/docs/tasks/traffic/circuit-breaker/), [retries](https://gateway.envoyproxy.io/docs/tasks/traffic/retry/), [backend mTLS](https://gateway.envoyproxy.io/docs/tasks/security/backend-mtls/), [failover](https://gateway.envoyproxy.io/docs/tasks/traffic/failover/), and more.
* Reliability: Added control-plane leader election, hardened the translation pipeline with robust error handling, and implemented various resilience and upgrade tests.
* Processes: Improved project hygiene through [image](https://github.com/envoyproxy/gateway/pull/3287), [dependency](https://github.com/envoyproxy/gateway/pull/3261), and [license](https://github.com/envoyproxy/gateway/pull/3407) scanning, and collaborated on developing a clear [security policy](https://github.com/envoyproxy/gateway/blob/main/SECURITY.md) and coordinated disclosure process.

After GA, we focused on high-impact capabilities requested by the Envoy Gateway and the Gateway API communities. Using Gateway API extension points, we introduced EG-native capabilities:
* Routing to non-Kubernetes backends (external domains, Unix domain sockets, etc.) using a custom [backend resource](https://gateway.envoyproxy.io/docs/tasks/traffic/backend/), later extended by the community for dynamic proxy use-cases.
* Advanced route actions via an Envoy Gateway [HTTPRouteFilter](https://gateway.envoyproxy.io/docs/api/extension_types/#httproutefilter), starting with [regex rewrites](https://github.com/envoyproxy/gateway/pull/4258) and subsequently expanded by the community to direct responses, credential injection, and more.

For scenarios beyond common API Gateway patterns, we invested in the extensibility of both the data plane and the control plane:
* Data Plane: We co-designed and co-implemented the [Envoy Extension Policy](https://gateway.envoyproxy.io/contributions/design/envoy-extension-policy/) resource with the community, enabling features such as [Ext-Proc](https://gateway.envoyproxy.io/latest/tasks/extensibility/ext-proc/) and [Wasm](https://gateway.envoyproxy.io/latest/tasks/extensibility/wasm/), filter ordering, and per-route extensions.
* Control Plane: We invested in the [Envoy Gateway Extension Server](https://gateway.envoyproxy.io/latest/tasks/extensibility/extension-server/), a programmable XDS mutation path. We added [Custom Policies](https://github.com/envoyproxy/gateway/issues/2975) to pass extension context via the same policy model used by the Gateway API. We strengthened [error handling](https://github.com/envoyproxy/gateway/pull/5540), [resilience](https://github.com/envoyproxy/gateway/issues/5612), and [security](https://github.com/envoyproxy/gateway/pull/5613) while exposing additional [deployment modes](https://github.com/envoyproxy/gateway/pull/3494).

These extensibility options became foundational for vendors aiming to deliver distinctive features, large-scale enterprises seeking additional configuration flexibility, and domain-specific projects such as [Envoy AI Gateway](https://aigateway.envoyproxy.io/). The community continues to invest in these extensibility options, introducing [Lua](https://gateway.envoyproxy.io/latest/tasks/extensibility/lua/) extensions and custom backend extension resources. 


{{% alert title="Major Milestones" color="primary" %}}
![KubeConNA24](/img/eg-kubecon-na-2024-update.png)
From [Envoy Gateway Project Update](https://sched.co/1iW9c) presented at KubeCon NA 2024. SAP contributed to key feature delivery at every milestone, becoming a [leading contributor](https://envoy.devstats.cncf.io/d/4/company-statistics-by-repository-group?var-period=d7&var-metric=activity&var-repogroup_name=envoyproxy%2Fgateway&var-companies=All) to the project.
{{% /alert %}}

## Adopting Envoy Gateway

In parallel, we executed a controlled rollout across diverse environments. We identified and resolved real-world resilience, scale, and performance bottlenecks along the way. The outcomes were substantial: significantly lower CPU and memory consumption, much larger configuration scale support, and markedly faster configuration programming times. Envoy Gateway now runs reliably across hundreds of clusters worldwide on multiple infrastructure providers.

Envoy Gateway is currently used to manage HTTPS traffic routing across multiple cloud products. The project's robust support for data plane and control plane extensibility has enabled SAP to implement organization-specific extensions, such as support for custom authentication, authorization, rate-limiting policies, request modification, and flexible dynamic routing capabilities. Moreover, control plane extensibility has allowed SAP to overcome limitations in the Gateway API and fine-tune low-level Envoy configuration options that are currently not exposed, demonstrating the versatility and power of Envoy Gateway in meeting the unique demands of an enterprise environment.

We are grateful to the contributors, reviewers, maintainers, and committee members who shaped this work. We’re genuinely excited to see our joint efforts in production and to watch the foundation we helped build continuously improve and be extended by end users and adopters. If you’re building on Envoy and value openness and collaboration, we invite you to join the [Envoy Gateway community](https://gateway.envoyproxy.io/).

