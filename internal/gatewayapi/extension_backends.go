// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

func (t *Translator) buildExtensionBackend(policy *egv1a1.EnvoyExtensionPolicy, backend egv1a1.ExtensionBackend,
	resources *resource.Resources, gateway *GatewayContext, moduleIndex, backendIndex int,
) (*ir.RouteDestination, error) {
	if problems := validation.IsDNS1123Subdomain(backend.Name); len(problems) > 0 {
		return nil, fmt.Errorf("cluster name %q is invalid: %s", backend.Name, strings.Join(problems, ", "))
	}
	// The service cluster exists in bootstrap before its Service appears in the resource tree.
	if backend.Name == t.getIRKey(gateway.Gateway) {
		return nil, fmt.Errorf("cluster name %q is reserved for Envoy Gateway's service cluster", backend.Name)
	}
	if backend.Name == "tracing" {
		return nil, fmt.Errorf("cluster name %q is reserved for Envoy Gateway's tracing cluster", backend.Name)
	}
	destination, err := t.translateExtServiceBackendRefs(policy,
		[]egv1a1.BackendRef{{BackendObjectReference: backend.BackendRef}}, ir.HTTP,
		resources, gateway, fmt.Sprintf("dynamic-module/%d", moduleIndex), backendIndex)
	if err != nil {
		return nil, err
	}
	destination.Name = backend.Name
	return destination, nil
}

type extensionBackendRoute struct {
	owner    types.NamespacedName
	consumer types.NamespacedName
	listener *ListenerContext
}

func (t *Translator) recordExtensionBackendRoute(route *ir.HTTPRoute, consumer, owner *egv1a1.EnvoyExtensionPolicy, listener *ListenerContext) {
	for _, dm := range route.EnvoyExtensions.DynamicModules {
		if len(dm.Backends) > 0 {
			t.extensionBackendRoutes[route] = extensionBackendRoute{
				owner: utils.NamespacedName(owner), consumer: utils.NamespacedName(consumer), listener: listener,
			}
			return
		}
	}
}

type extensionBackendRef struct {
	group     gwapiv1.Group
	kind      gwapiv1.Kind
	namespace string
	name      gwapiv1.ObjectName
	port      gwapiv1.PortNumber
}

func normalizeExtensionBackendRef(ref gwapiv1.BackendObjectReference, namespace string) extensionBackendRef {
	return extensionBackendRef{
		group: ptr.Deref(ref.Group, ""), kind: ptr.Deref(ref.Kind, resource.KindService),
		namespace: NamespaceDerefOr(ref.Namespace, namespace), name: ref.Name, port: ptr.Deref(ref.Port, 0),
	}
}

func (ref extensionBackendRef) String() string {
	backend := fmt.Sprintf("%s %s/%s", ref.kind, ref.namespace, ref.name)
	if ref.port != 0 {
		backend += fmt.Sprintf(":%d", ref.port)
	}
	return backend
}

type extensionBackendClaim struct {
	owner       *egv1a1.EnvoyExtensionPolicy
	location    string
	ref         extensionBackendRef
	destination *ir.RouteDestination
}

