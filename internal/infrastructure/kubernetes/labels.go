// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// envoyAppLabel returns the labels used for all Envoy resources.
func envoyAppLabel() map[string]string {
	return map[string]string{
		"app.gateway.envoyproxy.io/name": "envoy",
	}
}

// rateLimitLabels returns the labels used for all envoy rate limit resources.
func rateLimitLabels() map[string]string {
	return map[string]string{
		"app.gateway.envoyproxy.io/name": rateLimitInfraName,
	}
}

// getSelector returns a label selector used to select resources
// based on the provided lbls.
func getSelector(labels map[string]string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: labels,
	}
}

// envoyLabels returns the labels, including extraLbls, used for Envoy resources.
func envoyLabels(extraLbls map[string]string) map[string]string {
	lbls := envoyAppLabel()
	for k, v := range extraLbls {
		lbls[k] = v
	}

	return lbls
}
