---
name: review-envoy-gateway-pr
description: Review an Envoy Gateway pull request for essential API, implementation, status, and test coverage requirements.
metadata:
  short-description: Envoy Gateway PR review workflow
  version: "0.1"
---

# Envoy Gateway PR Review Skill

## Inputs
- GitHub PR URL (e.g. review PR: https://github.com/envoyproxy/gateway/pull/8237), or
- A local diff between commits (e.g. review change: git diff 4927877a HEAD)

## Review
- Check API changes, implementation changes, feature coverage, and release notes as applicable.
- Keep findings concise and actionable, with file references when possible.

## Checklist

### API changes
- Make sure changes under `api/` align with https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
- Make sure changes under `api/` are consistent with existing API patterns in the Gateway API project: https://github.com/kubernetes-sigs/gateway-api/tree/main/apis
- Make sure changes under `api/` are consistent with existing API patterns in this project.
- Try to reuse existing types.
- Keep naming consistent with this project.
- Try to add kubebuilder and CEL validations to catch errors.
- Make sure API validations are tested in `/test/cel-validation`.
- Check backward compatibility for API shape, CRD schema, defaults, versioned structs, and upgrade behavior.
- If a PR mixes API and implementation changes, say whether it should be split into separate PRs.

### Implementation changes
- For changes under `internal/gatewayapi`, check that user-visible errors are surfaced in status.
- For changes under `internal/gatewayapi`, check that status conditions follow the conventions in the Gateway API spec: https://gateway-api.sigs.k8s.io/geps/gep-1364/index.html
- For changes under `internal/gatewayapi`, check that `internal/gatewayapi/testdata` has coverage.
- For changes under `internal/xds/translator`, check that `internal/xds/translator/testdata` has coverage.

### Feature coverage
- For new user-facing features, check that `test/e2e` has coverage.

### Release notes
- Release notes should be added to `release-notes/current.yaml` for any user-facing changes.
- Bug fixes should be noted as "bug fix" and include a brief description of the issue and the fix.
- New features should be noted as "new feature" and include a brief description of the feature.
- Any breaking changes should be noted as "breaking change" and include a clear description of the change and its impact on users.
- Any change to generated Envoy config (xDS) that moves, removes, or modifies existing config content would break EnvoyPatchPolicies and Extension Servers, so it should be noted as a breaking change. Additions to generated xDS config do not need to be called out.
- Existing API changes should be noted.
- Existing behavior changes should be noted.
