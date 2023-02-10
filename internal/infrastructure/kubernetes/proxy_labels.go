// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

// envoyAppLabel returns the labels used for all Envoy resources.
func envoyAppLabel() map[string]string {
	return map[string]string{
		"app.gateway.envoyproxy.io/name": "envoy",
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
