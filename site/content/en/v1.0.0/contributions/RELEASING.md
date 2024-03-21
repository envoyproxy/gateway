---
title: "Release Process"
description: "This section tells the release process of Envoy Gateway."
---

This document guides maintainers through the process of creating an Envoy Gateway release.

- [Release Candidate](#release-candidate)
- [Minor Release](#minor-release)
- [Announce the Release](#announce-the-release)

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
2. Create a topic branch for adding the release notes and updating the [VERSION][] file with the release version. Refer to previous [release notes][] and [VERSION][] for additional details.
3. Sign, commit, and push your changes to your fork.
4. Submit a [Pull Request][] to merge the changes into the `main` branch. Do not proceed until your PR has merged and
   the [Build and Test][] has successfully completed.
5. Create a new release branch from `main`. The release branch should be named
   `release/v${MAJOR_VERSION}.${MINOR_VERSION}`, e.g. `release/v0.3`.

   ```shell
   git checkout -b release/v${MAJOR_VERSION}.${MINOR_VERSION}
   ```

6. Push the branch to the Envoy Gateway repo.

    ```shell
    git push ${GITHUB_REMOTE} release/v${MAJOR_VERSION}.${MINOR_VERSION}
    ```

7. Create a topic branch for updating the Envoy proxy image and Envoy Ratelimit image to the tag supported by the release. Reference [PR #2098][]
   for additional details on updating the image tag.
8. Sign, commit, and push your changes to your fork.
9. Submit a [Pull Request][] to merge the changes into the `release/v${MAJOR_VERSION}.${MINOR_VERSION}` branch. Do not
   proceed until your PR has merged into the release branch and the [Build and Test][] has completed for your PR.
10. Ensure your release branch is up-to-date and tag the head of your release branch with the release candidate number.

    ```shell
    git tag -a v${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER} -m 'Envoy Gateway v${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER} Release Candidate'
    ```

11. Push the tag to the Envoy Gateway repository.

    ```shell
    git push ${GITHUB_REMOTE} v${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER}
    ```

12. This will trigger the [release GitHub action][] that generates the release, release artifacts, etc.
13. Confirm that the [release workflow][] completed successfully.
14. Confirm that the Envoy Gateway [image][] with the correct release tag was published to Docker Hub.
15. Confirm that the [release][] was created.
16. Note that the [Quickstart Guide][] references are __not__ updated for release candidates. However, test
    the quickstart steps using the release candidate by manually updating the links.
17. [Generate][] the GitHub changelog.
18. Ensure you check the "This is a pre-release" checkbox when editing the GitHub release.
19. If you find any bugs in this process, please create an issue.

### Setup cherry picker action

After release branch cut, RM (Release Manager) should add job [cherrypick action](../../../.github/workflows/cherrypick.yaml) for target release.

Configuration looks like following:

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
          # put release manager here
          reviewers: |
            AliceProxy
```

Replace `v0.4` with real branch name, and `AliceProxy` with the real name of RM.

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
   2. Create a release announcement. Refer to [PR #635] as an example release announcement.
   3. Include the release in the compatibility matrix. Refer to [PR #1002] as an example.
   4. Generate the versioned release docs:

   ``` shell
      make docs-release TAG=v${MAJOR_VERSION}.${MINOR_VERSION}.0
   ```

   5. Update the `Get Started` and `Contributing` button referred link in `site/content/en/_index.md`:

   ```shell
      <a class="btn btn-lg btn-primary me-3 mb-4" href="/v0.5.0">
      Get Started <i class="fas fa-arrow-alt-circle-right ms-2"></i>
      </a>
      <a class="btn btn-lg btn-secondary me-3 mb-4" href="/v0.5.0/contributions">
      Contributing <i class="fa fa-heartbeat ms-2 "></i>
      </a>
   ```

   6. Uodate the `Documentation` referred link on the menu in `site/hugo.toml`:

   ```shell
   [[menu.main]]
      name = "Documentation"
      weight = -101
      pre = "<i class='fas fa-book pr-2'></i>"
      url = "/v0.5.0"
   ```

3. Sign, commit, and push your changes to your fork.
4. Submit a [Pull Request][] to merge the changes into the `main` branch. Do not proceed until all your PRs have merged
   and the [Build and Test][] has completed for your final PR.

5. Checkout the release branch.

   ```shell
   git checkout -b release/v${MAJOR_VERSION}.${MINOR_VERSION} $GITHUB_REMOTE/release/v${MAJOR_VERSION}.${MINOR_VERSION}
   ```

6. If the tip of the release branch does not match the tip of `main`, perform the following:

   1. Create a topic branch from the release branch.
   2. Cherry-pick the commits from `main` that differ from the release branch.
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

7. Tag the head of your release branch with the release tag. For example:

    ```shell
    git tag -a v${MAJOR_VERSION}.${MINOR_VERSION}.0 -m 'Envoy Gateway v${MAJOR_VERSION}.${MINOR_VERSION}.0 Release'
    ```

    __Note:__ The tag version differs from the release branch by including the `.0` patch version.

8. Push the tag to the Envoy Gateway repository.

     ```shell
     git push origin v${MAJOR_VERSION}.${MINOR_VERSION}.0
     ```

9. This will trigger the [release GitHub action][] that generates the release, release artifacts, etc.
10. Confirm that the [release workflow][] completed successfully.
11. Confirm that the Envoy Gateway [image][] with the correct release tag was published to Docker Hub.
12. Confirm that the [release][] was created.
13. Confirm that the steps in the [Quickstart Guide][] work as expected.
14. [Generate][] the GitHub changelog and include the following text at the beginning of the release page:

   ```console
   # Release Announcement

   Check out the [v${MAJOR_VERSION}.${MINOR_VERSION} release announcement]
   (https://gateway.envoyproxy.io/releases/v${MAJOR_VERSION}.${MINOR_VERSION}.html) to learn more about the release.
   ```

If you find any bugs in this process, please create an issue.

## Announce the Release

It's important that the world knows about the release. Use the following steps to announce the release.

1. Set the release information in the Envoy Gateway Slack channel. For example:

   ```shell
   Envoy Gateway v${MAJOR_VERSION}.${MINOR_VERSION} has been released: https://github.com/envoyproxy/gateway/releases/tag/v${MAJOR_VERSION}.${MINOR_VERSION}.0
   ```

2. Send a message to the Envoy Gateway Slack channel. For example:

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
[Quickstart Guide]: https://github.com/envoyproxy/gateway/blob/main/docs/user/quickstart.md
[Build and Test]: https://github.com/envoyproxy/gateway/blob/main/.github/workflows/build_and_test.yaml
[release GitHub action]: https://github.com/envoyproxy/gateway/blob/main/.github/workflows/release.yaml
[release workflow]: https://github.com/envoyproxy/gateway/actions/workflows/release.yaml
[image]: https://hub.docker.com/r/envoyproxy/gateway/tags
[release]: https://github.com/envoyproxy/gateway/releases
[Generate]: https://docs.github.com/en/repositories/releasing-projects-on-github/automatically-generated-release-notes
[PR #635]: https://github.com/envoyproxy/gateway/pull/635
[PR #2098]: https://github.com/envoyproxy/gateway/pull/2098
[PR #1002]: https://github.com/envoyproxy/gateway/pull/1002
[VERSION]: https://github.com/envoyproxy/gateway/blob/main/VERSION
