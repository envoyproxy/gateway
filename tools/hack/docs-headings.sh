#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Wrap sed to deal with GNU and BSD sed flags.
run::sed() {
  if sed --version </dev/null 2>&1 | grep -q GNU; then
    # GNU sed
    sed -i "$@"
  else
    # assume BSD sed
    sed -i '' "$@"
  fi
}

files=(docs/latest/api/config_types.md docs/latest/api/extension_types.md)

# Required since Sphinx mst does not link to h4 headings.
for file in "${files[@]}" ; do
  run::sed \
    "-es|####|##|" \
    "$file"
  echo "updated markdown headings for $file"
done
