// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1alpha1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

// CustomGRPCRouteContext wraps a CustomGRPCRoute and provides helper methods for
// accessing the route's parents.
type CustomGRPCRouteContext struct {
	*v1alpha2.CustomGRPCRoute

	parentRefs map[v1beta1.ParentReference]*RouteParentContext
}

func (g *CustomGRPCRouteContext) GetRouteType() string {
	return KindCustomGRPCRoute
}

func (g *CustomGRPCRouteContext) GetHostnames() []string {
	hostnames := make([]string, len(g.Spec.Hostnames))
	for idx, s := range g.Spec.Hostnames {
		hostnames[idx] = string(s)
	}
	return hostnames
}

func (g *CustomGRPCRouteContext) GetParentReferences() []v1beta1.ParentReference {
	return g.Spec.ParentRefs
}

func (g *CustomGRPCRouteContext) GetRouteStatus() *v1beta1.RouteStatus {
	return &g.Status.RouteStatus
}

func (g *CustomGRPCRouteContext) GetRouteParentContext(forParentRef v1beta1.ParentReference) *RouteParentContext {
	if g.parentRefs == nil {
		g.parentRefs = make(map[v1beta1.ParentReference]*RouteParentContext)
	}

	if ctx := g.parentRefs[forParentRef]; ctx != nil {
		return ctx
	}

	var parentRef *v1beta1.ParentReference
	for i, p := range g.Spec.ParentRefs {
		p := UpgradeParentReference(p)
		if reflect.DeepEqual(p, forParentRef) {
			upgraded := UpgradeParentReference(g.Spec.ParentRefs[i])
			parentRef = &upgraded
			break
		}
	}
	if parentRef == nil {
		panic("parentRef not found")
	}

	routeParentStatusIdx := -1
	for i := range g.Status.Parents {
		p := UpgradeParentReference(g.Status.Parents[i].ParentRef)
		defaultNamespace := v1beta1.Namespace(metav1.NamespaceDefault)
		if forParentRef.Namespace == nil {
			forParentRef.Namespace = &defaultNamespace
		}
		if p.Namespace == nil {
			p.Namespace = &defaultNamespace
		}
		if reflect.DeepEqual(p, forParentRef) {
			routeParentStatusIdx = i
			break
		}
	}
	if routeParentStatusIdx == -1 {
		rParentStatus := v1alpha2.RouteParentStatus{
			// TODO: get this value from the config
			ControllerName: v1alpha2.GatewayController(egv1alpha1.GatewayControllerName),
			ParentRef:      DowngradeParentReference(forParentRef),
		}
		g.Status.Parents = append(g.Status.Parents, rParentStatus)
		routeParentStatusIdx = len(g.Status.Parents) - 1
	}

	ctx := &RouteParentContext{
		ParentReference: parentRef,

		customgrpcRoute:      g.CustomGRPCRoute,
		routeParentStatusIdx: routeParentStatusIdx,
	}
	g.parentRefs[forParentRef] = ctx
	return ctx
}
