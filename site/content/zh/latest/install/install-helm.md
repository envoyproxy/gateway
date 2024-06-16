+++
title = "使用 Helm 安装"
weight = -100
+++

[Helm](https://helm.sh) 是 Kubernetes 的包管理器，可自动在 Kubernetes 上发布和管理软件。

Envoy Gateway 可以通过 Helm Chart 经过几个简单的步骤进行安装，
具体取决于您是首次部署、从现有安装升级 Envoy Gateway 还是从 Envoy Gateway 迁移。

## 开始之前 {#before-you-begin}

{{% alert title="兼容性矩阵" color="warning" %}}
请参阅[版本兼容性矩阵](./matrix)了解更多信息。
{{% /alert %}}

Envoy Gateway Helm Chart 托管在 DockerHub 中。

它发布在 `oci://docker.io/envoyproxy/gateway-helm`。

{{% alert title="注意" color="primary" %}}
我们使用 `v0.0.0-latest` 作为最新的开发版本。

您可以访问 [Envoy Gateway Helm Chart](https://hub.docker.com/r/envoyproxy/gateway-helm/tags) 了解更多版本。
{{% /alert %}}

## 使用 Helm 安装 {#install-with-helm}

Envoy Gateway 通常从命令行部署到 Kubernetes。如果您没有 Kubernetes，则应该使用 `kind` 来创建一个。

{{% alert title="开发者指南" color="primary" %}}
请参阅[开发者指南](../../contributions/develop)了解更多信息。
{{% /alert %}}

安装 Gateway API CRD 和 Envoy Gateway：

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace
```

等待 Envoy Gateway 变为可用：

```shell
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

安装 GatewayClass、Gateway、HTTPRoute 和示例应用程序：

```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/latest/quickstart.yaml -n default
```

**注意：**[`quickstart.yaml`] 定义 Envoy Gateway 将侦听 80 端口及其全局可路由 IP 地址的流量，
以便轻松使用浏览器测试 Envoy Gateway。当 Envoy Gateway 发现其侦听器正在使用特权端口（<1024）时，
它会在内部将其映射到非特权端口，以便 Envoy Gateway 不需要额外的特权。了解此映射很重要，因为您在调试时可能需要考虑它。

[`quickstart.yaml`]: https://github.com/envoyproxy/gateway/releases/download/latest/quickstart.yaml

## 自定义 Helm Chart {#helm-chart-customizations}

下面是使用 helm install 命令进行 Envoy Gateway 安装的一些快速方法。

### 增加副本数 {#increase-the-replicas}

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace --set deployment.replicas=2
```

### 更改 kubernetesClusterDomain 名称 {#change-the-kubernetesclusterdomain-name}

如果您使用不同的域名安装了集群，则可以使用以下命令。

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace --set kubernetesClusterDomain=<domain name>
```

**注意：**以上是我们可以直接用于自定义安装的一些方法。但如果您正在寻找更复杂的更改，
[values.yaml](https://helm.sh/docs/chart_template_guide/values_files/) 可以帮助您。

### 使用 values.yaml 文件进行复杂安装 {#using-values-yaml-file-for-complex-installation}

```yaml
deployment:
  envoyGateway:
    resources:
      limits:
        cpu: 700m
        memory: 128Mi
      requests:
        cpu: 10m
        memory: 64Mi
  ports:
    - name: grpc
      port: 18005
      targetPort: 18000
    - name: ratelimit
      port: 18006
      targetPort: 18001

config:
  envoyGateway:
    logging:
      level:
        default: debug
```

在这里，我们对 value.yaml 文件进行了三处更改。将 CPU 的资源限制增加到 `700m`，
将 gRPC 的端口更改为 `18005`，将限流端口更改为 `18006`，并将日志记录级别更新为 `debug`。

您可以通过以下命令使用 value.yaml 文件安装 Envoy Gateway。

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace -f values.yaml
```

{{% alert title="Helm Chart Values" color="primary" %}}
如果您想了解 values.yaml 文件中的所有可用字段，请参阅 [Helm Chart Values](./gateway-helm-api)。
{{% /alert %}}

## 开放端口 {#open-ports}

这些是 Envoy Gateway 和托管 Envoy 代理使用的端口。

### Envoy Gateway {#envoy-gateway}

| Envoy Gateway          |    地址    |  端口  |    是否可配置    |
|:----------------------:|:---------:|:------:|    :------:    |
| Xds EnvoyProxy Server  | 0.0.0.0   | 18000  |       No       |
| Xds RateLimit Server   | 0.0.0.0   | 18001  |       No       |
| Admin Server           | 127.0.0.1 | 19000  |       Yes      |
| Metrics Server         |  0.0.0.0  | 19001  |       No       |
| Health Check           | 127.0.0.1 |  8081  |       No       |

### EnvoyProxy {#envoyproxy}

| Envoy Proxy                       | 地址        | 端口    |
|:---------------------------------:|:-----------:| :-----: |
| Admin Server                      | 127.0.0.1   | 19000   |
| Heath Check                       | 0.0.0.0     | 19001   |

{{% alert title="后续步骤" color="warning" %}}
Envoy Gateway 现在应该已成功安装并运行。要体验 Envoy Gateway 的更多功能，请参阅[任务](../tasks)。
{{% /alert %}}
