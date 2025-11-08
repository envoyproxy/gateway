#!/usr/bin/env bash
set -euo pipefail

# DOC_DEST_DIR=/path/to/dir YAML_SRC_BASE_URL=... SYNC_FILES="a.md b.md" ./tools/hack/gwapi-doc-download-includes.sh

: "${DOC_DEST_DIR:=site/content/en/latest/api/gateway_api}"
: "${YAML_SRC_BASE_URL:=https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/main/examples}"
: "${SYNC_FILES:=gateway.md gatewayclass.md httproute.md grpcroute.md backendtlspolicy.md referencegrant.md}"
: "${LOG_TARGET:=echo \"[docs]\"}"

GREP=$(command -v grep)
SED=$(command -v sed)
# use ggrep/gsed on macOS if available
if [[ "$(uname)" == "Darwin" ]]; then
  if command -v ggrep &> /dev/null; then
    GREP=$(command -v ggrep)
  fi
  if command -v gsed &> /dev/null; then
    SED=$(command -v gsed)
  fi
fi

for file in $SYNC_FILES; do
  echo "processing: $file"
  src="$DOC_DEST_DIR/$file"
  if [ ! -f "$src" ]; then
    echo "  skip: $src not found"
    continue
  fi

  # Extract the path inside {% include 'path' %} and download each referenced file
  includes=$($GREP -oP "{% include '\K[^']+" "$src" || true)
  for include_path in $includes; do
    filename=$(basename "$include_path")
    dest="$DOC_DEST_DIR/$filename"
    dest_dir=$(dirname "$dest")
    mkdir -p "$dest_dir"
    url="$YAML_SRC_BASE_URL/$include_path"
    echo "downloading $url to $dest"
    curl -sSL -o "$dest" "$url"

    # Remove lines start with `#$`
    $SED -i '/^#\$/d' "$dest"
  done
done
