# Release notes for the next release

This directory holds release notes for the next release. Document each change by
adding a **new file** under the matching section directory, as described below.

## How to add an entry

1. Pick the section directory that matches your change:

   | Directory                  | Use for                                                                 |
   |----------------------------|-------------------------------------------------------------------------|
   | `breaking_changes/`        | Changes incompatible with previous versions (API deletions/changes).    |
   | `security_updates/`        | Fixes for vulnerabilities or compliance issues.                         |
   | `new_features/`            | New features or capabilities.                                           |
   | `bug_fixes/`               | Bug fixes.                                                              |
   | `performance_improvements/`| Performance improvements.                                               |
   | `deprecations/`            | Features or APIs marked deprecated.                                     |
   | `other_changes/`           | Anything notable not covered above.                                     |

2. Create a file named `<pr-number>-<short-slug>.md`, for example
   `bug_fixes/9241-backend-tls-panic.md`. The PR (or issue) number keeps the
   filename unique so two PRs never collide. The slug is lowercase letters,
   digits, and hyphens.

3. Put the release note as a **single entry** in the file — one change per file.
   Write it as one sentence/paragraph; it becomes one bullet in the compiled
   release notes. Example contents:

   ```
   Fixed the xDS server in GatewayNamespaceMode serving a stale certificate after
   cert-manager rotation by re-reading the cert from disk on every TLS handshake.
   ```

   If your PR has changes in more than one section, add one file per section.

## How these get released

At release time the maintainer runs:

```
make release-notes-gen RELEASE_NOTE_VERSION=vX.Y.Z RELEASE_NOTE_DATE="Month D, YYYY"

# for example:
make release-notes-gen RELEASE_NOTE_VERSION=v1.9.0 RELEASE_NOTE_DATE="June 23, 2026"
```

This compiles every fragment under `current/` into `release-notes/vX.Y.Z.yaml`
(the format consumed by the docs pipeline) and clears the fragment files, leaving
the empty section directories ready for the next development cycle.
