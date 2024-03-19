---
title: Envoy Gateway
---

{{< blocks/cover title="欢迎访问 Envoy Gateway！" image_anchor="top" height="full" >}}
<a class="btn btn-lg btn-primary me-3 mb-4" href="/v0.6.0">
  开始使用 <i class="fas fa-arrow-alt-circle-right ms-2"></i>
</a>
<a class="btn btn-lg btn-secondary me-3 mb-4" href="/v0.6.0/contributions">
  参与贡献 <i class="fa fa-heartbeat ms-2 "></i>
</a>
<p class="lead mt-5">将 Envoy 代理作为独立或基于 Kubernetes 的 API 网关进行管理</p>
{{< blocks/link-down color="white" >}}
{{< /blocks/cover >}}

{{% blocks/lead color="black" %}}
将 **Envoy 代理**作为**独立**或**基于 Kubernetes 的** API 网关进行管理。

**Gateway API** 用于**动态**提供和配置托管 Envoy 代理。
{{% /blocks/lead %}}

{{% blocks/section type="row" color="dark" %}}

{{% blocks/feature icon="fa fa-commenting" title="Expressive API" %}}
Based on Gateway API, with reasonable default settings to simplify the Envoy user experience, without knowing details of Envoy proxy.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa fa-battery-full" title="Batteries included" %}}
Automatically Envoy infrastructure provisioning and management.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa fa-tree" title="All environments" %}}
Support for heterogeneous environments. Initially, Kubernetes will receive the most focus.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa fa-cubes" title="Extensibility" %}}
Vendors will have the ability to provide value-added products built on the Envoy Gateway foundation.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa fa-lock" title="Security"%}}
Supports a variety of Security features, such as TLS, TLS pass-through, secure gRPC, authentication. rate-limiting, etc.
{{% /blocks/feature %}}

{{% blocks/feature icon="fa fa-bolt" title="High Performance"%}}
Built on top of the high-performance Envoy proxy, which can handle millions of requests per second.
{{% /blocks/feature %}}

{{% /blocks/section %}}

{{% blocks/lead color="black" %}}
Lower barriers to adoption through **Expressive, Extensible, Role-oriented APIs**

Support a multitude of **ingress** and **L7/L4** traffic routing

Common foundation for vendors to build **value-added** products

Without having to **re-engineer**
fundamental interactions.

{{% /blocks/lead %}}

{{% blocks/section type="row" color="dark" %}}

{{% blocks/feature icon="fab fa-app-store-ios" title="Download **from Github**" url="https://github.com/envoyproxy/gateway/releases" %}}
Try Envoy Gateway in GitHub Releases
{{% /blocks/feature %}}

{{% blocks/feature icon="fab fa-github" title="Contributions Welcome!"
    url="/latest/contributions/" %}}
We do a [Pull Request](https://github.com/envoyproxy/gateway/pulls)
contributions workflow on **GitHub**.
{{% /blocks/feature %}}

{{% blocks/feature icon="fab fa-slack" title="Contact us on Slack!"
    url="https://envoyproxy.slack.com/archives/C03E6NHLESV" %}}
For announcement of latest features etc.
{{% /blocks/feature %}}

{{% /blocks/section %}}

{{% blocks/lead type="row" color="white" %}}

<img src="/img/cncf.svg" alt="CNCF" width="40%">

---
Member of the [Envoy Proxy](https://www.envoyproxy.io/) family aimed at significantly **decreasing the barrier** to entry when using Envoy for **API Gateway**.
{{% /blocks/lead %}}
