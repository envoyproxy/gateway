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
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/provider/utils"
)

const (
	kindTLSRoute  = "TLSRoute"
	kindHTTPRoute = "HTTPRoute"
	kindSecret    = "Secret"

	gatewayClassFinalizer = gwapiv1b1.GatewayClassFinalizerGatewaysExist
)

// validateParentRefs validates the provided routeParentReferences, returning the
// referenced Gateways managed by Envoy Gateway. The only supported parentRef
// is a Gateway.
func validateParentRefs(ctx context.Context, client client.Client, namespace string,
	gatewayClassController gwapiv1b1.GatewayController,
	routeParentReferences []gwapiv1b1.ParentReference) ([]gwapiv1b1.Gateway, error) {

	var ret []gwapiv1b1.Gateway
	for i := range routeParentReferences {
		ref := routeParentReferences[i]
		if ref.Kind != nil && *ref.Kind != "Gateway" {
			return nil, fmt.Errorf("invalid Kind %q", *ref.Kind)
		}
		if ref.Group != nil && *ref.Group != gwapiv1b1.GroupName {
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

		gw := new(gwapiv1b1.Gateway)
		if err := client.Get(ctx, gwKey, gw); err != nil {
			return nil, fmt.Errorf("failed to get gateway %s/%s: %v", gwKey.Namespace, gwKey.Name, err)
		}

		gcKey := types.NamespacedName{Name: string(gw.Spec.GatewayClassName)}
		gc := new(gwapiv1b1.GatewayClass)
		if err := client.Get(ctx, gcKey, gc); err != nil {
			return nil, fmt.Errorf("failed to get gatewayclass %s: %v", gcKey.Name, err)
		}
		if gc.Spec.ControllerName == gatewayClassController {
			ret = append(ret, *gw)
		}
	}

	return ret, nil
}

// isRoutePresentInNamespace checks if any kind of Routes - HTTPRoute, TLSRoute
// exists in the namespace ns.
func isRoutePresentInNamespace(ctx context.Context, c client.Client, ns string) (bool, error) {
	tlsRouteList := &gwapiv1a2.TLSRouteList{}
	if err := c.List(ctx, tlsRouteList, &client.ListOptions{Namespace: ns}); err != nil {
		return false, fmt.Errorf("error listing tlsroutes")
	}

	httpRouteList := &gwapiv1b1.HTTPRouteList{}
	if err := c.List(ctx, httpRouteList, &client.ListOptions{Namespace: ns}); err != nil {
		return false, fmt.Errorf("error listing httproutes")
	}

	if len(tlsRouteList.Items)+len(httpRouteList.Items) > 0 {
		return true, nil
	}
	return false, nil
}

type controlledClasses struct {
	// matchedClasses holds all GatewayClass objects with matching controllerName.
	matchedClasses []*gwapiv1b1.GatewayClass

	// oldestClass stores the first GatewayClass encountered with matching
	// controllerName. This is maintained so that the oldestClass does not change
	// during reboots.
	oldestClass *gwapiv1b1.GatewayClass
}

func (cc *controlledClasses) addMatch(gc *gwapiv1b1.GatewayClass) {
	cc.matchedClasses = append(cc.matchedClasses, gc)

	switch {
	case cc.oldestClass == nil:
		cc.oldestClass = gc
	case gc.CreationTimestamp.Time.Before(cc.oldestClass.CreationTimestamp.Time):
		cc.oldestClass = gc
	case gc.CreationTimestamp.Time.Equal(cc.oldestClass.CreationTimestamp.Time) && gc.Name < cc.oldestClass.Name:
		// tie-breaker: first one in alphabetical order is considered oldest/accepted
		cc.oldestClass = gc
	}
}

func (cc *controlledClasses) removeMatch(gc *gwapiv1b1.GatewayClass) {
	// First remove gc from matchedClasses.
	for i, matchedGC := range cc.matchedClasses {
		if matchedGC.Name == gc.Name {
			cc.matchedClasses[i] = cc.matchedClasses[len(cc.matchedClasses)-1]
			cc.matchedClasses = cc.matchedClasses[:len(cc.matchedClasses)-1]
			break
		}
	}

	// If the oldestClass is removed, find the new oldestClass candidate
	// from matchedClasses.
	if cc.oldestClass != nil && cc.oldestClass.Name == gc.Name {
		if len(cc.matchedClasses) == 0 {
			cc.oldestClass = nil
			return
		}

		cc.oldestClass = cc.matchedClasses[0]
		for i := 1; i < len(cc.matchedClasses); i++ {
			current := cc.matchedClasses[i]
			if current.CreationTimestamp.Time.Before(cc.oldestClass.CreationTimestamp.Time) ||
				(current.CreationTimestamp.Time.Equal(cc.oldestClass.CreationTimestamp.Time) &&
					current.Name < cc.oldestClass.Name) {
				cc.oldestClass = current
				return
			}
		}
	}
}

func (cc *controlledClasses) acceptedClass() *gwapiv1b1.GatewayClass {
	return cc.oldestClass
}

func (cc *controlledClasses) notAcceptedClasses() []*gwapiv1b1.GatewayClass {
	var res []*gwapiv1b1.GatewayClass
	for _, gc := range cc.matchedClasses {
		// skip the oldest one since it will be accepted.
		if gc.Name != cc.oldestClass.Name {
			res = append(res, gc)
		}
	}

	return res
}

// isAccepted returns true if the provided gatewayclass contains the Accepted=true
// status condition.
func isAccepted(gc *gwapiv1b1.GatewayClass) bool {
	if gc == nil {
		return false
	}
	for _, cond := range gc.Status.Conditions {
		if cond.Type == string(gwapiv1b1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

// gatewaysOfClass returns a list of gateways that reference gc from the provided gwList.
func gatewaysOfClass(gc *gwapiv1b1.GatewayClass, gwList *gwapiv1b1.GatewayList) []gwapiv1b1.Gateway {
	var ret []gwapiv1b1.Gateway
	if gwList == nil || gc == nil {
		return ret
	}
	for i := range gwList.Items {
		gw := gwList.Items[i]
		if string(gw.Spec.GatewayClassName) == gc.Name {
			ret = append(ret, gw)
		}
	}
	return ret
}

// terminatesTLS returns true if the provided gateway contains a listener configured
// for TLS termination.
func terminatesTLS(listener *gwapiv1b1.Listener) bool {
	if listener.TLS != nil &&
		listener.Protocol == gwapiv1b1.HTTPSProtocolType &&
		listener.TLS.Mode != nil &&
		*listener.TLS.Mode == gwapiv1b1.TLSModeTerminate {
		return true
	}
	return false
}

// refsSecret returns true if ref refers to a Secret.
func refsSecret(ref *gwapiv1b1.SecretObjectReference) bool {
	return (ref.Group == nil || *ref.Group == corev1.GroupName) &&
		(ref.Kind == nil || *ref.Kind == gatewayapi.KindSecret)
}

func infraServiceName(gateway *gwapiv1b1.Gateway) string {
	infraName := utils.GetHashedName(fmt.Sprintf("%s-%s", gateway.Namespace, gateway.Name))
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, infraName)
}

func infraDeploymentName(gateway *gwapiv1b1.Gateway) string {
	infraName := utils.GetHashedName(fmt.Sprintf("%s-%s", gateway.Namespace, gateway.Name))
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, infraName)
}

// findGatewayReferencesFromRefGrant helps in finding and aggregating all the
// Gateway references if present in the ReferenceGrant object ref.
func findGatewayReferencesFromRefGrant(ref *gwapiv1a2.ReferenceGrant) []types.NamespacedName {
	refs := []types.NamespacedName{}

	for _, to := range ref.Spec.To {
		if to.Group == gwapiv1a2.GroupName && to.Kind == gatewayapi.KindGateway && to.Name != nil {
			refs = append(refs, types.NamespacedName{
				Namespace: ref.Namespace,
				Name:      string(*to.Name),
			})
		}
	}
	for _, from := range ref.Spec.From {
		if from.Group == gwapiv1a2.GroupName && from.Kind == gatewayapi.KindGateway {
			refs = append(refs, types.NamespacedName{
				Namespace: string(from.Namespace),
				Name:      ref.Name,
			})
		}
	}

	return refs
}

// validateBackendRef validates that ref is a reference to a local Service.
// TODO: Add support for:
//   - Validating weights.
//   - Validating ports.
//   - Referencing HTTPRoutes.
//   - Referencing Services/HTTPRoutes from other namespaces using ReferenceGrant.
func validateBackendRef(ref *gwapiv1b1.BackendRef) error {
	switch {
	case ref == nil:
		return nil
	case ref.Group != nil && *ref.Group != corev1.GroupName:
		return fmt.Errorf("invalid group; must be nil or empty string")
	case ref.Kind != nil && *ref.Kind != gatewayapi.KindService:
		return fmt.Errorf("invalid kind %q; must be %q",
			*ref.BackendObjectReference.Kind, gatewayapi.KindService)
	case ref.Namespace != nil:
		return fmt.Errorf("invalid namespace; must be nil")
	}

	return nil
}
