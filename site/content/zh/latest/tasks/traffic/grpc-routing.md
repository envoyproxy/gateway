---
title: "GRPC 路由"
---

[GRPCRoute][] 资源允许用户通过匹配 HTTP/2 流量并将其转发到后端 gRPC 服务器来配置 gRPC 路由。
要了解有关 gRPC 路由的更多信息，请参阅[Gateway API 文档][]。

## 先决条件 {#prerequisites}

按照[快速入门](../quickstart)中的步骤安装 Envoy Gateway 和示例清单。
在继续之前，您应该能够使用 HTTP 查询示例程序后端。

## 安装 {#installation}

安装 gRPC 路由示例资源：

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/grpc-routing.yaml
```

该清单安装 [GatewayClass][]、[Gateway][]、Deployment、Service 和 GRPCRoute 资源。
GatewayClass 是集群范围的资源，表示可以被实例化的一类 Gateway。

**注意：**Envoy Gateway 默认被配置为使用 `controllerName: gateway.envoyproxy.io/gatewayclass-controller` 管理 GatewayClass。

## 验证 {#verification}

检查 GatewayClass 的状态：

```shell
kubectl get gc --selector=example=grpc-routing
```

状态应反映为 `Accepted=True`，表示 Envoy Gateway 正在管理 GatewayClass。

Gateway 代表基础设施的配置。创建 Gateway 时，[Envoy 代理][]基础设施由 Envoy Gateway 预配或配置。
`gatewayClassName` 定义此 Gateway 使用的 GatewayClass 的名称。检查 Gateway 状态：

```shell
kubectl get gateways --selector=example=grpc-routing
```

状态应反映为 `Ready=True`，表示 Envoy 代理基础设施已被配置。
该状态还提供 Gateway 的地址。该地址稍后用于测试与代理后端服务的连接。

检查 GRPCRoute 的状态：

```shell
kubectl get grpcroutes --selector=example=grpc-routing -o yaml
```

GRPCRoute 的状态应显示 `Accepted=True` 和引用示例 Gateway 的 `parentRef`。
`example-route` 匹配 `grpc-example.com` 的任何流量并将其转发到 `yages` 服务。

## 测试配置 {#testing-the-configuration}

在测试到 `yages` 后端的 GRPC 路由之前，请获取 Gateway 的地址。

```shell
export GATEWAY_HOST=$(kubectl get gateway/example-gateway -o jsonpath='{.status.addresses[0].value}')
```

使用 [grpcurl][] 命令测试到 `yages` 后端的 GRPC 路由。

```shell
grpcurl -plaintext -authority=grpc-example.com ${GATEWAY_HOST}:80 yages.Echo/Ping
```

您应该看到以下响应：

```shell
{
  "text": "pong"
}
```

Envoy Gateway 还支持此配置的 [gRPC-Web][] 请求。下面的 `curl` 命令可用于通过 HTTP/2 发送 grpc-Web 请求。
您应该收到与上一个命令相同的响应。

正文 `AAAAAAA=` 中的数据是 Ping RPC 接受的空消息（数据长度为 0）的 Base64 编码表示。

```shell
curl --http2-prior-knowledge -s ${GATEWAY_HOST}:80/yages.Echo/Ping -H 'Host: grpc-example.com'   -H 'Content-Type: application/grpc-web-text'   -H 'Accept: application/grpc-web-text' -XPOST -d'AAAAAAA=' | base64 -d
```

## GRPCRoute 匹配 {#grpcroute-match}

`matches` 字段可用于根据 GRPC 的服务和/或方法名称将路由限制到一组特定的请求。
它支持两种匹配类型：`Exact`（精准）和 `RegularExpression`（正则）。

### 精准 {#exact}

`Exact`（精准）匹配是默认匹配类型。

以下示例显示如何根据 `grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo` 的服务和方法名称来匹配请求，
以及如何在我们的部署中匹配方法名称为 `Ping` 且与 `yages.Echo/Ping` 匹配的所有服务。

{{< tabpane text=true >}}
{{% tab header="通过标准输入应用" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GRPCRoute
metadata:
  name: yages
  labels:
    example: grpc-routing
spec:
  parentRefs:
    - name: example-gateway
  hostnames:
    - "grpc-example.com"
  rules:
    - matches:
      - method:
          method: ServerReflectionInfo
          service: grpc.reflection.v1alpha.ServerReflection
      - method:
          method: Ping
      backendRefs:
        - group: ""
          kind: Service
          name: yages
          port: 9000
          weight: 1
EOF
```

