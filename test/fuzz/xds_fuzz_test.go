// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package fuzz

import (
	"testing"
	"unicode/utf8"

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

func FuzzGatewayClassToXDS(f *testing.F) {
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

// Following code implements a temporary fuzzer to simulate an error to verify OSS-Fuzz integration
// this test will be removed once the OSS-Fuzz integration is complete
// This example is based on the Go fuzzing example: https://go.dev/doc/tutorial/fuzz#code_to_test
// This test should generate errors when the input string is not a valid UTF-8 string

func ReverseStringBugSimulator(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < len(b)/2; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

func FuzzReverseStringBugSimulator(f *testing.F) {
	f.Fuzz(func(t *testing.T, orig string) {
		rev := ReverseStringBugSimulator(orig)
		doubleRev := ReverseStringBugSimulator(rev)
		if orig != doubleRev {
			t.Errorf("Before: %q, after: %q", orig, doubleRev)
		}
		if utf8.ValidString(orig) && !utf8.ValidString(rev) {
			t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
		}
	})
}