func (t *Translator) resolveExtensionBackendConflicts(policies []*egv1a1.EnvoyExtensionPolicy) {
	if len(t.extensionBackendRoutes) == 0 {
		return
	}
	byDeployment := make(map[string]map[types.NamespacedName][]*ir.HTTPRoute)
	for route, attachment := range t.extensionBackendRoutes {
		key := t.getIRKey(attachment.listener.gateway.Gateway)
		if byDeployment[key] == nil {
			byDeployment[key] = make(map[types.NamespacedName][]*ir.HTTPRoute)
		}
		byDeployment[key][attachment.owner] = append(byDeployment[key][attachment.owner], route)
	}
	byName := make(map[types.NamespacedName]*egv1a1.EnvoyExtensionPolicy, len(policies))
	for _, policy := range policies {
		byName[utils.NamespacedName(policy)] = policy
	}
	// Attachment traversal favors specific targets. Cluster ownership must depend only on age.
	policies = slices.Clone(policies)
	slices.SortFunc(policies, func(a, b *egv1a1.EnvoyExtensionPolicy) int {
		if order := a.CreationTimestamp.Compare(b.CreationTimestamp.Time); order != 0 {
			return order
		}
		if order := cmp.Compare(a.Namespace, b.Namespace); order != 0 {
			return order
		}
		return cmp.Compare(a.Name, b.Name)
	})
	for _, owners := range byDeployment {
		claimed := make(map[string]extensionBackendClaim)
		for _, policy := range policies {
			routes := owners[utils.NamespacedName(policy)]
			if len(routes) == 0 {
				continue
			}
			slices.SortFunc(routes, func(a, b *ir.HTTPRoute) int {
				if order := cmp.Compare(irListenerName(t.extensionBackendRoutes[a].listener), irListenerName(t.extensionBackendRoutes[b].listener)); order != 0 {
					return order
				}
				return cmp.Compare(a.Name, b.Name)
			})
			pending := make(map[string]extensionBackendClaim)
			var message string
			reason := gwapiv1.PolicyReasonConflicted
			for moduleIndex, dm := range routes[0].EnvoyExtensions.DynamicModules {
				for backendIndex, destination := range dm.Backends {
					backend := policy.Spec.DynamicModule[moduleIndex].Backends[backendIndex]
					claim := extensionBackendClaim{
						owner: policy, location: fmt.Sprintf("dynamicModule[%d].backends[%d]", moduleIndex, backendIndex),
						ref: normalizeExtensionBackendRef(backend.BackendRef, policy.Namespace), destination: destination,
					}
					previous, exists := pending[destination.Name]
					if exists && previous.ref != claim.ref {
						reason = gwapiv1.PolicyReasonInvalid
					} else {
						previous, exists = claimed[destination.Name]
					}
					if exists && previous.ref != claim.ref {
						message = fmt.Sprintf("%s: cluster %q references %s, but EnvoyExtensionPolicy %s/%s %s already declares it for %s. Choose another cluster name and update the module configuration to match, or use the same backend reference.",
							claim.location, destination.Name, claim.ref, previous.owner.Namespace, previous.owner.Name, previous.location, previous.ref)
						break
					}
					pending[destination.Name] = claim
				}
				if message != "" {
					break
				}
			}
			if message != "" {
				for _, route := range routes {
					attachment := t.extensionBackendRoutes[route]
					route.EnvoyExtensions = nil
					route.DirectResponse = &ir.CustomResponse{StatusCode: new(uint32(500))}
					t.setExtensionBackendConflict(policy, attachment.listener, reason, message)
					if attachment.consumer != attachment.owner {
						t.setExtensionBackendConflict(byName[attachment.consumer], attachment.listener, reason, message)
					}
				}
				continue
			}
			// Publish only after every declaration passes, so rejected policies reserve no names.
			for name, claim := range pending {
				if _, exists := claimed[name]; !exists {
					claimed[name] = claim
				}
			}
			for _, route := range routes {
				for _, dm := range route.EnvoyExtensions.DynamicModules {
					for index, destination := range dm.Backends {
						dm.Backends[index] = claimed[destination.Name].destination
					}
				}
			}
		}
	}
}

func (t *Translator) setExtensionBackendConflict(policy *egv1a1.EnvoyExtensionPolicy, listener *ListenerContext,
	reason gwapiv1.PolicyConditionReason, message string,
) {
	for _, ancestor := range policy.Status.Ancestors {
		ref := ancestor.AncestorRef
		if ref.SectionName != nil && *ref.SectionName != listener.Name {
			continue
		}
		namespace := NamespaceDerefOr(ref.Namespace, policy.Namespace)
		matches := ptr.Deref(ref.Kind, resource.KindGateway) == resource.KindGateway &&
			namespace == listener.gateway.Namespace && string(ref.Name) == listener.gateway.Name
		if ptr.Deref(ref.Kind, resource.KindGateway) == resource.KindListenerSet && listener.listenerSet != nil {
			matches = namespace == listener.listenerSet.Namespace && string(ref.Name) == listener.listenerSet.Name
		}
		if matches && string(ancestor.ControllerName) == t.GatewayControllerName {
			status.SetConditionForPolicyAncestor(&policy.Status, &ref, t.GatewayControllerName,
				gwapiv1.PolicyConditionAccepted, metav1.ConditionFalse, reason, message, policy.Generation)
		}
	}
}
