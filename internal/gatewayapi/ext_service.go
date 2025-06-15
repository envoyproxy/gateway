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
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

// translateExtServiceBackendRefs translates external service backend references to route destinations.
func (t *Translator) translateExtServiceBackendRefs(
	policy client.Object,
	backendRefs []egv1a1.BackendRef,
	protocol ir.AppProtocol,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
	configType string,
	index int, // index is used to differentiate between multiple external services in the same policy
) (*ir.RouteDestination, error) {
	var (
		rs  *ir.RouteDestination
		ds  []*ir.DestinationSetting
		err error
	)

	if len(backendRefs) == 0 {
		return nil, errors.New("no backendRefs found for external service")
	}

	pnn := utils.NamespacedName(policy)
	destName := irIndexedExtServiceDestinationName(pnn, policy.GetObjectKind().GroupVersionKind().Kind, configType, index)
	for i, backendRef := range backendRefs {
		if err = t.validateExtServiceBackendReference(
			&backendRef.BackendObjectReference,
			policy.GetNamespace(),
			policy.GetObjectKind().GroupVersionKind().Kind,
			resources); err != nil {
			return nil, err
		}

		settingName := irDestinationSettingName(destName, i)
		var extServiceDest *ir.DestinationSetting
		if extServiceDest, err = t.processExtServiceDestination(
			settingName,
			&backendRef,
			pnn,
			policy.GetObjectKind().GroupVersionKind().Kind,
			protocol,
			resources,
			envoyProxy,
		); err != nil {
			return nil, err
		}
		ds = append(ds, extServiceDest)
	}

	rs = &ir.RouteDestination{
		Name:     destName,
		Settings: ds,
		Metadata: buildResourceMetadata(policy, nil),
	}

	if validationErr := rs.Validate(); validationErr != nil {
		return nil, validationErr
	}
	// TODO: Support mixed destinations for ext service
	if rs.HasMixedEndpoints() {
		return nil, errors.New("external service destinations having multiple endpoint types are not supported")
	}
	return rs, nil
}

func (t *Translator) processExtServiceDestination(
	settingName string,
	backendRef *egv1a1.BackendRef,
	policyNamespacedName types.NamespacedName,
	policyKind string,
	protocol ir.AppProtocol,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.DestinationSetting, error) {
	var (
		backendTLS *ir.TLSUpstreamConfig
		ds         *ir.DestinationSetting
		err        error
	)

	backendNamespace := NamespaceDerefOr(backendRef.Namespace, policyNamespacedName.Namespace)

	switch KindDerefOr(backendRef.Kind, resource.KindService) {
	case resource.KindService:
		ds, err = t.processServiceDestinationSetting(settingName, backendRef.BackendObjectReference, backendNamespace, protocol, resources, envoyProxy)
		if err != nil {
			return nil, err
		}
	case egv1a1.KindBackend:
		if !t.BackendEnabled {
			return nil, fmt.Errorf("resource %s of type Backend cannot be used since Backend is disabled in Envoy Gateway configuration", string(backendRef.Name))
		}
		ds = t.processBackendDestinationSetting(settingName, backendRef.BackendObjectReference, backendNamespace, protocol, resources)
		// Dynamic resolver destinations are not supported for none-route destinations
		if ds.IsDynamicResolver {
			return nil, errors.New("dynamic resolver destinations are not supported")
		}
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

	backendTLS, err = t.applyBackendTLSSetting(
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
		false,
	)
	if err != nil {
		return nil, err
	}

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

func irIndexedExtServiceDestinationName(policyNamespacedName types.NamespacedName, policyKind, configType string, idx int) string {
	return strings.ToLower(fmt.Sprintf(
		"%s/%s/%s/%s/%d",
		policyKind,
		policyNamespacedName.Namespace,
		policyNamespacedName.Name,
		configType,
		idx))
}
