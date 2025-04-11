#!/bin/bash

# Copyright Envoy Gateway Authors
# SPDX-License-Identifier: Apache-2.0
# The full text of the Apache license is available in the LICENSE file at
# the root of the repo.

# This file contains code derived from Istio, https://github.com/istio/istio
# from the source file https://github.com/istio/istio/blob/master/tests/fuzz/oss_fuzz_build.sh
# and is provided here subject to the following: Copyright Istio Authors SPDX-License-Identifier: Apache-2.0

# Required by `compile_native_go_fuzzer`
# Ref: https://google.github.io/oss-fuzz/getting-started/new-project-guide/go-lang/#buildsh
cd "$SRC"
git clone --depth=1 https://github.com/AdamKorcz/go-118-fuzz-build --branch=include-all-test-files
cd go-118-fuzz-build
go build .
mv go-118-fuzz-build /root/go/bin/

cd "$SRC"/gateway

set -o nounset
set -o pipefail
set -o errexit
set -x

# Create empty file that imports "github.com/AdamKorcz/go-118-fuzz-build/testing"
# This is a small hack to install this dependency, since it is not used anywhere,
# and Go would therefore remove it from go.mod once we run "go mod tidy && go mod vendor".
printf "package envoygateway\nimport _ \"github.com/AdamKorcz/go-118-fuzz-build/testing\"\n" > register.go
go mod edit -replace github.com/AdamKorcz/go-118-fuzz-build="$SRC"/go-118-fuzz-build
go mod tidy


# compile native-format fuzzers
compile_native_go_fuzzer github.com/envoyproxy/gateway/test/fuzz FuzzGatewayAPIToXDS FuzzGatewayAPIToXDS
compile_native_go_fuzzer github.com/envoyproxy/gateway/test/fuzz FuzzGatewayAPIToXDSWithGatewayClass FuzzGatewayAPIToXDSWithGatewayClass

