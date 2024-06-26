#!/usr/bin/env bash
# shellcheck disable=SC2044

# This script checks that all aliases in the documentation are set up.

readonly CONTENT_DIR=${CONTENT_DIR:-"site/content/en/latest/tasks"}

FAILED=0

red='\e[0;31m'
clr='\e[0m'

error() {
  echo -e "${red}$*${clr}"
}


for file in $(find "${CONTENT_DIR}" -type f -name '*.md' \( ! \( -name '_index.md' \) \) ); do
   if ! grep -Hn '^aliases:' "$file" ; then
        error "Aliases not found in $file"
        FAILED=1
     fi
done

if [[ ${FAILED} -eq 1 ]]; then
    error "LINTING ALIASES FAILED"
    exit 1
fi