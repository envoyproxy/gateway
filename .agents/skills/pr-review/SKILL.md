---
name: review-envoy-gateway-pr
description: Review an Envoy Gateway pull request using a consistent, severity-ordered checklist. Input a PR URL (preferred) or diff + context.
metadata:
  short-description: Envoy Gateway PR review workflow
  version: "1.0"
---

# Envoy Gateway PR Review Skill

## Inputs
- GitHub PR URL (e.g. review PR: https://github.com/envoyproxy/gateway/pull/8237), or
- A local diff between commits (e.g. review change: git diff 4927877a HEAD)

## Workflow
1. **Collect context (if PR URL provided)**
   Use `gh` to fetch PR metadata + diff. Include:
   - title/body, labels, base/head branches
   - file list, commits, additions/deletions
   - full diff (or at least relevant hunks)
   - any CI failures, coverage notes, or reviewer guidance from the author

   Example commands:
   - `gh pr view <url|number> -R envoyproxy/gateway --json number,title,author,body,files,commits,additions,deletions,changedFiles,baseRefName,headRefName,labels,reviewRequests,url`
   - `gh pr diff <url|number> -R envoyproxy/gateway`

2. **Review and report**
   - Keep it concise and actionable.
   - Order findings by **severity**: `Blocker > High > Medium > Low > Nit`.
   - Include **file path + line numbers** where possible (or closest diff hunk anchors).
   - Prefer **concrete suggestions** (what to change / what test to add / what invariant to document).

3. **Explicitly list assumptions and open questions**
   - Call out any missing context required to be confident.

---

## Review Rubric

### A) Core Review (required)
You are a coding agent reviewing the Envoy Gateway pull request using the supplied context and diff.

**Priorities**
1. **Correctness & safety**: logic, edge cases, reconcilers, translation, status updates
2. **API compatibility**: Gateway API behavior, CRD changes/defaults, backward compatibility
3. **Operational risk**: upgrade safety, rollout behavior, managed dataplane stability
4. **Security**: RBAC, secrets, SSRF, header/route escaping, admin exposure
5. **Performance**: reconciler churn, queue load, xDS generation cost, memory usage, cpu usage
6. **Tests & docs**: coverage quality, conformance gaps, docs/examples/release notes

**Output format**
- Findings grouped by severity with bullets:
  - **[Severity]** summary — *impact* — *recommendation* — *file:line*
- Then:
  - **Open questions**
  - **Assumptions made**
  - **Suggested tests/docs**

---

## B) Targeted Component Deep-Dive (run if applicable)
Perform a focused review on the most impacted component(s) based on file changes (or user-specified focus).

Focus on:
- Reconciler invariants: informer lifetimes, goroutine ownership, context cancellation, retries/backoff
- K8s object semantics: owner refs, label selectors, immutable fields, rollout behavior
- Gateway API / CRD compatibility: defaulting, validation, conversion, version guards
- xDS / translation correctness: IR generation, ExtensionManager integration, Envoy config consistency
- Multi-namespace and multi-gateway edge cases
- “Managed dataplane friendliness”: hot-restart expectations, no disruptive changes by default

Return concrete findings + questions with file/line anchors.

---

## C) Tests & Docs Completeness Audit (required)
Inspect the PR for completeness.

Tasks:
- Identify missing/weak tests (unit/integration/e2e/conformance) for:
  - happy paths and failure paths
  - multi-namespace references
  - upgrade/downgrade and defaulting behavior (if API/CRD touched)
  - regression coverage for the reported bug (if applicable)
- Check docs/examples/release notes:
  - reference docs for changed behavior
  - example manifests
  - mention behavior changes in release notes if user-visible

Deliver actionable bullets with file references.

---

## D) Last-Mile Checklist (required)
Return **pass/fail/unknown** for each item:

1. No debug artifacts left
2. CRDs + generated files updated as needed (DeepCopy, manifests, schema)
3. Ownership/concurrency handled correctly (goroutines, informers, channels, context cancellation)
4. Status conditions follow Gateway API expectations:
   - correct condition types/reasons
   - correct use of ObservedGeneration
   - aligns with Gateway API guidance (e.g., GEP-1364 where relevant)
5. RBAC/PodSecurity implications are correct and minimal
6. CI/test commands likely to pass locally (`go test ./...`, lint, etc.)
7. API naming consistency + validation markers present (if API changed)
8. Do not introduce unnecessary changes that are not related to the PR

---

## Constraints
- Follow Envoy Gateway CONTRIBUTING and compatibility expectations:
  - https://github.com/envoyproxy/gateway/blob/main/CONTRIBUTING.md
- Ensure behavior aligns with Gateway API spec:
  - https://gateway-api.sigs.k8s.io/
- Do not use Map in the internal/ir/xds.go file.
- Stay concise; don’t summarize the PR unless needed to justify a finding.
- If file/line numbers aren’t available, reference the closest diff hunk and filename.

---

## Completion Statement (required)
At the end, explicitly confirm completion of:
- A) Core Review
- B) Targeted Deep-Dive (or “skipped: not applicable”)
- C) Tests & Docs Audit
- D) Last-Mile Checklist

And list the resulting findings.
