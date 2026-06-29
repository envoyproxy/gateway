<!--
Your PR title should be descriptive, and generally start with a type that contains a subsystem name with `()` if necessary
and a summary followed by a colon. format `chore/docs/api/feat/fix/refactor/style/test: summary`.
Examples:
* "docs: fix grammar error"
* "feat(translator): add new feature"
* "fix: fix xx bug"
* "chore: change ci & build tools etc"
* "api: add xxx fields in ClientTrafficPolicy"
-->

<!--
NOTE: If your PR contains any API changes (changes under `/api`), the API must be discussed and
agreed before the implementation. We strongly recommend separating API changes into their own PR so
we can review the API first, but the API may also live in the same PR as long as it was agreed
beforehand. This will save you a lot of implementation time if the API gets accepted.
-->

**What this PR does / why we need it**:
<!--
Briefly describe what this PR changes and the motivation behind it. Include enough context for a
reviewer to understand the problem being solved and the approach taken, e.g. what behavior changes,
any alternatives considered, and anything reviewers should pay special attention to.
-->

**Which issue(s) this PR fixes**:
<!--
*Automatically closes linked issue when PR is merged.
Usage: `Fixes #<issue number>`, or `Fixes (paste link of issue)`.
-->
Fixes #

---

**PR Checklist**
<!--
Please tick the boxes below before requesting a review. PRs that leave required items unchecked may
be delayed or closed. Replace `[ ]` with `[x]` to check a box. If an item does not apply, check it
and add "N/A: <reason>".
-->

- [ ] **Authorship & ownership**: Coding agents / AI assistants are welcome, but I have reviewed every change, understand how and why it works, can explain and maintain it, and take full responsibility for this PR. I have not submitted generated output I do not understand.
- [ ] **DCO**: All commits are signed off (`git commit -s`). See [DCO: Sign your work](https://gateway.envoyproxy.io/community/contributing/#dco-sign-your-work).
- [ ] **API agreed first**: If this PR contains API changes (changes under `/api`), the API was discussed and agreed **before** the implementation. The API change can be in a separate PR, or in the same PR, but the API must be agreed before implementation. N/A if this PR does not contain API changes.
- [ ] **Required checks pass**: `make generate gen-check`, `make lint`, and the unit-test/coverage build pass. (Flaky e2e failures are not considered breakages, but `gen-check`, `lint`, and coverage **MUST** pass.)
- [ ] **Tests added/updated**: New/changed code is covered by appropriate tests. N/A if this PR does not contain code changes.
- [ ] **Docs**: User-facing changes update the [docs](https://github.com/envoyproxy/gateway/tree/main/site), either in this PR or a follow-up PR. N/A if this PR does not contain user-facing changes.
- [ ] **Release notes**: For any non-trivial change, added a release-note fragment under `release-notes/current/<section>/<pr-number>-<slug>.md` (see `release-notes/current/README.md` for sections and naming). N/A if this PR does not contain non-trivial changes.
- [ ] **Generated files committed**: Ran `make gen-check` and committed the result if API/helm charts/modules changed.
- [ ] **Scope & compatibility**: The PR is reasonably scoped (no unrelated changes) and preserves backward compatibility, or any breaking change is called out above and documented in `release-notes/current/breaking_changes/`.
- [ ] **Codex review**: Requested a Codex review and addressed all of its comments.
- [ ] **Copilot review**: Requested a Copilot review and addressed all of its comments.
