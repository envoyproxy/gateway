// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"crypto/x509"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

// GatewayContext wraps a Gateway and provides helper methods for
// setting conditions, accessing Listeners, etc.
type GatewayContext struct {
	*gwapiv1.Gateway

	listeners  []*ListenerContext
	envoyProxy *egv1a1.EnvoyProxy
}

// ResetListeners resets the listener statuses and re-generates the GatewayContext
// ListenerContexts from the Gateway spec.
func (g *GatewayContext) ResetListeners(resource *resource.Resources) {
	numListeners := len(g.Spec.Listeners)
	g.Status.Listeners = make([]gwapiv1.ListenerStatus, numListeners)
	g.listeners = make([]*ListenerContext, numListeners)
	for i := range g.Spec.Listeners {
		listener := &g.Spec.Listeners[i]
		g.Status.Listeners[i] = gwapiv1.ListenerStatus{Name: listener.Name}
		g.listeners[i] = &ListenerContext{
			Listener:          listener,
			gateway:           g,
			listenerStatusIdx: i,
		}
	}

	g.attachEnvoyProxy(resource)
}

func (g *GatewayContext) attachEnvoyProxy(resources *resource.Resources) {
	if g.Spec.Infrastructure != nil && g.Spec.Infrastructure.ParametersRef != nil && !IsMergeGatewaysEnabled(resources) {
		ref := g.Spec.Infrastructure.ParametersRef
		if string(ref.Group) == egv1a1.GroupVersion.Group && ref.Kind == egv1a1.KindEnvoyProxy {
			ep := resources.GetEnvoyProxy(g.Namespace, ref.Name)
			if ep != nil {
				g.envoyProxy = ep
				return
			}
		}
		// not found, fallthrough to use envoyProxy attached to gatewayclass
	}

	g.envoyProxy = resources.EnvoyProxyForGatewayClass
}

// ListenerContext wraps a Listener and provides helper methods for
// setting conditions and other status information on the associated
// Gateway, etc.
type ListenerContext struct {
	*gwapiv1.Listener

	gateway           *GatewayContext
	listenerStatusIdx int
	namespaceSelector labels.Selector
	tlsSecrets        []*corev1.Secret
	tlsCertificates   []*x509.Certificate

	httpIR *ir.HTTPListener
}

func (l *ListenerContext) SetSupportedKinds(kinds ...gwapiv1.RouteGroupKind) {
	l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds = make([]gwapiv1.RouteGroupKind, 0, len(kinds))
	l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds = append(l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds, kinds...)
}

func (l *ListenerContext) IncrementAttachedRoutes() {
	l.gateway.Status.Listeners[l.listenerStatusIdx].AttachedRoutes++
}

func (l *ListenerContext) AttachedRoutes() int32 {
	return l.gateway.Status.Listeners[l.listenerStatusIdx].AttachedRoutes
}

func (l *ListenerContext) AllowsKind(kind gwapiv1.RouteGroupKind) bool {
	for _, allowed := range l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds {
		if GroupDerefOr(allowed.Group, "") == GroupDerefOr(kind.Group, "") &&
			allowed.Kind == kind.Kind {
			return true
		}
	}

	return false
}

func (l *ListenerContext) AllowsNamespace(namespace *corev1.Namespace) bool {
	if namespace == nil {
		return false
	}

	if l.AllowedRoutes == nil || l.AllowedRoutes.Namespaces == nil || l.AllowedRoutes.Namespaces.From == nil {
		return l.gateway.Namespace == namespace.Name
	}

	switch *l.AllowedRoutes.Namespaces.From {
	case gwapiv1.NamespacesFromAll:
		return true
	case gwapiv1.NamespacesFromSelector:
		if l.namespaceSelector == nil {
			return false
		}
		return l.namespaceSelector.Matches(labels.Set(namespace.Labels))
	default:
		// NamespacesFromSame is the default
		return l.gateway.Namespace == namespace.Name
	}
}

func (l *ListenerContext) IsReady() bool {
	for _, cond := range l.gateway.Status.Listeners[l.listenerStatusIdx].Conditions {
		if cond.Type == string(gwapiv1.ListenerConditionProgrammed) && cond.Status == metav1.ConditionTrue {
			return true
		}
	}

	return false
}

func (l *ListenerContext) GetConditions() []metav1.Condition {
	return l.gateway.Status.Listeners[l.listenerStatusIdx].Conditions
}

func (l *ListenerContext) SetTLSSecrets(tlsSecrets []*corev1.Secret) {
	l.tlsSecrets = tlsSecrets
}

func (l *ListenerContext) SetTLSCertificates(certs []*x509.Certificate) {
	l.tlsCertificates = certs
}

// RouteContext represents a generic Route object (HTTPRoute, TLSRoute, etc.)
// that can reference Gateway objects.
type RouteContext interface {
	client.Object
}

