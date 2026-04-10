---
name: review-envoy-gateway-pr
description: Review an Envoy Gateway pull request for essential API, implementation, status, and test coverage requirements.
metadata:
  short-description: Envoy Gateway PR review workflow
  version: "1.0"
---

# Envoy Gateway PR Review Skill

## Inputs
- GitHub PR URL (e.g. review PR: https://github.com/envoyproxy/gateway/pull/8237), or
- A local diff between commits (e.g. review change: git diff 4927877a HEAD)

## Review
- Check correctness, API compatibility, status behavior, and test coverage.
- Keep findings concise and actionable, with file references when possible.

## API Changes
- make sure they align with https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
- make sure they are consistent with existing API patterns in Gateway API project: https://github.com/kubernetes-sigs/gateway-api/tree/main/apis
- make sure they are consistent with existing API patterns in this project,
- try to add kubebuilder and CEL validations to catch errors
- make sure these are tested in /test/cel-validation
- split up API and implementation in separate PRs
- try and reuse existing types
- keep naming consistent to this project
- Backward compatibility for API shape, CRD schema, defaults, versioned structs, and upgrade behavior.

## Implementation Changes
- For changes under `internal/gatewayapi`, check that user-visible errors are surfaced in status.
- For changes under `internal/gatewayapi`, check that status conditions follow the conventions in Gateway API spec: https://gateway-api.sigs.k8s.io/geps/gep-1364/index.html
- For changes under `internal/gatewayapi`, check that `internal/gatewayapi/testdata` has coverage.
- For changes under `internal/xds/translator`, check that `internal/xds/translator/testdata` has coverage.

## Feature Coverage
- For new user-facing features, check that `test/e2e` has coverage.

## Relese Notes
- Release notes should be added to release-notes/current.yaml for any user-facing changes.
  * Bug fixes should be noted as "bug fix" and include a brief description of the issue and the fix.
  * New features should be noted as "new feature".
  * Any breaking changes should be noted as "breaking change" and include a clear description of the change and its impact on users.
    - existing xDS change(move/remove/modify, not include additional) would break EnvoyPatchPolicies and Extension Servers.
    - existing API changes.
    - exiting behavior changes.
