#!/bin/bash

REPO=${REPO:-"envoyproxy/gateway"}
LABEL=${LABEL:-"cherrypick/release-v1.5.1"}
DRY_RUN=${DRY_RUN:-true}
# SKIP specific PR numbers, e.g. "123 456" after resolved conflicts.
SKIP_PR_NUMBERS=${SKIP_PR_NUMBERS:-""}

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is not installed, exiting."
  exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is not installed, exiting."
  exit 1
fi

if [[ -z "$GITHUB_TOKEN" ]]; then
  echo "GITHUB_TOKEN not found, exiting."
  exit 1
fi

# ensure we are in a git repo and not on the main branch
if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "Not inside a git repository, exiting."
  exit 1
fi

CURRENT_BRANCH=$(git symbolic-ref --short HEAD)
if [[ "$CURRENT_BRANCH" == "main" || "$CURRENT_BRANCH" == "master" ]]; then
  echo "On main or master branch, please switch to a feature or release branch."
  exit 1
fi

# current branch should create from the target release branch
# e.g. for LABEL=cherrypick/release-v1.5.1
# git checkout release/v1.5
# git checkout -b cherry-pick/v1.5.1
BRANCH_SUFFIX=${LABEL#cherrypick/release-}
TARGET_BRANCH="cherry-pick/$BRANCH_SUFFIX"
if [[ "$CURRENT_BRANCH" != "$TARGET_BRANCH" ]]; then
  echo "Current branch $CURRENT_BRANCH does not match $TARGET_BRANCH, exiting."
  exit 1
fi

# get all merged PR numbers with the specified label, sorted by merged date ascending
PR_NUMBERS=($(curl -s -H "Authorization: token $GITHUB_TOKEN" \
  "https://api.github.com/repos/$REPO/pulls?state=closed&per_page=100" | \
  jq -r ".[] | select(.labels[].name==\"$LABEL\" and .merged_at!=null) | {number, merged_at}" | \
  jq -s "sort_by(.merged_at) | .[].number"))

echo "Total PRs to cherry-pick: ${#PR_NUMBERS[@]}"
echo "PR_NUMBERS: ${PR_NUMBERS[@]}"

# cherry-pick each PR by its merge commit SHA
for PR in "${PR_NUMBERS[@]}"; do
  if [[ "$SKIP_PR_NUMBERS" == *"$PR"* ]]; then
    echo "Skipping PR #$PR as it is in SKIP_PR_NUMBERS"
    continue
  fi

  SHA=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
    "https://api.github.com/repos/$REPO/pulls/$PR" | \
    jq -r ".merge_commit_sha")

  echo "PR #$PR merge commit SHA: $SHA"
  if [[ -z "$SHA" || "$SHA" == "null" ]]; then
    echo "PR #$PR merge commit $SHA not found"
    exit 1
  fi
  echo "Cherry-pick PR #$PR commit $SHA"
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "Dry run: git cherry-pick PR #$PR with SHA: $SHA"
    continue
  fi
  git cherry-pick "$SHA"
  make build
  if [[ $? -ne 0 ]]; then
    echo "Cherry-pick $SHA failed, please resolve conflicts and run the script again."
    exit 1
  fi
done
