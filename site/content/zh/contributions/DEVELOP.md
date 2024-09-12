---
title: "开发者指南"
description: "本节介绍如何开发 Envoy Gateway。"
weight: 2
---

Envoy Gateway 使用基于 [make][] 的构建系统进行构建。我们的 CI 使用基于 [Github Actions][] 的[工作流][]建设。

## 先决条件 {#prerequisites}

### go {#go}

* 版本：1.20
* 安装指南：https://go.dev/doc/install

### make {#make}

* 推荐版本：4.0 或更高
* 安装指南：https://www.gnu.org/software/make

### docker {#docker}

* 当您想要构建 Docker 镜像或在 Docker 内运行 `make` 时可选。
* 推荐版本：20.10.16
* 安装指南：https://docs.docker.com/engine/install

### python3 {#python3}

* 需要一个 `python3` 程序
* 必须有一个正常运行的 `venv` 模块；这是标准库的一部分，但某些发行版使用 stub 将其替换（例如 Debian 和 Ubuntu），
  并要求您单独安装 `python3-venv` 包。

## 快速开始 {#quickstart}

* 运行 `make help` 以查看构建、测试和运行 Envoy Gateway 的所有可用目标。

### 构建 {#building}

* 运行 `make build` 来构建所有二进制文件。
* 运行 `make build BINS="envoy-gateway"` 来构建 Envoy Gateway 二进制文件。
* 运行 `make build BINS="egctl"` 来构建 egctl 二进制文件。

**注意：**二进制文件在 `bin/$OS/$ARCH` 目录中生成，例如，`bin/linux/amd64/`。

### 测试 {#testing}

* 运行 `make test` 来运行 golang 测试。

* 运行 `make testdata` 生成标准的 YAML 测试数据文件。

### 运行 Linter {#running-linters}

* 运行 `make lint` 以确保您的代码通过所有 Linter 检查。

