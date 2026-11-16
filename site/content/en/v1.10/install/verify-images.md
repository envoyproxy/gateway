---
title: "Verify Image Signatures"
weight: -60
---

The release workflow signs the Envoy Gateway image `docker.io/envoyproxy/gateway` with [cosign][] keyless signing.
Keyless signing uses the GitHub Actions OIDC identity of the release workflow, so there is no private key to manage.
The signature is recorded in the [Rekor][] public transparency log.

Only releases published after signing was added have a signature.

## Prerequisites

Install [cosign][cosign-install] v3.0 or later.

## Verify the image

Set `VERSION` to the release you want to check, then run `cosign verify` with the identity of the release workflow for that tag:

```shell
VERSION=vX.Y.Z
cosign verify "docker.io/envoyproxy/gateway:${VERSION}" \
  --certificate-identity "https://github.com/envoyproxy/gateway/.github/workflows/release.yaml@refs/tags/${VERSION}" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
```

The command fails if the image has no valid signature from that workflow and tag.
The signature is on the multi-arch image index, so verify the tag or the index digest, not a per-platform digest.

[cosign]: https://docs.sigstore.dev/cosign/signing/overview/
[cosign-install]: https://docs.sigstore.dev/cosign/system_config/installation/
[Rekor]: https://docs.sigstore.dev/logging/overview/
