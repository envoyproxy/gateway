// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package net

import (
	"fmt"
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
		// TODO: support http and https scheme
		return "", "", fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	case "unix":
		// For Unix Domain Socket, return the path as host, empty port.
		// unix:///absolute/path has empty host and path="/absolute/path".
		// unix://./relative/path (SPIFFE workload API format) has host="." and path="/relative/path";
		// combine them so the caller receives "./relative/path".
		if u.Host != "" && u.Host != "." {
			return "", "", fmt.Errorf("unix URL must not contain a host, use unix:///absolute/path")
		}
		if u.Path == "" && u.Host != "." {
			return "", "", fmt.Errorf("unix URL must contain a path")
		}

		return u.Scheme, u.Host + u.Path, nil
	default:
		return "", "", fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}
}
