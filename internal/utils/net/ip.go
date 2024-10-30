// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package net

import (
	"net"
	"os"
)

func IsIPv6(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}
	return ip.To4() == nil
}

// IsIPv6Pod returns true if the POD_IP environment variable is an IPv6 address.
// WARNING: This function is only intended to be used in the context of Kubernetes.
func IsIPv6Pod() bool {
	return IsIPv6(os.Getenv("POD_IP"))
}
