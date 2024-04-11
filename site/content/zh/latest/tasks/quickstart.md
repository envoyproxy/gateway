---
title: "快速开始"
weight: 1
description: 只需几个简单的步骤即可开始使用 Envoy Gateway。
---

本指南将帮助您通过几个简单的步骤开始使用 Envoy Gateway。

## 前置条件 {#prerequisites}

一个 Kubernetes 集群。

**注意：** 请参考[兼容性表格](../install/matrix)来查看所支持的 Kubernetes 版本。

**注意：** 如果您的 Kubernetes 集群没有负载均衡器实现，我们建议安装一个
`Gateway` 资源有一个与其关联的地址。我们推荐使用 [MetalLB](https://metallb.universe.tf/installation/)。

## 安装 {#installation}

安装 Gateway API CRD 和 Envoy Gateway：

```shell
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace
```

等待 Envoy Gateway 至可用后：

```shell
kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
```

安装 GatewayClass，Gateway，HTTPRoute 和示例应用：

```shell
kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/latest/quickstart.yaml -n default
```

**注意：** [`quickstart.yaml`] 定义了 Envoy Gateway 将侦听其全局可路由 IP 地址上端口
80 上的流量，以便轻松使用浏览器测试 Envoy Gateway。当 Envoy Gateway 看到它的 listener
使用特权端口（<1024），它将在内部映射到非特权端口，因此 Envoy Gateway 不需要额外的特权。
了解此映射很重要，当您调试时您可能需要将其考虑在内。

[`quickstart.yaml`]: https://github.com/envoyproxy/gateway/releases/download/latest/quickstart.yaml

## 测试配置 {#testing-the-configuration}

获取由示例 Gateway 创建的 Envoy 服务的名称：

```shell
export ENVOY_SERVICE=$(kubectl get svc -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

端口转发到 Envoy 服务：

```shell
kubectl -n envoy-gateway-system port-forward service/${ENVOY_SERVICE} 8888:80 &
```

通过 Envoy 代理，使用 curl 测试示例应用：

```shell
curl --verbose --header "Host: www.example.com" http://localhost:8888/get
```

### 外部负载均衡器支持 {#external-loadbalancer-support}

您还可以通过将流量发送到外部 IP 来测试相同的功能。获取外部 IP Envoy 服务，运行：

```shell
export GATEWAY_HOST=$(kubectl get svc/${ENVOY_SERVICE} -n envoy-gateway-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

在某些环境中，负载均衡器可能会使用主机名而不是 IP 地址来公开，如果是这样，将上述命令中的 `ip` 替换为 `hostname` 。

通过 Envoy Proxy，使用 curl 操作示例应用：

```shell
curl --verbose --header "Host: www.example.com" http://$GATEWAY_HOST/get
```

## 清理 {#clean-up}

使用本节（快速开始）的步骤来卸载与 Envoy Gateway 相关的一切。

删除 GatewayClass，Gateway，HTTPRoute 和示例应用：

```shell
kubectl delete -f https://github.com/envoyproxy/gateway/releases/download/latest/quickstart.yaml --ignore-not-found=true
```

删除 Gateway API CRD 和 Envoy Gateway：

```shell
helm uninstall eg -n envoy-gateway-system
```

## 接下来 {#next-steps}

查看[开发者指南](../contributions/develop) 来参与该项目。
