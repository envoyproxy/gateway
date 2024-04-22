---
title: 开发者指南
description: 本节告知开发者如何开发 Envoy Gateway。
weight: 2
---

Envoy Gateway 使用基于 [make][] 的构建系统构建。我们的 CI 基于 [Github Actions][]，使用 [workflows][]。

## 前置条件 {#prerequisites}

### go {#go}

* 版本：1.22
* 安装指南：https://go.dev/doc/install

### make {#make}

* 推荐版本：4.0 or later
* 安装指南：https://www.gnu.org/software/make

### docker {#docker}

* 如果您想构建一个 Docker 镜像或者在 Docker 内部运行 `make`
* 推荐版本：20.10.16
* 安装指南：https://docs.docker.com/engine/install

### python3 {#python3}

* 需要 `python3` 程序
* 一个可正常执行的 `venv` module 是必要的；这是标准的一部分，
  但某些发行版（例如 Debian 和 Ubuntu）使用 stub 来替代它，并要求您单独安装 `python3-venv` 包。

## 快速开始 {#quickstart}

* 运行 `make help` 来查看所有可构建，可测试，可运行的的目标，并运行 Envoy Gateway。

### 构建 {#building}

* 运行 `make build` 来构建所有的二进制文件。
* 运行 `make build BINS="envoy-gateway"` 来构建 Envoy Gateway 库。
* 运行 `make build BINS="egctl"` 来构建 egctl 库。

**注意：** 上述二进制文件会在 `bin/$OS/$ARCH` 目录下生成，例如, `bin/linux/amd64/`。

### 测试 {#testing}

* 运行 `make test` 来运行 golang 测试。

* 运行 `make testdata` 来生成 golden YAML 测试数据文件。

### 运行代码检查器（Linters） {#running-linters}

