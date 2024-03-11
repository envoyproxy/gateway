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
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"

	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/status"
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
	gatewayapi.Resources
	XdsIR   gatewayapi.XdsIRMap    `json:"xdsIR,omitempty" yaml:"xdsIR,omitempty"`
	InfraIR gatewayapi.InfraIRMap  `json:"infraIR,omitempty" yaml:"infraIR,omitempty"`
	Xds     map[string]interface{} `json:"xds,omitempty"`
}

func newTranslateCommand() *cobra.Command {
	var (
		inFile, inType, output, resourceType string
		addMissingResources                  bool
		outTypes                             []string
		dnsDomain                            string
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

  # Translate Gateway API Resources into IR in YAML output,
  egctl experimental translate --from gateway-api --to ir --output yaml --file <input file>
	`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return translate(cmd.OutOrStdout(), inFile, inType, outTypes, output, resourceType, addMissingResources, dnsDomain)
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

func translate(w io.Writer, inFile, inType string, outTypes []string, output, resourceType string, addMissingResources bool, dnsDomain string) error {
	if err := validate(inFile, inType, outTypes, resourceType); err != nil {
		return err
	}

	inBytes, err := getInputBytes(inFile)
	if err != nil {
		return fmt.Errorf("unable to read input file: %w", err)
	}

	if inType == gatewayAPIType {
		// Unmarshal input
		resources, err := kubernetesYAMLToResources(string(inBytes), addMissingResources)
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
				res, err := translateGatewayAPIToXds(dnsDomain, resourceType, resources)
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
		if err = printOutput(w, result, output); err != nil {
			return fmt.Errorf("failed to print result, error:%w", err)
		}

		return nil
	}
	return fmt.Errorf("unable to find translate from input type %s to output type %s", inType, outTypes)
}

func translateGatewayAPIToIR(resources *gatewayapi.Resources) (*gatewayapi.TranslateResult, error) {
	if resources.GatewayClass == nil {
		return nil, fmt.Errorf("the GatewayClass resource is required")
	}

	t := &gatewayapi.Translator{
		GatewayControllerName:   egv1a1.GatewayControllerName,
		GatewayClassName:        gwapiv1.ObjectName(resources.GatewayClass.Name),
		GlobalRateLimitEnabled:  true,
		EndpointRoutingDisabled: true,
		EnvoyPatchPolicyEnabled: true,
	}

	// Fix the services in the resources section so that they have an IP address - this prevents nasty
	// errors in the translation.
	for _, svc := range resources.Services {
		if svc.Spec.ClusterIP == "" {
			svc.Spec.ClusterIP = "10.96.1.2"
		}
	}

	result := t.Translate(resources)

	return result, nil
}

func translateGatewayAPIToGatewayAPI(resources *gatewayapi.Resources) (gatewayapi.Resources, error) {
	if resources.GatewayClass == nil {
		return gatewayapi.Resources{}, fmt.Errorf("the GatewayClass resource is required")
	}

	// Translate from Gateway API to Xds IR
	gTranslator := &gatewayapi.Translator{
		GatewayControllerName:   egv1a1.GatewayControllerName,
		GatewayClassName:        gwapiv1.ObjectName(resources.GatewayClass.Name),
		GlobalRateLimitEnabled:  true,
		EndpointRoutingDisabled: true,
		EnvoyPatchPolicyEnabled: true,
	}
	gRes := gTranslator.Translate(resources)
	// Update the status of the GatewayClass based on EnvoyProxy validation
	epInvalid := false
	if resources.EnvoyProxy != nil {
		if err := validation.ValidateEnvoyProxy(resources.EnvoyProxy); err != nil {
			epInvalid = true
			msg := fmt.Sprintf("%s: %v", status.MsgGatewayClassInvalidParams, err)
			status.SetGatewayClassAccepted(resources.GatewayClass, false, string(gwapiv1.GatewayClassReasonInvalidParameters), msg)
		}
		gRes.EnvoyProxy = resources.EnvoyProxy
	}
	if !epInvalid {
		status.SetGatewayClassAccepted(resources.GatewayClass, true, string(gwapiv1.GatewayClassReasonAccepted), status.MsgValidGatewayClass)
	}

	gRes.GatewayClass = resources.GatewayClass
	return gRes.Resources, nil
}

func translateGatewayAPIToXds(dnsDomain string, resourceType string, resources *gatewayapi.Resources) (map[string]any, error) {
	if resources.GatewayClass == nil {
		return nil, fmt.Errorf("the GatewayClass resource is required")
	}

	// Translate from Gateway API to Xds IR
	gTranslator := &gatewayapi.Translator{
		GatewayControllerName:   egv1a1.GatewayControllerName,
		GatewayClassName:        gwapiv1.ObjectName(resources.GatewayClass.Name),
		GlobalRateLimitEnabled:  true,
		EndpointRoutingDisabled: true,
		EnvoyPatchPolicyEnabled: true,
	}
	gRes := gTranslator.Translate(resources)

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
				ServiceURL: ratelimit.GetServiceURL("envoy-gateway", dnsDomain),
			},
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
func printOutput(w io.Writer, result TranslationResult, output string) error {
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
func constructConfigDump(resources *gatewayapi.Resources, tCtx *xds_types.ResourceVersionTable) (*adminv3.ConfigDump, error) {
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
	if resources.EnvoyProxy != nil && resources.EnvoyProxy.Spec.Bootstrap != nil {
		bootstrapConfigurations, err = bootstrap.ApplyBootstrapConfig(resources.EnvoyProxy.Spec.Bootstrap, bootstrapConfigurations)
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

func addMissingServices(requiredServices map[string]*v1.Service, obj interface{}) {
	var objNamespace string
	protocol := v1.Protocol(gatewayapi.TCPProtocol)

	refs := []gwapiv1.BackendRef{}
	switch route := obj.(type) {
	case *gwapiv1.HTTPRoute:
		objNamespace = route.Namespace
		for _, rule := range route.Spec.Rules {
			for _, httpBakcendRef := range rule.BackendRefs {
				refs = append(refs, httpBakcendRef.BackendRef)
			}
		}
	case *gwapiv1a2.GRPCRoute:
		objNamespace = route.Namespace
		for _, rule := range route.Spec.Rules {
			for _, gRPCBakcendRef := range rule.BackendRefs {
				refs = append(refs, gRPCBakcendRef.BackendRef)
			}
		}
	case *gwapiv1a2.TLSRoute:
		objNamespace = route.Namespace
		for _, rule := range route.Spec.Rules {
			refs = append(refs, rule.BackendRefs...)
		}
	case *gwapiv1a2.TCPRoute:
		objNamespace = route.Namespace
		for _, rule := range route.Spec.Rules {
			refs = append(refs, rule.BackendRefs...)
		}
	case *gwapiv1a2.UDPRoute:
		protocol = v1.Protocol(gatewayapi.UDPProtocol)
		objNamespace = route.Namespace
		for _, rule := range route.Spec.Rules {
			refs = append(refs, rule.BackendRefs...)
		}
	}

	for _, ref := range refs {
		if ref.Kind == nil || *ref.Kind != gatewayapi.KindService {
			continue
		}

		ns := objNamespace
		if ref.Namespace != nil {
			ns = string(*ref.Namespace)
		}
		name := string(ref.Name)
		key := ns + "/" + name

		port := int32(*ref.Port)
		servicePort := v1.ServicePort{
			Name:     fmt.Sprintf("%s-%d", protocol, port),
			Protocol: protocol,
			Port:     port,
		}
		if service, found := requiredServices[key]; !found {
			service := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: ns,
				},
				Spec: v1.ServiceSpec{
					// Just a dummy IP
					ClusterIP: "127.0.0.1",
					Ports:     []v1.ServicePort{servicePort},
				},
			}
			requiredServices[key] = service

		} else {
			inserted := false
			for _, port := range service.Spec.Ports {
				if port.Protocol == servicePort.Protocol && port.Port == servicePort.Port {
					inserted = true
					break
				}
			}

			if !inserted {
				service.Spec.Ports = append(service.Spec.Ports, servicePort)
			}
		}
	}
}

// kubernetesYAMLToResources converts a Kubernetes YAML string into GatewayAPI Resources
func kubernetesYAMLToResources(str string, addMissingResources bool) (*gatewayapi.Resources, error) {
	resources := gatewayapi.NewResources()
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
			return nil, err
		}
		un := unstructured.Unstructured{Object: obj}
		gvk := un.GroupVersionKind()
		name, namespace := un.GetName(), un.GetNamespace()
		if namespace == "" {
			// When kubectl applies a resource in yaml which doesn't have a namespace,
			// the current namespace is applied. Here we do the same thing before translating
			// the GatewayAPI resource. Otherwise, the resource can't pass the namespace validation
			useDefaultNamespace = true
			namespace = config.DefaultNamespace
		}
		requiredNamespaceMap[namespace] = struct{}{}
		kobj, err := combinedScheme.New(gvk)
		if err != nil {
			return nil, err
		}
		err = combinedScheme.Convert(&un, kobj, nil)
		if err != nil {
			return nil, err
		}

		objType := reflect.TypeOf(kobj)
		if objType.Kind() != reflect.Ptr {
			return nil, fmt.Errorf("expected pointer type, but got %s", objType.Kind().String())
		}
		kobjVal := reflect.ValueOf(kobj).Elem()
		spec := kobjVal.FieldByName("Spec")

		switch gvk.Kind {
		case gatewayapi.KindEnvoyProxy:
			typedSpec := spec.Interface()
			envoyProxy := &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(egv1a1.EnvoyProxySpec),
			}
			resources.EnvoyProxy = envoyProxy
		case gatewayapi.KindGatewayClass:
			typedSpec := spec.Interface()
			gatewayClass := &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1.GatewayClassSpec),
			}
			resources.GatewayClass = gatewayClass
		case gatewayapi.KindGateway:
			typedSpec := spec.Interface()
			gateway := &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1.GatewaySpec),
			}
			resources.Gateways = append(resources.Gateways, gateway)
		case gatewayapi.KindTCPRoute:
			typedSpec := spec.Interface()
			tcpRoute := &gwapiv1a2.TCPRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: gatewayapi.KindTCPRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1a2.TCPRouteSpec),
			}
			resources.TCPRoutes = append(resources.TCPRoutes, tcpRoute)
		case gatewayapi.KindUDPRoute:
			typedSpec := spec.Interface()
			udpRoute := &gwapiv1a2.UDPRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: gatewayapi.KindUDPRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1a2.UDPRouteSpec),
			}
			resources.UDPRoutes = append(resources.UDPRoutes, udpRoute)
		case gatewayapi.KindTLSRoute:
			typedSpec := spec.Interface()
			tlsRoute := &gwapiv1a2.TLSRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: gatewayapi.KindTLSRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1a2.TLSRouteSpec),
			}
			resources.TLSRoutes = append(resources.TLSRoutes, tlsRoute)
		case gatewayapi.KindHTTPRoute:
			typedSpec := spec.Interface()
			httpRoute := &gwapiv1.HTTPRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: gatewayapi.KindHTTPRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1.HTTPRouteSpec),
			}
			resources.HTTPRoutes = append(resources.HTTPRoutes, httpRoute)
		case gatewayapi.KindGRPCRoute:
			typedSpec := spec.Interface()
			grpcRoute := &gwapiv1a2.GRPCRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: gatewayapi.KindGRPCRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1a2.GRPCRouteSpec),
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
		case egv1a1.KindEnvoyPatchPolicy:
			typedSpec := spec.Interface()
			envoyPatchPolicy := &egv1a1.EnvoyPatchPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyPatchPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      name,
				},
				Spec: typedSpec.(egv1a1.EnvoyPatchPolicySpec),
			}
			resources.EnvoyPatchPolicies = append(resources.EnvoyPatchPolicies, envoyPatchPolicy)
		case egv1a1.KindClientTrafficPolicy:
			typedSpec := spec.Interface()
			clientTrafficPolicy := &egv1a1.ClientTrafficPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindClientTrafficPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      name,
				},
				Spec: typedSpec.(egv1a1.ClientTrafficPolicySpec),
			}
			resources.ClientTrafficPolicies = append(resources.ClientTrafficPolicies, clientTrafficPolicy)
		case egv1a1.KindBackendTrafficPolicy:
			typedSpec := spec.Interface()
			backendTrafficPolicy := &egv1a1.BackendTrafficPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindBackendTrafficPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      name,
				},
				Spec: typedSpec.(egv1a1.BackendTrafficPolicySpec),
			}
			resources.BackendTrafficPolicies = append(resources.BackendTrafficPolicies, backendTrafficPolicy)
		case egv1a1.KindSecurityPolicy:
			typedSpec := spec.Interface()
			securityPolicy := &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      name,
				},
				Spec: typedSpec.(egv1a1.SecurityPolicySpec),
			}
			resources.SecurityPolicies = append(resources.SecurityPolicies, securityPolicy)
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

		requiredServiceMap := map[string]*v1.Service{}
		for _, route := range resources.TCPRoutes {
			addMissingServices(requiredServiceMap, route)
		}
		for _, route := range resources.UDPRoutes {
			addMissingServices(requiredServiceMap, route)
		}
		for _, route := range resources.TLSRoutes {
			addMissingServices(requiredServiceMap, route)
		}
		for _, route := range resources.HTTPRoutes {
			addMissingServices(requiredServiceMap, route)
		}
		for _, route := range resources.GRPCRoutes {
			addMissingServices(requiredServiceMap, route)
		}

		providedServiceMap := map[string]*v1.Service{}
		for _, service := range resources.Services {
			providedServiceMap[service.Namespace+"/"+service.Name] = service
		}

		for key, service := range requiredServiceMap {
			if provided, found := providedServiceMap[key]; !found {
				resources.Services = append(resources.Services, service)
			} else {
				providedPorts := sets.NewString()
				for _, port := range provided.Spec.Ports {
					portKey := fmt.Sprintf("%s-%d", port.Protocol, port.Port)
					providedPorts.Insert(portKey)
				}

				for _, port := range service.Spec.Ports {
					name := fmt.Sprintf("%s-%d", port.Protocol, port.Port)
					if !providedPorts.Has(name) {
						servicePort := v1.ServicePort{
							Name:     name,
							Protocol: port.Protocol,
							Port:     port.Port,
						}
						provided.Spec.Ports = append(provided.Spec.Ports, servicePort)
					}
				}
			}
		}

		// Add EnvoyProxy if it does not exist
		if resources.EnvoyProxy == nil {
			if err := addDefaultEnvoyProxy(resources); err != nil {
				return nil, err
			}
		}
	}

	return resources, nil
}

func addDefaultEnvoyProxy(resources *gatewayapi.Resources) error {
	if resources.GatewayClass == nil {
		return fmt.Errorf("the GatewayClass resource is required")
	}

	defaultEnvoyProxyName := "default-envoy-proxy"
	namespace := resources.GatewayClass.Namespace
	defaultBootstrapStr, err := bootstrap.GetRenderedBootstrapConfig(nil)
	if err != nil {
		return err
	}
	ep := &egv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      defaultEnvoyProxyName,
		},
		Spec: egv1a1.EnvoyProxySpec{
			Bootstrap: &egv1a1.ProxyBootstrap{
				Value: defaultBootstrapStr,
			},
		},
	}
	resources.EnvoyProxy = ep
	ns := gwapiv1.Namespace(namespace)
	resources.GatewayClass.Spec.ParametersRef = &gwapiv1.ParametersReference{
		Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
		Kind:      gatewayapi.KindEnvoyProxy,
		Name:      defaultEnvoyProxyName,
		Namespace: &ns,
	}
	return nil
}
