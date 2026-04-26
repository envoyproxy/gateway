// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package net

import (
	"testing"

	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestBackendHostAndPort(t *testing.T) {
	tests := []struct {
		name             string
		backendRef       gwapiv1.BackendObjectReference
		defaultNamespace string
		wantHost         string
		wantPort         uint32
	}{
		{
			name: "explicit namespace",
			backendRef: gwapiv1.BackendObjectReference{
				Name:      "my-service",
				Namespace: ptr.To(gwapiv1.Namespace("custom-ns")),
				Port:      ptr.To[gwapiv1.PortNumber](8080),
			},
			defaultNamespace: "default",
			wantHost:         "my-service.custom-ns.svc",
			wantPort:         8080,
		},
		{
			name: "nil namespace uses default",
			backendRef: gwapiv1.BackendObjectReference{
				Name: "my-service",
				Port: ptr.To[gwapiv1.PortNumber](443),
			},
			defaultNamespace: "default",
			wantHost:         "my-service.default.svc",
			wantPort:         443,
		},
		{
			name: "explicit empty namespace overrides default",
			backendRef: gwapiv1.BackendObjectReference{
				Name:      "my-service",
				Namespace: ptr.To(gwapiv1.Namespace("")),
				Port:      ptr.To[gwapiv1.PortNumber](3000),
			},
			defaultNamespace: "default",
			wantHost:         "my-service",
			wantPort:         3000,
		},
		{
			name: "empty default namespace returns name only",
			backendRef: gwapiv1.BackendObjectReference{
				Name: "my-service",
				Port: ptr.To[gwapiv1.PortNumber](9090),
			},
			defaultNamespace: "",
			wantHost:         "my-service",
			wantPort:         9090,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort := BackendHostAndPort(tt.backendRef, tt.defaultNamespace)
			if gotHost != tt.wantHost {
				t.Errorf("BackendHostAndPort() host = %q, want %q", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("BackendHostAndPort() port = %d, want %d", gotPort, tt.wantPort)
			}
		})
	}
}
