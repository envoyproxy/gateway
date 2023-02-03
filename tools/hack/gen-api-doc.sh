#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOPATH=${GOPATH:-$(go env GOPATH)}

readonly HERE=$(cd $(dirname $0) && pwd)
readonly REPO=$(cd ${HERE}/../.. && pwd)

# Exec the doc generator.
gendoc() {
    local readonly CFG_DIR="${REPO}/tools/api-docs"

    tools/bin/gen-crd-api-reference-docs \
        -template-dir ${CFG_DIR} \
        -config ${CFG_DIR}/config.json \
        "$@"
}

gendoc \
    -api-dir "github.com/envoyproxy/gateway/api/v1alpha1/" \
    -out-file "docs/latest/api-references/gateway.v1alpha1.md"

gendoc \
    -api-dir "github.com/envoyproxy/gateway/api/config/v1alpha1" \
    -out-file "docs/latest/api-references/config.v1alpha1.md"