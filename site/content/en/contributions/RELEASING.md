---
title: "Release Process"
description: "This section tells the release process of Envoy Gateway."
---

This document guides maintainers through the process of creating an Envoy Gateway release.

- [Release Candidate](#release-candidate)
  - [Prerequisites](#prerequisites)
- [Minor Release](#minor-release)
  - [Prerequisites](#prerequisites-1)
  - [Announce the Release](#announce-the-release)
- [Patch Release](#patch-release)
  - [Prerequisites](#prerequisites-2)
  - [Announce the Release](#announce-the-release-1)

## Release Candidate

The following steps should be used for creating a release candidate.

### Prerequisites

- Permissions to push to the Envoy Gateway repository.

Set environment variables for use in subsequent steps:

```shell
export MAJOR_VERSION=0
export MINOR_VERSION=3
export RELEASE_CANDIDATE_NUMBER=1
export GITHUB_REMOTE=origin
```

1. Clone the repo, checkout the `main` branch, ensure it’s up-to-date, and your local branch is clean.
2. Create a topic branch for adding the release notes and updating the [VERSION][] file with the release version. Refer to previous [release notes][] and [VERSION][] for additional details. The latest changes are already accumulated in the current.yaml file. Copy the content of the current.yaml file to the release notes file and clear the current.yaml file.

   ```shell
   echo "v${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER}" > VERSION
   ```

   __Note:__ The release candidate version should be in the format `${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER}`.
3. Sign, commit, and push your changes to your fork.
4. Submit a [Pull Request][] to merge the changes into the `main` branch. 
5. Do not proceed until your PR has merged and the [Build and Test][] has successfully completed.
6. Create a new release branch from `main`. The release branch should be named
   `release/v${MAJOR_VERSION}.${MINOR_VERSION}`, e.g. `release/v0.3`.

   ```shell
   git checkout -b release/v${MAJOR_VERSION}.${MINOR_VERSION}
   ```

7. Push the branch to the Envoy Gateway repo.

    ```shell
    git push ${GITHUB_REMOTE} release/v${MAJOR_VERSION}.${MINOR_VERSION}
    ```

8. Create a topic branch for updating the [Envoy proxy image][] and [Envoy Ratelimit image][] to the tag supported by the release.
 Please note that the tags should be updated in both the source code and the Helm chart. Reference [PR #5872][]
   for additional details on updating the image tag.
9. Sign, commit, and push your changes to your fork.
10. Submit a [Pull Request][] to merge the changes into the `release/v${MAJOR_VERSION}.${MINOR_VERSION}` branch. 
11. Do not proceed until your PR has merged into the release branch and the [Build and Test][] has completed for your PR.
12. Ensure your release branch is up-to-date and tag the head of your release branch with the release candidate number.

    ```shell
    git tag -a v${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER} -m 'Envoy Gateway v${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER} Release Candidate'
    ```

13. Push the tag to the Envoy Gateway repository.

    ```shell
    git push ${GITHUB_REMOTE} v${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER}
    ```

14. This will trigger the [release GitHub action][] that generates the release, release artifacts, etc.
15. Confirm that the [release workflow][] completed successfully.
16. Confirm that the Envoy Gateway [image][] with the correct release tag was published to Docker Hub.
17. Confirm that the [release][] was created.
18. Note that the [Quickstart][] references are __not__ updated for release candidates. However, test
    the quickstart steps using the release candidate by manually updating the links.
19. [Generate][] the GitHub changelog.
20. Ensure you check the "This is a pre-release" checkbox when editing the GitHub release.
21. If you find any bugs in this process, please create an issue.

## Minor Release

The following steps should be used for creating a minor release.

### Prerequisites

- Permissions to push to the Envoy Gateway repository.
- A release branch that has been cut from the corresponding release candidate. Refer to the
  [Release Candidate](#release-candidate) section for additional details on cutting a release candidate.

Set environment variables for use in subsequent steps:

```shell
export MAJOR_VERSION=0
export MINOR_VERSION=3
export GITHUB_REMOTE=origin
```

1. Clone the repo, checkout the `main` branch, ensure it’s up-to-date, and your local branch is clean.
2. Create a topic branch for adding the release notes, release announcement, and versioned release docs.

   1. Create the release notes. Reference previous [release notes][] for additional details. __Note:__  The release
      notes should be an accumulation of the release candidate release notes and any changes since the release
      candidate.
   1. Create a release announcement. Refer to [PR #635] as an example release announcement.
   1. Include the release in the compatibility matrix. Refer to [PR #1002] as an example.
   1. Generate the versioned release docs:

      ``` shell
      make docs-release TAG=v${MAJOR_VERSION}.${MINOR_VERSION}.0
      ```

   1. Update `site/layouts/shortcodes/helm-version.html`, add the latest version of the minor release, and update the short code for `{{- with (strings.HasPrefix $pagePrefix "doc") -}}` to the latest minor version.

      ```console
      {{- $pagePrefix := (index (split $.Page.File.Dir "/") 0) -}}
      {{- with (eq $pagePrefix "latest") -}}
      {{- "v0.0.0-latest" -}}
      {{- end -}}
      {{- with (strings.HasPrefix $pagePrefix "v1.1") -}}
      {{- "v1.1.3" -}}
      {{- end -}}
      {{- with (strings.HasPrefix $pagePrefix "v1.2") -}}
      {{- "v1.2.0" -}}
      {{- end -}}
      {{- with (strings.HasPrefix $pagePrefix "docs") -}}
      {{- "v1.2.0" -}}
      {{- end -}}
      ```

   1. Update `site/layouts/shortcodes/yaml-version.html`, add the latest version of the minor release, and update the short code for `{{- with (strings.HasPrefix $pagePrefix "doc") -}}` to the latest minor version.

      ```console
      {{- $pagePrefix := (index (split $.Page.File.Dir "/") 0) -}}
      {{- with (eq $pagePrefix "latest") -}}
      {{- "latest" -}}
      {{- end -}}
      {{- with (strings.HasPrefix $pagePrefix "v1.1") -}}
      {{- "v1.1.3" -}}
      {{- end -}}
      {{- with (strings.HasPrefix $pagePrefix "v1.2") -}}
      {{- "v1.2.0" -}}
      {{- end -}}
      {{- with (strings.HasPrefix $pagePrefix "docs") -}}
      {{- "v1.2.0" -}}
      {{- end -}}
      ```

   1. Update `site/hugo.toml`, add the new version to the `params.versions` section.

      ```console
      [[params.versions]]
        version = "v1.3"
        url = "/v1.3"
        eol = "2025-07-30"
      ```

   1. Update `site/hugo.toml`, change the version to current major version.

      ```console
      # The version number for the version of the docs represented in this doc set.
      # Used in the "version-banner" partial to display a version number for the
      # current doc set.
      version = "v1.3"
      ```

3. Sign, commit, and push your changes to your fork.
4. Submit a [Pull Request][] to merge the changes into the `main` branch.

5. Do not proceed until all your PRs have merged and the [Build and Test][] has completed for your final PR.

6. Checkout the release branch.

   ```shell
   git checkout release/v${MAJOR_VERSION}.${MINOR_VERSION} $GITHUB_REMOTE/release/v${MAJOR_VERSION}.${MINOR_VERSION}
   ```

7. If the tip of the release branch does not match the tip of `main`, perform the following:

   1. Create a topic branch from the release branch.
   2. Cherry-pick the commits from `main` that differ from the release branch, e.g. `git cherry-pick <present commit in rc>..<latest commit on main> -s`
   3. Run tests locally, e.g. `make lint`.
   4. Sign, commit, and push your topic branch to your Envoy Gateway fork.
   5. Submit a PR to merge the topic from of your fork into the Envoy Gateway release branch.
   6. Do not proceed until the PR has merged and CI passes for the merged PR.
   7. If you are still on your topic branch, change to the release branch:

      ```shell
      git checkout release/v${MAJOR_VERSION}.${MINOR_VERSION}
      ```

   8. Ensure your local release branch is up-to-date:

      ```shell
      git pull $GITHUB_REMOTE release/v${MAJOR_VERSION}.${MINOR_VERSION}
      ```

   9. If upstream has updated the [Envoy proxy image][] or [Envoy Ratelimit image][] tag supported by the release,
   you should also create a topic branch for bumping these tags.
   Please note that the tags should be updated in both the source code and the Helm chart. Reference [PR #5872][]

8. Tag the head of your release branch with the release tag. For example:

    ```shell
    git tag -a v${MAJOR_VERSION}.${MINOR_VERSION}.0 -m 'Envoy Gateway v${MAJOR_VERSION}.${MINOR_VERSION}.0 Release'
    ```

    __Note:__ The tag version differs from the release branch by including the `.0` patch version.

9. Push the tag to the Envoy Gateway repository.

     ```shell
     git push origin v${MAJOR_VERSION}.${MINOR_VERSION}.0
     ```

10. This will trigger the [release GitHub action][] that generates the release, release artifacts, etc.
11. Confirm that the [release workflow][] completed successfully.
12. Confirm that the Envoy Gateway [image][] with the correct release tag was published to Docker Hub.
13. Confirm that the [release][] was created.
14. Confirm that the steps in the [Quickstart][] work as expected.
15. [Generate][] the GitHub changelog and include the following text at the beginning of the release page:

   ```console
   # Release Announcement

   Check out the [v${MAJOR_VERSION}.${MINOR_VERSION} release announcement]
   (https://gateway.envoyproxy.io/news/releases/notes/v${MAJOR_VERSION}.${MINOR_VERSION}.html) to learn more about the release.
   ```

16. Update the `lastVersionTag` in `test/e2e/tests/eg_upgrade.go` to reflect the latest prior release. Refer to [PR #4666] as an example.

If you find any bugs in this process, please create an issue.

### Announce the Release

It's important that the world knows about the release. Use the following steps to announce the release.

1. Set the release information in the Envoy Gateway Slack channel. For example:

   ```shell
   Envoy Gateway v${MAJOR_VERSION}.${MINOR_VERSION} has been released: https://github.com/envoyproxy/gateway/releases/tag/v${MAJOR_VERSION}.${MINOR_VERSION}.0
   ```

2. Send a message to the Envoy Gateway Slack channel and the [Google Group](https://groups.google.com/g/envoy-gateway-announce). For example:

   ```shell
   On behalf of the entire Envoy Gateway community, I am pleased to announce the release of Envoy Gateway
   v${MAJOR_VERSION}.${MINOR_VERSION}. A big thank you to all the contributors that made this release possible.
   Refer to the official v${MAJOR_VERSION}.${MINOR_VERSION} announcement for release details and the project docs
   to start using Envoy Gateway.
   ...
   ```

   Link to the GitHub release and release announcement page that highlights the release.

## Patch Release

The following steps should be used for creating a patch release.

### Prerequisites

- Permissions to push to the Envoy Gateway repository.
- A minor release has already been released. Refer to the [Minor Release](#minor-candidate) section for additional details on releasing a minor release.

Set environment variables for use in subsequent steps:

```shell
export MAJOR_VERSION=1
export MINOR_VERSION=2
export PATCH_VERSION=1
export GITHUB_REMOTE=origin
```

1. Clone the repo, checkout the `main` branch, ensure it’s up-to-date, and your local branch is clean.
2. Create a topic branch for adding the release notes.

   1. Create the release notes. The release note should only include the changes since the last minor or patch release.
   1. Update `site/layouts/shortcodes/helm-version.html`, update the short code for `{{- with (strings.HasPrefix $pagePrefix "doc") -}}` to the latest patch version. For example:

      ```console
      {{- $pagePrefix := (index (split $.Page.File.Dir "/") 0) -}}
      {{- with (eq $pagePrefix "latest") -}}
      {{- "v0.0.0-latest" -}}
      {{- end -}}
      {{- with (strings.HasPrefix $pagePrefix "v1.1") -}}
      {{- "v1.1.3" -}}
      {{- end -}}
      {{- with (strings.HasPrefix $pagePrefix "v1.2") -}}
      {{- "v1.2.1" -}}
      {{- end -}}
      {{- with (strings.HasPrefix $pagePrefix "doc") -}}
      {{- "v1.2.1" -}}
      {{- end -}}
      ```

   1. Update `site/layouts/shortcodes/yaml-version.html`, update the short code for `{{- with (strings.HasPrefix $pagePrefix "doc") -}}` to the latest patch version. For example:

      ```console
      {{- $pagePrefix := (index (split $.Page.File.Dir "/") 0) -}}
      {{- with (eq $pagePrefix "latest") -}}
      {{- "latest" -}}
      {{- end -}}
      {{- with (strings.HasPrefix $pagePrefix "v1.1") -}}
      {{- "v1.1.3" -}}
      {{- end -}}
      {{- with (strings.HasPrefix $pagePrefix "v1.2") -}}
      {{- "v1.2.1" -}}
      {{- end -}}
      {{- with (strings.HasPrefix $pagePrefix "doc") -}}
      {{- "v1.2.1" -}}
      {{- end -}}
      ```

3. Sign, commit, and push your changes to your fork.
4. Submit a [Pull Request][] to merge the changes into the `main` branch.
5. Do not proceed until all your PRs have merged and the [Build and Test][] has completed for your final PR.
6. Checkout the release branch.

   ```shell
   git checkout release/v${MAJOR_VERSION}.${MINOR_VERSION} $GITHUB_REMOTE/release/v${MAJOR_VERSION}.${MINOR_VERSION}
   ```

7. Cherry-pick the release note that you created in the previous step to the release branch. The release note will be included in the release artifacts.
   1. Create a topic branch from the release branch.
   2. Cherry-pick the release note and release announcement commit from `main` to the topic branch.
   3. Submit a PR to merge the topic from of your fork into the release branch.

8. Cherry-pick the commits that you want to include in the patch release.
   1. Create a topic branch from the release branch.
   2. Cherry-pick the commits from `main` that you want to include in the patch release.
   3. Run tests locally, e.g. `make lint`.
   4. Sign, commit, and push your topic branch to your Envoy Gateway fork.
   5. Submit a PR to merge the topic from of your fork into the release branch.
   6. Do not proceed until the PR has merged and CI passes for the merged PR.
   7. If you are still on your topic branch, change to the release branch:

      ```shell
      git checkout release/v${MAJOR_VERSION}.${MINOR_VERSION}
      ```

   8. Ensure your local release branch is up-to-date:

      ```shell
      git pull $GITHUB_REMOTE release/v${MAJOR_VERSION}.${MINOR_VERSION}
      ```

   9. If upstream has updated the [Envoy proxy image][] or [Envoy Ratelimit image][] tag supported by the release,
   you should also create a topic branch for bumping these tags.
   Please note that the tags should be updated in both the source code and the Helm chart. Reference [PR #5872][]

9. Tag the head of your release branch with the release tag. For example:

    ```shell
    git tag -a v${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION} -m 'Envoy Gateway v${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION} Release'
    ```

10. Push the tag to the Envoy Gateway repository.

      ```shell
      git push origin v${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION}
      ```

11. This will trigger the [release GitHub action][] that generates the release, release artifacts, etc.
12. Confirm that the [release workflow][] completed successfully.
13. Confirm that the Envoy Gateway [image][] with the correct release tag was published to Docker Hub.
14. Confirm that the [release][] was created.
15. Confirm that the steps in the [Quickstart][] work as expected.
16. [Generate][] the GitHub changelog and include the following text at the beginning of the release page:

   ```console
   # Release Announcement

   Check out the [v${MAJOR_VERSION}.${MINOR_VERSION}.${MINOR_VERSION}  release announcement]
   (https://gateway.envoyproxy.io/news/releases/notes/v${MAJOR_VERSION}.${MINOR_VERSION}.${MINOR_VERSION}.html) to learn more about the release.
   ```

17. If this patch release is the latest release, update the `lastVersionTag` in `test/e2e/tests/eg_upgrade.go` to reflect the latest prior release. Refer to [PR #4666] as an example.

### Announce the Release

It's important that the world knows about the release. Use the following steps to announce the release.

1. Set the release information in the Envoy Gateway Slack channel. For example:

   ```shell
   Envoy Gateway v${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION} has been released: https://github.com/envoyproxy/gateway/releases/tag/v${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION}
   ```

2. Send a message to the Envoy Gateway Slack channel and the [Google Group](https://groups.google.com/g/envoy-gateway-announce). For example:

   ```shell
   On behalf of the entire Envoy Gateway community, I am pleased to announce the release of Envoy Gateway
   v${MAJOR_VERSION}.${MINOR_VERSION}. A big thank you to all the contributors that made this release possible.
   Refer to the official v${MAJOR_VERSION}.${MINOR_VERSION} announcement for release details and the project docs
   to start using Envoy Gateway.
   ...
   ```

   Link to the GitHub release and release announcement page that highlights the release.

[release notes]: https://github.com/envoyproxy/gateway/tree/main/release-notes
[Pull Request]: https://github.com/envoyproxy/gateway/pulls
[Quickstart]: https://github.com/envoyproxy/gateway/blob/main/docs/user/quickstart.md
[Build and Test]: https://github.com/envoyproxy/gateway/blob/main/.github/workflows/build_and_test.yaml
[release GitHub action]: https://github.com/envoyproxy/gateway/blob/main/.github/workflows/release.yaml
[release workflow]: https://github.com/envoyproxy/gateway/actions/workflows/release.yaml
[image]: https://hub.docker.com/r/envoyproxy/gateway/tags
[release]: https://github.com/envoyproxy/gateway/releases
[Generate]: https://docs.github.com/en/repositories/releasing-projects-on-github/automatically-generated-release-notes
[PR #635]: https://github.com/envoyproxy/gateway/pull/635
[PR #5872]: https://github.com/envoyproxy/gateway/pull/5872
[PR #1002]: https://github.com/envoyproxy/gateway/pull/1002
[PR #4666]: https://github.com/envoyproxy/gateway/pull/4666
[VERSION]: https://github.com/envoyproxy/gateway/blob/main/VERSION
[Envoy proxy image]: https://hub.docker.com/r/envoyproxy/envoy/tags?name=distroless
[Envoy Ratelimit image]: https://hub.docker.com/r/envoyproxy/ratelimit
