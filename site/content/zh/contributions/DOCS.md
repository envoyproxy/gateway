---
title: "参与 Envoy Gateway 文档工作"
description: "本节讲述 Envoy Gateway 文档的开发工作。"
---

Envoy Gateway 的文档位于 `site/content/en` 目录中（中文内容位于 `site/content/zh` 目录中）。
任何单独的文档都可以使用 [Markdown] 编写。

## 文档结构 {#documentation-structure}

我们现在支持版本化的文档，文档下的目录名代表了文档的版本。
最新站点的根目录位于 `site/content/en/latest` 中。
如果您想了解这些内容如何组合在一起，可以由此处开始。

请注意，新内容应被添加到 `site/content/en/latest` 中，
并将在下一个版本中被截断。`site/content/en/v0.5.0` 下的内容是自动生成的，
通常不需要对其进行更改，除非您发现当前发布页面有一些不正确的内容。
如果是这样，您应该提交 PR 来一并更新 `site/content/en/latest` 和 `site/content/en/v0.5.0` 的内容。

您可以默认访问代表当前版本的网站内容，
也可以在[此处][latest-website]或页面的页脚处访问包含最新版本变更的网站。

## 文档工作流程 {#documentation-workflow}

要参与文档工作，只需编辑 `site/content/en/latest` 中的 Markdown 文件，然后运行：

```bash
make docs
```

这将使用被构建的 HTML 页面创建 `site/public`。您可以通过运行以下命令来预览它：

```shell
make docs-serve
```

如果您想生成文档的新发布版本，例如 `v0.6.0`，请运行：

```bash
make docs-release TAG=v0.6.0
```

该操作将更新项目根目录下的 VERSION 文件，该文件记录当前发布的版本，
并将在页面版本上下文和二进制版本输出中被使用。此外，这将生成新的目录 `site/content/en/v0.6.0`，
其中包含 v0.6.0 的文档，如 `/api`、`/install` 等。

## 发布文档 {#publishing-docs}

每当文档被推送到 `main` 分支时，CI 都会将构建的文档发布到 GitHub Pages。
有关更多详细信息，请参阅 `.github/workflows/docs.yaml`。

## 参考 {#reference}

前往 [Hugo](https://gohugo.io) 和 [Docsy](https://www.docsy.dev/docs) 了解更多信息。

如果您希望参与中文内容翻译或贡献，请先阅读[规范][docs-standard]以帮助您更好的参与内容贡献。

[Markdown]: https://daringfireball.net/projects/markdown/syntax
[latest-website]: /zh/latest
[docs-standard]: ../docs_standard
