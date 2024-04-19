---
title: "文档内容编写规范（中文）"
description: "本节讲述 Envoy Gateway 文档的编写或翻译规范，包括：格式，书写习惯建议等。"
---

中文内容的位置在 `site/content/zh` 目录中，具体内容和结构应与英文版本保持一致。

## 规范 {#standard}

以下定义了一些编写规范，并提供示例和对应 Markdown 写法。

### 中英文混排 {#mixed-chinese-and-english}

当内容中存在中文和英文混排的时候中英文间需要加一个空格，
中文与中文之间无需空格（即使某个中文是链路的一部分也不需要空格）：

**示例：**

这里有中文和 English 混排。
可以跳转到[规范](#standard)处重新阅读。

```md
这里有中文和 English 混排。
可以跳转到[规范](#standard)处重新阅读。
```

### 粗体 {#bold}

在中文内容中建议一律使用两个星号表示粗体：

**示例：**

其中**非常重要**的内容是

```md
其中**非常重要**的内容是
```

### 标点符号 {#punctuation}

中文内容中遇到的标点符号都要用全角，
无论中英文内容与全角标点符号之间无需空格（即使某个中文是在加粗格式之中也不需要空格）

**示例：**

无法确认吗？确认无误（没有任何拼写错误）后，可以继续后面的步骤。
安装 Envoy Gateway（通过命令）**结束后**，查看相关 Service。

```md
无法确认吗？确认无误（没有任何拼写错误）后，可以继续后面的步骤。
安装 Envoy Gateway（通过命令）**结束后**，查看相关 Service。
```

### 锚 {#anchor}

标题要添加英文锚（anchor），用于在其他内容中保持引用定位的一致性：

**示例：**

#### 示例 {#example}

```md
#### 示例 {#example}
```

### 无需翻译的情况 {#no-need-for-translation}

如果内容中的英文是命令，特定专有名词，如 CRD 名称，协议名称等，请保持大小写与原本定义一致：

**示例：**

使用 curl 命令发起 HTTP 请求。
在 kind 集群中添加 Gateway 资源，创建 Service 和 Deployment。

```md
使用 curl 命令发起 HTTP 请求。
在 kind 集群中添加 Gateway 资源，创建 Service 和 Deployment。
```

### 英文复数处理 {#english-plural}

大部分情况遇到英文的复数形式请按需调整为单数：

**示例：**

英文：Delete all Services in Kubernetes.

中文：删除所有 Kubernetes 中的 Service。

```md
英文：Delete all Services in Kubernetes.
中文：删除所有 Kubernetes 中的 Service。
```

### 换行断句 {#line-break}

如果一句话很长，尽量按照本身的句子进行断句换行，通常一行内容保持在 40-60 字符，方便 Review：

**示例：**

这是一个长句子，在网页中查看

```md
至此，您已经了解了中文文档内容编写规范！
```

### 使用“您” {#second-person}

第二人称代词统一使用“您”：

**示例：**

至此，您已经了解了中文文档内容编写规范！

```md
至此，您已经了解了中文文档内容编写规范！
```

### 必要时意译 {#free-translation}

翻译时有些情况需要使用意译代替直译：

**示例：**

英文原文：He is as cool as a cucumber.

直译成中文：他像黄瓜一样冷静。

更好的翻译可能是：他泰然自若，从容不迫。

## 欢迎贡献 {#welcome-to-contribute}

以上规范参考了一些开源社区中文内容贡献者日常贡献中的约定俗成的做法，如有不适合或需要改进的，欢迎提出您的建议！
