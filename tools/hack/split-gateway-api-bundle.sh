#!/usr/bin/env sh

set -eu

if [ "$#" -ne 3 ]; then
  echo "usage: $0 <input> <crds-output> <validating-admission-policy-output>" >&2
  exit 1
fi

input=$1
crds_output=$2
vap_output=$3
script_dir=$(dirname "$0")
supporting_resources_header=$script_dir/gatewayapi-supporting-resources-header.tpl

tmp_crds=$(mktemp)
tmp_vap=$(mktemp)
trap 'rm -f "$tmp_crds" "$tmp_vap"' EXIT

awk -v crds="$tmp_crds" -v vap="$tmp_vap" '
function flush_doc() {
  if (doc == "") {
    return
  }

  if (doc ~ /\nkind:[[:space:]]*CustomResourceDefinition[[:space:]]*(\n|$)/ || doc !~ /\nkind:[[:space:]]*/) {
    printf "%s", doc >> crds
  } else {
    printf "%s", doc >> vap
  }

  doc = ""
}

/^---[[:space:]]*$/ {
  flush_doc()
}

{
  doc = doc $0 "\n"
}

END {
  flush_doc()
}
' "$input"

mv "$tmp_crds" "$crds_output"

{
  cat "$supporting_resources_header"
  cat "$tmp_vap"
  echo '{{- end }}'
  echo '{{- end }}'
} > "$vap_output"
