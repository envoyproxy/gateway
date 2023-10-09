---
title: "Roadmap"
weight: -1
description: "This section records the roadmap of Envoy Gateway."
---

This document serves as a high-level reference for Envoy Gateway users and contributors to understand the direction of
the project.

## Contributing to the Roadmap

- To add a feature to the roadmap, create an [issue][issue] or join a [community meeting][meeting] to discuss your use
  case. If your feature is accepted, a maintainer will assign your issue to a [release milestone][milestones] and update
  this document accordingly.
- To help with an existing roadmap item, comment on or assign yourself to the associated issue.
- If a roadmap item doesn't have an issue, create one, assign yourself to the issue, and reference this document. A
  maintainer will submit a [pull request][PR] to add the feature to the roadmap. __Note:__ The feature should be
  discussed in an issue or a community meeting before implementing it.

If you don't know where to start contributing, help is needed to reduce technical, automation, and documentation debt.
Look for issues with the `help wanted` label to get started.

## Details

Roadmap features and timelines may change based on feedback, community contributions, etc. If you depend on a specific
roadmap item, you're encouraged to attend a community meeting to discuss the details, or help us deliver the feature by
contributing to the project.

`Last Updated: April 2023`

### [v0.2.0][v0.2.0]: Establish a Solid Foundation

- Complete the core Envoy Gateway implementation- [Issue #60][60].
- Establish initial testing, e2e, integration, etc- [Issue #64][64].
- Establish user and developer project documentation- [Issue #17][17].
- Achieve Gateway API conformance (e.g. routing, LB, Header transformation, etc.)- [Issue #65][65].
- Setup a CI/CD pipeline- [Issue #63][63].

### [v0.3.0][v0.3.0]: Drive Advanced Features through Extension Mechanisms

- Support extended Gateway API fields [Issue #707][707].
- Support experimental Gateway APIs such as TCPRoute [Issue #643][643], UDPRoute [Issue #641][641] and GRPCRoute [Issue #642][642].
- Establish guidelines for leveragaing Gateway API extensions [Issue #675][675].
- Rate Limiting [Issue #670][670].
- Authentication [Issue #336][336].

### [v0.4.0][v0.4.0]: Customizing Envoy Gateway

- Extending Envoy Gateway control plane [Issue #20][20]
- Helm based installation for Envoy Gateway [Issue #650][650]
- Customizing managed Envoy Proxy Kubernetes resource fields [Issue #648][648] 
- Configuring xDS Bootstrap [Issue #31][31]

### [v0.5.0][v0.5.0]: Observability and Scale

- Observability for control plane and data plane [Issue #701][701]. 
- Compute and document Envoy Gateway performance [Issue #1365][1365].
- Allow users to configure xDS Resources [Issue #24][24].

### [v0.6.0][v0.6.0]: Preparation for GA

- Envoy Gateway meets readiness criteria [Issue #1160][1160]. 

[issue]: https://github.com/envoyproxy/gateway/issues
[meeting]: https://docs.google.com/document/d/1leqwsHX8N-XxNEyTflYjRur462ukFxd19Rnk3Uzy55I/edit?usp=sharing
[pr]: https://github.com/envoyproxy/gateway/compare
[milestones]: https://github.com/envoyproxy/gateway/milestones
[v0.2.0]: https://github.com/envoyproxy/gateway/milestone/1
[v0.3.0]: https://github.com/envoyproxy/gateway/milestone/7
[v0.4.0]: https://github.com/envoyproxy/gateway/milestone/12
[v0.5.0]: https://github.com/envoyproxy/gateway/milestone/13
[v0.6.0]: https://github.com/envoyproxy/gateway/milestone/15
[17]: https://github.com/envoyproxy/gateway/issues/17
[20]: https://github.com/envoyproxy/gateway/issues/20
[24]: https://github.com/envoyproxy/gateway/issues/24
[31]: https://github.com/envoyproxy/gateway/issues/31
[60]: https://github.com/envoyproxy/gateway/issues/60
[63]: https://github.com/envoyproxy/gateway/issues/63
[64]: https://github.com/envoyproxy/gateway/issues/64
[65]: https://github.com/envoyproxy/gateway/issues/65
[336]: https://github.com/envoyproxy/gateway/issues/336
[641]: https://github.com/envoyproxy/gateway/issues/641
[642]: https://github.com/envoyproxy/gateway/issues/642
[648]: https://github.com/envoyproxy/gateway/issues/648
[650]: https://github.com/envoyproxy/gateway/issues/650
[643]: https://github.com/envoyproxy/gateway/issues/643
[670]: https://github.com/envoyproxy/gateway/issues/670
[675]: https://github.com/envoyproxy/gateway/issues/675
[701]: https://github.com/envoyproxy/gateway/issues/701
[707]: https://github.com/envoyproxy/gateway/issues/707
[1160]: https://github.com/envoyproxy/gateway/issues/1160
[1365]: https://github.com/envoyproxy/gateway/issues/1365
