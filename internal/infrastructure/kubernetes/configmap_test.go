package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestExpectedConfigMap(t *testing.T) {
	expected := `
admin:
  access_log_path: /dev/null
  address:
    socket_address:
      address: 127.0.0.1
      port_value: 19000
dynamic_resources:
  cds_config:
    resource_api_version: V3
    api_config_source:
      api_type: GRPC
      transport_api_version: V3
      grpc_services:
      - envoy_grpc:
          cluster_name: xds_cluster
      set_node_on_first_message_only: true
  lds_config:
    resource_api_version: V3
    api_config_source:
      api_type: GRPC
      transport_api_version: V3
      grpc_services:
      - envoy_grpc:
          cluster_name: xds_cluster
      set_node_on_first_message_only: true
node:
  cluster: envoy-gateway-system
  id: envoy-default
static_resources:
  clusters:
  - connect_timeout: 1s
    load_assignment:
      cluster_name: xds_cluster
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: envoy-gateway
                port_value: 18000
    http2_protocol_options: {}
    name: xds_cluster
    type: STRICT_DNS
layered_runtime:
  layers:
    - name: runtime-0
      rtds_layer:
        rtds_config:
          resource_api_version: V3
          api_config_source:
            transport_api_version: V3
            api_type: GRPC
            grpc_services:
              envoy_grpc:
                cluster_name: xds_cluster
        name: runtime-0
`

	// Setup the infra.
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli)
	kube.Namespace = "foo"
	infra := ir.NewInfra()

	// Create the configmap data.
	cfg := &envoyConfigMap{
		Key: envoyCfgFileName,
		Envoy: envoyConfig{
			XdsServerAddress: envoyGatewayService,
		},
	}

	cm, err := kube.expectedConfigMap(infra, cfg)
	require.NoError(t, err)
	require.Equal(t, "envoy-default", cm.Name)
	require.Equal(t, "foo", cm.Namespace)
	require.Contains(t, cm.Data, "envoy.yaml")
	assert.Equal(t, expected, cm.Data["envoy.yaml"])
}
