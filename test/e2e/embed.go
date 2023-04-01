// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package e2e

import "embed"

//go:embed testdata/*.yaml base/*
var Manifests embed.FS
