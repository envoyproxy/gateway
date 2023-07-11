#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

gitroot=$(git rev-parse --show-toplevel)
#Install latest helm-docs cli
go install github.com/norwoodj/helm-docs/cmd/helm-docs@latest

#Update the README.md file with the latest helm chart documentation
helm-docs charts/gateway-helm/ -f values.tmpl.yaml -o api.md
mv "$gitroot"/charts/gateway-helm/api.md "$gitroot"/docs/latest/helm/api.md