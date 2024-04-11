---
title: 关于 Envoy Gateway
linkTitle: 关于
---

{{% blocks/cover title="关于 Envoy Gateway" height="auto" %}}

Envoy Gateway 是一个开源项目，用于将 Envoy Proxy 作为独立或基于 Kubernetes 的 API 网关进行管理。

Gateway API 资源用于动态配置和管理 Envoy 代理。

请继续阅读以了解更多信息，或访问我们的[文档](/zh/latest/)并开始使用！

{{% /blocks/cover %}}

{{% blocks/section color="black" %}}

## 目标 {#objectives}

---

Envoy Gateway 的目标是通过支持多种 Ingress 和 L7/L4 流量路由用例的 API 来降低 Envoy 的使用障碍，这种富有表现力、可扩展、面向角色的 API 可以吸引更多用户使用 Envoy。
从而吸引更多用户使用 Envoy，并为供应商构建增值产品提供共同基础，而无需对基本交互进行重新设计。

{{% /blocks/section %}}

{{% blocks/section color="dark" %}}

### **富有表现力的 API** {#expressive-api}

Envoy Gateway 项目将公开一个简单且富有表现力的 API，并为许多能力设置推荐的默认值。

该 API 将是 Kubernetes 原生的 [Gateway API](https://gateway-api.sigs.k8s.io)，
加上 Envoy 特定的扩展和扩展点。这种富有表现力且广为人知的 API 将使更多用户（尤其是应用程序开发人员）可以使用 Envoy，
并使 Envoy 成为与其他代理相比更强大的“入门”选择。应用程序开发人员将使用开箱即用的 API，
无需深入了解 Envoy 代理的概念或使用 OSS 包装器。该 API 将使用[用户](#personas)可以理解并熟悉的名词。

核心的全功能 Envoy xDS API 将仍然可供那些需要更多功能的人以及那些在
Envoy Gateway 之上添加功能的人使用，例如商业 API 网关产品。

这种富有表现力的 API 不会由 Envoy Proxy 实现，而是由官方支持的顶部翻译层实现。

---

### **包含的能力** {#batteries-included}

Envoy Gateway 将简化 Envoy 的部署和管理方式，使应用程序开发人员能够专注于提供核心业务价值。

该项目计划包含用户所需的额外基础设施组件，以满足其 Ingress 和 API 网关需求：
它将处理 Envoy 基础设施配置（例如 Kubernetes Service、Deployment 等），
以及可能的相关 Sidecar 服务的基础设施配置。它将包括具有覆盖能力的合理默认值。
它将包括通过 API 条件和 Kubernetes 状态子资源公开状态来改进操作的渠道。

对于任何开发人员来说，使应用程序易于访问是一项繁琐的任务。
同样，基础设施管理员将享受简化的管理模型，无需深入了解解决方案的架构即可操作。

---

### **所有环境** {#all-environments}

Envoy Gateway 将支持在 Kubernetes 原生环境以及非 Kubernetes 部署中运行。

初期，Kubernetes 将受到最大的关注，目标是让 Envoy Gateway 支持以
[Gateway API](https://gateway-api.sigs.k8s.io/) 为 Kubernetes 入口的事实标准。
其他目标包括多集群支持和各种运行时环境。

---

### **可扩展性** {#extensibility}

供应商将有能力提供基于 Envoy Gateway 基础的增值产品。

最终用户仍然可以轻松利用常见的 Envoy 代理扩展点，例如提供身份验证方法和限流的实现。
对于高级用例，用户将能够使用 xDS 的全部功能。

由于通用 API 无法解决所有用例，Envoy Gateway 将提供额外的扩展点以实现灵活性。
因此，Envoy Gateway 将构成供应商提供的托管控制平面解决方案的基础，使供应商能够转向更高的管理层。

{{% /blocks/section %}}
{{% blocks/section color="black" %}}

## 非目标 {#non-objectives}

---

### 蚕食供应商模型 {#cannibalize-vendor-models}

供应商需要有能力推动商业价值，因此目标不是蚕食任何现有的供应商货币化模型，尽管某些供应商可能会受到影响。

---

### 破坏当前 Envoy 使用模式 {#disrupt-current-envoy-usage-patterns}

Envoy Gateway 只是一个附加的便利层，并不意味着破坏任何使用
Envoy Proxy、xDS 或 go-control-plane 的用户使用模式。

{{% /blocks/section %}}

{{% blocks/section color="dark" %}}

## 用户模型 {#personas}

**按优先级顺序**

---

### 应用程序开发人员 {#application-developer}

应用程序开发人员将大部分时间花在开发业务逻辑代码上。
他们需要能够管理对其应用程序的访问。

---

### 基础设施管理员 {#infrastructure-administrators}

基础设施管理员负责基础设施中 API 网关设备的安装、维护和操作，
例如 CRD、角色、服务帐户、证书等。基础设施管理员通过管理 Envoy Gateway 实例来支持应用程序开发人员的需求。

---

{{% /blocks/section %}}

{{% blocks/lead type="row" color="white" %}}

<img src="/img/cncf.svg" alt="CNCF" width="40%">

---
[Envoy Proxy](https://www.envoyproxy.io/) 家族的成员，
旨在显着降低使用 Envoy 作为 **API 网关** 时的使用门槛。
{{% /blocks/lead %}}
