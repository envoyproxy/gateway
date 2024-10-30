// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package net

import (
	"net"
	"os"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

const (
	IPv4ListenerAddress = "0.0.0.0"
	IPv6ListenerAddress = "::"
)

func IsIPv6(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}
	return ip.To4() == nil
}

// IsIPv6FirstPod returns true if the POD_IP environment variable is an IPv6 address.
// WARNING: This function is only intended to be used in the context of Kubernetes.
func IsIPv6FirstPod() bool {
	return IsIPv6(os.Getenv("POD_IP"))
}

func PreferIPFamily(ipv6First bool, envoyProxy *egv1a1.EnvoyProxy) egv1a1.IPFamily {
	if ipv6First {
		// return IPv4 if envoy proxy specifies IPv4
		if envoyProxy != nil && envoyProxy.Spec.IPFamily != nil && *envoyProxy.Spec.IPFamily == egv1a1.IPv4 {
			return egv1a1.IPv4
		}

		return egv1a1.IPv6
	}

	// return IPv6 if envoy proxy specifies IPv6
	if envoyProxy != nil && envoyProxy.Spec.IPFamily != nil && *envoyProxy.Spec.IPFamily == egv1a1.IPv6 {
		return egv1a1.IPv6
	}

	return egv1a1.IPv4
}
