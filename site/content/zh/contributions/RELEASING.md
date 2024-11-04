---
title: "发布流程"
description: "本文描述了 Envoy Gateway 的发布流程。"
---

本文档指导维护人员完成创建 Envoy Gateway 版本的过程。

- [候选版本](#release-candidate)
  - [先决条件](#prerequisites)
  - [设置 Cherry Picker Action](#setup-cherry-picker-action)
- [次要版本](#minor-release)
  - [先决条件](#prerequisites-1)
- [公告发布](#announce-the-release)

## 候选版本 {#release-candidate}

应使用以下步骤创建候选版本。

### 先决条件 {#prerequisites}

- 具备推送到 Envoy Gateway 仓库的权限。

设置环境变量以供后续步骤使用：

```shell
export MAJOR_VERSION=0
export MINOR_VERSION=3
export RELEASE_CANDIDATE_NUMBER=1
export GITHUB_REMOTE=origin
```

1. 克隆仓库，迁出 `main` 分支，确保它是最新的，并且您的本地分支是干净的。
2. 创建一个主题分支，用于添加发布说明并使用发布版本更新 [VERSION][] 文件。
   请参阅之前的[发布说明][]和 [VERSION][] 了解更多详细信息。
3. 签名、提交更改并将其推送到您 Fork 的分支。
4. 提交 [Pull Request][] 将更改合并到 `main` 分支中。
   在您的 PR 合并并且[构建和测试][]成功完成之前，请勿继续。
5. 从 `main` 创建一个新的发布分支。发布分支应命名为
   `release/v${MAJOR_VERSION}.${MINOR_VERSION}`，例如 `release/v0.3`。

   ```shell
   git checkout -b release/v${MAJOR_VERSION}.${MINOR_VERSION}
   ```

6. 将分支推送到 Envoy Gateway 仓库。

    ```shell
    git push ${GITHUB_REMOTE} release/v${MAJOR_VERSION}.${MINOR_VERSION}
    ```

7. 创建主题分支，用于将 Envoy Proxy 镜像和 Envoy Ratelimit 镜像更新为版本支持的 Tag。
   有关更新镜像 Tag 的更多详细信息，请参阅 [PR #2098][]。
8. 签名、提交更改并将其推送到您 Fork 的分支。
9. 提交 [Pull Request][] 将更改合并到 `release/v${MAJOR_VERSION}.${MINOR_VERSION}` 分支中。
   在您的 PR 已合并到发布分支并且 PR 的[构建和测试][]已完成之前，请勿继续。
10. 确保您的发布分支是最新的，并使用候选版本编号为发布分支的头部打 Tag。

    ```shell
    git tag -a v${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER} -m 'Envoy Gateway v${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER} Release Candidate'
    ```

11. 将 Tag 推送到 Envoy Gateway 仓库。

    ```shell
    git push ${GITHUB_REMOTE} v${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER}
    ```

12. 这将触发[发布 GitHub Action][] 并生成发布版本、发布制品等内容。
13. 确认[发布工作流程][]已成功完成。
14. 确认具有正确发布标签的 Envoy Gateway [镜像][]已发布到 Docker Hub。
15. 确认[版本][]已被创建。
16. 请注意，[快速入门][]参考资料并未针对候选版本进行更新。
    但是，请通过手动更新链接来使用候选版本测试快速入门步骤。
17. [生成][] GitHub 变更日志。
18. 确保在编辑 GitHub 版本时选中 "这是一个 pre-release" 复选框。
19. 如果您在此过程中发现任何错误，请创建 Issue。

### 设置 Cherry Picker Action {#setup-cherry-picker-action}

在发布分支切分后，RM（发布经理）应添加
[Cherrypick Action](https://github.com/envoyproxy/gateway/blob/main/.github/workflows/cherrypick.yaml) Job 以进行目标发布。

配置如下所示：

```yaml
  cherry_pick_release_v0_4:
    runs-on: ubuntu-latest
    name: Cherry pick into release-v0.4
    if: ${{ contains(github.event.pull_request.labels.*.name, 'cherrypick/release-v0.4') && github.event.pull_request.merged == true }}
    steps:
      - name: Checkout
        uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11  # v4.1.1
        with:
          fetch-depth: 0
      - name: Cherry pick into release/v0.4
        uses: carloscastrojumo/github-cherry-pick-action@a145da1b8142e752d3cbc11aaaa46a535690f0c5  # v1.0.9
        with:
          branch: release/v0.4
          title: "[release/v0.4] {old_title}"
          body: "Cherry picking #{old_pull_request_id} onto release/v0.4"
          labels: |
            cherrypick/release-v0.4
          # 将发布经理名字放在这里
          reviewers: |
            Alice-Lilith
```

将 `v0.4` 替换为真实的分支名称，并将 `Alice-Lilith` 替换为 RM 的真实名称。

## 次要版本 {#minor-release}

应使用以下步骤创建次要版本。

### 先决条件 {#prerequisites-1}

- 具备推送到 Envoy Gateway 仓库的权限。
- 版本分支是从相应候选版本中切分的。
  有关切分候选版本的更多详细信息，请参阅[候选版本](#release-candidate)部分。

设置环境变量以供后续步骤使用：

```shell
export MAJOR_VERSION=0
export MINOR_VERSION=3
export GITHUB_REMOTE=origin
```

1. 克隆仓库，迁出 `main` 分支，确保它是最新的，并且您的本地分支是干净的。
2. 创建主题分支以添加发布说明、发布公告和具体版本的发布文档。

   1. 创建发布说明。请参阅之前的[发布说明][]以了解更多详细信息。
      **注意：**发布说明应该是候选版本发布说明以及自发布候选版以来的任何更改的累积。
   2. 创建发布公告。请参阅 [PR #635] 作为发布公告示例。
   3. 将版本包含在兼容性矩阵中。请参阅 [PR #1002] 作为示例。
   4. 生成具体版本的发布文档：

   ``` shell
      make docs-release TAG=v${MAJOR_VERSION}.${MINOR_VERSION}.0
   ```

   5. 更新 `site/content/en/_index.md` 中的 `Get Started` 和 `Contributing` 按钮引用链接：

   ```shell
      <a class="btn btn-lg btn-primary me-3 mb-4" href="/v0.5.0">
      Get Started <i class="fas fa-arrow-alt-circle-right ms-2"></i>
      </a>
      <a class="btn btn-lg btn-secondary me-3 mb-4" href="/v0.5.0/contributions">
      Contributing <i class="fa fa-heartbeat ms-2 "></i>
      </a>
   ```

   6. 更新 `site/hugo.toml` 中菜单上的 `Documentation` 引用链接：

   ```shell
   [[menu.main]]
      name = "Documentation"
      weight = -101
      pre = "<i class='fas fa-book pr-2'></i>"
      url = "/v0.5.0"
   ```

3. 签名、提交更改并将其推送到您 Fork 的分支。
4. 提交 [Pull Request][] 将更改合并到 `main` 分支中。
   在您的所有 PR 合并并且最终 PR 的[构建和测试][]完成之前，请勿继续。

5. 迁出发布分支。

   ```shell
   git checkout -b release/v${MAJOR_VERSION}.${MINOR_VERSION} $GITHUB_REMOTE/release/v${MAJOR_VERSION}.${MINOR_VERSION}
   ```

6. 如果发布分支的提示与 `main` 的提示不匹配，则执行以下操作：

   1. 从发布分支创建主题分支。
   2. 从 `main` 中 Cherry-pick 与发布分支不同的提交。
   3. 在本地运行测试，例如 `make lint`。
   4. 签名、提交主题分支并将其推送到您 Fork 的 Envoy Gateway 分支。
   5. 提交 PR，将您的 Fork 分支中的主题合并到 Envoy Gateway 发布分支中。
   6. 在 PR 合并以及该合并的 PR 的 CI 通过之前，不要继续。
   7. 如果您仍在主题分支，请切换到发布分支：

      ```shell
      git checkout release/v${MAJOR_VERSION}.${MINOR_VERSION}
      ```

   8. 确保您的本地发布分支是最新的：

      ```shell
      git pull $GITHUB_REMOTE release/v${MAJOR_VERSION}.${MINOR_VERSION}
      ```

7. 使用发布信息为发布分支的头部打 Tag。例如：

    ```shell
    git tag -a v${MAJOR_VERSION}.${MINOR_VERSION}.0 -m 'Envoy Gateway v${MAJOR_VERSION}.${MINOR_VERSION}.0 Release'
    ```

    **注意：**Tag 版本与发布分支的不同之处在于包含 `.0` 补丁版本。

8. 将标签推送到 Envoy Gateway 仓库。

     ```shell
     git push origin v${MAJOR_VERSION}.${MINOR_VERSION}.0
     ```

9. 这将触发[发布 GitHub Action][] 并生成发布版本、发布制品等内容。
10. 确认[发布工作流程][]已成功完成。
11. 确认具有正确发布标签的 Envoy Gateway [镜像][]已发布到 Docker Hub。
12. 确认[版本][]已被创建。
13. 确认[快速入门][]中的步骤按预期工作。
14. [生成][] GitHub 变更日志并保持发布页面的开头包含以下文本：

   ```console
   # Release Announcement

   Check out the [v${MAJOR_VERSION}.${MINOR_VERSION} release announcement]
   (https://gateway.envoyproxy.io/releases/v${MAJOR_VERSION}.${MINOR_VERSION}.html) to learn more about the release.
   ```

如果您在此过程中发现任何错误，请创建 Issue。

## 公告发布 {#announce-the-release}

让所有人都知道这次发布非常重要。使用以下步骤来公告发布。

1. 在 Envoy Gateway Slack 频道中设置发布信息。例如：

   ```shell
   Envoy Gateway v${MAJOR_VERSION}.${MINOR_VERSION} has been released: https://github.com/envoyproxy/gateway/releases/tag/v${MAJOR_VERSION}.${MINOR_VERSION}.0
   ```

2. 向 Envoy Gateway Slack 频道发送消息。例如：

   ```shell
   On behalf of the entire Envoy Gateway community, I am pleased to announce the release of Envoy Gateway
   v${MAJOR_VERSION}.${MINOR_VERSION}. A big thank you to all the contributors that made this release possible.
   Refer to the official v${MAJOR_VERSION}.${MINOR_VERSION} announcement for release details and the project docs
   to start using Envoy Gateway.
   ...
   ```

   链接到 GitHub 版本和突出显示该版本的版本公告页面。

[发布说明]: https://github.com/envoyproxy/gateway/tree/main/release-notes
[Pull Request]: https://github.com/envoyproxy/gateway/pulls
[快速入门]: https://github.com/envoyproxy/gateway/blob/main/docs/user/quickstart.md
[构建和测试]: https://github.com/envoyproxy/gateway/blob/main/.github/workflows/build_and_test.yaml
[发布 GitHub Action]: https://github.com/envoyproxy/gateway/blob/main/.github/workflows/release.yaml
[发布工作流程]: https://github.com/envoyproxy/gateway/actions/workflows/release.yaml
[镜像]: https://hub.docker.com/r/envoyproxy/gateway/tags
[版本]: https://github.com/envoyproxy/gateway/releases
[生成]: https://docs.github.com/en/repositories/releasing-projects-on-github/automatically-generated-release-notes
[PR #635]: https://github.com/envoyproxy/gateway/pull/635
[PR #2098]: https://github.com/envoyproxy/gateway/pull/2098
[PR #1002]: https://github.com/envoyproxy/gateway/pull/1002
[VERSION]: https://github.com/envoyproxy/gateway/blob/main/VERSION
