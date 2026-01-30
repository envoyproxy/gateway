// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build ignore

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

//nolint:unused

// Specifically block type aliases that point to maps in internal/ir package
func forbidMapAlias(m dsl.Matcher) {
	// Matches: type MyAlias = map[string]int
	m.Match(`type $_ = map[$_]$_`).
		Where(m.File().PkgPath.Matches(`internal/ir($|/)`)).
		Report("Do not alias maps in internal/ir package.")
}

// Block any usage of the built-in map keyword in internal/ir package
func forbidMap(m dsl.Matcher) {
	// Matches map[key]value anywhere in the code
	m.Match(`map[$_]$_`).
		Where(m.File().PkgPath.Matches(`internal/ir($|/)`)).
		Report("Built-in maps are forbidden in internal/ir package.")
}
