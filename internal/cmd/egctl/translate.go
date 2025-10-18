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
	"sort"

	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	xds_types "github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	gatewayAPIType = "gateway-api"
	xdsType        = "xds"
	irType         = "ir"
)

type TranslationResult struct {
	resource.Resources
	XdsIR   resource.XdsIRMap      `json:"xdsIR,omitempty" yaml:"xdsIR,omitempty"`
	InfraIR resource.InfraIRMap    `json:"infraIR,omitempty" yaml:"infraIR,omitempty"`
	Xds     map[string]interface{} `json:"xds,omitempty"`
}

func newTranslateCommand() *cobra.Command {
	var (
		inFile, inType, output, resourceType string
		addMissingResources                  bool
		outTypes                             []string
		dnsDomain                            string
		namespace                            string
	)

	translateCommand := &cobra.Command{
		Use:   "translate",
		Short: "Translate Configuration from an input type to an output type",
		Example: `  # Translate Gateway API Resources into All xDS Resources.
  egctl experimental translate --from gateway-api --to xds --file <input file> -n <namespace>

  # Translate Gateway API Resources into All xDS Resources in JSON output.
  egctl experimental translate --from gateway-api --to xds --type all --output json --file <input file> -n <namespace>

  # Translate Gateway API Resources into All xDS Resources in YAML output.
  egctl experimental translate --from gateway-api --to xds --type all --output yaml --file <input file> -n <namespace>

  # Translate Gateway API Resources into Bootstrap xDS Resources.
  egctl experimental translate --from gateway-api --to xds --type bootstrap --file <input file> -n <namespace>

  # Translate Gateway API Resources into Cluster xDS Resources.
  egctl experimental translate --from gateway-api --to xds --type cluster --file <input file> -n <namespace>

  # Translate Gateway API Resources into Listener xDS Resources.
  egctl experimental translate --from gateway-api --to xds --type listener --file <input file> -n <namespace>

  # Translate Gateway API Resources into Route xDS Resources.
  egctl experimental translate --from gateway-api --to xds --type route --file <input file> -n <namespace>

  # Translate Gateway API Resources into Cluster xDS Resources with short syntax.
  egctl x translate --from gateway-api --to xds -t cluster -o yaml -f <input file> -n <namespace>

  # Translate Gateway API Resources into All xDS Resources with dummy resources added.
  egctl x translate --from gateway-api --to xds -t cluster --add-missing-resources -f <input file> -n <namespace>

  # Translate Gateway API Resources into All xDS Resources in YAML output,
  # also print the Gateway API Resources with updated status in the same output.
  egctl experimental translate --from gateway-api --to gateway-api,xds --type all --output yaml --file <input file> -n <namespace>

  # Translate Gateway API Resources into IR in YAML output,
  egctl experimental translate --from gateway-api --to ir --output yaml --file <input file>
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return translate(cmd.OutOrStdout(), inFile, inType, outTypes, output, resourceType, addMissingResources, namespace, dnsDomain)
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
	translateCommand.PersistentFlags().StringVarP(&dnsDomain, "dns-domain", "", "cluster.local", "DNS domain used by k8s services, default is cluster.local")
	translateCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "envoy-gateway-system", "Namespace where envoy gateway is installed.")

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
	return []string{xdsType, gatewayAPIType, irType}
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
	return []envoyConfigType{
		BootstrapEnvoyConfigType,
		EndpointEnvoyConfigType,
		ClusterEnvoyConfigType,
		ListenerEnvoyConfigType,
		RouteEnvoyConfigType,
		AllEnvoyConfigType,
	}
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

func translate(w io.Writer, inFile, inType string, outTypes []string, output, resourceType string, addMissingResources bool, namespace, dnsDomain string) error {
	if err := validate(inFile, inType, outTypes, resourceType); err != nil {
		return err
	}

	inBytes, err := getInputBytes(inFile)
	if err != nil {
		return fmt.Errorf("unable to read input file: %w", err)
	}

	if inType == gatewayAPIType {
		// Unmarshal input
		resources, err := resource.LoadResourcesFromYAMLBytes(inBytes, addMissingResources)
		if err != nil {
			return fmt.Errorf("unable to unmarshal input: %w", err)
		}

		var result TranslationResult
		for _, outType := range outTypes {
			// Translate
			if outType == gatewayAPIType {
				result.Resources, err = translateGatewayAPIToGatewayAPI(resources)
				if err != nil {
					return err
				}
			}
			if outType == xdsType {
				res, err := TranslateGatewayAPIToXds(namespace, dnsDomain, resourceType, resources)
				if err != nil {
					return err
				}
				result.Xds = res
			}
			if outType == irType {
				res, err := translateGatewayAPIToIR(resources)
				if err != nil {
					return err
				}
				result.Resources = res.Resources
				result.XdsIR = res.XdsIR
				result.InfraIR = res.InfraIR
			}
		}
		// Print
		if err = printOutput(w, &result, output); err != nil {
			return fmt.Errorf("failed to print result, error:%w", err)
		}

		return nil
	}
	return fmt.Errorf("unable to find translate from input type %s to output type %s", inType, outTypes)
}

func translateGatewayAPIToIR(resources *resource.Resources) (*gatewayapi.TranslateResult, error) {
	if resources.GatewayClass == nil {
		return nil, fmt.Errorf("the GatewayClass resource is required")
	}

	t := &gatewayapi.Translator{
		GatewayControllerName:   string(resources.GatewayClass.Spec.ControllerName),
		GatewayClassName:        gwapiv1.ObjectName(resources.GatewayClass.Name),
		GlobalRateLimitEnabled:  true,
		EndpointRoutingDisabled: true,
		EnvoyPatchPolicyEnabled: true,
		BackendEnabled:          true,
	}

	// Fix the services in the resources section so that they have an IP address - this prevents nasty
	// errors in the translation.
	for _, svc := range resources.Services {
		if svc.Spec.ClusterIP == "" {
			svc.Spec.ClusterIP = "10.96.1.2"
		}
	}

	result, _ := t.Translate(resources)

	return result, nil
}

func translateGatewayAPIToGatewayAPI(resources *resource.Resources) (resource.Resources, error) {
	if resources.GatewayClass == nil {
		return resource.Resources{}, fmt.Errorf("the GatewayClass resource is required")
	}

	// Translate from Gateway API to Xds IR
	gTranslator := &gatewayapi.Translator{
		GatewayControllerName:   string(resources.GatewayClass.Spec.ControllerName),
		GatewayClassName:        gwapiv1.ObjectName(resources.GatewayClass.Name),
		GlobalRateLimitEnabled:  true,
		EndpointRoutingDisabled: true,
		EnvoyPatchPolicyEnabled: true,
		BackendEnabled:          true,
	}
	gRes, _ := gTranslator.Translate(resources)
	// Update the status of the GatewayClass based on EnvoyProxy validation
	epInvalid := false
	if resources.EnvoyProxyForGatewayClass != nil {
		if err := validation.ValidateEnvoyProxy(resources.EnvoyProxyForGatewayClass); err != nil {
			epInvalid = true
			msg := fmt.Sprintf("%s: %v", status.MsgGatewayClassInvalidParams, err)
			status.SetGatewayClassAccepted(resources.GatewayClass, false, string(gwapiv1.GatewayClassReasonInvalidParameters), msg)
		}
		if err := bootstrap.Validate(resources.EnvoyProxyForGatewayClass.Spec.Bootstrap); err != nil {
			epInvalid = true
			msg := fmt.Sprintf("%s: %v", status.MsgGatewayClassInvalidParams, err)
			status.SetGatewayClassAccepted(resources.GatewayClass, false, string(gwapiv1.GatewayClassReasonInvalidParameters), msg)
		}
		gRes.EnvoyProxyForGatewayClass = resources.EnvoyProxyForGatewayClass
	}

	if !epInvalid {
		status.SetGatewayClassAccepted(resources.GatewayClass, true, string(gwapiv1.GatewayClassReasonAccepted), status.MsgValidGatewayClass)
	}

	gRes.GatewayClass = resources.GatewayClass
	return gRes.Resources, nil
}

func TranslateGatewayAPIToXds(namespace, dnsDomain, resourceType string, resources *resource.Resources) (map[string]any, error) {
	if resources.GatewayClass == nil {
		return nil, fmt.Errorf("the GatewayClass resource is required")
	}

	// Translate from Gateway API to Xds IR
	gTranslator := &gatewayapi.Translator{
		GatewayControllerName:   string(resources.GatewayClass.Spec.ControllerName),
		GatewayClassName:        gwapiv1.ObjectName(resources.GatewayClass.Name),
		GlobalRateLimitEnabled:  true,
		EndpointRoutingDisabled: true,
		EnvoyPatchPolicyEnabled: true,
		BackendEnabled:          true,
	}
	gRes, _ := gTranslator.Translate(resources)

	keys := []string{}
	for key := range gRes.XdsIR {
		keys = append(keys, key)
	}
	// Make output stable since XdsIR is a map
	sort.Strings(keys)

	// Translate from Xds IR to Xds
	result := make(map[string]interface{})
	for _, key := range keys {
		val := gRes.XdsIR[key]
		xTranslator := &translator.Translator{
			// Set some default settings for translation
			GlobalRateLimit: &translator.GlobalRateLimitSettings{
				ServiceURL: ratelimit.GetServiceURL(namespace, dnsDomain),
			},
		}
		if resources.EnvoyProxyForGatewayClass != nil {
			xTranslator.FilterOrder = resources.EnvoyProxyForGatewayClass.Spec.FilterOrder
		}
		xRes, err := xTranslator.Translate(val)
		if err != nil {
			return nil, fmt.Errorf("failed to translate xds ir for key %s value %+v, error:%w", key, val, err)
		}

		globalConfigs, err := constructConfigDump(resources, xRes)
		if err != nil {
			return nil, err
		}

		wrapper := map[string]any{}
		var data protoreflect.ProtoMessage
		rType := envoyConfigType(resourceType)
		if rType == AllEnvoyConfigType {
			data = globalConfigs
		} else {
			// Find resource
			xdsResources, err := findXDSResourceFromConfigDump(rType, globalConfigs)
			if err != nil {
				return nil, err
			}
			data = xdsResources
		}

		out, err := protojson.Marshal(data)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(out, &wrapper); err != nil {
			return nil, err
		}

		result[key] = wrapper
	}

	return result, nil
}

// printOutput prints the echo-backed gateway API and xDS output
func printOutput(w io.Writer, result *TranslationResult, output string) error {
	var (
		out []byte
		err error
	)
	switch output {
	case yamlOutput:
		out, err = yaml.Marshal(result)
	case jsonOutput:
		out, err = json.Marshal(result)
	default:
		out, err = yaml.Marshal(result)
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(out))
	return err
}

// constructConfigDump constructs configDump from ResourceVersionTable and BootstrapConfig
func constructConfigDump(resources *resource.Resources, tCtx *xds_types.ResourceVersionTable) (*adminv3.ConfigDump, error) {
	globalConfigs := &adminv3.ConfigDump{}
	bootstrapConfigs := &adminv3.BootstrapConfigDump{}
	proxyBootstrap := &bootstrapv3.Bootstrap{}
	listenerConfigs := &adminv3.ListenersConfigDump{}
	routeConfigs := &adminv3.RoutesConfigDump{}
	clusterConfigs := &adminv3.ClustersConfigDump{}
	endpointConfigs := &adminv3.EndpointsConfigDump{}

	// construct bootstrap config
	var bootstrapConfigurations string
	var err error
	if bootstrapConfigurations, err = bootstrap.GetRenderedBootstrapConfig(nil); err != nil {
		return nil, err
	}

	// Apply Bootstrap from EnvoyProxy API if set by the user
	// The config should have been validated already
	if resources.EnvoyProxyForGatewayClass != nil && resources.EnvoyProxyForGatewayClass.Spec.Bootstrap != nil {
		bootstrapConfigurations, err = bootstrap.ApplyBootstrapConfig(resources.EnvoyProxyForGatewayClass.Spec.Bootstrap, bootstrapConfigurations)
		if err != nil {
			return nil, err
		}
	}

	jsonData, err := yaml.YAMLToJSON([]byte(bootstrapConfigurations))
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
