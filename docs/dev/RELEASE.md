## Introduction
This document guides maintainers through the process of creating an Envoy Gateway release.

## Prerequisites
- Permissions to push to the Envoy Gateway repository.

## Creating a Minor Release

1. Clone the repo, checkout the `main` branch, ensure it’s up-to-date, and your local branch is clean.
2. Create the release notes corresponding to the release number. Reference previous [release notes][eg_notes]
   for additional details.
3. Submit a [Pull Request][pr] to merge the release notes into the main branch. This should be the last commit to main
   before cutting the release.
4. Create a new release branch from `main`. The release branch should be named
   `release/v${MAJOR_VERSION}.${MINOR_VERSION}.0`, e.g. `release/v0.3.0`.
   ```shell
   git checkout -b release/v0.3.0
   ```
5. Push the branch to the Envoy Gateway repo.
6. Create a topic branch and update the release tag references in the [Quickstart Guide][quickstart].
7. Sign, commit, and push your changes to your fork. Send a PR to get your changes merged to the release branch.
   Do not proceed until your PR is merged into the release branch.
8. Confirm that the [release workflow][release_workflow] for your PR completed successfully.
9. Tag the head of your release branch with the release tag. For example:
   ```shell
   git tag -a v0.3.0 -m 'Envoy Gateway v0.3.0 Release'
   ```
10. Push the tag to the Envoy Gateway repository.
    ```shell
    git push --tags
    ```
11. This will trigger the [release GitHub action][release_gha] that generates the release, release artifacts, etc.
12. Confirm that the [release workflow][release_workflow] completed successfully.
13. Confirm that the [Envoy Gateway image][image] with the correct release tag was published to Docker Hub.
14. Confirm that the [release][release] was created.
15. Confirm that the steps in the [Quickstart Guide][quickstart] work as expected.
16. [Generate][release_notes] the release notes.
17. Create a ## Release Notes
18. Submit a PR to merge the Quickstart Guide changes from the release branch into the main branch.
19. If you find any bugs in this process, please create an issue.

## Creating a Release Candidate

1. Clone the repo, checkout the `main` branch, ensure it’s up-to-date, and your local branch is clean.
2. Create a changelog corresponding to the release candidate that summarizes the changes included in the
   release candidate. Reference previous [changelogs][changelogs] for additional details.
3. Submit a [Pull Request][pr] to merge the changelog into the main branch. This should be the last commit to main
   before cutting the release candidate.
4. Tag the head of the main branch with the release candidate number. The tag should be named
   `v${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER}`. For example:
   ```shell
   git tag -a v0.3.0-rc.1 -m 'Envoy Gateway v0.3.0-rc.1 Release Candidate'
   ```
5. Push the tag to the Envoy Gateway repository.
   ```shell
   git push --tags
   ```
6. This will trigger the [release GitHub action][release_gha] that generates the release, release artifacts, etc.
7. Confirm that the [release workflow][release_workflow] completed successfully.
8. Confirm that the [Envoy Gateway image][image] with the correct release tag was published to Docker Hub.
9. Confirm that the [release][release] was created.
10. Note that the [Quickstart Guide][quickstart] references are __not__ updated for release candidates. However, test
    the quickstart steps using the release candidate by manually updating the links.
11. [Generate][release_notes] the release notes. Add a `## Release Notes` section that links to the Envoy Gateway
    [release notes][eg_notes].
12. Ensure you check the "This is a pre-release" checkbox when editing the release.
13. If you find any bugs in this process, please create an issue.

## Announcing the Release
It's important that the world knows about the release. Follow the steps to announce the release.
1. Set the release information in the Envoy Gateway Slack channel. For example:
   ```shell
   Envoy Gateway v0.3.0 has been released: https://github.com/envoyproxy/gateway/releases/tag/v0.3.0
   ```
2. Send a message to the Envoy Gateway Slack channel. For example:
   ```shell
   I am pleased to announce the release of Envoy Gateway v${MAJOR_VERSION}.${MINOR_VERSION}.0. The release would not be
   possible without all the support from the Envoy Gateway community...
   ```
   Include a sentence or two that highlights key aspects of the release.

[eg_notes]: https://github.com/envoyproxy/gateway/tree/main/release-notes
[pr]: https://github.com/envoyproxy/gateway/pulls
[quickstart]: https://github.com/envoyproxy/gateway/blob/main/docs/user/QUICKSTART.md
[release_gha]: https://github.com/envoyproxy/gateway/blob/main/.github/workflows/release.yaml
[release_workflow]: https://github.com/envoyproxy/gateway/actions/workflows/release.yaml
[image]: https://hub.docker.com/r/envoyproxy/gateway/tags
[release]: https://github.com/envoyproxy/gateway/releases
[release_notes]: https://docs.github.com/en/repositories/releasing-projects-on-github/automatically-generated-release-notes
