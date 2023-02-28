// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package net

import (
	"fmt"
	"net"
)

const (
	DefaultLocalAddress = "localhost"
)

func LocalAvailablePort() (int, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:0", DefaultLocalAddress))
	if err != nil {
		return 0, err
	}

	return l.Addr().(*net.TCPAddr).Port, l.Close()
}