{{% /tab %}}
{{% tab header="通过文件应用" %}}
保存以下资源并将其应用到您的集群：

```yaml
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GRPCRoute
metadata:
  name: yages
  labels:
    example: grpc-routing
spec:
  parentRefs:
    - name: example-gateway
  hostnames:
    - "grpc-example.com"
  rules:
    - matches:
      - method:
          method: ServerReflectionInfo
          service: grpc.reflection.v1alpha.ServerReflection
      - method:
          method: Ping
      backendRefs:
        - group: ""
          kind: Service
          name: yages
          port: 9000
          weight: 1
```

{{% /tab %}}
{{< /tabpane >}}

验证 GRPCRoute 状态：

```shell
kubectl get grpcroutes --selector=example=grpc-routing -o yaml
```

使用 [grpcurl][] 命令测试到 `yages` 后端的 GRPC 路由。

```shell
grpcurl -plaintext -authority=grpc-example.com ${GATEWAY_HOST}:80 yages.Echo/Ping
```

### 正则 {#regularexpression}

以下示例演示如何根据服务和方法名称将请求与匹配类型 `RegularExpression` 进行匹配。
它与模式 `/.*.Echo/Pin.+` 匹配所有服务和方法，该模式与我们部署中的 `yages.Echo/Ping` 匹配。

{{< tabpane text=true >}}
{{% tab header="通过标准输入应用" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GRPCRoute
metadata:
  name: yages
  labels:
    example: grpc-routing
spec:
  parentRefs:
    - name: example-gateway
  hostnames:
    - "grpc-example.com"
  rules:
    - matches:
      - method:
          method: ServerReflectionInfo
          service: grpc.reflection.v1alpha.ServerReflection
      - method:
          method: "Pin.+"
          service: ".*.Echo"
          type: RegularExpression
      backendRefs:
        - group: ""
          kind: Service
          name: yages
          port: 9000
          weight: 1
EOF
```

{{% /tab %}}
{{% tab header="通过文件应用" %}}
保存以下资源并将其应用到您的集群：

```yaml
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: GRPCRoute
metadata:
  name: yages
  labels:
    example: grpc-routing
spec:
  parentRefs:
    - name: example-gateway
  hostnames:
    - "grpc-example.com"
  rules:
    - matches:
      - method:
          method: ServerReflectionInfo
          service: grpc.reflection.v1alpha.ServerReflection
      - method:
          method: "Pin.+"
          service: ".*.Echo"
          type: RegularExpression
      backendRefs:
        - group: ""
          kind: Service
          name: yages
          port: 9000
          weight: 1
```

{{% /tab %}}
{{< /tabpane >}}

检查 GRPCRoute 状态：

```shell
kubectl get grpcroutes --selector=example=grpc-routing -o yaml
```

使用 [grpcurl][] 命令测试到 `yages` 后端的 GRPC 路由。

```shell
grpcurl -plaintext -authority=grpc-example.com ${GATEWAY_HOST}:80 yages.Echo/Ping
```

[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute/
[Gateway API 文档]: https://gateway-api.sigs.k8s.io/
[GatewayClass]: https://gateway-api.sigs.k8s.io/api-types/gatewayclass/
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway/
[Envoy 代理]: https://www.envoyproxy.io/
[grpcurl]: https://github.com/fullstorydev/grpcurl
[gRPC-Web]: https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md#protocol-differences-vs-grpc-over-http2
