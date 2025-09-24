#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OWNERS_FILE="$REPO_ROOT/OWNERS"
CODEOWNERS_FILE="$REPO_ROOT/site/content/en/contributions/CODEOWNERS.md"

maintainers=()
emeritus=()
current_section=""

while IFS= read -r line; do
    line="${line#"${line%%[![:space:]]*}"}" # trim leading whitespace
    if [[ "$line" == "maintainers:" ]]; then
        current_section="maintainers"
        continue
    elif [[ "$line" == "emeritus-maintainers:" ]]; then
        current_section="emeritus"
        continue
    elif [[ "$line" =~ ^[a-zA-Z0-9_-]+: ]]; then
        current_section=""
        continue
    fi

    if [[ "$line" == -* ]]; then
        entry="${line#- }"
        if [[ "$current_section" == "maintainers" ]]; then
            maintainers+=("$entry")
        elif [[ "$current_section" == "emeritus" ]]; then
            emeritus+=("$entry")
        fi
    fi
done < "$OWNERS_FILE"

# Remove emeritus from maintainers
for e in "${emeritus[@]}"; do
    maintainers=("${maintainers[@]/$e}")
done

# Output to CODEOWNERS.md
{
    echo "---"
    echo 'title: "Maintainers"'
    echo 'description: "This section includes Maintainers of Envoy Gateway."'
    echo "---"
    echo
    echo "## The following maintainers, listed in alphabetical order, own everything"
    echo
    for m in $(printf "%s\n" "${maintainers[@]}" | sort -f); do
        echo "- @$m"
    done
    echo
    echo "## Emeritus Maintainers"
    echo
    for e in $(printf "%s\n" "${emeritus[@]}" | sort -f); do
        echo "- @$e"
    done
} > "$CODEOWNERS_FILE"

echo "✅ Synced maintainers into $CODEOWNERS_FILE"