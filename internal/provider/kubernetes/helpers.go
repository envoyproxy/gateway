package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
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
