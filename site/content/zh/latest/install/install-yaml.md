+++
title = "使用 Kubernetes YAML 安装"
weight = -99
+++

此任务将引导您完成在 Kubernetes 集群中安装 Envoy Gateway。

手动安装过程不允许像 [Helm 安装方法](./install-helm)那样对配置进行更多控制，
因此如果您需要对 Envoy Gateway 安装进行更多控制，建议您使用 Helm。

## 开始之前 {#before-you-begin}

Envoy Gateway 设计为在 Kubernetes 中运行以进行生产。最重要的要求是：

* Kubernetest 1.25+ 版本
* `kubectl` 命令行工具

{{% alert title="兼容性矩阵" color="warning" %}}
请参阅[版本兼容性矩阵](./matrix)了解更多信息。
{{% /alert %}}

## 使用 YAML 安装 {#install-with-yaml}

Envoy Gateway 通常从命令行部署到 Kubernetes。如果您没有 Kubernetes，则应该使用 `kind` 来创建一个。

{{% alert title="开发者指南" color="primary" %}}
请参阅[开发者指南](../../contributions/develop)了解更多信息。
{{% /alert %}}

1. 在终端中，运行以下命令：

    ```shell
    kubectl apply -f https://github.com/envoyproxy/gateway/releases/download/latest/install.yaml
    ```

2. 后续步骤

   Envoy Gateway 现在应该已成功安装并运行，但是为了体验 Envoy Gateway 的更多功能，您可以参考[任务](/latest/tasks)。
