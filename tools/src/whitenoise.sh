#!/bin/bash
TRAILING_WHITESPACE=0
# only check changed docs in pr
git fetch
for file in $(git diff --cached --name-only --diff-filter=ACRMTU origin/$GITHUB_BASE_REF| grep "\.md"); do
  if grep -r '[[:blank:]]$' "$1" > /dev/null; then
    echo "trailing whitespace: ${1}" >&2
    ERRORS=yes
    ((TRAILING_WHITESPACE=TRAILING_WHITESPACE+1))
  fi
done
if [[ -n "$ERRORS" ]]; then
  echo >&2
  echo "ERRORS found" >&2
  echo "${TRAILING_WHITESPACE} files with trailing whitespace" >&2
  echo >&2
  exit 1
fi
