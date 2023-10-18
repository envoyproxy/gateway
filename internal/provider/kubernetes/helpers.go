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
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	mcsapi "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/provider/utils"
)

const (
	gatewayClassFinalizer = gwapiv1b1.GatewayClassFinalizerGatewaysExist
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
	gatewayClassController gwapiv1b1.GatewayController,
	routeParentReferences []gwapiv1b1.ParentReference) ([]gwapiv1b1.Gateway, error) {

	var gateways []gwapiv1b1.Gateway
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
			gateways = append(gateways, *gw)
		}
	}

	return gateways, nil
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
	var gateways []gwapiv1b1.Gateway
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
func terminatesTLS(listener *gwapiv1b1.Listener) bool {
	if listener.TLS != nil &&
		(listener.Protocol == gwapiv1b1.HTTPSProtocolType ||
			listener.Protocol == gwapiv1b1.TLSProtocolType) &&
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
	infraName := utils.GetHashedName(fmt.Sprintf("%s/%s", gateway.Namespace, gateway.Name))
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, infraName)
}

func infraDeploymentName(gateway *gwapiv1b1.Gateway) string {
	infraName := utils.GetHashedName(fmt.Sprintf("%s/%s", gateway.Namespace, gateway.Name))
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, infraName)
}

// validateBackendRef validates that ref is a reference to a local Service.
// TODO: Add support for:
//   - Validating weights.
//   - Validating ports.
//   - Referencing HTTPRoutes.
func validateBackendRef(ref *gwapiv1b1.BackendRef) error {
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
func classRefsEnvoyProxy(gc *gwapiv1b1.GatewayClass, ep *egv1a1.EnvoyProxy) bool {
	if gc == nil || ep == nil {
		return false
	}

	return refsEnvoyProxy(gc) &&
		string(*gc.Spec.ParametersRef.Namespace) == ep.Namespace &&
		gc.Spec.ParametersRef.Name == ep.Name
}

// refsEnvoyProxy returns true if the provided GatewayClass references an EnvoyProxy.
func refsEnvoyProxy(gc *gwapiv1b1.GatewayClass) bool {
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
func classAccepted(gc *gwapiv1b1.GatewayClass) bool {
	if gc == nil {
		return false
	}

	for _, cond := range gc.Status.Conditions {
		if cond.Type == string(gwapiv1b1.GatewayClassConditionStatusAccepted) &&
			cond.Status == metav1.ConditionTrue {
			return true
		}
	}

	return false
}