* 运行 `make lint` 来确保您的代码可以通过所有的代码检查工具检查。
  **注意：**`golangci-lint` 在[这里](https://github.com/envoyproxy/gateway/blob/main/tools/linter/golangci-lint/.golangci.yml)。

### 构建和推送镜像 {#building-and-pushing-the-image}

* 运行 `IMAGE=docker.io/you/gateway-dev make image` 来构建 Docker 镜像。
* 运行 `IMAGE=docker.io/you/gateway-dev make push-multiarch` 来构建和推送支持多架构的 Docker 镜像。

**注意：** 使用您注册的镜像名称来替代 `IMAGE`。

### 为测试或开发部署 Envoy Gateway {#deploying-envoy-gateway-for-test-dev}

* 运行 `make create-cluster` 来创建一个 [Kind][] 集群。

#### 可选 1：使用最新的 [gateway-dev][] 镜像 {#use-the-latest-gateway-dev-image}

* 运行 `TAG=latest make kube-deploy` 来使用最新的镜像在 Kind 集群中部署 Envoy Gateway。
  替换 `latest` 来使用不同的镜像标签。

#### 可选 2：使用定制的镜像 {#use-a-custom-image}

* 运行 `make kube-install-image` 来从当前分支来构建一个镜像，然后将镜像载入 Kind 集群中。
* 运行 `IMAGE_PULL_POLICY=IfNotPresent make kube-deploy` 来使用定制化镜像将 Envoy Gateway 安装到 Kind 集群中。

### 在 Kubernetes 中部署 Envoy Gateway {#deploying-envoy-gateway-inkubernetes}

* 运行 `TAG=latest make kube-deploy` 使用最新镜像将 Envoy Gateway 部署到 Kubernetes 集群中（当前 kube 上下文指向的集群）。
  在命令前面加上 `IMAGE` 或替换 `TAG` 以使用不同的 Envoy Gateway 镜像或标签。
* 运行 `make kube-undeploy` 在集群中卸载 Envoy Gateway。

**注意：** Envoy Gateway 针对 Kubernetes v1.24.0 进行了测试。

### 创建示例 {#demo-setup}

* 运行 `make kube-demo` 来部署一个示例后端服务，
  GatewayClass，Gateway 和 HTTPRoute 资源（类似于[快速开始][]文档中概述的步骤）并且测试配置。
* 运行 `make kube-demo-undeploy` 来删除由 `make kube-demo` 命令创建的资源。

### 运行 Gateway API 一致性测试 {#run-gateway-api-conformance-tests}

以下命令将 Envoy Gateway 部署到 Kubernetes 集群并运行 Gateway API 一致性测试。
请参阅 Gateway API [一致性主页][]了解有关测试的更多信息。如果 Envoy Gateway 已安装，
请运行 `TAG=latest make run-conformance` 运行一致性测试。

#### 在 Linux 主机中 {#on-a-linux-host}

* 运行 `TAG=latest make conformance` 来创建一个 Kind 集群, 使用最新的 [gateway-dev][] 镜像安装 Envoy Gateway，
  然后运行 Gateway API 一致性测试。

#### 在 Mac 主机中 {#on-a-machost}

由于 Mac 不支持将 Docker 网络[直接暴露][]到 Mac 主机，因此请使用以下方法之一来运行一致性测试：

* 在 [Kubernetes 支持][]下部署 Kubernetes 集群或使用 Docker Desktop 然后运行
  `TAG=latest make kube-deploy run-conformance`。
  这将使用最新的 [gateway-dev][] 镜像安装 Envoy Gateway 到当前 kubectl 上下文连接到的 Kubernetes 集群中，并运行一致性测试。
  使用 `make kube-undeploy` 来卸载 Envoy Gateway。
* 安装并执行 [Docker Mac Net Connect][mac_connect] 然后运行 `TAG=latest make conformance`。

**注意：** 在命令前加上 `IMAGE` 或替换 `TAG` 以使用不同的 Envoy Gateway 镜像或标签。
如果未指定 `TAG` ，则会默认使用当前分支的短 SHA。

### 调试 Envoy 配置 {#debugging-the-envoy-config}

查看 Envoy Gateway 正在使用的 Envoy 配置的一种简单方法是将 Envoy 的管理端口（当前为 `19000`）转发到一个本地端口上，这样就可以直接在本地进行访问。

获取 Envoy 部署的名称。下面是 Gateway `eg` 在 `default` 命名空间中的例子：

```shell
export ENVOY_DEPLOYMENT=$(kubectl get deploy -n envoy-gateway-system --selector=gateway.envoyproxy.io/owning-gateway-namespace=default,gateway.envoyproxy.io/owning-gateway-name=eg -o jsonpath='{.items[0].metadata.name}')
```

对其管理端口进行端口转发：

```shell
kubectl port-forward deploy/${ENVOY_DEPLOYMENT} -n envoy-gateway-system 19000:19000
```

现在您可以访问 `127.0.0.1:19000/config_dump` 来查看 Envoy 正在使用的配置。


[Envoy 管理接口][]上还有许多其他端点，这些端点在调试时可能会有所帮助。

### JWT 测试 {#jwt-testing}

示例 [JSON Web Token（JWT）][jwt] 和 [JSON Web Key Set（JWKS）][jwks] 用于[请求认证][]任务。
JWT 由 [JWT 调试器][]使用 `RS256` 算法创建。来自 JWT 的公钥验证签名已复制到 [JWK Creator][] 以生成 JWK。
JWK Creator 配置了匹配的设置，即 `Signing` 公钥使用和 `RS256` 算法。
生成的 JWK 包装在 JWKS 结构中并被托管在仓库中。

[快速开始]: ../../tasks/quickstart
[make]: https://www.gnu.org/software/make/
[Github Actions]: https://docs.github.com/en/actions
[workflows]: https://github.com/envoyproxy/gateway/tree/main/.github/workflows
[Kind]: https://kind.sigs.k8s.io/
[一致性主页]: https://gateway-api.sigs.k8s.io/concepts/conformance/
[直接暴露]: https://kind.sigs.k8s.io/docs/user/loadbalancer/
[Kubernetes 支持]: https://docs.docker.com/desktop/kubernetes/
[gateway-dev]: https://hub.docker.com/r/envoyproxy/gateway-dev/tags
[mac_connect]: https://github.com/chipmk/docker-mac-net-connect
[Envoy 管理接口]: https://www.envoyproxy.io/docs/envoy/latest/operations/admin#operations-admin-interface
[jwt]: https://tools.ietf.org/html/rfc7519
[jwks]: https://tools.ietf.org/html/rfc7517
[请求认证]: https://gateway.envoyproxy.io/latest/tasks/security/jwt-authentication/
[JWT 调试器]: https://jwt.io/
[JWK Creator]: https://russelldavies.github.io/jwk-creator/
