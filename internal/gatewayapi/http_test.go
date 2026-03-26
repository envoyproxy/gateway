// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestBuildIRHTTP2Settings(t *testing.T) {
	cases := []struct {
		name                       string
		http2Settings              *egv1a1.HTTP2Settings
		wantInitialStreamWindow    *uint32
		wantInitialConnWindow      *uint32
		wantMaxConcurrentStreams   *uint32
		wantResetStreamOnError     *bool
		wantErr                    string
		wantMultipleErrorFragments []string
	}{
		{
			name:          "nil settings returns nil",
			http2Settings: nil,
		},
		{
			name: "valid InitialStreamWindowSize at min boundary",
			http2Settings: &egv1a1.HTTP2Settings{
				InitialStreamWindowSize: quantityPtr(MinHTTP2InitialStreamWindowSize),
			},
			wantInitialStreamWindow: ptr.To(uint32(MinHTTP2InitialStreamWindowSize)),
		},
		{
			name: "valid InitialStreamWindowSize at max boundary",
			http2Settings: &egv1a1.HTTP2Settings{
				InitialStreamWindowSize: quantityPtr(MaxHTTP2InitialStreamWindowSize),
			},
			wantInitialStreamWindow: ptr.To(uint32(MaxHTTP2InitialStreamWindowSize)),
		},
		{
			name: "InitialStreamWindowSize below min",
			http2Settings: &egv1a1.HTTP2Settings{
				InitialStreamWindowSize: quantityPtr(MinHTTP2InitialStreamWindowSize - 1),
			},
			wantErr: "InitialStreamWindowSize value",
		},
		{
			name: "InitialStreamWindowSize above max",
			http2Settings: &egv1a1.HTTP2Settings{
				InitialStreamWindowSize: quantityPtr(MaxHTTP2InitialStreamWindowSize + 1),
			},
			wantErr: "InitialStreamWindowSize value",
		},
		{
			name: "valid InitialConnectionWindowSize at min boundary",
			http2Settings: &egv1a1.HTTP2Settings{
				InitialConnectionWindowSize: quantityPtr(MinHTTP2InitialConnectionWindowSize),
			},
			wantInitialConnWindow: ptr.To(uint32(MinHTTP2InitialConnectionWindowSize)),
		},
		{
			name: "InitialConnectionWindowSize below min",
			http2Settings: &egv1a1.HTTP2Settings{
				InitialConnectionWindowSize: quantityPtr(MinHTTP2InitialConnectionWindowSize - 1),
			},
			wantErr: "InitialConnectionWindowSize value",
		},
		{
			name: "MaxConcurrentStreams passthrough",
			http2Settings: &egv1a1.HTTP2Settings{
				MaxConcurrentStreams: ptr.To(uint32(200)),
			},
			wantMaxConcurrentStreams: ptr.To(uint32(200)),
		},
		{
			name: "OnInvalidMessage TerminateStream sets ResetStreamOnError true",
			http2Settings: &egv1a1.HTTP2Settings{
				OnInvalidMessage: ptr.To(egv1a1.InvalidMessageActionTerminateStream),
			},
			wantResetStreamOnError: ptr.To(true),
		},
		{
			name: "OnInvalidMessage TerminateConnection sets ResetStreamOnError false",
			http2Settings: &egv1a1.HTTP2Settings{
				OnInvalidMessage: ptr.To(egv1a1.InvalidMessageActionTerminateConnection),
			},
			wantResetStreamOnError: ptr.To(false),
		},
		{
			name: "both window sizes invalid accumulates errors",
			http2Settings: &egv1a1.HTTP2Settings{
				InitialStreamWindowSize:     quantityPtr(0),
				InitialConnectionWindowSize: quantityPtr(0),
			},
			wantMultipleErrorFragments: []string{
				"InitialStreamWindowSize value",
				"InitialConnectionWindowSize value",
			},
		},
		{
			name: "all valid settings combined",
			http2Settings: &egv1a1.HTTP2Settings{
				InitialStreamWindowSize:     quantityPtr(100000),
				InitialConnectionWindowSize: quantityPtr(200000),
				MaxConcurrentStreams:        ptr.To(uint32(50)),
				OnInvalidMessage:            ptr.To(egv1a1.InvalidMessageActionTerminateStream),
			},
			wantInitialStreamWindow:  ptr.To(uint32(100000)),
			wantInitialConnWindow:    ptr.To(uint32(200000)),
			wantMaxConcurrentStreams: ptr.To(uint32(50)),
			wantResetStreamOnError:   ptr.To(true),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildIRHTTP2Settings(tc.http2Settings)

			if tc.wantMultipleErrorFragments != nil {
				require.Error(t, err)
				for _, fragment := range tc.wantMultipleErrorFragments {
					require.Contains(t, err.Error(), fragment)
				}
				return
			}

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)

			if tc.http2Settings == nil {
				require.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			require.Equal(t, tc.wantInitialStreamWindow, got.InitialStreamWindowSize)
			require.Equal(t, tc.wantInitialConnWindow, got.InitialConnectionWindowSize)
			require.Equal(t, tc.wantMaxConcurrentStreams, got.MaxConcurrentStreams)
			require.Equal(t, tc.wantResetStreamOnError, got.ResetStreamOnError)
		})
	}
}

func quantityPtr(val int64) *resource.Quantity {
	q := resource.NewQuantity(val, resource.DecimalSI)
	return q
}
