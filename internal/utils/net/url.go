// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package net

import (
	"fmt"
	"net"
	"net/url"
)

// ParseURL return host and port if the URL is a valid HTTP, HTTPS or UDS,
// return the scheme and host:port if it's valid.
func ParseURL(urlStr string) (scheme, hostAndPort string, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL: %w", err)
	}

	switch u.Scheme {
	case "http", "https":
		host := u.Hostname()
		port := u.Port()
		if host == "" {
			return "", "", fmt.Errorf("URL must contain a host")
		}
		// Set default port if not specified
		if port == "" {
			if u.Scheme == "http" {
				port = "80"
			} else {
				port = "443"
			}
		}
		return u.Scheme, net.JoinHostPort(host, port), nil
	case "unix":
		// For Unix Domain Socket, return the path as host, empty port
		if u.Path == "" {
			return "", "", fmt.Errorf("unix URL must contain a path")
		}

		return u.Scheme, u.Path, nil
	default:
		return "", "", fmt.Errorf("unsupported URL scheme: %s (supported: http, https, unix)", u.Scheme)
	}
}
