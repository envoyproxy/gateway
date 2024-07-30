// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type DNS struct {
	// DNSRefreshRate specifies the rate at which DNS records should be refreshed.
	// Defaults to 30 seconds.
	DNSRefreshRate *metav1.Duration `json:"dnsRefreshRate,omitempty"`
	// RespectDNSTTL indicates whether the DNS Time-To-Live (TTL) should be respected.
	// If the value is set to true, the DNS refresh rate will be set to the resource record’s TTL.
	// Defaults to true.
	RespectDNSTTL *bool `json:"respectDnsTtl,omitempty"`
}
