// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"

	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	infra "github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	xds_types "github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	gatewayAPIType = "gateway-api"
	xdsType        = "xds"
)

func NewTranslateCommand() *cobra.Command {
	var (
		inFile, inType, outType, output, resourceType string
	)

	translateCommand := &cobra.Command{
		Use:   "translate",
		Short: "Translate Configuration from an input type to an output type",
		RunE: func(cmd *cobra.Command, args []string) error {
			return translate(cmd.OutOrStdout(), inFile, inType, outType, output, resourceType)
		},
	}

	translateCommand.PersistentFlags().StringVarP(&inFile, "file", "f", "", "Location of input file.")
	if err := translateCommand.MarkPersistentFlagRequired("file"); err != nil {
		return nil
	}
	translateCommand.PersistentFlags().StringVarP(&inType, "from", "", gatewayAPIType, getValidInputTypesStr())
	translateCommand.PersistentFlags().StringVarP(&outType, "to", "", xdsType, getValidOutputTypesStr())
	translateCommand.PersistentFlags().StringVarP(&output, "output", "o", yamlOutput, "One of 'yaml' or 'json'")
	translateCommand.PersistentFlags().StringVarP(&resourceType, "type", "t", string(AllEnvoyConfigType), getValidResourceTypesStr())
	return translateCommand
}

func validInputTypes() []string {
	return []string{gatewayAPIType}
}

func isValidInputType(inType string) bool {
	for _, vType := range validInputTypes() {
		if inType == vType {
			return true
		}
	}
	return false
}

func getValidInputTypesStr() string {
	return fmt.Sprintf("Valid types are %v.", validInputTypes())
}

func validOutputTypes() []string {
	return []string{xdsType}
}

func isValidOutputType(outType string) bool {
	for _, vType := range validOutputTypes() {
		if outType == vType {
			return true
		}
	}
	return false
}

func getValidOutputTypesStr() string {
	return fmt.Sprintf("Valid types are %v.", validOutputTypes())
}

type envoyConfigType string

var (
	BootstrapEnvoyConfigType envoyConfigType = "bootstrap"
	ClusterEnvoyConfigType   envoyConfigType = "cluster"
	ListenerEnvoyConfigType  envoyConfigType = "listener"
	RouteEnvoyConfigType     envoyConfigType = "route"
	AllEnvoyConfigType       envoyConfigType = "all"
)

func validResourceTypes() []envoyConfigType {
	return []envoyConfigType{BootstrapEnvoyConfigType,
		ClusterEnvoyConfigType,
		ListenerEnvoyConfigType,
		RouteEnvoyConfigType,
		AllEnvoyConfigType}
}

func isValidResourceType(outType envoyConfigType) bool {
	for _, vType := range validResourceTypes() {
		if outType == vType {
			return true
		}
	}
	return false
}

func getValidResourceTypesStr() string {
	return fmt.Sprintf("Valid types are %v.", validResourceTypes())
}

func getInputBytes(inFile string) ([]byte, error) {

	// Get input from stdin
	if inFile == "-" {
		scanner := bufio.NewScanner(os.Stdin)
		var input string
		for {
			if !scanner.Scan() {
				break
			}
			input += scanner.Text() + "\n"
		}
		return []byte(input), nil
	}
	// Get input from file
	return os.ReadFile(inFile)
}

func validate(inFile, inType, outType, resourceType string) error {
	if !isValidInputType(inType) {
		return fmt.Errorf("%s is not a valid input type. %s", inType, getValidInputTypesStr())
	}
	if !isValidOutputType(outType) {
		return fmt.Errorf("%s is not a valid output type. %s", outType, getValidOutputTypesStr())
	}
	if !isValidResourceType(envoyConfigType(resourceType)) {
		return fmt.Errorf("%s is not a valid output type. %s", resourceType, getValidResourceTypesStr())
	}
	if inFile == "" {
		return fmt.Errorf("--file must be specified")
	}

	return nil
}

func translate(w io.Writer, inFile, inType, outType, output, resourceType string) error {
	if err := validate(inFile, inType, outType, resourceType); err != nil {
		return err
	}

	inBytes, err := getInputBytes(inFile)
	if err != nil {
		return fmt.Errorf("unable to read input file: %w", err)
	}

	if inType == gatewayAPIType && outType == xdsType {
		return translateFromGatewayAPIToXds(w, inBytes, output, resourceType)
	}

	return fmt.Errorf("unable to find translate from input type %s to output type %s", inType, outType)
}

