// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type DNSSettings struct {
	// DNSRefreshRate specifies the rate at which DNS records should be refreshed.
	DNSRefreshRate *metav1.Duration `json:"dnsRefreshRate,omitempty"`
	// RespectDNSTTL indicates whether the DNS Time-To-Live (TTL) should be respected.
	RespectDNSTTL *bool `json:"respectDnsTtl,omitempty"`
}
