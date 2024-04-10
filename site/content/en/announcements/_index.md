+++
title = "Release Announcement"
description = "Envoy Gateway Release Announcement"
linktitle = "Announcement"

[[cascade]]
type = "docs"
+++

This document provides details for Envoy Gateway releases. Envoy Gateway follows the Semantic Versioning [v2.0.0 spec][]
for release versioning. Since Envoy Gateway is a new project, minor releases are the only defined releases. Envoy
Gateway maintainers will establish additional release details, e.g. patch releases, at a future date.

## Stable Releases

Stable releases of Envoy Gateway include:

* Minor Releases- A new release branch and corresponding tag are created from the `main` branch. A minor release
  is supported for 6 months following the release date. As the project matures, Envoy Gateway maintainers will reassess
  the support timeframe.

Minor releases happen quarterly and follow the schedule below.

## Release Management

Minor releases are handled by a designated Envoy Gateway maintainer. This maintainer is considered the Release Manager
for the release. The details for creating a release are outlined in the [release guide][].  The Release Manager is
responsible for coordinating the overall release. This includes identifying issues to be fixed in the release,
communications with the Envoy Gateway community, and the mechanics of the release.

| Quarter |                        Release Manager                         |
|:-------:|:--------------------------------------------------------------:|
| 2022 Q4 |    Daneyon Hansen ([danehans](https://github.com/danehans))    |
| 2023 Q1 |    Xunzhuo Liu ([Xunzhuo](https://github.com/Xunzhuo))         |
| 2023 Q2 |    Alice Wasko ([AliceProxy](https://github.com/AliceProxy))   |
| 2023 Q3 |    Arko Dasgupta ([arkodg](https://github.com/arkodg))         |
| 2023 Q4 |    Arko Dasgupta ([arkodg](https://github.com/arkodg))         |
| 2024 Q1 |    Xunzhuo Liu ([Xunzhuo](https://github.com/Xunzhuo))         |

## Release Schedule

In order to align with the Envoy Proxy [release schedule][], Envoy Gateway releases are produced on a fixed schedule
(the 22nd day of each quarter), with an acceptable delay of up to 2 weeks, and a hard deadline of 3 weeks.

| Version |  Expected   |   Actual    | Difference  | End of Life |
|:-------:|:-----------:|:-----------:|:-----------:|:-----------:|
|  0.2.0  | 2022/10/22  | 2022/10/20  |   -2 days   |  2023/4/20  |
|  0.3.0  | 2023/01/22  | 2023/02/09  |   +17 days  |  2023/08/09 |
|  0.4.0  | 2023/04/22  | 2023/04/24  |   +2 days   |  2023/10/24 |
|  0.5.0  | 2023/07/22  | 2023/08/02  |   +10 days  |  2024/01/02 |
|  0.6.0  | 2023/10/22  | 2023/11/02  |   +10 days  |  2024/05/02 |

[v2.0.0 spec]: https://semver.org/spec/v2.0.0.html
[release guide]: ../latest/contributions/releasing
[release schedule]: https://github.com/envoyproxy/envoy/blob/main/RELEASES.md#major-release-schedule
