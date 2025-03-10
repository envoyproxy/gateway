// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package fuzz

import (
	"testing"

	fuzz "github.com/AdaLogics/go-fuzz-headers"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/cmd/egctl"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func FuzzGatewayAPIToIR(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		fc := fuzz.NewConsumer(data)
		resources := &resource.Resources{}
		if err := fc.GenerateStruct(resources); err != nil {
			return
		}
		addMissingResources, err := fc.GetBool()
		if err != nil {
			return
		}

		// Populate default values
		yamlBytes, err := yaml.Marshal(resources)
		if err != nil {
			return
		}
		rs, err := resource.LoadResourcesFromYAMLBytes(yamlBytes, addMissingResources)
		if err != nil {
			return
		}

		_, _ = egctl.TranslateGatewayAPIToIR(rs)
	})
}
