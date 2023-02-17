// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	infra "github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	xds_types "github.com/envoyproxy/gateway/internal/xds/types"
	"github.com/envoyproxy/gateway/internal/xds/utils"
)

const (
	gatewayAPIType = "gateway-api"
	xdsType        = "xds"
)

func NewTranslateCommand() *cobra.Command {
	var (
		inFile, inType, outType string
	)

	translateCommand := &cobra.Command{
		Use:   "translate",
		Short: "Translate Configuration from an input type to an output type",
		RunE: func(cmd *cobra.Command, args []string) error {
			return translate(cmd.OutOrStdout(), inFile, inType, outType)
		},
	}

	translateCommand.PersistentFlags().StringVarP(&inFile, "file", "f", "", "Location of input file.")
	if err := translateCommand.MarkPersistentFlagRequired("file"); err != nil {
		return nil
	}
	translateCommand.PersistentFlags().StringVarP(&inType, "from", "", gatewayAPIType, getValidInputTypesStr())
	translateCommand.PersistentFlags().StringVarP(&outType, "to", "", xdsType, getValidOutputTypesStr())
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

func translate(w io.Writer, inFile, inType, outType string) error {
	if !isValidInputType(inType) {
		return fmt.Errorf("%s is not a valid input type. %s", inType, getValidInputTypesStr())
	}
	if !isValidOutputType(outType) {
		return fmt.Errorf("%s is not a valid output type. %s", outType, getValidOutputTypesStr())
	}
	inBytes, err := os.ReadFile(inFile)
	if err != nil {
		return fmt.Errorf("unable to read input file: %w", err)
	}

	if inType == gatewayAPIType && outType == xdsType {
		return translateFromGatewayAPIToXds(w, inBytes)
	}

	return fmt.Errorf("unable to find translate from input type %s to output type %s", inType, outType)
}

func translateFromGatewayAPIToXds(w io.Writer, inBytes []byte) error {

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
	fmt.Fprintf(w, "\nxDS\n")
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
		if err := printXds(w, key, xRes); err != nil {
			return fmt.Errorf("failed to print result for key %s value %+v, error:%w", key, val, err)
		}
	}

	return nil
}

// printXds prints the xDS Output
func printXds(w io.Writer, key string, tCtx *xds_types.ResourceVersionTable) error {
	fmt.Fprintf(w, "\nKey: %s\n", key)
	bootsrapYAML, err := infra.GetRenderedBootstrapConfig()
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "\nBootstrap:\n%s", bootsrapYAML)
	listeners := tCtx.XdsResources[resource.ListenerType]
	listenersYAML, err := resourcesToYAMLString(listeners)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "\nListeners:\n%s", listenersYAML)
	routes := tCtx.XdsResources[resource.RouteType]
	routesYAML, err := resourcesToYAMLString(routes)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "\nRoutes:\n%s", routesYAML)
	clusters := tCtx.XdsResources[resource.ClusterType]
	clustersYAML, err := resourcesToYAMLString(clusters)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "\nClusters:\n%s", clustersYAML)

	return nil
}

// resourcesToYAMLString converts xDS Resource types into YAML string
func resourcesToYAMLString(resources []types.Resource) (string, error) {
	jsonBytes, err := utils.MarshalResourcesToJSON(resources)
	if err != nil {
		return "", err
	}
	data, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		return "", err
	}
	return string(data), nil
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
