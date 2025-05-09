// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package fuzz

import (
	"context"
	"testing"

	fuzz "github.com/AdaLogics/go-fuzz-headers"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/cmd/egctl"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/message"
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

		pResources := new(message.ProviderResources)
		loadAndReconcile := egctl.NewSimpleController(context.Background(), pResources, namespace)
		err = loadAndReconcile(rs)
		require.NoError(t, err)

		resources := pResources.GetResources()
		require.NotEmpty(t, resources)

		_, _ = egctl.TranslateGatewayAPIToXds(namespace, dnsDomain, resourceType, resources[0])
	})
}
