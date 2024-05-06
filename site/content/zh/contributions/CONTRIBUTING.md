---
title: "贡献"
description: "本节介绍如何为 Envoy Gateway 做出贡献。"
weight: 3
---

我们欢迎来自社区的贡献。请仔细查看[项目目标](/zh/about)和以下指南，以简化您的贡献流程。

## 沟通 {#communication}

* 在开始开发主要功能之前，请通过 GitHub 或 Slack 联系我们。
  我们将确保没有其他人正在处理此问题，并要求您创建 GitHub Issue。
* “主要功能”定义为超过 100 个 LOC 的任意更改（不包括测试），
  或更改任何面向用户的行为。我们将使用 GitHub Issue 来讨论该功能并达成一致。
  这是为了防止浪费您和我们的时间。主要功能的 GitHub 审核流程也很重要，
  以便[与提交权限的干系者](../codeowners)可以就设计达成一致。
  如果适合编写设计文档，则该文档必须托管在 GitHub Issue 中，或者托管在公开可读的位置并链接到 Issue 中。
* 小补丁和错误修复不需要事先沟通。

## 包容性 {#inclusivity}

Envoy Gateway 社区的一个明确的目标是完全包容性。
因此，所有 PR 的所有代码、API 和文档都必须遵守以下准则：

* 不允许使用以下单词和短语：
  * **Whitelist**：使用 allowlist 代替。
  * **Blacklist**：使用 denylist 或 blocklist 代替。
  * **Master**：使用 primary 代替。
  * **Slave**：使用 secondary 或 replica 代替。
