// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

func TestNodeDetailsAddressStore(t *testing.T) {
	store := newProviderStore()
	testCases := []struct {
		name              string
		nodeObject        *corev1.Node
		expectedAddresses status.NodeAddresses
	}{
		{
			name: "No node addresses",
			nodeObject: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node1"},
				Status:     corev1.NodeStatus{Addresses: []corev1.NodeAddress{{}}},
			},
			expectedAddresses: status.NodeAddresses{},
		},
		{
			name: "only external address",
			nodeObject: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node1"},
				Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{{
					Address: "1.1.1.1",
					Type:    corev1.NodeExternalIP,
				}}},
			},
			expectedAddresses: status.NodeAddresses{
				IPv4: []string{"1.1.1.1"},
			},
		},
		{
			name: "only internal address",
			nodeObject: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node1"},
				Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{{
					Address: "1.1.1.1",
					Type:    corev1.NodeInternalIP,
				}}},
			},
			expectedAddresses: status.NodeAddresses{
				IPv4: []string{"1.1.1.1"},
			},
		},
		{
			name: "prefer external address",
			nodeObject: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node1"},
				Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{
					{
						Address: "1.1.1.1",
						Type:    corev1.NodeExternalIP,
					},
					{
						Address: "2.2.2.2",
						Type:    corev1.NodeInternalIP,
					},
				}},
			},
			expectedAddresses: status.NodeAddresses{
				IPv4: []string{"1.1.1.1"},
			},
		},
		{
			name: "all external addresses",
			nodeObject: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node1"},
				Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{{
					Address: "1.1.1.1",
					Type:    corev1.NodeExternalIP,
				}, {
					Address: "2606:4700:4700::1111",
					Type:    corev1.NodeExternalIP,
				}}},
			},
			expectedAddresses: status.NodeAddresses{
				IPv4: []string{"1.1.1.1"},
				IPv6: []string{"2606:4700:4700::1111"},
			},
		},
		{
			name: "all internal addresses",
			nodeObject: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node1"},
				Status: corev1.NodeStatus{Addresses: []corev1.NodeAddress{{
					Address: "1.1.1.1",
					Type:    corev1.NodeInternalIP,
				}, {
					Address: "2606:4700:4700::1111",
					Type:    corev1.NodeInternalIP,
				}}},
			},
			expectedAddresses: status.NodeAddresses{
				IPv4: []string{"1.1.1.1"},
				IPv6: []string{"2606:4700:4700::1111"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			store.addNode(tc.nodeObject)
			assert.Equal(t, tc.expectedAddresses, store.listNodeAddresses())
			store.removeNode(tc.nodeObject)
		})
	}
}

func TestRace(t *testing.T) {
	s := newProviderStore()

	go func() {
		for {
			s.addNode(&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node1"},
				Status:     corev1.NodeStatus{Addresses: []corev1.NodeAddress{{}}},
			})
		}
	}()

	_ = s.listNodeAddresses()
}
