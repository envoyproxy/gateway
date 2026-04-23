#!/bin/bash

# Copyright Envoy Gateway Authors
# SPDX-License-Identifier: Apache-2.0
# The full text of the Apache license is available in the LICENSE file at
# the root of the repo.

# This file contains code derived from Istio, https://github.com/istio/istio
# from the source file https://github.com/istio/istio/blob/master/tests/fuzz/oss_fuzz_build.sh
# and is provided here subject to the following: Copyright Istio Authors SPDX-License-Identifier: Apache-2.0

# Using `compile_native_go_fuzzer_v2`
# Ref: https://github.com/google/oss-fuzz/pull/13835

cd "$SRC"/gateway

set -o nounset
set -o pipefail
set -o errexit
set -x

# compile native-format fuzzers
compile_native_go_fuzzer_v2 github.com/envoyproxy/gateway/test/fuzz FuzzGatewayAPIToXDS FuzzGatewayAPIToXDS

# add seed corpus
zip -j $OUT/FuzzGatewayAPIToXDS_seed_corpus.zip "$SRC"/gateway/test/fuzz/testdata/FuzzGatewayAPIToXDS/*
