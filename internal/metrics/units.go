// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

// Unit encodes the standard name for describing the quantity
// measured by a Metric (if applicable).
type Unit string

// Predefined units for use with the metrics package.
const (
	None         Unit = "1"
	Bytes        Unit = "By"
	Seconds      Unit = "s"
	Milliseconds Unit = "ms"
)