func translateFromGatewayAPIToXds(w io.Writer, inBytes []byte, output, resourceType string) error {
	// Unmarshal input
	gcName, resources, err := kubernetesYAMLToResources(string(inBytes))
	if err != nil {
		return fmt.Errorf("unable to unmarshal input: %w", err)
	}

	// Translate from Gateway API to Xds IR
	gTranslator := &gatewayapi.Translator{
		GatewayClassName:       v1beta1.ObjectName(gcName),
		GlobalRateLimitEnabled: true,
	}
	gRes := gTranslator.Translate(resources)

	// Translate from Xds IR to Xds
	for key, val := range gRes.XdsIR {
		xTranslator := &translator.Translator{
			// Set some default settings for translation
			GlobalRateLimit: &translator.GlobalRateLimitSettings{
				ServiceURL: infra.GetRateLimitServiceURL("envoy-gateway"),
			},
		}
		xRes, err := xTranslator.Translate(val)
		if err != nil {
			return fmt.Errorf("failed to translate xds ir for key %s value %+v, error:%w", key, val, err)
		}

		// Print results
		if err := printXds(w, key, xRes, output, envoyConfigType(resourceType)); err != nil {
			return fmt.Errorf("failed to print result for key %s value %+v, error:%w", key, val, err)
		}
	}

	return nil
}

// printXds prints the xDS Output
func printXds(w io.Writer, key string, tCtx *xds_types.ResourceVersionTable, output string, resourceType envoyConfigType) (err error) {
	globalConfigs, err := constructConfigDump(tCtx)
	if err != nil {
		return err
	}

	var (
		out, data []byte
	)
	switch resourceType {
	case AllEnvoyConfigType:
		data, err = protojson.Marshal(globalConfigs)
		if err != nil {
			return err
		}
	case BootstrapEnvoyConfigType:
		for _, config := range globalConfigs.Configs {
			if config.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.BootstrapConfigDump" {
				data, err = protojson.Marshal(config)
				if err != nil {
					return err
				}
			}
		}
	case ClusterEnvoyConfigType:
		for _, config := range globalConfigs.Configs {
			if config.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.ClustersConfigDump" {
				data, err = protojson.Marshal(config)
				if err != nil {
					return err
				}
			}
		}
	case ListenerEnvoyConfigType:
		for _, config := range globalConfigs.Configs {
			if config.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.ListenersConfigDump" {
				data, err = protojson.Marshal(config)
				if err != nil {
					return err
				}
			}
		}
	case RouteEnvoyConfigType:
		for _, config := range globalConfigs.Configs {
			if config.GetTypeUrl() == "type.googleapis.com/envoy.admin.v3.RoutesConfigDump" {
				data, err = protojson.Marshal(config)
				if err != nil {
					return err
				}
			}
		}
	default:
		return fmt.Errorf("unknown resourceType %s", resourceType)
	}

	wrapper := map[string]any{}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}

	wrapper["configKey"] = key
	wrapper["resourceType"] = resourceType

	switch output {
	case yamlOutput:
		out, err = yaml.Marshal(wrapper)
	case jsonOutput:
		out, err = json.Marshal(wrapper)
	default:
		out, err = yaml.Marshal(wrapper)
	}
	if err != nil {
		return err
	}

	fmt.Fprintln(w, string(out))
	return nil
}

