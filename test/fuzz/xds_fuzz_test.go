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

	"github.com/envoyproxy/gateway/internal/cmd/egctl"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func FuzzGatewayAPIToXDS(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		rs, err := resource.LoadResourcesFromYAMLBytes(b, true)
		if err != nil {
			return
		}
		fc := fuzz.NewConsumer(b)
		namespace, _ := fc.GetString()
		dnsDomain, _ := fc.GetString()
		resourceType, _ := fc.GetString()

		_, _ = egctl.TranslateGatewayAPIToXds(namespace, dnsDomain, resourceType, rs)
	})
}

func FuzzGatewayClassToXDS(f *testing.F) {
	f.Fuzz(func(t *testing.T, b []byte) {
		fc := fuzz.NewConsumer(b)
		namespace, _ := fc.GetString()
		name, _ := fc.GetString()
		controllerName, _ := fc.GetString()
		dnsDomain, _ := fc.GetString()
		resourceType, _ := fc.GetString()

		rs := &resource.Resources{
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

		_, _ = egctl.TranslateGatewayAPIToXds(namespace, dnsDomain, resourceType, rs)
	})
}
