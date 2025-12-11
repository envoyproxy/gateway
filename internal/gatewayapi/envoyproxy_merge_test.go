// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// TestEnvoyProxyTemplate tests the merging of EnvoyProxyTemplate with GatewayClass
// and Gateway-level EnvoyProxy configurations.
//
// Test cases are defined as YAML files in testdata/envoyproxy-merge/ directory:
// - <test-name>.in.yaml: Input containing envoyGateway config (with optional template) plus optional gatewayClass and gateway envoyProxies.
// - <test-name>.out.yaml: Expected merged EnvoyProxy output
func TestEnvoyProxyMerge(t *testing.T) {
	// Find all test input files
	inputFiles, err := filepath.Glob(filepath.Join("testdata", "envoyproxy-merge", "*.in.yaml"))
	require.NoError(t, err)
	require.NotEmpty(t, inputFiles, "No test input files found in testdata/envoyproxy-merge/")

	for _, inputFile := range inputFiles {
		// Extract test name from file path
		testName := filepath.Base(inputFile)
		testName = testName[:len(testName)-len(".in.yaml")]

		t.Run(testName, func(t *testing.T) {
			// Read input file
			inputData, err := os.ReadFile(inputFile)
			require.NoError(t, err)

			// Parse input
			input := &EnvoyProxyTemplateTestInput{}
			err = yaml.UnmarshalStrict(inputData, input)
			require.NoError(t, err)

			// Read expected output file
			outputFile := filepath.Join("testdata", "envoyproxy-merge", testName+".out.yaml")
			outputData, err := os.ReadFile(outputFile)
			require.NoError(t, err)

			// Parse expected output
			expected := &egv1a1.EnvoyProxy{}
			err = yaml.UnmarshalStrict(outputData, expected)
			require.NoError(t, err)

			// Extract template from EnvoyGateway config
			var template *egv1a1.EnvoyProxySpec
			if input.EnvoyGateway != nil && input.EnvoyGateway.GetEnvoyProxyTemplate() != nil {
				template = input.EnvoyGateway.GetEnvoyProxyTemplate()
			}

			// Get Gateway-level EnvoyProxy (use first one if multiple)
			var gatewayProxy *egv1a1.EnvoyProxy
			if len(input.EnvoyProxiesForGateways) > 0 {
				gatewayProxy = input.EnvoyProxiesForGateways[0]
			}

			// Perform the 3-level merge
			actual, err := MergeEnvoyProxyConfigs(
				template,
				input.EnvoyProxyForGatewayClass,
				gatewayProxy,
			)
			require.NoError(t, err)

			// Compare actual vs expected
			// Note: We only compare the Spec since metadata might differ
			require.Equal(t, expected.Spec, actual.Spec, "EnvoyProxy specs should match")
		})
	}
}

// EnvoyProxyTemplateTestInput represents the input for EnvoyProxyTemplate tests.
// This structure matches the YAML input files in testdata/envoyproxy-merge/
type EnvoyProxyTemplateTestInput struct {
	// EnvoyGateway config containing the EnvoyProxyTemplate
	EnvoyGateway *egv1a1.EnvoyGateway `json:"envoyGateway,omitempty" yaml:"envoyGateway,omitempty"`

	// GatewayClass-level EnvoyProxy (referenced via parametersRef)
	EnvoyProxyForGatewayClass *egv1a1.EnvoyProxy `json:"envoyProxyForGatewayClass,omitempty" yaml:"envoyProxyForGatewayClass,omitempty"`

	// Gateway-level EnvoyProxies (referenced via infrastructure.parametersRef)
	EnvoyProxiesForGateways []*egv1a1.EnvoyProxy `json:"envoyProxiesForGateways,omitempty" yaml:"envoyProxiesForGateways,omitempty"`
}
