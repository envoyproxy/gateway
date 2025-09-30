// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"strconv"
	"sync/atomic"
)

// Version uniquely identifies a watchable message update as it flows between runners.
// The zero value represents an unspecified version.
type Version uint64

var globalVersionCounter atomic.Uint64

// NextVersion returns the next monotonically increasing Version value.
func NextVersion() Version {
	return Version(globalVersionCounter.Add(1))
}

// Versioned marks types that can expose the version associated with them.
type Versioned interface {
	MessageVersion() Version
}

// String implements fmt.Stringer for Version.
func (v Version) String() string {
	if v == 0 {
		return ""
	}
	return strconv.FormatUint(uint64(v), 10)
}
