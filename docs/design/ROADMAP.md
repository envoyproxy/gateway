## Introduction

This document serves as a high-level reference for Envoy Gateway users and contributors to understand the direction of
the project.

## Contributing to the Roadmap

- To propose a roadmap item, create an [issue][issue] or join a [community meeting][meeting] to discuss your desired
  feature. Update this document if your requested feature is accepted for a particular [milestone][milestones].
- To help with an existing roadmap item, comment on the associated issue. If an item doesn't have an issue, create one
  and open a [pull request][PR] linking the issue to an exiting roadmap item.
- Submit a [pull request][PR] to add a feature to the roadmap. __Note:__ The feature should be discussed in an issue or
  a community meeting before implementing the feature.

If you don't know where to start contributing, help is needed to reduce technical, automation, and documentation debt.
Look for issues with the `help wanted` label to get started.

## Roadmap
Roadmap features and timelines may change based on feedback, community contributions, etc. If you depend on a specific
roadmap item, you're encouraged to attend a community meeting to discuss the details, or help us deliver the feature by
contributing to the project.

`Last Updated: October 2022`

### [v0.2.0][v0.2.0]: Establish a Solid Foundation
- Complete the core Envoy Gateway implementation- [Issue #60][60].
- Establish initial testing, e2e, integration, etc- [Issue #64][64].
- Establish user and developer project documentation- [Issue #17][17].
- Achieve Gateway API conformance (e.g. routing, LB, Header transformation, etc.)- [Issue #65][65].
- Setup a CI/CD pipeline- [Issue #63][63].

### [v0.3.0][v0.3.0]: Drive Advanced Features through Extension Mechanisms
- Global Rate Limiting
- AuthN/AuthZ- [Issue #336][336].
- Lets Encrypt Integration

### [v0.4.0][v0.4.0]: Manageability and Scale
- Tooling for devs/infra admins to aid in managing/maintaining EG
- Support advanced provisioning use cases (e.g. multi-cluster, serverless, etc.)
- Perf testing (EG specifically)
- Support for Chaos engineering?

[eg_board]: https://github.com/orgs/envoyproxy/projects/1/views/1?layout=board
[issue]: https://github.com/envoyproxy/gateway/issues
[meeting]: https://docs.google.com/document/d/1leqwsHX8N-XxNEyTflYjRur462ukFxd19Rnk3Uzy55I/edit?usp=sharing
[pr]: https://github.com/envoyproxy/gateway/compare
[milestones]: https://github.com/envoyproxy/gateway/milestones
[v0.2.0]: https://github.com/envoyproxy/gateway/milestone/1
[v0.3.0]: https://github.com/envoyproxy/gateway/milestone/9
[v0.4.0]: https://github.com/envoyproxy/gateway/milestone/10
[60]: https://github.com/envoyproxy/gateway/issues/60
[64]: https://github.com/envoyproxy/gateway/issues/64
[17]: https://github.com/envoyproxy/gateway/issues/17
[65]: https://github.com/envoyproxy/gateway/issues/65
[63]: https://github.com/envoyproxy/gateway/issues/63
[336]: https://github.com/envoyproxy/gateway/issues/336
