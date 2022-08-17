package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"hash"
	"hash/fnv"
	"text/template"

	"github.com/davecgh/go-spew/spew"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	// ConfigMapHashAnnotation defines an annotation key that holds a hash
	// of the generated ConfigMap. The value is updated when the ConfigMap
	// changes.
	ConfigMapHashAnnotation = "gateway.envoy.io/hash"
)

var (
	// envoyGatewayService is the name of the Envoy Gateway service.
	envoyGatewayService = "envoy-gateway"
)

var envoyTmpl = template.Must(template.New(envoyCfgFileName).Parse(`
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
                address: {{ .XdsServerAddress }}
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
`))

// envoyConfigMap defines the managed Envoy ConfigMap object.
type envoyConfigMap struct {
	Key   string
	Envoy envoyConfig
}

// envoyConfig contains Envoy configuration parameters.
type envoyConfig struct {
	// XdsServerAddress is the address of the XDS Server that Envoy is managed by.
	XdsServerAddress string
}

// createConfigMapIfNeeded creates a ConfigMap based on the provided infra, if
// it doesn't exist in the kube api server.
func (i *Infra) createConfigMapIfNeeded(ctx context.Context, infra *ir.Infra) error {
	current, err := i.getConfigMap(ctx, infra)
	if err != nil {
		if kerrors.IsNotFound(err) {
			cm, err := i.createConfigMap(ctx, infra)
			if err != nil {
				return err
			}
			if err := i.addResource(cm); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if err := i.addResource(current); err != nil {
		return err
	}

	return nil
}

// getConfigMap gets the ConfigMap for the provided infra from the Kube API server.
func (i *Infra) getConfigMap(ctx context.Context, infra *ir.Infra) (*corev1.ConfigMap, error) {
	ns := i.Namespace
	name := infra.GetProxyInfra().ObjectName()
	key := types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}
	cm := new(corev1.ConfigMap)
	if err := i.Client.Get(ctx, key, cm); err != nil {
		return nil, fmt.Errorf("failed to get configmap %s/%s: %w", ns, name, err)
	}

	return cm, nil
}

// createConfigMap creates a ConfigMap in the Kube api server based on the provided
// infra, if it doesn't exist.
func (i *Infra) createConfigMap(ctx context.Context, infra *ir.Infra) (*corev1.ConfigMap, error) {
	cfg := &envoyConfigMap{
		Key: envoyCfgFileName,
		Envoy: envoyConfig{
			XdsServerAddress: envoyGatewayService,
		},
	}

	expected, err := i.expectedConfigMap(infra, cfg)
	if err != nil {
		return nil, err
	}

	if err := i.Client.Create(ctx, expected); err != nil {
		if kerrors.IsAlreadyExists(err) {
			return expected, nil
		}
		return nil, fmt.Errorf("failed to create deployment %s/%s: %w",
			expected.Namespace, expected.Name, err)
	}

	return expected, nil
}

// expectedConfigMap returns the expected ConfigMap based on the provided infra.
func (i *Infra) expectedConfigMap(infra *ir.Infra, cmCfg *envoyConfigMap) (*corev1.ConfigMap, error) {
	ns := i.Namespace
	name := infra.GetProxyInfra().ObjectName()

	buf := new(bytes.Buffer)
	if err := envoyTmpl.Execute(buf, cmCfg.Envoy); err != nil {
		return nil, fmt.Errorf("failed to render configmap %s/%s: %v", ns, name, err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Data: map[string]string{
			cmCfg.Key: buf.String(),
		},
	}

	// Compute the hash for topology spread constraints and possibly
	// affinity policy now, after all the other fields have been computed,
	// and inject it into the appropriate fields.
	hash := configMapHash(cm)
	cm.Annotations[ConfigMapHashAnnotation] = hash
	i.ConfigHash = hash

	return cm, nil
}

// configMapHash returns a hash value for the managed ConfigMap fields that,
// if changed, should trigger an update.
func configMapHash(cm *corev1.ConfigMap) string {
	hasher := fnv.New32a()
	deepHashObject(hasher, hashableConfigMap(cm))
	return rand.SafeEncodeString(fmt.Sprint(hasher.Sum32()))
}

// hashableConfigMap returns a copy of the given configmap with the fields that
// should be used for computing its hash copied over. In particular, these are
// the fields that expectedConfigMap sets.
func hashableConfigMap(cm *corev1.ConfigMap) *corev1.ConfigMap {
	var hashableConfigMap corev1.ConfigMap

	hashableConfigMap.Namespace = cm.Namespace
	hashableConfigMap.Name = cm.Name
	hashableConfigMap.Data[envoyCfgFileName] = cm.Data[envoyCfgFileName]

	return &hashableConfigMap

}

// deepHashObject writes a specified object to a hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
//
// Copied from github.com/kubernetes/kubernetes/pkg/util/hash/hash.go.
func deepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", objectToWrite)
}