* 文档应该以包容的风格编写。
  [Google 开发人员文档](https://developers.google.com/style/inclusive-documentation)包含有关此主题的出色参考。
* 上述政策并非最终政策，未来可能会随着行业最佳实践的发展而进行修改。
  维护人员在代码审查期间可能会提供有关此主题的其他评论。

## 提交一个 PR {#submitting-a-pr}

* Fork 该仓库。
* 修改。
* 对每次提交都进行 DCO 签署。这可以通过 `git commit -s` 来完成。
* 提交您的 PR。
* 将为您自动运行测试。
* 我们**不会**合并任何未通过测试的 PR。
* PR 预计对添加的代码具有 100% 的测试覆盖率。这可以通过覆盖范围构建来验证。
  如果您的 PR 由于某种原因无法被 100% 覆盖，请在创建时明确解释原因。
* 任何更改面向用户行为的 PR **必须**在仓库中的
  [docs](https://github.com/envoyproxy/gateway/tree/main/site) 文件夹中具有关联的文档以及[变更日志](./RELEASING)。
* 所有代码注释和文档均应具有正确的英语语法和标点符号。
  如果您的英语不流利（或者是一个糟糕的写作者 ;-)），请告诉我们，我们会尽力为您寻求帮助，但不能保证。
* 您的 PR 标题应该是描述性的，通常以包含子系统名称的类型开头，
  如有必要，带有 `()`，摘要后跟冒号。格式例子如下：
  * "docs: fix grammar error"
  * "feat(translator): add new feature"
  * "fix: fix xx bug"
  * "chore: change ci & build tools etc"
* 当您的 PR 被合并时，您的 PR 提交消息将用作其提交消息。
  如果您的 PR 在审核期间出现分歧，您应该更新此字段。
* 您的 PR 描述应详细说明 PR 的用途。如果它修复了现有问题，则应以“Fixes #XXX”结尾。
* 如果您的 PR 是共同创作的或基于其他贡献者的早期 PR，
  请注明 `Co-authored-by: name <name@example.com>`。
  有关更多详细信息，请参阅 GitHub 的[多作者指南](https://docs.github.com/zh/pull-requests/committing-changes-to-your-project/creating-and-editing-commits/creating-a-commit-with-multiple-authors)。
* 当所有测试都通过并且满足本文所述的所有其他条件时，将指派维护人员审查并合并 PR。
* 一旦您提交了 PR，**请不要对其进行 Rebase**。
  如果后续提交是新提交和/或合并，则检查起来要容易得多。我们会压缩并合并，因此 PR 中的提交数量并不重要。
* 我们预计，一旦 PR 被开启，它将被积极处理，直到它被合并或关闭。
  我们保留关闭没有取得进展的 PR 的权利。这通常定义为 7 天内没有变化。
  显然，由于缺乏活跃度而被关闭的 PR 可以稍后重新开启。
  关闭过时的 PR 有助于我们掌控当前正在进行的所有工作。

## 维护者 PR 审查策略 {#maintainer-pr-review-policy}

* 请参阅 [CODEOWNERS.md](./codeowners) 了解当前的维护者列表。
* 需要由代表与 PR 所有者不同的隶属关系的维护者来审查和批准 PR。
* 当项目成熟时，预计 PR 涉及代码的“领域专家”应该对 PR 进行审查。
  此人不需要提交权限，只需要领域知识。
* 对于仅更新文档或评论，
  或对测试和工具进行细微更改（其中细微更改由相关维护者决定）的 PR，可以免除上述规则。
* 如果有关于谁应该审查 PR 的问题，请在 Slack 中讨论。
* 欢迎任何人审查他们想要审查的任何 PR，无论他们是否是维护者。
* 如果 PR 在审核期间发生重大变化，请确保更新 PR 标题、提交消息和描述。
* 合并前请**清理标题和正文**。默认情况下，GitHub 使用原始标题填充压缩合并标题，
  并使用 PR 中的每个单独提交填充提交正文。进行合并的维护者应确保标题遵循上述准则，
  并应使用 PR 的原始提交消息覆盖正文（如有必要，请将其清理），同时保留 PR 作者的最终 DCO 签名。

## 决策 {#decision-making}

这是一个新的、复杂的项目，我们需要很快做出很多决定。
为此，我们确定了做出（可能有争议的）决定的流程：

* 对于需要记录的决策，我们会创建一个 Issue。
* 在该 Issue 中，我们讨论想法，然后维护者可以在评论中要求投票。
* 维护者可以通过在另外的评论中做出反馈或回复来对该评论进行具有约束力的投票。
* 欢迎非维护者社区成员通过这两种方法进行不具约束力的投票。
* 投票将通过简单多数达成决定。
* 如果出现僵局，这个问题将被转移到 Steering 处理。

## DCO：签署您的工作 {#dco-sign-your-work}

签署是补丁说明末尾的一行简单内容，它证明您编写了该补丁或有权将其作为开源补丁传递。
规则非常简单：如果您可以证明以下内容（来自 [developercertificate.org](https://developercertificate.org/)）：

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
660 York Street, Suite 102,
San Francisco, CA 94110 USA

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

然后您只需在每个 git 提交消息中使用您的真实姓名（抱歉，不支持化名或匿名贡献）添加一行：

    Signed-off-by: Joe Smith <joe@gmail.com>

您可以在通过 `git commit -s` 创建 git 提交时添加签署。

如果您希望这是自动的，您可以设置一些别名：

```bash
git config --add alias.amend "commit -s --amend"
git config --add alias.c "commit -s"
```

## 修复 DCO {#fixing-dco}

如果您的 PR 未通过 DCO 检查，则有必要修复 PR 中的整个提交历史记录。
最佳实践是 [squash](https://docs.github.com/zh/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/about-pull-request-merges#squash-and-merge-your-commits)
将提交历史记录合并到单个提交中，如上所述附加 DCO 签名，
然后[强制推送](https://git-scm.com/docs/git-push/zh_HANS-CN#git-push---force)。
例如，如果您的历史记录中有 2 次提交：

```bash
git rebase -i HEAD^^
(interactive squash + DCO append)
git push origin -f
```

请注意，一般来说，以这种方式重写历史记录会阻碍审核过程，并且只能在纠正 DCO 错误时才这样做。
