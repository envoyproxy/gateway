// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	mcsapi "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
)

const (
	gatewayClassFinalizer = gwapiv1.GatewayClassFinalizerGatewaysExist
)

type ObjectKindNamespacedName struct {
	kind      string
	namespace string
	name      string
}

// validateParentRefs validates the provided routeParentReferences, returning the
// referenced Gateways managed by Envoy Gateway. The only supported parentRef
// is a Gateway.
func validateParentRefs(ctx context.Context, client client.Client, namespace string,
	gatewayClassController gwapiv1.GatewayController,
	routeParentReferences []gwapiv1.ParentReference) ([]gwapiv1.Gateway, error) {

	var gateways []gwapiv1.Gateway
	for i := range routeParentReferences {
		ref := routeParentReferences[i]
		if ref.Kind != nil && *ref.Kind != "Gateway" {
			return nil, fmt.Errorf("invalid Kind %q", *ref.Kind)
		}
		if ref.Group != nil && *ref.Group != gwapiv1.GroupName {
			return nil, fmt.Errorf("invalid Group %q", *ref.Group)
		}

		// Ensure the referenced Gateway exists, using the route's namespace unless
		// specified by the parentRef.
		ns := namespace
		if ref.Namespace != nil {
			ns = string(*ref.Namespace)
		}
		gwKey := types.NamespacedName{
			Namespace: ns,
			Name:      string(ref.Name),
		}

		gw := new(gwapiv1.Gateway)
		if err := client.Get(ctx, gwKey, gw); err != nil {
			return nil, fmt.Errorf("failed to get gateway %s/%s: %w", gwKey.Namespace, gwKey.Name, err)
		}

		gcKey := types.NamespacedName{Name: string(gw.Spec.GatewayClassName)}
		gc := new(gwapiv1.GatewayClass)
		if err := client.Get(ctx, gcKey, gc); err != nil {
			return nil, fmt.Errorf("failed to get gatewayclass %s: %w", gcKey.Name, err)
		}
		if gc.Spec.ControllerName == gatewayClassController {
			gateways = append(gateways, *gw)
		}
	}

	return gateways, nil
}

type controlledClasses struct {
	// matchedClasses holds all GatewayClass objects with matching controllerName.
	matchedClasses []*gwapiv1.GatewayClass
}

func (cc *controlledClasses) addMatch(gc *gwapiv1.GatewayClass) {
	cc.matchedClasses = append(cc.matchedClasses, gc)
}

func (cc *controlledClasses) removeMatch(gc *gwapiv1.GatewayClass) {
	// First remove gc from matchedClasses.
	for i, matchedGC := range cc.matchedClasses {
		if matchedGC.Name == gc.Name {
			cc.matchedClasses[i] = cc.matchedClasses[len(cc.matchedClasses)-1]
			cc.matchedClasses = cc.matchedClasses[:len(cc.matchedClasses)-1]
			break
		}
	}
}

// isAccepted returns true if the provided gatewayclass contains the Accepted=true
// status condition.
func isAccepted(gc *gwapiv1.GatewayClass) bool {
	if gc == nil {
		return false
	}
	for _, cond := range gc.Status.Conditions {
		if cond.Type == string(gwapiv1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

// gatewaysOfClass returns a list of gateways that reference gc from the provided gwList.
func gatewaysOfClass(gc *gwapiv1.GatewayClass, gwList *gwapiv1.GatewayList) []gwapiv1.Gateway {
	var gateways []gwapiv1.Gateway
	if gwList == nil || gc == nil {
		return gateways
	}
	for i := range gwList.Items {
		gw := gwList.Items[i]
		if string(gw.Spec.GatewayClassName) == gc.Name {
			gateways = append(gateways, gw)
		}
	}
	return gateways
}

// terminatesTLS returns true if the provided gateway contains a listener configured
// for TLS termination.
func terminatesTLS(listener *gwapiv1.Listener) bool {
	if listener.TLS != nil &&
		(listener.Protocol == gwapiv1.HTTPSProtocolType ||
			listener.Protocol == gwapiv1.TLSProtocolType) &&
		listener.TLS.Mode != nil &&
		*listener.TLS.Mode == gwapiv1.TLSModeTerminate {
		return true
	}
	return false
}

// refsSecret returns true if ref refers to a Secret.
func refsSecret(ref *gwapiv1.SecretObjectReference) bool {
	return (ref.Group == nil || *ref.Group == corev1.GroupName) &&
		(ref.Kind == nil || *ref.Kind == gatewayapi.KindSecret)
}

// infraName returns expected name for the EnvoyProxy infra resources.
// By default it returns hashed string from {GatewayNamespace}/{GatewayName},
// but if mergeGateways is set, it will return hashed string of {GatewayClassName}.
func infraName(gateway *gwapiv1.Gateway, merged bool) string {
	if merged {
		return proxy.ExpectedResourceHashedName(string(gateway.Spec.GatewayClassName))
	}
	infraName := fmt.Sprintf("%s/%s", gateway.Namespace, gateway.Name)
	return proxy.ExpectedResourceHashedName(infraName)
}

// validateBackendRef validates that ref is a reference to a local Service.
// TODO: Add support for:
//   - Validating weights.
//   - Validating ports.
//   - Referencing HTTPRoutes.
func validateBackendRef(ref *gwapiv1.BackendRef) error {
	switch {
	case ref == nil:
		return nil
	case gatewayapi.GroupDerefOr(ref.Group, corev1.GroupName) != corev1.GroupName && gatewayapi.GroupDerefOr(ref.Group, corev1.GroupName) != mcsapi.GroupName:
		return fmt.Errorf("invalid group; must be nil, empty string or %q", mcsapi.GroupName)
	case gatewayapi.KindDerefOr(ref.Kind, gatewayapi.KindService) != gatewayapi.KindService && gatewayapi.KindDerefOr(ref.Kind, gatewayapi.KindService) != gatewayapi.KindServiceImport:
		return fmt.Errorf("invalid kind %q; must be %q or %q",
			*ref.BackendObjectReference.Kind, gatewayapi.KindService, gatewayapi.KindServiceImport)
	}

	return nil
}

// classRefsEnvoyProxy returns true if the provided GatewayClass references the provided EnvoyProxy.
func classRefsEnvoyProxy(gc *gwapiv1.GatewayClass, ep *egv1a1.EnvoyProxy) bool {
	if gc == nil || ep == nil {
		return false
	}

	return refsEnvoyProxy(gc) &&
		string(*gc.Spec.ParametersRef.Namespace) == ep.Namespace &&
		gc.Spec.ParametersRef.Name == ep.Name
}

// refsEnvoyProxy returns true if the provided GatewayClass references an EnvoyProxy.
func refsEnvoyProxy(gc *gwapiv1.GatewayClass) bool {
	if gc == nil {
		return false
	}

	return gc.Spec.ParametersRef != nil &&
		string(gc.Spec.ParametersRef.Group) == egv1a1.GroupVersion.Group &&
		gc.Spec.ParametersRef.Kind == egv1a1.KindEnvoyProxy &&
		gc.Spec.ParametersRef.Namespace != nil &&
		len(gc.Spec.ParametersRef.Name) > 0
}

// classAccepted returns true if the provided GatewayClass is accepted.
func classAccepted(gc *gwapiv1.GatewayClass) bool {
	if gc == nil {
		return false
	}

	for _, cond := range gc.Status.Conditions {
		if cond.Type == string(gwapiv1.GatewayClassConditionStatusAccepted) &&
			cond.Status == metav1.ConditionTrue {
			return true
		}
	}

	return false
}
