// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package fuzz

import (
	"strings"
	"testing"

	fuzz "github.com/AdaLogics/go-fuzz-headers"

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

		_, err = egctl.TranslateGatewayAPIToXds(namespace, dnsDomain, resourceType, rs)
		if err != nil && strings.Contains(err.Error(), "failed to translate xds") {
			t.Fatalf("%v", err)
		}
	})
}