// consructConfigDump constructs configDump from ResourceVersionTable and BootstrapConfig
func constructConfigDump(tCtx *xds_types.ResourceVersionTable) (*adminv3.ConfigDump, error) {
	globalConfigs := &adminv3.ConfigDump{}
	bootstrapConfigs := &adminv3.BootstrapConfigDump{}
	bstrap := &bootstrapv3.Bootstrap{}
	listenerConfigs := &adminv3.ListenersConfigDump{}
	routeConfigs := &adminv3.RoutesConfigDump{}
	clusterConfigs := &adminv3.ClustersConfigDump{}

	// construct bootstrap config
	bootsrapYAML, err := bootstrap.GetRenderedBootstrapConfig()
	if err != nil {
		return nil, err
	}
	jsonData, err := yaml.YAMLToJSON([]byte(bootsrapYAML))
	if err != nil {
		return nil, err
	}

	if err := protojson.Unmarshal(jsonData, bstrap); err != nil {
		return nil, err
	}
	bootstrapConfigs.Bootstrap = bstrap
	if err := bootstrapConfigs.Validate(); err != nil {
		return nil, err
	}
	if configs, err := anypb.New(bootstrapConfigs); err == nil {
		globalConfigs.Configs = append(globalConfigs.Configs, configs)
	}

	// construct clusters config
	clusters := tCtx.XdsResources[resource.ClusterType]
	for _, cluster := range clusters {
		c, err := anypb.New(cluster)
		if err != nil {
			return nil, err
		}
		clusterConfigs.DynamicActiveClusters = append(clusterConfigs.DynamicActiveClusters, &adminv3.ClustersConfigDump_DynamicCluster{
			Cluster: c,
		})
	}
	if err := clusterConfigs.Validate(); err != nil {
		return nil, err
	}
	if configs, err := anypb.New(clusterConfigs); err == nil {
		globalConfigs.Configs = append(globalConfigs.Configs, configs)
	}

	// construct listeners config
	listeners := tCtx.XdsResources[resource.ListenerType]
	for _, listener := range listeners {
		l, err := anypb.New(listener)
		if err != nil {
			return nil, err
		}
		listenerConfigs.DynamicListeners = append(listenerConfigs.DynamicListeners, &adminv3.ListenersConfigDump_DynamicListener{
			ActiveState: &adminv3.ListenersConfigDump_DynamicListenerState{
				Listener: l,
			},
		})
	}
	if err := listenerConfigs.Validate(); err != nil {
		return nil, err
	}
	if configs, err := anypb.New(listenerConfigs); err == nil {
		globalConfigs.Configs = append(globalConfigs.Configs, configs)
	}

	// construct routes config
	routes := tCtx.XdsResources[resource.RouteType]
	for _, route := range routes {
		r, err := anypb.New(route)
		if err != nil {
			return nil, err
		}
		routeConfigs.DynamicRouteConfigs = append(routeConfigs.DynamicRouteConfigs, &adminv3.RoutesConfigDump_DynamicRouteConfig{
			RouteConfig: r,
		})
	}
	if err := routeConfigs.Validate(); err != nil {
		return nil, err
	}
	if configs, err := anypb.New(routeConfigs); err == nil {
		globalConfigs.Configs = append(globalConfigs.Configs, configs)
	}

	return globalConfigs, nil
}

// kubernetesYAMLToResources converts a Kubernetes YAML string into GatewayAPI Resources
func kubernetesYAMLToResources(str string) (string, *gatewayapi.Resources, error) {
	resources := gatewayapi.NewResources()
	var gcName string
	yamls := strings.Split(str, "\n---")
	combinedScheme := envoygateway.GetScheme()
	for _, y := range yamls {
		if strings.TrimSpace(y) == "" {
			continue
		}
		var obj map[string]interface{}
		err := yaml.Unmarshal([]byte(y), &obj)
		if err != nil {
			return "", nil, err
		}
		un := unstructured.Unstructured{Object: obj}
		gvk := un.GroupVersionKind()
		name, namespace := un.GetName(), un.GetNamespace()
		kobj, err := combinedScheme.New(gvk)
		if err != nil {
			return "", nil, err
		}
		err = combinedScheme.Convert(&un, kobj, nil)
		if err != nil {
			return "", nil, err
		}

		objType := reflect.TypeOf(kobj)
		if objType.Kind() != reflect.Ptr {
			return "", nil, fmt.Errorf("expected pointer type, but got %s", objType.Kind().String())
		}
		kobjVal := reflect.ValueOf(kobj).Elem()
		spec := kobjVal.FieldByName("Spec")

		switch gvk.Kind {
		case gatewayapi.KindGatewayClass:
			gcName = name
		case gatewayapi.KindGateway:
			typedSpec := spec.Interface()
			gateway := &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(v1beta1.GatewaySpec),
			}
			resources.Gateways = append(resources.Gateways, gateway)
		case gatewayapi.KindHTTPRoute:
			typedSpec := spec.Interface()
			httpRoute := &v1beta1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(v1beta1.HTTPRouteSpec),
			}
			resources.HTTPRoutes = append(resources.HTTPRoutes, httpRoute)
		case gatewayapi.KindNamespace:
			namespace := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}
			resources.Namespaces = append(resources.Namespaces, namespace)
		case gatewayapi.KindService:
			typedSpec := spec.Interface()
			service := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(v1.ServiceSpec),
			}
			resources.Services = append(resources.Services, service)
		}
	}

	return gcName, resources, nil
}
