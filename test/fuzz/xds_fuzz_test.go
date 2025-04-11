// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package fuzz

import (
	"testing"

	fuzz "github.com/AdaLogics/go-fuzz-headers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/cmd/egctl"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func FuzzGatewayAPIToXDS(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		fc := fuzz.NewConsumer(data)
		resources := &resource.Resources{}
		if err := fc.GenerateStruct(resources); err != nil {
			return
		}
		namespace, err := fc.GetString()
		if err != nil {
			return
		}
		dnsDomain, err := fc.GetString()
		if err != nil {
			return
		}
		resourceType, err := fc.GetString()
		if err != nil {
			return
		}

		// Populate default values
		yamlBytes, err := yaml.Marshal(resources)
		if err != nil {
			return
		}
		addMissingResources, err := fc.GetBool()
		if err != nil {
			return
		}
		rs, err := resource.LoadResourcesFromYAMLBytes(yamlBytes, addMissingResources)
		if err != nil {
			return
		}

		_, _ = egctl.TranslateGatewayAPIToXds(namespace, dnsDomain, resourceType, rs)
	})
}

func FuzzGatewayAPIToXDSWithGatewayClass(f *testing.F) {
	f.Fuzz(func(t *testing.T, name, controllerName, namespace, dnsDomain, resourceType string) {
		resources := &resource.Resources{
			GatewayClass: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gwapiv1.GatewayController(controllerName),
				},
			},
		}

		_, _ = egctl.TranslateGatewayAPIToXds(namespace, dnsDomain, resourceType, resources)
	})
}
