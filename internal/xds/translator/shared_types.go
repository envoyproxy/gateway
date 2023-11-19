// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const (
	defaultPort = 443
)

// urlCluster is a cluster that is created from a URL.
type urlCluster struct {
	name         string
	hostname     string
	port         uint32
	endpointType EndpointType
}

// url2Cluster returns a urlCluster from the provided url.
func url2Cluster(strURL string) (*urlCluster, error) {
	epType := EndpointTypeDNS

	// The URL should have already been validated in the gateway API translator.
	u, err := url.Parse(strURL)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URI scheme %s", u.Scheme)
	}

	port := defaultPort
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return nil, err
		}
	}

	name := fmt.Sprintf("%s_%d", strings.ReplaceAll(u.Hostname(), ".", "_"), port)

	if ip := net.ParseIP(u.Hostname()); ip != nil {
		if v4 := ip.To4(); v4 != nil {
			epType = EndpointTypeStatic
		}
	}

	return &urlCluster{
		name:         name,
		hostname:     u.Hostname(),
		port:         uint32(port),
		endpointType: epType,
	}, nil
}
