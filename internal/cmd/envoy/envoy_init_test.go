// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoy

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_buildLocalityZone(t *testing.T) {
	tests := []struct {
		name    string
		node    *corev1.Node
		want    string
		wantErr bool
	}{
		{
			name: "zone1",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						corev1.LabelTopologyZone: "zone1",
					},
				},
			},
			want:    "zone1",
			wantErr: false,
		},
		{
			name: "no-zone-label",
			node: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildLocalityZone(tt.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildLocalityZone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildLocalityZone() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildBootstrapCfg(t *testing.T) {
	testCases := []struct {
		name    string
		zone    string
		want    []byte
		wantErr bool
	}{
		{
			name: "build-bootstrap",
			zone: "zone1",
			want: ([]byte)(`{
  "node": {
    "locality": {
      "zone": "zone1"
    }
  },
  "static_resources": {
    "clusters": [
      {
        "name": "local_cluster",
        "type": "STATIC",
        "connect_timeout": "1s",
        "load_assignment": {
          "cluster_name": "local_cluster",
          "endpoints": [
            {
              "locality": {
                "zone": "zone1"
              },
              "lb_endpoints": [
                {
                  "endpoint": {
                    "address": {
                      "socket_address": {
                        "address": "0.0.0.0",
                        "port_value": 10080
                      }
                    }
                  },
                  "load_balancing_weight": 1
                }
              ],
              "load_balancing_weight": 1
            }
          ]
        }
      }
    ]
  },
  "cluster_manager": {
    "local_cluster_name": "local_cluster"
  }
}`),
			wantErr: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildBootstrapCfg(tt.zone)
			require.NoError(t, err)
			require.Equal(t, tt.want, got, "Expected resources to be equal\n%s", cmp.Diff(tt.want, got))
		})
	}
}
