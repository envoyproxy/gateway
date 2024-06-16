---
title: "路线图"
weight: -1
description: "本部分记录了 Envoy Gateway 的路线图。"
---

本文档可作为 Envoy Gateway 用户和贡献者了解项目方向的高级参考。

## 为路线图做出贡献 {#contributing-to-the-roadmap}

- 要将功能添加到路线图中，请创建 [Issue][issue] 或加入[社区会议][meeting]来讨论您的用例。
  如果您的功能被接受，维护人员会将您的 Issue 分配到[发布里程碑][milestones]并相应地更新此文档。
- 为了帮助解决现有的路线图项，请对相关 Issue 进行评论或将自己分配到该 Issue。
- 如果路线图项没有 Issue，请创建一个 Issue，将自己分配到该 Issue，并参考此文档。
  维护者将提交 [Pull Request][PR] 以将该功能添加到路线图中。
  **注意：**在实现该功能之前，应在 Issue 或社区会议上讨论该功能。

如果您不知道从哪里开始贡献，需要帮助来减少技术、自动化和文档债务。查找带有 `help wanted` 标签的 Issue 以开始。

## 细节 {#details}

路线图功能和时间表可能会根据反馈、社区贡献等而改变。
如果您依赖特定的路线图项，我们鼓励您参加社区会议讨论细节，或者通过为项目做出贡献来帮助我们提供该功能。

`最后更新时间：2023 年 4 月`

### [v0.2.0][v0.2.0]: 建立坚实的基础

- 完成核心 Envoy Gateway 实现 - [Issue #60][60]。
- 建立初步测试、e2e、集成等 - [Issue #64][64]。
- 建立用户和开发人员项目文档 - [Issue #17][17]。
- 实现 Gateway API 一致性（例如路由、LB、标头转换等）- [Issue #65][65]。
- 设置 CI/CD 流程 - [Issue #63][63]。

### [v0.3.0][v0.3.0]: 通过扩展机制驱动高级功能

- 支持扩展 Gateway API 字段 [Issue #707][707]。
- 支持实验性 Gateway API，例如 TCPRoute [Issue #643][643]、UDPRoute [Issue #641][641] 和 GRPCRoute [Issue #642][642]。
- 制定利用 Gateway API 扩展的指南 [Issue #675][675]。
- 限流 [Issue #670][670]。
- 认证 [Issue #336][336]。

### [v0.4.0][v0.4.0]: 自定义 Envoy Gateway

- 扩展 Envoy Gateway 控制平面 [Issue #20][20]
- 基于 Helm 的 Envoy Gateway 安装 [Issue #650][650]
- 自定义被管理的 Envoy Proxy Kubernetes 资源字段 [Issue #648][648]
- 配置 xDS Bootstrap [Issue #31][31]

### [v0.5.0][v0.5.0]: 可观察性和扩缩容

- 数据平面的可观察性 [Issue #699][699]。
- 允许用户配置 xDS 资源 [Issue #24][24]。

### [v0.6.0][v0.6.0]: 为 GA 做准备

- 控制平面的可观察性 [Issue #700][700]。
- 计算并记录 Envoy Gateway 性能 [Issue #1365][1365]。
- 添加 TrafficPolicy API 以实现高级功能 [Issue #1492][1492]。
- Envoy Gateway 满足就绪标准 [Issue #1160][1160]。

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
[699]: https://github.com/envoyproxy/gateway/issues/699
[700]: https://github.com/envoyproxy/gateway/issues/700
[707]: https://github.com/envoyproxy/gateway/issues/707
[1160]: https://github.com/envoyproxy/gateway/issues/1160
[1365]: https://github.com/envoyproxy/gateway/issues/1365
[1492]: https://github.com/envoyproxy/gateway/issues/1492
