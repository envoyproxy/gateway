// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import "fmt"

// GetServiceURL returns the URL for the rate limit service.
// TODO: support custom trust domain
func GetServiceURL(namespace string) string {
	return fmt.Sprintf("grpc://%s.%s.svc.cluster.local:%d", InfraName, namespace, InfraGRPCPort)
}
