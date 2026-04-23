// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package console

import (
	"embed"
)

// Embed static files and templates
//
//go:embed static
var staticFiles embed.FS

//go:embed templates
var templateFiles embed.FS