// HTTPRouteContext wraps an HTTPRoute and provides helper methods for
// accessing the route's parents.
type HTTPRouteContext struct {
	// GatewayControllerName is the name of the Gateway API controller.
	GatewayControllerName string

	*gwapiv1.HTTPRoute

	ParentRefs map[gwapiv1.ParentReference]*RouteParentContext
}

// GRPCRouteContext wraps a GRPCRoute and provides helper methods for
// accessing the route's parents.
type GRPCRouteContext struct {
	// GatewayControllerName is the name of the Gateway API controller.
	GatewayControllerName string

	*gwapiv1.GRPCRoute

	ParentRefs map[gwapiv1.ParentReference]*RouteParentContext
}

// TLSRouteContext wraps a TLSRoute and provides helper methods for
// accessing the route's parents.
type TLSRouteContext struct {
	// GatewayControllerName is the name of the Gateway API controller.
	GatewayControllerName string

	*gwapiv1a2.TLSRoute

	ParentRefs map[gwapiv1.ParentReference]*RouteParentContext
}

// UDPRouteContext wraps a UDPRoute and provides helper methods for
// accessing the route's parents.
type UDPRouteContext struct {
	// GatewayControllerName is the name of the Gateway API controller.
	GatewayControllerName string

	*gwapiv1a2.UDPRoute

	ParentRefs map[gwapiv1.ParentReference]*RouteParentContext
}

// TCPRouteContext wraps a TCPRoute and provides helper methods for
// accessing the route's parents.
type TCPRouteContext struct {
	// GatewayControllerName is the name of the Gateway API controller.
	GatewayControllerName string

	*gwapiv1a2.TCPRoute

	ParentRefs map[gwapiv1.ParentReference]*RouteParentContext
}

// GetRouteType returns the Kind of the Route object, HTTPRoute,
// TLSRoute, TCPRoute, UDPRoute etc.
func GetRouteType(route RouteContext) gwapiv1.Kind {
	rv := reflect.ValueOf(route).Elem()
	return gwapiv1.Kind(rv.FieldByName("Kind").String())
}

// GetHostnames returns the hosts targeted by the Route object.
func GetHostnames(route RouteContext) []string {
	rv := reflect.ValueOf(route).Elem()
	kind := rv.FieldByName("Kind").String()
	if kind == resource.KindTCPRoute || kind == resource.KindUDPRoute {
		return nil
	}

	hs := rv.FieldByName("Spec").FieldByName("Hostnames")
	hostnames := make([]string, hs.Len())
	for i := 0; i < len(hostnames); i++ {
		hostnames[i] = hs.Index(i).String()
	}
	return hostnames
}

// GetParentReferences returns the ParentReference of the Route object.
func GetParentReferences(route RouteContext) []gwapiv1.ParentReference {
	rv := reflect.ValueOf(route).Elem()
	pr := rv.FieldByName("Spec").FieldByName("ParentRefs")
	return pr.Interface().([]gwapiv1.ParentReference)
}

// GetRouteStatus returns the RouteStatus object associated with the Route.
func GetRouteStatus(route RouteContext) *gwapiv1.RouteStatus {
	rv := reflect.ValueOf(route).Elem()
	rs := rv.FieldByName("Status").FieldByName("RouteStatus").Interface().(gwapiv1.RouteStatus)
	return &rs
}

// GetRouteParentContext returns RouteParentContext by using the Route objects' ParentReference.
// It creates a new RouteParentContext and add a new RouteParentStatus to the Route's Status if the ParentReference is not found.
func GetRouteParentContext(route RouteContext, forParentRef gwapiv1.ParentReference) *RouteParentContext {
	rv := reflect.ValueOf(route).Elem()
	pr := rv.FieldByName("ParentRefs")

	// If the ParentRefs field is nil, initialize it.
	if pr.IsNil() {
		mm := reflect.MakeMap(reflect.TypeOf(map[gwapiv1.ParentReference]*RouteParentContext{}))
		pr.Set(mm)
	}

	// If the RouteParentContext is already in the RouteContext, return it.
	if p := pr.MapIndex(reflect.ValueOf(forParentRef)); p.IsValid() && !p.IsZero() {
		ctx := p.Interface().(*RouteParentContext)
		return ctx
	}

	// Verify that the ParentReference is present in the Route.Spec.ParentRefs.
	// This is just a sanity check, the parentRef should always be present, otherwise it's a programming error.
	var parentRef *gwapiv1.ParentReference
	specParentRefs := rv.FieldByName("Spec").FieldByName("ParentRefs")
	for i := 0; i < specParentRefs.Len(); i++ {
		p := specParentRefs.Index(i).Interface().(gwapiv1.ParentReference)
		if reflect.DeepEqual(p, forParentRef) {
			parentRef = &p
			break
		}
	}
	if parentRef == nil {
		panic("parentRef not found")
	}

	// Find the parent in the Route's Status.
	routeParentStatusIdx := -1
	statusParents := rv.FieldByName("Status").FieldByName("Parents")

	for i := 0; i < statusParents.Len(); i++ {
		p := statusParents.Index(i).FieldByName("ParentRef").Interface().(gwapiv1.ParentReference)
		if isParentRefEqual(p, *parentRef, route.GetNamespace()) {
			routeParentStatusIdx = i
			break
		}
	}

	// If the parent is not found in the Route's Status, create a new RouteParentStatus and add it to the Route's Status.
	if routeParentStatusIdx == -1 {
		rParentStatus := gwapiv1a2.RouteParentStatus{
			ControllerName: gwapiv1a2.GatewayController(rv.FieldByName("GatewayControllerName").String()),
			ParentRef:      forParentRef,
		}
		statusParents.Set(reflect.Append(statusParents, reflect.ValueOf(rParentStatus)))
		routeParentStatusIdx = statusParents.Len() - 1
	}

	// Also add the RouteParentContext to the RouteContext.
	ctx := &RouteParentContext{
		ParentReference:      parentRef,
		routeParentStatusIdx: routeParentStatusIdx,
	}
	rctx := reflect.ValueOf(ctx)
	rctx.Elem().FieldByName(string(GetRouteType(route))).Set(rv.Field(1))
	pr.SetMapIndex(reflect.ValueOf(forParentRef), rctx)
	return ctx
}

