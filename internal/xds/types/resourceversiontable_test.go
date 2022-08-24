package types

import (
	"testing"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
)

var (
	testListener = &listenerv3.Listener{
		Name: "test-listener",
	}
	testSecret = &tlsv3.Secret{
		Name: "test-secret",
	}
)

func TestDeepCopy(t *testing.T) {
	testCases := []struct {
		name string
		in   *ResourceVersionTable
		out  *ResourceVersionTable
	}{
		{
			name: "nil",
			in:   nil,
			out:  nil,
		},
		{
			name: "listener",
			in: &ResourceVersionTable{
				XdsResources: XdsResources{
					resource.ListenerType: []types.Resource{testListener},
				},
			},
			out: &ResourceVersionTable{
				XdsResources: XdsResources{
					resource.ListenerType: []types.Resource{testListener},
				},
			},
		},
		{
			name: "kitchen-sink",
			in: &ResourceVersionTable{
				XdsResources: XdsResources{
					resource.ListenerType: []types.Resource{testListener},
					resource.SecretType:   []types.Resource{testSecret},
				},
			},
			out: &ResourceVersionTable{
				XdsResources: XdsResources{
					resource.ListenerType: []types.Resource{testListener},
					resource.SecretType:   []types.Resource{testSecret},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.out == nil {
				require.Nil(t, tc.in.DeepCopy())
			} else {
				diff := cmp.Diff(tc.out, tc.in.DeepCopy(), protocmp.Transform())
				require.Empty(t, diff)
			}
		})
	}
}
