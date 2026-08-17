// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

var tracer = otel.Tracer("envoy-gateway/gateway-api/translator")

// inputSizeAttrs reports the size of the resource tree a translation is about to
// process, so that a slow translation can be attributed to a bigger input instead
// of being correlated against other signals by hand.
//
// These are the counts of the resources as they arrive. Counts derived during the
// translation, for example the gateways that were accepted or the routes that
// attached to a listener, use distinct keys so that the two are never conflated.
func inputSizeAttrs(resources *resource.Resources) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Int("gateways.count", len(resources.Gateways)),
		attribute.Int("listener-sets.count", len(resources.ListenerSets)),
		attribute.Int("http-routes.count", len(resources.HTTPRoutes)),
		attribute.Int("grpc-routes.count", len(resources.GRPCRoutes)),
		attribute.Int("tls-routes.count", len(resources.TLSRoutes)),
		attribute.Int("tcp-routes.count", len(resources.TCPRoutes)),
		attribute.Int("udp-routes.count", len(resources.UDPRoutes)),
		attribute.Int("backends.count", len(resources.Backends)),
		attribute.Int("services.count", len(resources.Services)),
		attribute.Int("endpoint-slices.count", len(resources.EndpointSlices)),
		attribute.Int("client-traffic-policies.count", len(resources.ClientTrafficPolicies)),
		attribute.Int("backend-traffic-policies.count", len(resources.BackendTrafficPolicies)),
		attribute.Int("security-policies.count", len(resources.SecurityPolicies)),
		attribute.Int("backend-tls-policies.count", len(resources.BackendTLSPolicies)),
		attribute.Int("envoy-extension-policies.count", len(resources.EnvoyExtensionPolicies)),
		attribute.Int("extension-server-policies.count", len(resources.ExtensionServerPolicies)),
		attribute.Int("envoy-patch-policies.count", len(resources.EnvoyPatchPolicies)),
	}
}