func isParentRefEqual(ref1, ref2 gwapiv1.ParentReference, routeNS string) bool {
	defaultGroup := (*gwapiv1.Group)(&gwapiv1.GroupVersion.Group)
	if ref1.Group == nil {
		ref1.Group = defaultGroup
	}
	if ref2.Group == nil {
		ref2.Group = defaultGroup
	}

	defaultKind := gwapiv1.Kind(resource.KindGateway)
	if ref1.Kind == nil {
		ref1.Kind = &defaultKind
	}
	if ref2.Kind == nil {
		ref2.Kind = &defaultKind
	}

	// If the parent's namespace is not set, default to the namespace of the Route.
	defaultNS := gwapiv1.Namespace(routeNS)
	if ref1.Namespace == nil {
		ref1.Namespace = &defaultNS
	}
	if ref2.Namespace == nil {
		ref2.Namespace = &defaultNS
	}
	return reflect.DeepEqual(ref1, ref2)
}

// RouteParentContext wraps a ParentReference and provides helper methods for
// setting conditions and other status information on the associated
// HTTPRoute, TLSRoute etc.
type RouteParentContext struct {
	*gwapiv1.ParentReference

	// TODO: [v1alpha2-gwapiv1] This can probably be replaced with
	// a single field pointing to *gwapiv1.RouteStatus.
	HTTPRoute *gwapiv1.HTTPRoute
	GRPCRoute *gwapiv1.GRPCRoute
	TLSRoute  *gwapiv1a2.TLSRoute
	TCPRoute  *gwapiv1a2.TCPRoute
	UDPRoute  *gwapiv1a2.UDPRoute

	routeParentStatusIdx int
	listeners            []*ListenerContext
}

// GetGateway returns the GatewayContext if parent resource is a gateway.
func (r *RouteParentContext) GetGateway() *GatewayContext {
	if r == nil || len(r.listeners) == 0 {
		return nil
	}
	return r.listeners[0].gateway
}

func (r *RouteParentContext) SetListeners(listeners ...*ListenerContext) {
	r.listeners = append(r.listeners, listeners...)
}

func (r *RouteParentContext) ResetConditions(route RouteContext) {
	routeStatus := GetRouteStatus(route)
	routeStatus.Parents[r.routeParentStatusIdx].Conditions = make([]metav1.Condition, 0)
}

func (r *RouteParentContext) HasCondition(route RouteContext, condType gwapiv1.RouteConditionType, status metav1.ConditionStatus) bool {
	var conditions []metav1.Condition
	routeStatus := GetRouteStatus(route)
	conditions = routeStatus.Parents[r.routeParentStatusIdx].Conditions
	for _, c := range conditions {
		if c.Type == string(condType) && c.Status == status {
			return true
		}
	}
	return false
}

// BackendRefContext represents a generic BackendRef object (HTTPBackendRef, GRPCBackendRef or BackendRef itself)
type BackendRefContext any

func GetBackendRef(b BackendRefContext) *gwapiv1.BackendRef {
	rv := reflect.ValueOf(b)
	br := rv.FieldByName("BackendRef")
	if br.IsValid() {
		backendRef := br.Interface().(gwapiv1.BackendRef)
		return &backendRef
	}

	backendRef := b.(gwapiv1.BackendRef)
	return &backendRef
}

func GetFilters(b BackendRefContext) any {
	filters := reflect.ValueOf(b).FieldByName("Filters")
	if !filters.IsValid() {
		return nil
	}
	return filters.Interface()
}
