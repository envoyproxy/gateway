// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

// TODO: zhaohuabing combine this function with the one in the route translator
func (t *Translator) processExtServiceDestination(
	backendRef *egv1a1.BackendRef,
	policyNamespacedName types.NamespacedName,
	policyKind string,
	protocol ir.AppProtocol,
	resources *Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.DestinationSetting, error) {
	var (
		backendTLS *ir.TLSUpstreamConfig
		ds         *ir.DestinationSetting
	)

	backendNamespace := NamespaceDerefOr(backendRef.Namespace, policyNamespacedName.Namespace)

	switch KindDerefOr(backendRef.Kind, KindService) {
	case KindService:
		ds = t.processServiceDestinationSetting(backendRef.BackendObjectReference, backendNamespace, protocol, resources, envoyProxy)
	case egv1a1.KindBackend:
		if !t.BackendEnabled {
			return nil, fmt.Errorf("resource %s of type Backend cannot be used since Backend is disabled in Envoy Gateway configuration", string(backendRef.Name))
		}
		ds = t.processBackendDestinationSetting(backendRef.BackendObjectReference, backendNamespace, resources)
		ds.Protocol = protocol
	}

	if ds == nil {
		return nil, errors.New(
			"failed to translate external service backendRef")
	}

	// TODO: support mixed endpointslice address type for the same backendRef
	if !t.IsEnvoyServiceRouting(envoyProxy) && ds.AddressType != nil && *ds.AddressType == ir.MIXED {
		return nil, errors.New(
			"mixed endpointslice address type for the same backendRef is not supported")
	}

	backendTLS = t.applyBackendTLSSetting(
		backendRef.BackendObjectReference,
		backendNamespace,
		// Gateway is not the appropriate parent reference here because the owner
		// of the BackendRef is the policy, and there is no hierarchy
		// relationship between the policy and a gateway.
		// The owner policy of the BackendRef is used as the parent reference here.
		gwapiv1a2.ParentReference{
			Group:     ptr.To(gwapiv1.Group(egv1a1.GroupName)),
			Kind:      ptr.To(gwapiv1.Kind(policyKind)),
			Namespace: ptr.To(gwapiv1.Namespace(policyNamespacedName.Namespace)),
			Name:      gwapiv1.ObjectName(policyNamespacedName.Name),
		},
		resources,
		envoyProxy,
	)

	ds.TLS = backendTLS

	// TODO: support weighted non-xRoute backends
	ds.Weight = ptr.To(uint32(1))
	if backendRef.Fallback != nil {
		// set only the secondary priority, the backend defaults to a primary priority if unset.
		if ptr.Deref(backendRef.Fallback, false) {
			ds.Priority = ptr.To(uint32(1))
		}
	}
	return ds, nil
}

// TODO: also refer to extension type, as Wasm may also introduce destinations
func irIndexedExtServiceDestinationName(policyNamespacedName types.NamespacedName, policyKind string, idx int) string {
	return strings.ToLower(fmt.Sprintf(
		"%s/%s/%s/%d",
		policyKind,
		policyNamespacedName.Namespace,
		policyNamespacedName.Name,
		idx))
}
