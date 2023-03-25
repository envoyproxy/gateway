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
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"

	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
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
		inFile, inType, output, resourceType string
		addMissingResources                  bool
		outTypes                             []string
	)

	translateCommand := &cobra.Command{
		Use:   "translate",
		Short: "Translate Configuration from an input type to an output type",
		Example: `  # Translate Gateway API Resources into All xDS Resources.
  egctl experimental translate --from gateway-api --to xds --file <input file>

  # Translate Gateway API Resources into All xDS Resources in JSON output.
  egctl experimental translate --from gateway-api --to xds --type all --output json --file <input file>

  # Translate Gateway API Resources into All xDS Resources in YAML output.
  egctl experimental translate --from gateway-api --to xds --type all --output yaml --file <input file>

  # Translate Gateway API Resources into Bootstrap xDS Resources.
  egctl experimental translate --from gateway-api --to xds --type bootstrap --file <input file>

  # Translate Gateway API Resources into Cluster xDS Resources.
  egctl experimental translate --from gateway-api --to xds --type cluster --file <input file>

  # Translate Gateway API Resources into Listener xDS Resources.
  egctl experimental translate --from gateway-api --to xds --type listener --file <input file>

  # Translate Gateway API Resources into Route xDS Resources.
  egctl experimental translate --from gateway-api --to xds --type route --file <input file>

  # Translate Gateway API Resources into Cluster xDS Resources with short syntax.
  egctl x translate --from gateway-api --to xds -t cluster -o yaml -f <input file>

  # Translate Gateway API Resources into All xDS Resources with dummy resources added.
  egctl x translate --from gateway-api --to xds -t cluster --add-missing-resources -f <input file>

  # Translate Gateway API Resources into All xDS Resources in YAML output,
  # also print the Gateway API Resources with updated status in the same output.
  egctl experimental translate --from gateway-api --to gateway-api,xds --type all --output yaml --file <input file>
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return translate(cmd.OutOrStdout(), inFile, inType, outTypes, output, resourceType, addMissingResources)
		},
	}

	translateCommand.PersistentFlags().StringVarP(&inFile, "file", "f", "", "Location of input file.")
	if err := translateCommand.MarkPersistentFlagRequired("file"); err != nil {
		return nil
	}
	translateCommand.PersistentFlags().StringVarP(&inType, "from", "", gatewayAPIType, getValidInputTypesStr())
	translateCommand.PersistentFlags().StringSliceVarP(&outTypes, "to", "", []string{gatewayAPIType, xdsType}, getValidOutputTypesStr())
	translateCommand.PersistentFlags().StringVarP(&output, "output", "o", yamlOutput, "One of 'yaml' or 'json'")
	translateCommand.PersistentFlags().StringVarP(&resourceType, "type", "t", string(AllEnvoyConfigType), getValidResourceTypesStr())
	translateCommand.PersistentFlags().BoolVarP(&addMissingResources, "add-missing-resources", "", false, "Provides dummy resources if missed")
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
	return []string{xdsType, gatewayAPIType}
}

func findInvalidOutputType(outTypes []string) string {
	for _, oType := range outTypes {
		found := false
		for _, vType := range validOutputTypes() {
			if oType == vType {
				found = true
				break
			}
		}
		if !found {
			return oType
		}
	}
	return ""
}

func getValidOutputTypesStr() string {
	return fmt.Sprintf("Valid types are %v.", validOutputTypes())
}

func validResourceTypes() []envoyConfigType {
	return []envoyConfigType{BootstrapEnvoyConfigType,
		EndpointEnvoyConfigType,
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

func validate(inFile, inType string, outTypes []string, resourceType string) error {
	if !isValidInputType(inType) {
		return fmt.Errorf("%s is not a valid input type. %s", inType, getValidInputTypesStr())
	}
	invalidType := findInvalidOutputType(outTypes)
	if invalidType != "" {
		return fmt.Errorf("%s is not a valid output type. %s", invalidType, getValidOutputTypesStr())
	}
	if !isValidResourceType(envoyConfigType(resourceType)) {
		return fmt.Errorf("%s is not a valid output type. %s", resourceType, getValidResourceTypesStr())
	}
	if inFile == "" {
		return fmt.Errorf("--file must be specified")
	}

	return nil
}

func translate(w io.Writer, inFile, inType string, outTypes []string, output, resourceType string, addMissingResources bool) error {
	if err := validate(inFile, inType, outTypes, resourceType); err != nil {
		return err
	}

	inBytes, err := getInputBytes(inFile)
	if err != nil {
		return fmt.Errorf("unable to read input file: %w", err)
	}

	if inType == gatewayAPIType {
		var toXds, echoGatewayAPI bool
		for _, outType := range outTypes {
			if outType == xdsType {
				toXds = true
			} else if outType == gatewayAPIType {
				echoGatewayAPI = true
			}
		}

		// Unmarshal input
		gcName, resources, err := kubernetesYAMLToResources(string(inBytes), addMissingResources)
		if err != nil {
			return fmt.Errorf("unable to unmarshal input: %w", err)
		}

		// Translate from Gateway API to Xds IR
		gTranslator := &gatewayapi.Translator{
			GatewayClassName:       v1beta1.ObjectName(gcName),
			GlobalRateLimitEnabled: true,
		}
		gRes := gTranslator.Translate(resources)

		results := []interface{}{}
		if echoGatewayAPI {
			results = append(results, resources)
		}

		if toXds {
			keys := []string{}
			for key := range gRes.XdsIR {
				keys = append(keys, key)
			}
			// Make output stable since XdsIR is a map
			sort.Strings(keys)

			// Translate from Xds IR to Xds
			for _, key := range keys {
				val := gRes.XdsIR[key]
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

				// Collect results
				if err := collectXds(&results, key, xRes, envoyConfigType(resourceType)); err != nil {
					return fmt.Errorf("failed to collect result for key %s value %+v, error:%w", key, val, err)
				}
			}
		}

		if err := printOutput(w, results, output); err != nil {
			return fmt.Errorf("failed to print result, error:%w", err)
		}

		return nil
	}

	return fmt.Errorf("unable to find translate from input type %s to output type %s", inType, outTypes)
}

// printOutput prints the echo-backed gateway API and xDS output
func printOutput(w io.Writer, results []interface{}, output string) (err error) {
	var (
		out       []byte
		firstItem = true
	)

	if output == jsonOutput {
		fmt.Fprint(w, "[")
	}

	for _, wrapper := range results {
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

		if firstItem {
			switch output {
			case yamlOutput:
				fmt.Fprintln(w, string(out))
			case jsonOutput:
				fmt.Fprint(w, string(out))
			default:
				fmt.Fprintln(w, string(out))
			}

			firstItem = false
		} else {
			switch output {
			case yamlOutput:
				fmt.Fprintln(w, "---")
				fmt.Fprintln(w, string(out))
			case jsonOutput:
				fmt.Fprintln(w, ",")
				fmt.Fprint(w, string(out))
			default:
				fmt.Fprintln(w, "---")
				fmt.Fprintln(w, string(out))
			}
		}
	}

	if output == jsonOutput {
		fmt.Fprintln(w, "]")
	}

	return nil
}

// collectXds collects xDS which will be output later
func collectXds(results *[]interface{}, key string, tCtx *xds_types.ResourceVersionTable, resourceType envoyConfigType) (err error) {
	globalConfigs, err := constructConfigDump(tCtx)
	if err != nil {
		return err
	}

	var (
		data []byte
	)

	if resourceType == AllEnvoyConfigType {
		data, err = protojson.Marshal(globalConfigs)
		if err != nil {
			return err
		}
	} else {
		xdsResources, err := findXDSResourceFromConfigDump(resourceType, globalConfigs)
		if err != nil {
			return err
		}
		data, err = protojson.Marshal(xdsResources)
		if err != nil {
			return err
		}
	}

	wrapper := map[string]any{}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}

	wrapper["configKey"] = key
	wrapper["resourceType"] = resourceType
	*results = append(*results, wrapper)
	return nil
}

// constructConfigDump constructs configDump from ResourceVersionTable and BootstrapConfig
func constructConfigDump(tCtx *xds_types.ResourceVersionTable) (*adminv3.ConfigDump, error) {
	globalConfigs := &adminv3.ConfigDump{}
	bootstrapConfigs := &adminv3.BootstrapConfigDump{}
	proxyBootstrap := &bootstrapv3.Bootstrap{}
	listenerConfigs := &adminv3.ListenersConfigDump{}
	routeConfigs := &adminv3.RoutesConfigDump{}
	clusterConfigs := &adminv3.ClustersConfigDump{}
	endpointConfigs := &adminv3.EndpointsConfigDump{}

	// construct bootstrap config
	bootstrapYAML, err := bootstrap.GetRenderedBootstrapConfig()
	if err != nil {
		return nil, err
	}
	jsonData, err := yaml.YAMLToJSON([]byte(bootstrapYAML))
	if err != nil {
		return nil, err
	}

	if err := protojson.Unmarshal(jsonData, proxyBootstrap); err != nil {
		return nil, err
	}
	bootstrapConfigs.Bootstrap = proxyBootstrap
	if err := bootstrapConfigs.Validate(); err != nil {
		return nil, err
	}
	if configs, err := anypb.New(bootstrapConfigs); err == nil {
		globalConfigs.Configs = append(globalConfigs.Configs, configs)
	}

	// construct endpoints config
	endpoints := tCtx.XdsResources[resourcev3.EndpointType]
	for _, endpoint := range endpoints {
		c, err := anypb.New(endpoint)
		if err != nil {
			return nil, err
		}
		endpointConfigs.DynamicEndpointConfigs = append(endpointConfigs.DynamicEndpointConfigs, &adminv3.EndpointsConfigDump_DynamicEndpointConfig{
			EndpointConfig: c,
		})
	}
	if err := endpointConfigs.Validate(); err != nil {
		return nil, err
	}
	if configs, err := anypb.New(endpointConfigs); err == nil {
		globalConfigs.Configs = append(globalConfigs.Configs, configs)
	}

	// construct clusters config
	clusters := tCtx.XdsResources[resourcev3.ClusterType]
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
	listeners := tCtx.XdsResources[resourcev3.ListenerType]
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
	routes := tCtx.XdsResources[resourcev3.RouteType]
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
func kubernetesYAMLToResources(str string, addMissingResources bool) (string, *gatewayapi.Resources, error) {
	resources := gatewayapi.NewResources()
	var gcName string
	var useDefaultNamespace bool
	providedNamespaceMap := map[string]struct{}{}
	requiredNamespaceMap := map[string]struct{}{}
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
		if namespace == "" {
			// When kubectl apply a resource in yaml which doesn't have a namespace,
			// the current namespace is applied. Here we do the same thing before translating
			// the GatewayAPI resource. Otherwise the resource can't pass the namespace validation
			useDefaultNamespace = true
			namespace = config.DefaultNamespace
		}
		requiredNamespaceMap[namespace] = struct{}{}
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
		case gatewayapi.KindGRPCRoute:
			typedSpec := spec.Interface()
			grpcRoute := &v1alpha2.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(v1alpha2.GRPCRouteSpec),
			}
			resources.GRPCRoutes = append(resources.GRPCRoutes, grpcRoute)
		case gatewayapi.KindNamespace:
			namespace := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}
			resources.Namespaces = append(resources.Namespaces, namespace)
			providedNamespaceMap[name] = struct{}{}
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

	if useDefaultNamespace {
		if _, found := providedNamespaceMap[config.DefaultNamespace]; !found {
			namespace := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: config.DefaultNamespace,
				},
			}
			resources.Namespaces = append(resources.Namespaces, namespace)
			providedNamespaceMap[config.DefaultNamespace] = struct{}{}
		}
	}

	if addMissingResources {
		for ns := range requiredNamespaceMap {
			if _, found := providedNamespaceMap[ns]; !found {
				namespace := &v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: ns,
					},
				}
				resources.Namespaces = append(resources.Namespaces, namespace)
			}
		}
	}

	return gcName, resources, nil
}
