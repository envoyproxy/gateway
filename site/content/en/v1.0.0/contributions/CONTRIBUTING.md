---
title: "Contributing"
description: "This section tells how to contribute to Envoy Gateway."
weight: 3
---

We welcome contributions from the community. Please carefully review the [project goals](/about)
and following guidelines to streamline your contributions.

## Communication

* Before starting work on a major feature, please contact us via GitHub or Slack. We will ensure no
  one else is working on it and ask you to open a GitHub issue.
* A "major feature" is defined as any change that is > 100 LOC altered (not including tests), or
  changes any user-facing behavior. We will use the GitHub issue to discuss the feature and come to
  agreement. This is to prevent your time being wasted, as well as ours. The GitHub review process
  for major features is also important so that [affiliations with commit access](./codeowners) can
  come to agreement on the design. If it's appropriate to write a design document, the document must
  be hosted either in the GitHub issue, or linked to from the issue and hosted in a world-readable
  location.
* Small patches and bug fixes don't need prior communication.

## Inclusivity

The Envoy Gateway community has an explicit goal to be inclusive to all. As such, all PRs must adhere
to the following guidelines for all code, APIs, and documentation:

* The following words and phrases are not allowed:
  * *Whitelist*: use allowlist instead.
  * *Blacklist*: use denylist or blocklist instead.
  * *Master*: use primary instead.
  * *Slave*: use secondary or replica instead.
* Documentation should be written in an inclusive style. The [Google developer
  documentation](https://developers.google.com/style/inclusive-documentation) contains an excellent
  reference on this topic.
* The above policy is not considered definitive and may be amended in the future as industry best
  practices evolve. Additional comments on this topic may be provided by maintainers during code
  review.

## Submitting a PR

* Fork the repo.
* Hack
* DCO sign-off each commit. This can be done with `git commit -s`.
* Submit your PR.
* Tests will automatically run for you.
* We will **not** merge any PR that is not passing tests.
* PRs are expected to have 100% test coverage for added code. This can be verified with a coverage
  build. If your PR cannot have 100% coverage for some reason please clearly explain why when you
  open it.
* Any PR that changes user-facing behavior **must** have associated documentation in the [docs](https://github.com/envoyproxy/gateway/tree/main/site) folder of the repo as
  well as the [changelog](../releases).
* All code comments and documentation are expected to have proper English grammar and punctuation.
  If you are not a fluent English speaker (or a bad writer ;-)) please let us know and we will try
  to find some help but there are no guarantees.
* Your PR title should be descriptive, and generally start with type that contains a subsystem name with `()` if necessary 
  and summary followed by a colon. format `chore/docs/feat/fix/refactor/style/test: summary`.
  Examples:
  * "docs: fix grammar error"
  * "feat(translator): add new feature"
  * "fix: fix xx bug"
  * "chore: change ci & build tools etc"
* Your PR commit message will be used as the commit message when your PR is merged. You should
  update this field if your PR diverges during review.
* Your PR description should have details on what the PR does. If it fixes an existing issue it
  should end with "Fixes #XXX".
* If your PR is co-authored or based on an earlier PR from another contributor,
  please attribute them with `Co-authored-by: name <name@example.com>`. See
  GitHub's [multiple author
  guidance](https://help.github.com/en/github/committing-changes-to-your-project/creating-a-commit-with-multiple-authors)
  for further details.
* When all tests are passing and all other conditions described herein are satisfied, a maintainer
  will be assigned to review and merge the PR.
* Once you submit a PR, *please do not rebase it*. It's much easier to review if subsequent commits
  are new commits and/or merges. We squash and merge so the number of commits you have in the PR
  doesn't matter.
* We expect that once a PR is opened, it will be actively worked on until it is merged or closed.
  We reserve the right to close PRs that are not making progress. This is generally defined as no
  changes for 7 days. Obviously PRs that are closed due to lack of activity can be reopened later.
  Closing stale PRs helps us to keep on top of all the work currently in flight.

## Maintainer PR Review Policy

* See [CODEOWNERS.md](../codeowners) for the current list of maintainers.
* A maintainer representing a different affiliation from the PR owner is required to review and
  approve the PR.
* When the project matures, it is expected that a "domain expert" for the code the PR touches should
  review the PR. This person does not require commit access, just domain knowledge.
* The above rules may be waived for PRs which only update docs or comments, or trivial changes to
  tests and tools (where trivial is decided by the maintainer in question).
* If there is a question on who should review a PR please discuss in Slack.
* Anyone is welcome to review any PR that they want, whether they are a maintainer or not.
* Please make sure that the PR title, commit message, and description are updated if the PR changes
  significantly during review.
* Please **clean up the title and body** before merging. By default, GitHub fills the squash merge
  title with the original title, and the commit body with every individual commit from the PR.
  The maintainer doing the merge should make sure the title follows the guidelines above and should
  overwrite the body with the original commit message from the PR (cleaning it up if necessary)
  while preserving the PR author's final DCO sign-off.

## Decision making

This is a new and complex project, and we need to make a lot of decisions very quickly.
To this end, we've settled on this process for making (possibly contentious) decisions:

* For decisions that need a record, we create an issue.
* In that issue, we discuss opinions, then a maintainer can call for a vote in a comment.
* Maintainers can cast binding votes on that comment by reacting or replying in another comment.
* Non-maintainer community members are welcome to cast non-binding votes by either of these methods.
* Voting will be resolved by simple majority.
* In the event of deadlocks, the question will be put to steering instead.

## DCO: Sign your work

The sign-off is a simple line at the end of the explanation for the
patch, which certifies that you wrote it or otherwise have the right to
pass it on as an open-source patch. The rules are pretty simple: if you
can certify the below (from
[developercertificate.org](https://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
660 York Street, Suite 102,
San Francisco, CA 94110 USA

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

then you just add a line to every git commit message:

    Signed-off-by: Joe Smith <joe@gmail.com>

using your real name (sorry, no pseudonyms or anonymous contributions.)

You can add the sign-off when creating the git commit via `git commit -s`.

If you want this to be automatic you can set up some aliases:

```bash
git config --add alias.amend "commit -s --amend"
git config --add alias.c "commit -s"
```

## Fixing DCO

If your PR fails the DCO check, it's necessary to fix the entire commit history in the PR. Best
practice is to [squash](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/incorporating-changes-from-a-pull-request/about-pull-request-merges#squash-and-merge-your-commits)
the commit history to a single commit, append the DCO sign-off as described above, and [force
push](https://git-scm.com/docs/git-push#git-push---force). For example, if you have 2 commits in
your history:

```bash
git rebase -i HEAD^^
(interactive squash + DCO append)
git push origin -f
```

Note, that in general rewriting history in this way is a hindrance to the review process and this
should only be done to correct a DCO mistake.
