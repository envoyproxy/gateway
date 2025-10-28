#!/usr/bin/env bash

# Copyright Envoy Gateway Authors
# SPDX-License-Identifier: Apache-2.0
# The full text of the Apache license is available in the LICENSE file at
# the root of the repo.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..
cd "${SCRIPT_ROOT}"

# Generate clientset, informers, and listers for v1alpha1 API
echo "Generating Kubernetes clients, informers, and listers..."

# Use the code-generator from tools/go.mod
GO_TOOL="go tool -modfile=tools/go.mod"

# client-gen: Generate typed clientsets
echo "Running client-gen..."
${GO_TOOL} client-gen \
  --clientset-name versioned \
  --input-base "" \
  --input github.com/envoyproxy/gateway/api/v1alpha1 \
  --go-header-file "${SCRIPT_ROOT}/tools/boilerplate/boilerplate.generatego.txt" \
  --output-pkg github.com/envoyproxy/gateway/pkg/client/clientset \
  --output-dir "${SCRIPT_ROOT}/pkg/client/clientset" \
  --plural-exceptions "EnvoyProxy:EnvoyProxies" \
  github.com/envoyproxy/gateway/api/v1alpha1

# lister-gen: Generate listers
echo "Running lister-gen..."
${GO_TOOL} lister-gen \
  --go-header-file "${SCRIPT_ROOT}/tools/boilerplate/boilerplate.generatego.txt" \
  --output-pkg github.com/envoyproxy/gateway/pkg/client/listers \
  --output-dir "${SCRIPT_ROOT}/pkg/client/listers" \
  --plural-exceptions "EnvoyProxy:EnvoyProxies" \
  github.com/envoyproxy/gateway/api/v1alpha1

# informer-gen: Generate informers
echo "Running informer-gen..."
${GO_TOOL} informer-gen \
  --versioned-clientset-package github.com/envoyproxy/gateway/pkg/client/clientset/versioned \
  --listers-package github.com/envoyproxy/gateway/pkg/client/listers \
  --go-header-file "${SCRIPT_ROOT}/tools/boilerplate/boilerplate.generatego.txt" \
  --output-pkg github.com/envoyproxy/gateway/pkg/client/informers \
  --output-dir "${SCRIPT_ROOT}/pkg/client/informers" \
  --plural-exceptions "EnvoyProxy:EnvoyProxies" \
  github.com/envoyproxy/gateway/api/v1alpha1

echo "Code generation complete!"

