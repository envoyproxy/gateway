// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	infra "github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	"github.com/envoyproxy/gateway/internal/xds/types"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
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
	translateCommand.MarkPersistentFlagRequired("file")
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
		return fmt.Errorf("%s is not a valid input type. %s", getValidInputTypesStr)
	}
	if !isValidOutputType(outType) {
		return fmt.Errorf("%s is not a valid output type. %s", getValidOutputTypesStr)
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
	resources := &gatewayapi.Resources{}

	// Unmarshal input
	err := yaml.UnmarshalStrict(inBytes, resources, yaml.DisallowUnknownFields)
	if err != nil {
		return fmt.Errorf("unable to unmarshal input: %w", err)
	}

	// Translate from Gateway API to Xds IR
	gTranslator := &gatewayapi.Translator{
		GatewayClassName:       resources.GatewayClass.Name,
		GlobalRateLimitEnabled: true,
	}
	gRes := gTranslator.Translate(resources)

	// Translate from Xds IR to Xds
	for key, val := range gRes.XdsIR {
		xTranslator := &translator.Translator{
			// Set some default settings for translation
			GlobalRateLimit: &xds.GlobalRateLimitSettings{
				ServiceURL: infra.GetRateLimitServiceURL("envoy-gateway"),
			},
		}
		xRes, err := xTranslator.Translate(val)
		if err != nil {
			return fmt.Errorf("failed to translate xds ir for key %s value %+v, error:%w", key, value, err)
		}

		// Print results
		printXds(w, xRes)
	}

	return nil
}

func printXds(w io.Writer, key string, tCtx *types.ResourceVersionTable) {
	fmt.Fprintf(w, "\nKey: %s", key)
	fmt.Fprintf(w, "\nTODO: Bootstrap:\n")
	listeners := tCtx.XdsResources[resource.ListenerType]
	fmt.Fprintf(w, "\nListeners:\n%s", resourcesToYAMLString(listeners))
	routes := tCtx.XdsResources[resource.RouteType]
	fmt.Fprintf(w, "\nRoutes:\n%s", resourcesToYAMLString(routes))
	clusters := tCtx.XdsResources[resource.ClusterType]
	fmt.Fprintf(w, "\nClusters:\n%s", resourcesToYAMLString(clusters))
}

func resourcesToYAMLString(resources []types.Resource) string {
	jsonBytes, err := xds_translator.MarshalResourcesToJSON(resources)
	data, err := yaml.JSONToYAML(jsonBytes)
	return string(data)
}
