---
title: "JWT 身份验证"
---

此任务提供有关配置 [JSON Web Token（JWT）][jwt]身份验证的说明。
JWT 身份验证在将请求路由到后端服务之前检查传入请求是否具有有效的 JWT。
目前，Envoy Gateway 仅支持通过 HTTP 标头验证 JWT，例如 `Authorization: Bearer <token>`。

Envoy Gateway 引入了一个名为 [SecurityPolicy][SecurityPolicy] 的新 CRD，允许用户配置 JWT 身份验证。
该实例化资源可以链接到 [Gateway][Gateway]、[HTTPRoute][HTTPRoute] 或 [GRPCRoute][GRPCRoute] 资源。

## 先决条件 {#prerequisites}

按照[快速入门](../quickstart)中的步骤安装 Envoy Gateway 和示例清单。
对于 GRPC - 请按照 [GRPC 路由](../traffic/grpc-routing)示例中的步骤操作。
在继续之前，您应该能够使用 HTTP 或 GRPC 查询示例程序后端。

## 配置 {#configuration}

通过创建 [SecurityPolicy][SecurityPolicy] 并将其附加到示例 HTTPRoute 或 GRPCRoute，允许使用具有有效 JWT 的请求。

### HTTPRoute {#httproute}

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/jwt/jwt.yaml
```

已创建两个 HTTPRoute，一个用于 `/foo`，另一个用于 `/bar`。
已创建 SecurityPolicy 并以 HTTPRoute foo 为目标来验证对 `/foo` 的请求。
HTTPRoute bar 不是 SecurityPolicy 的目标，并且将允许未经身份验证的请求发送到 `/bar`。

验证 HTTPRoute 配置和状态：

```shell
kubectl get httproute/foo -o yaml
kubectl get httproute/bar -o yaml
```

SecurityPolicy 配置为 JWT 身份验证，并使用单个 [JSON Web Key Set（JWKS）][jwks]提供程序来对 JWT 进行身份验证。

验证 SecurityPolicy 配置：

```shell
kubectl get securitypolicy/jwt-example -o yaml
```

### GRPCRoute {#grpcroute}

```shell
kubectl apply -f https://raw.githubusercontent.com/envoyproxy/gateway/latest/examples/kubernetes/jwt/grpc-jwt.yaml
```

已创建 SecurityPolicy 并针对 GRPCRoute yages 来验证 `yages` 服务的所有请求。

验证 GRPCRoute 配置和状态：

```shell
kubectl get grpcroute/yages -o yaml
```

SecurityPolicy 配置为 JWT 身份验证，并使用单个 [JSON Web Key Set（JWKS）][jwks]提供程序来对 JWT 进行身份验证。

验证 SecurityPolicy 配置：

```shell
kubectl get securitypolicy/jwt-example -o yaml
```

## 测试 {#testing}

确保设置了[快速入门](../../quickstart) 中的 `GATEWAY_HOST` 环境变量。如果没有，请按照快速入门说明设置变量。

```shell
echo $GATEWAY_HOST
```

### HTTPRoute {#httproute-1}

验证在没有 JWT 的情况下对 `/foo` 的请求是否被拒绝：

```shell
curl -sS -o /dev/null -H "Host: www.example.com" -w "%{http_code}\n" http://$GATEWAY_HOST/foo
```

应返回一个 `401` HTTP 响应码。

获取用于测试请求身份验证的 JWT：

```shell
TOKEN=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/test.jwt -s) && echo "$TOKEN" | cut -d '.' -f2 - | base64 --decode
```

**注意：**上述命令解码并返回令牌的有效内容。您可以将 `f2` 替换为 `f1` 来查看令牌的标头。

验证是否允许使用有效 JWT 向 `/foo` 发出请求：

```shell
curl -sS -o /dev/null -H "Host: www.example.com" -H "Authorization: Bearer $TOKEN" -w "%{http_code}\n" http://$GATEWAY_HOST/foo
```

应返回一个 `200` HTTP 响应码。

验证是否允许在**没有** JWT 的情况下向 `/bar` 发出请求：

```shell
curl -sS -o /dev/null -H "Host: www.example.com" -w "%{http_code}\n" http://$GATEWAY_HOST/bar
```

### GRPCRoute {#grpcroute-1}

验证是否在没有 JWT 的情况下拒绝对 `yages` 服务的请求：

```shell
grpcurl -plaintext -authority=grpc-example.com ${GATEWAY_HOST}:80 yages.Echo/Ping
```

您应该看到以下响应：

```shell
Error invoking method "yages.Echo/Ping": rpc error: code = Unauthenticated desc = failed to query for service descriptor "yages.Echo": Jwt is missing
```

获取用于测试请求身份验证的 JWT：

```shell
TOKEN=$(curl https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/test.jwt -s) && echo "$TOKEN" | cut -d '.' -f2 - | base64 --decode
```

**注意：**上述命令解码并返回令牌的有效内容。您可以将 `f2` 替换为 `f1` 来查看令牌的标头。

验证是否允许使用有效 JWT 向 `yages` 服务发出请求：

```shell
grpcurl -plaintext -H "authorization: Bearer $TOKEN" -authority=grpc-example.com ${GATEWAY_HOST}:80 yages.Echo/Ping
```

您应该看到以下响应：

```shell
{
  "text": "pong"
}
```

## 清理 {#clean-up}

按照[快速入门](../../quickstart) 中的步骤卸载 Envoy Gateway 和示例清单。

删除 SecurityPolicy：

```shell
kubectl delete securitypolicy/jwt-example
```

## 后续步骤 {#next-steps}

查看[开发者指南](../../../contributions/develop)参与该项目。

[SecurityPolicy]: ../../../contributions/design/security-policy
[jwt]: https://tools.ietf.org/html/rfc7519
[jwks]: https://tools.ietf.org/html/rfc7517
[Gateway]: https://gateway-api.sigs.k8s.io/api-types/gateway
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute
[GRPCRoute]: https://gateway-api.sigs.k8s.io/api-types/grpcroute