**注意：**`golangci-lint` 配置位于[此处](https://github.com/envoyproxy/gateway/blob/main/tools/linter/golangci-lint/.golangci.yml)。

### 构建并推送镜像 {#building-and-pushing-the-image}

* 运行 `IMAGE=docker.io/you/gateway-dev make image` 来构建 Docker 镜像。
* 运行 `IMAGE=docker.io/you/gateway-dev make push-multiarch` 来构建并推送多架构 Docker 镜像。

**注意：**将 `IMAGE` 替换为您的仓库的镜像名称。

### 部署 Envoy Gateway 进行测试/开发 {#deploying-envoy-gateway-for-testdev}

* 运行 `make create-cluster` 创建一个 [Kind][] 集群。

#### 选项 1：使用最新的 [gateway-dev][] 镜像 {#option-1-use-the-latest-gateway-dev-image}

* 运行 `TAG=latest make kube-deploy` 以使用最新镜像在 Kind 集群中部署 Envoy Gateway。
  替换 `latest` 以使用不同的镜像标签。

#### 选项 2：使用自定义镜像 {#option-2-use-a-custom-image}

* 运行 `make kube-install-image` 从当前分支的最新构建镜像并将其加载到 Kind 集群中。
* 运行 `IMAGE_PULL_POLICY=IfNotPresent make kube-deploy` 以使用自定义镜像将 Envoy Gateway 安装到 Kind 集群中。

### 在 Kubernetes 中部署 Envoy Gateway {#deploying-envoy-gateway-in-kubernetes}

* 运行 `TAG=latest make kube-deploy` 以使用最新镜像将 Envoy Gateway 部署到 Kubernetes 集群（链接到当前 kube 上下文）。
  在命令前面加上 `IMAGE` 或替换 `TAG` 以使用不同的 Envoy Gateway 镜像或标签。
* 运行 `make kube-undeploy` 从集群中卸载 Envoy Gateway。

**注意：**Envoy Gateway 针对 Kubernetes v1.24.0 进行了测试。

### 演示设置 {#demo-setup}

* 运行 `make kube-demo` 来部署演示后端服务、GatewayClass、Gateway 和 HTTPRoute 资源
  （类似于[快速入门][]文档中概述的步骤）并测试配置。
* 运行 `make kube-demo-undeploy` 删除 `make kube-demo` 命令创建的资源。

### 运行 Gateway API 一致性测试 {#run-gateway-api-conformance-tests}

通过以下命令将 Envoy Gateway 部署到 Kubernetes 集群并运行 Gateway API 一致性测试。
请参阅 Gateway API [一致性主页][]以了解有关测试的更多信息。如果 Envoy Gateway 已安装，
请运行 `TAG=latest make run-conformance` 来运行一致性测试。

#### 在 Linux 主机上 {#on-a-linux-host}

* 运行 `TAG=latest make conformance` 来创建 Kind 集群，
  使用最新的 [gateway-dev][] 镜像安装 Envoy Gateway，并运行 Gateway API 一致性测试。

#### 在 Mac 主机上 {#on-a-mac-host}

由于 Mac 不支持将 Docker 网络[直接暴露][]到 Mac 主机，因此请使用以下解决方法之一来运行一致性测试：

* 部署您自己的 Kubernetes 集群或使用具有 [Kubernetes 支持][] 的 Docker Desktop，
  然后运行 `TAG=latest make kube-deploy run-conformance`。这将使用最新的 [gateway-dev][] 镜像将 Envoy Gateway
  安装到使用当前 kubectl 上下文的 Kubernetes 集群，并运行一致性测试。使用 `make kube-undeploy` 卸载 Envoy Gateway。
* 安装并运行 [Docker Mac Net Connect][mac_connect]，然后运行 `TAG=latest make conformance`。

**注意：**在命令前加上 `IMAGE` 或替换 `TAG` 以使用不同的 Envoy Gateway 镜像或标签。如果未指定 `TAG`，则使用当前分支的短 SHA。

### 调试 Envoy 配置 {#debugging-the-envoy-config}

查看 Envoy Gateway 正在使用的 Envoy 配置的一种简单方法是将端口转发到与 Gateway
对应的 Envoy 部署上的管理界面端口（当前为 `19000`），以便可以在本地访问它。

获取 Envoy 部署的名称。以下示例适用于 `default` 命名空间中的网关 `eg`：

```shell
export ENVOY_DEPLOYMENT=$(kubectl get deploy -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

通过端口转发管理接口端口：

```shell
kubectl port-forward deploy/${ENVOY_DEPLOYMENT} -n envoy-gateway-system 19000:19000
```

现在，您可以通过导航到 `127.0.0.1:19000/config_dump` 来查看正在运行的 Envoy 配置。

[Envoy 管理接口][]上还有许多其他端点在调试时可能会有所帮助。

### JWT 测试 {#jwt-testing}

[JSON Web Token（JWT）][jwt]和 [JSON Web Key Set（JWKS）][jwks]示例用于[请求身份验证][]任务。
JWT 由 [JWT Debugger][] 使用 `RS256` 算法创建。JWT 验证签名中的公钥已复制到 [JWK Creator][] 以生成 JWK。
JWK Creator 配置了匹配的设置，即 `Signing` 公钥使用和 `RS256` 算法。生成的 JWK 包装在 JWKS 结构中并托管在仓库中。

[快速入门]: https://github.com/envoyproxy/gateway/blob/main/docs/latest/user/quickstart.md
[make]: https://www.gnu.org/software/make/
[Github Actions]: https://docs.github.com/en/actions
[工作流]: https://github.com/envoyproxy/gateway/tree/main/.github/workflows
[Kind]: https://kind.sigs.k8s.io/
[一致性主页]: https://gateway-api.sigs.k8s.io/concepts/conformance/
[直接暴露]: https://kind.sigs.k8s.io/docs/user/loadbalancer/
[Kubernetes 支持]: https://docs.docker.com/desktop/kubernetes/
[gateway-dev]: https://hub.docker.com/r/envoyproxy/gateway-dev/tags
[mac_connect]: https://github.com/chipmk/docker-mac-net-connect
[Envoy 管理接口]: https://www.envoyproxy.io/docs/envoy/latest/operations/admin#operations-admin-interface
[jwt]: https://tools.ietf.org/html/rfc7519
[jwks]: https://tools.ietf.org/html/rfc7517
[请求身份验证]: ../latest/tasks/security/jwt-authentication
[JWT Debugger]: https://jwt.io/
[JWK Creator]: https://russelldavies.github.io/jwk-creator/
