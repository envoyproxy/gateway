+++
title = "发布公告"
description = "Envoy Gateway 发布公告"
linktitle = "发布"

[[cascade]]
type = "docs"
+++

本文档提供了 Envoy Gateway 版本的详细信息。
Envoy Gateway 遵循语义版本控制 [v2.0.0 规范][]进行发布版本控制。
由于 Envoy Gateway 是一个新项目，因此次要版本是唯一被定义的版本。
Envoy Gateway 维护人员在未来的某个日期将建立额外的发布详细信息，例如补丁版本。

## 稳定版本 {#stable-releases}

Envoy Gateway 的稳定版本包括：

* 次要版本 - 从 `main` 分支创建新的版本分支和相应的 Tag。
  次要版本将在发布日期后的 6 个月内受到支持。
  随着项目的成熟，Envoy Gateway 维护人员将重新评估支持时间范围。

次要版本每季度发布一次，并遵循以下时间表。

## 发布管理 {#release-management}

次要版本由指定的 Envoy Gateway 维护人员处理。
该维护者被视为该版本的发布经理。[发布指南][]中描述了创建发布的详细信息。
发布经理负责协调整体发布。这包括确定版本中要解决的问题、与 Envoy Gateway 社区的沟通以及发布版本的机制。

|   季度   |                            发布经理                             |
|:-------:|:--------------------------------------------------------------:|
| 2022 Q4 |    Daneyon Hansen ([danehans](https://github.com/danehans))    |
| 2023 Q1 |    Xunzhuo Liu ([Xunzhuo](https://github.com/Xunzhuo))         |
| 2023 Q2 |    Alice Wasko ([AliceProxy](https://github.com/AliceProxy))   |
| 2023 Q3 |    Arko Dasgupta ([arkodg](https://github.com/arkodg))         |
| 2023 Q4 |    Arko Dasgupta ([arkodg](https://github.com/arkodg))         |
| 2024 Q1 |    Xunzhuo Liu ([Xunzhuo](https://github.com/Xunzhuo))         |

## 发布时间表 {#release-schedule}

为了与 Envoy Proxy [发布时间表][]保持一致，
Envoy Gateway 版本按固定时间表（每个季度的第 22 天）生成，
可接受的延迟最多为 2 周，硬性截止日期为 3 周。

|   版本   |  预期时间    |   实际时间   |     偏差     | 生命周期结束 |
|:-------:|:-----------:|:-----------:|:-----------:|:-----------:|
|  0.2.0  | 2022/10/22  | 2022/10/20  |   -2 天   |  2023/4/20  |
|  0.3.0  | 2023/01/22  | 2023/02/09  |   +17 天  |  2023/08/09 |
|  0.4.0  | 2023/04/22  | 2023/04/24  |   +2 天   |  2023/10/24 |
|  0.5.0  | 2023/07/22  | 2023/08/02  |   +10 天  |  2024/01/02 |
|  0.6.0  | 2023/10/22  | 2023/11/02  |   +10 天  |  2024/05/02 |

[v2.0.0 规范]: https://semver.org/lang/zh-CN/
[发布指南]: ../../contributions/releasing
[发布时间表]: https://github.com/envoyproxy/envoy/blob/main/RELEASES.md#major-release-schedule
