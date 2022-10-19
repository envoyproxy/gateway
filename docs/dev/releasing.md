# Release Process

This document guides maintainers through the process of creating an Envoy Gateway release.

## Prerequisites

- Permissions to push to the Envoy Gateway repository.

## Creating a Minor Release

1. Clone the repo, checkout the `main` branch, ensure it’s up-to-date, and your local branch is clean.
2. Create the release notes corresponding to the release number. Reference previous [release notes][]
   for additional details.
3. Submit a [Pull Request][] to merge the release notes into the main branch. This should be the last commit to main
   before cutting the release.
4. Create a new release branch from `main`. The release branch should be named
   `release/v${MAJOR_VERSION}.${MINOR_VERSION}`, e.g. `release/v0.3`.

   ```shell
   git checkout -b release/v0.3
   ```

5. Push the branch to the Envoy Gateway repo.
6. Create a topic branch and update the release tag references in the [Quickstart Guide][].

   ```shell
   make update-quickstart TAG=v0.3.0
   ```

7. Sign, commit, and push your changes to your fork. Send a PR to get your changes merged into the release branch.
   Do not proceed until your PR is merged.
8. Tag the head of your release branch with the release tag. For example:

   ```shell
   git tag -a v0.3.0 -m 'Envoy Gateway v0.3.0 Release'
   ```

   __Note:__ The tag version differs from the release branch by including the `.0` patch version.

9. Push the tag to the Envoy Gateway repository.

    ```shell
    git push v0.3.0
    ```

10. This will trigger the [release GitHub action][] that generates the release, release artifacts, etc.
11. Confirm that the [release workflow][] completed successfully.
12. Confirm that the Envoy Gateway [image][] with the correct release tag was published to Docker Hub.
13. Confirm that the [release][] was created.
14. Confirm that the steps in the [Quickstart Guide][] work as expected.
15. [Generate][] the GitHub changelog.
16. If you find any bugs in this process, please create an issue.

## Creating a Release Candidate

1. Clone the repo, checkout the `main` branch, ensure it’s up-to-date, and your local branch is clean.
2. Create the release notes corresponding to the release candidate that summarizes the changes included in the
   release candidate. Reference previous [release notes][] for additional details.
3. Submit a [Pull Request][] to merge the changelog into the main branch. This should be the last commit to main
   before cutting the release candidate.
4. Tag the head of the main branch with the release candidate number. The tag should be named
   `v${MAJOR_VERSION}.${MINOR_VERSION}.0-rc.${RELEASE_CANDIDATE_NUMBER}`. For example:

   ```shell
   git tag -a v0.3.0-rc.1 -m 'Envoy Gateway v0.3.0-rc.1 Release Candidate'
   ```

5. Push the tag to the Envoy Gateway repository.

   ```shell
   git push v0.3.0-rc.1
   ```

6. This will trigger the [release GitHub action][] that generates the release, release artifacts, etc.
7. Confirm that the [release workflow][] completed successfully.
8. Confirm that the Envoy Gateway [image][] with the correct release tag was published to Docker Hub.
9. Confirm that the [release][] was created.
10. Note that the [Quickstart Guide][] references are __not__ updated for release candidates. However, test
    the quickstart steps using the release candidate by manually updating the links.
11. [Generate][] the GitHub changelog.
12. Ensure you check the "This is a pre-release" checkbox when editing the GitHub release.
13. If you find any bugs in this process, please create an issue.

## Announcing the Release

It's important that the world knows about the release. Use the following steps to announce the release.

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

[release notes]: https://github.com/envoyproxy/gateway/tree/main/release-notes
[Pull Request]: https://github.com/envoyproxy/gateway/pulls
[Quickstart Guide]: https://github.com/envoyproxy/gateway/blob/main/docs/user/quickstart.md
[release GitHub action]: https://github.com/envoyproxy/gateway/blob/main/.github/workflows/release.yaml
[release workflow]: https://github.com/envoyproxy/gateway/actions/workflows/release.yaml
[image]: https://hub.docker.com/r/envoyproxy/gateway/tags
[release]: https://github.com/envoyproxy/gateway/releases
[Generate]: https://docs.github.com/en/repositories/releasing-projects-on-github/automatically-generated-release-notes
