#!/usr/bin/env bash
# Ensure all GitHub Actions workflow jobs use 'ubuntu-latest' for runs-on.
# Pinning to a specific version (e.g. ubuntu-latest) causes jobs to fall behind
# and should be avoided.

set -euo pipefail

if matches=$(
  grep -rn --include="*.yml" --include="*.yaml" \
    'runs-on:' .github/workflows \
  | grep -Ev "runs-on:[[:space:]]*['\"]?ubuntu-latest['\"]?$"
); then
  echo "ERROR: found workflows not using 'ubuntu-latest':"
  echo "$matches"
  exit 1
fi

echo "OK: all workflow jobs use 'ubuntu-latest'"
