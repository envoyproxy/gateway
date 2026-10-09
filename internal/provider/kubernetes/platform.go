// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"strings"
	"sync/atomic"

	corev1 "k8s.io/api/core/v1"
)

const awsZoneIDLabel = "topology.k8s.aws/zone-id"

// clusterPlatform shares cloud detection across proxy and endpoint locality lookups.
type clusterPlatform struct {
	aws atomic.Bool
}

// zoneID detects the platform from Node metadata and returns the zone ID for this Node.
func (p *clusterPlatform) zoneID(node *corev1.Node) string {
	if !p.aws.Load() {
		if !strings.HasPrefix(node.Spec.ProviderID, "aws://") && node.Labels[awsZoneIDLabel] == "" {
			return ""
		}
		p.aws.Store(true)
	}
	return node.Labels[awsZoneIDLabel]
}
