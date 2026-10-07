#!/usr/bin/env bash

# Copyright Envoy Gateway Authors
# SPDX-License-Identifier: Apache-2.0
# The full text of the Apache license is available in the LICENSE file at
# the root of the repo.

# Synchronize a subset of the Gateway API api-types documentation into the
# Envoy Gateway site.
#
# The upstream docs moved from MkDocs (site-src/api-types/*.md) to Hugo
# (site/content/en/reference/api-types/**.md) in Gateway API v1.6, which changed
# both the source paths and the markup (Hugo shortcodes instead of MkDocs
# admonitions and `{% include %}`). This script downloads the files for the
# Gateway API version Envoy Gateway bundles and rewrites them so they render
# with the Envoy Gateway site's own shortcode set.

set -euo pipefail

: "${GATEWAY_API_VERSION:=v1.6.1}"
: "${GATEWAY_API_MINOR_VERSION:=1.6}"
: "${RAW_BASE_URL:=https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/${GATEWAY_API_VERSION}}"
: "${DOC_SRC_PATH:=site/content/en/reference/api-types}"
: "${DOC_DEST_DIR:=site/content/en/latest/api/gateway_api}"
# Paths are relative to DOC_SRC_PATH. Only the basename is kept in DOC_DEST_DIR.
: "${SYNC_FILES:=gateway.md gatewayclass.md httproute.md grpcroute.md referencegrant.md policy/backendtlspolicy.md}"

UPSTREAM_SITE_URL="https://gateway-api.sigs.k8s.io"
SPEC_URL="${UPSTREAM_SITE_URL}/reference/api-spec/${GATEWAY_API_MINOR_VERSION}/spec/"

YAML_TMP_DIR=$(mktemp -d)
trap 'rm -rf "${YAML_TMP_DIR}"' EXIT

mkdir -p "${DOC_DEST_DIR}"

for src_path in ${SYNC_FILES}; do
  file=$(basename "${src_path}")
  dest="${DOC_DEST_DIR}/${file}"
  echo "syncing ${src_path} -> ${dest}"

  # --fail is required: without it curl happily writes a "404: Not Found" body
  # into the destination file when the upstream layout changes.
  curl -sSfL -o "${dest}" "${RAW_BASE_URL}/${DOC_SRC_PATH}/${src_path}"

  # Download every YAML example referenced by a readfile shortcode. The file
  # attribute is repo-root relative, e.g. "/examples/standard/basic-http.yaml".
  while IFS= read -r yaml_path; do
    [ -n "${yaml_path}" ] || continue
    yaml_dest="${YAML_TMP_DIR}/${yaml_path//\//_}"
    [ -f "${yaml_dest}" ] && continue
    echo "  downloading ${yaml_path}"
    curl -sSfL -o "${yaml_dest}" "${RAW_BASE_URL}${yaml_path}"
    # Drop the upstream test harness directives.
    perl -i -ne 'print unless /^#\$/' "${yaml_dest}"
  done < <(perl -ne 'print "$1\n" while /\{\{<\s*readfile\s+file="([^"]+)"/g' "${dest}")

  YAML_TMP_DIR="${YAML_TMP_DIR}" \
  SPEC_URL="${SPEC_URL}" \
  UPSTREAM_SITE_URL="${UPSTREAM_SITE_URL}" \
  GATEWAY_API_MINOR_VERSION="${GATEWAY_API_MINOR_VERSION}" \
  perl -0777 -i -pe '
    # Hugo YAML front matter -> the TOML front matter used across the EG site.
    # The upstream weight is dropped so ordering stays owned by this repo.
    s{\A---\n(.*?)\n---\n}{
      my $fm = $1;
      my ($title) = $fm =~ /^title:\s*"?(.*?)"?\s*$/m;
      "+++\ntitle = \"" . ($title // "") . "\"\n+++\n";
    }se;

    # readfile shortcode -> inline fenced code block.
    s{^\{\{<\s*readfile\s+file="([^"]+)"[^>]*>\}\}[ \t]*\n}{
      (my $key = $1) =~ s|/|_|g;
      # Read with an explicit handle: touching @ARGV/<> here would hijack the
      # -i in-place edit and truncate the file being processed.
      my $body = "";
      if (open my $fh, "<", "$ENV{YAML_TMP_DIR}/$key") {
        local $/;
        $body = <$fh>;
        close $fh;
      }
      $body =~ s/\s+\z//;
      "```yaml\n$body\n```\n";
    }gme;

    # Drop the shortcode wrappers EG does not define, keeping their content.
    s{^\{\{<\s*/?details[^>]*>\}\}[ \t]*\n}{}gm;
    s{^\{\{%\s*/?alert[^%]*%\}\}[ \t]*\n}{}gm;

    # Point the API spec links at the published upstream reference.
    s{(\]\()(?:\.\./)*reference/spec\.md(\))}{$1$ENV{SPEC_URL}$2}g;
    s{(^\[[^\]]+\]:\s*)(?:\.\./|/)reference/spec\.md\s*$}{$1$ENV{SPEC_URL}}gm;

    # Absolute and relative upstream links -> absolute upstream site links.
    s{(\]\()(/(?!/)[^)]*)(\))}{$1$ENV{UPSTREAM_SITE_URL}$2$3}g;
    s{(\]\()((?:\.\./)+[^):]*)(\))}{$1 . $ENV{UPSTREAM_SITE_URL} . "/" . ($2 =~ s{^(?:\.\./)+}{}r) . $3}ge;
    s{(^\[[^\]]+\]:\s*)(/(?!/)\S*)\s*$}{$1$ENV{UPSTREAM_SITE_URL}$2}gm;
    s{(^\[[^\]]+\]:\s*)((?:\.\./)+\S*)\s*$}{$1 . $ENV{UPSTREAM_SITE_URL} . "/" . ($2 =~ s{^(?:\.\./)+}{}r)}gme;

    # Upstream serves extensionless URLs.
    s{(\Q$ENV{UPSTREAM_SITE_URL}\E[^\s)]*)\.md}{$1}g;

    # Pin the API spec references to the bundled Gateway API version instead of
    # following the upstream main branch.
    s{(\Q$ENV{UPSTREAM_SITE_URL}\E/reference/api-spec/)main(/spec)}{$1$ENV{GATEWAY_API_MINOR_VERSION}$2}g;

    # Collapse the blank lines left behind by the removed shortcodes.
    s{\n{3,}}{\n\n}g;
  ' "${dest}"
done

echo "gateway-api docs synced from ${GATEWAY_API_VERSION}"
