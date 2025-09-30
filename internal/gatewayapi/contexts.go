// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
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
	certDNSNames      []string

	httpIR *ir.HTTPListener
}

func (l *ListenerContext) SetSupportedKinds(kinds ...gwapiv1.RouteGroupKind) {
	if len(kinds) > 0 {
		l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds = make([]gwapiv1.RouteGroupKind, len(kinds))
		copy(l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds, kinds)
	} else {
		l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds = []gwapiv1.RouteGroupKind{}
	}
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

// RouteContext represents a generic Route object (HTTPRoute, TLSRoute, etc.)
// that can reference Gateway objects.
type RouteContext interface {
	client.Object
	GetRouteType() gwapiv1.Kind
	HasRuleNames(sectionName gwapiv1.SectionName) bool
	GetHostnames() []string
	GetParentReferences() []gwapiv1.ParentReference
	GetRouteStatus() *gwapiv1.RouteStatus
	GetRouteParentContext(forParentRef gwapiv1.ParentReference) *RouteParentContext
	SetRouteParentContext(forParentRef gwapiv1.ParentReference, ctx *RouteParentContext)
}

// HTTPRouteContext wraps an HTTPRoute and provides helper methods for
// accessing the route's parents.
type HTTPRouteContext struct {
	*gwapiv1.HTTPRoute

	ParentRefs map[gwapiv1.ParentReference]*RouteParentContext
}

func (r *HTTPRouteContext) GetRouteType() gwapiv1.Kind {
	return resource.KindHTTPRoute
}

func (r *HTTPRouteContext) HasRuleNames(sectionName gwapiv1.SectionName) bool {
	rs := r.Spec.Rules
	for _, rule := range rs {
		if rule.Name != nil {
			if *rule.Name == sectionName {
				return true
			}
		}
	}
	return false
}

func (r *HTTPRouteContext) GetHostnames() []string {
	hs := r.Spec.Hostnames
	hostnames := make([]string, len(hs))
	for i := range hs {
		hostnames[i] = string(hs[i])
	}
	return hostnames
}

func (r *HTTPRouteContext) GetParentReferences() []gwapiv1.ParentReference {
	return r.Spec.ParentRefs
}

func (r *HTTPRouteContext) GetRouteStatus() *gwapiv1.RouteStatus {
	return &r.Status.RouteStatus
}

func (r *HTTPRouteContext) GetRouteParentContext(forParentRef gwapiv1.ParentReference) *RouteParentContext {
	if r.ParentRefs == nil {
		r.ParentRefs = make(map[gwapiv1.ParentReference]*RouteParentContext)
	}
	return r.ParentRefs[forParentRef]
}

func (r *HTTPRouteContext) SetRouteParentContext(forParentRef gwapiv1.ParentReference, ctx *RouteParentContext) {
	if r.ParentRefs == nil {
		r.ParentRefs = make(map[gwapiv1.ParentReference]*RouteParentContext)
	}
	r.ParentRefs[forParentRef] = ctx
}

// GRPCRouteContext wraps a GRPCRoute and provides helper methods for
// accessing the route's parents.
type GRPCRouteContext struct {
	*gwapiv1.GRPCRoute

	ParentRefs map[gwapiv1.ParentReference]*RouteParentContext
}

func (r *GRPCRouteContext) GetRouteType() gwapiv1.Kind {
	return resource.KindGRPCRoute
}

func (r *GRPCRouteContext) HasRuleNames(sectionName gwapiv1.SectionName) bool {
	rs := r.Spec.Rules
	for _, rule := range rs {
		if rule.Name != nil {
			if *rule.Name == sectionName {
				return true
			}
		}
	}
	return false
}

func (r *GRPCRouteContext) GetHostnames() []string {
	hs := r.Spec.Hostnames
	hostnames := make([]string, len(hs))
	for i := range hs {
		hostnames[i] = string(hs[i])
	}
	return hostnames
}

func (r *GRPCRouteContext) GetParentReferences() []gwapiv1.ParentReference {
	return r.Spec.ParentRefs
}

func (r *GRPCRouteContext) GetRouteStatus() *gwapiv1.RouteStatus {
	return &r.Status.RouteStatus
}

func (r *GRPCRouteContext) GetRouteParentContext(forParentRef gwapiv1.ParentReference) *RouteParentContext {
	if r.ParentRefs == nil {
		r.ParentRefs = make(map[gwapiv1.ParentReference]*RouteParentContext)
	}
	return r.ParentRefs[forParentRef]
}

func (r *GRPCRouteContext) SetRouteParentContext(forParentRef gwapiv1.ParentReference, ctx *RouteParentContext) {
	if r.ParentRefs == nil {
		r.ParentRefs = make(map[gwapiv1.ParentReference]*RouteParentContext)
	}
	r.ParentRefs[forParentRef] = ctx
}

// TLSRouteContext wraps a TLSRoute and provides helper methods for
// accessing the route's parents.
type TLSRouteContext struct {
	*gwapiv1a2.TLSRoute

	ParentRefs map[gwapiv1.ParentReference]*RouteParentContext
}

func (r *TLSRouteContext) GetRouteType() gwapiv1.Kind {
	return resource.KindTLSRoute
}

func (r *TLSRouteContext) HasRuleNames(sectionName gwapiv1.SectionName) bool {
	rs := r.Spec.Rules
	for _, rule := range rs {
		if rule.Name != nil {
			if *rule.Name == sectionName {
				return true
			}
		}
	}
	return false
}

func (r *TLSRouteContext) GetHostnames() []string {
	hs := r.Spec.Hostnames
	hostnames := make([]string, len(hs))
	for i := range hs {
		hostnames[i] = string(hs[i])
	}
	return hostnames
}

func (r *TLSRouteContext) GetParentReferences() []gwapiv1.ParentReference {
	return r.Spec.ParentRefs
}

func (r *TLSRouteContext) GetRouteStatus() *gwapiv1.RouteStatus {
	return &r.Status.RouteStatus
}

func (r *TLSRouteContext) GetRouteParentContext(forParentRef gwapiv1.ParentReference) *RouteParentContext {
	if r.ParentRefs == nil {
		r.ParentRefs = make(map[gwapiv1.ParentReference]*RouteParentContext)
	}
	return r.ParentRefs[forParentRef]
}

func (r *TLSRouteContext) SetRouteParentContext(forParentRef gwapiv1.ParentReference, ctx *RouteParentContext) {
	if r.ParentRefs == nil {
		r.ParentRefs = make(map[gwapiv1.ParentReference]*RouteParentContext)
	}
	r.ParentRefs[forParentRef] = ctx
}

// UDPRouteContext wraps a UDPRoute and provides helper methods for
// accessing the route's parents.
type UDPRouteContext struct {
	*gwapiv1a2.UDPRoute

	ParentRefs map[gwapiv1.ParentReference]*RouteParentContext
}

func (r *UDPRouteContext) GetRouteType() gwapiv1.Kind {
	return resource.KindUDPRoute
}

func (r *UDPRouteContext) HasRuleNames(sectionName gwapiv1.SectionName) bool {
	rs := r.Spec.Rules
	for _, rule := range rs {
		if rule.Name != nil {
			if *rule.Name == sectionName {
				return true
			}
		}
	}
	return false
}

func (r *UDPRouteContext) GetHostnames() []string {
	// UDPRoute doesn't have hostnames, return empty slice
	return []string{}
}

func (r *UDPRouteContext) GetParentReferences() []gwapiv1.ParentReference {
	return r.Spec.ParentRefs
}

func (r *UDPRouteContext) GetRouteStatus() *gwapiv1.RouteStatus {
	return &r.Status.RouteStatus
}

func (r *UDPRouteContext) GetRouteParentContext(forParentRef gwapiv1.ParentReference) *RouteParentContext {
	if r.ParentRefs == nil {
		r.ParentRefs = make(map[gwapiv1.ParentReference]*RouteParentContext)
	}
	return r.ParentRefs[forParentRef]
}

func (r *UDPRouteContext) SetRouteParentContext(forParentRef gwapiv1.ParentReference, ctx *RouteParentContext) {
	if r.ParentRefs == nil {
		r.ParentRefs = make(map[gwapiv1.ParentReference]*RouteParentContext)
	}
	r.ParentRefs[forParentRef] = ctx
}

func (r *UDPRouteContext) GetParentRefs() map[gwapiv1.ParentReference]*RouteParentContext {
	return r.ParentRefs
}

// TCPRouteContext wraps a TCPRoute and provides helper methods for
// accessing the route's parents.
type TCPRouteContext struct {
	*gwapiv1a2.TCPRoute

	ParentRefs map[gwapiv1.ParentReference]*RouteParentContext
}

func (r *TCPRouteContext) GetRouteType() gwapiv1.Kind {
	return resource.KindTCPRoute
}

func (r *TCPRouteContext) HasRuleNames(sectionName gwapiv1.SectionName) bool {
	rs := r.Spec.Rules
	for _, rule := range rs {
		if rule.Name != nil {
			if *rule.Name == sectionName {
				return true
			}
		}
	}
	return false
}

func (r *TCPRouteContext) GetHostnames() []string {
	// TCPRoute doesn't have hostnames, return empty slice
	return []string{}
}

func (r *TCPRouteContext) GetParentReferences() []gwapiv1.ParentReference {
	return r.Spec.ParentRefs
}

func (r *TCPRouteContext) GetRouteStatus() *gwapiv1.RouteStatus {
	return &r.Status.RouteStatus
}

func (r *TCPRouteContext) GetRouteParentContext(forParentRef gwapiv1.ParentReference) *RouteParentContext {
	if r.ParentRefs == nil {
		r.ParentRefs = make(map[gwapiv1.ParentReference]*RouteParentContext)
	}
	return r.ParentRefs[forParentRef]
}

func (r *TCPRouteContext) SetRouteParentContext(forParentRef gwapiv1.ParentReference, ctx *RouteParentContext) {
	if r.ParentRefs == nil {
		r.ParentRefs = make(map[gwapiv1.ParentReference]*RouteParentContext)
	}
	r.ParentRefs[forParentRef] = ctx
}

// GetHostnames returns the hosts targeted by the Route object.
func GetHostnames(route RouteContext) []string {
	return route.GetHostnames()
}

// GetParentReferences returns the ParentReference of the Route object.
func GetParentReferences(route RouteContext) []gwapiv1.ParentReference {
	return route.GetParentReferences()
}

// GetRouteStatus returns the RouteStatus object associated with the Route.
func GetRouteStatus(route RouteContext) *gwapiv1.RouteStatus {
	return route.GetRouteStatus()
}

// GetRouteParentContext returns RouteParentContext by using the Route objects' ParentReference.
// It creates a new RouteParentContext and add a new RouteParentStatus to the Route's Status if the ParentReference is not found.
func GetRouteParentContext(route RouteContext, forParentRef gwapiv1.ParentReference, controllerName string) *RouteParentContext {
	// If the RouteParentContext is already in the RouteContext, return it.
	if existingCtx := route.GetRouteParentContext(forParentRef); existingCtx != nil {
		return existingCtx
	}

	// Verify that the ParentReference is present in the Route.Spec.ParentRefs.
	// This is just a sanity check, the parentRef should always be present, otherwise it's a programming error.
	var parentRef *gwapiv1.ParentReference
	specParentRefs := route.GetParentReferences()
	for _, p := range specParentRefs {
		if IsParentRefEqual(p, forParentRef, route.GetNamespace()) {
			parentRef = &p
			break
		}
	}
	if parentRef == nil {
		panic("parentRef not found")
	}

	// Find the parent in the Route's Status.
	routeParentStatusIdx := -1
	routeStatus := route.GetRouteStatus()

	for i, parent := range routeStatus.Parents {
		if IsParentRefEqual(parent.ParentRef, *parentRef, route.GetNamespace()) {
			routeParentStatusIdx = i
			break
		}
	}

	// If the parent is not found in the Route's Status, create a new RouteParentStatus and add it to the Route's Status.
	if routeParentStatusIdx == -1 {
		rParentStatus := gwapiv1.RouteParentStatus{
			ControllerName: gwapiv1.GatewayController(controllerName),
			ParentRef:      forParentRef,
		}
		routeStatus.Parents = append(routeStatus.Parents, rParentStatus)
		routeParentStatusIdx = len(routeStatus.Parents) - 1
	}

	// Also add the RouteParentContext to the RouteContext.
	ctx := &RouteParentContext{
		ParentReference:      parentRef,
		routeParentStatusIdx: routeParentStatusIdx,
	}

	// Set the appropriate route field based on the route type
	switch route.GetRouteType() {
	case resource.KindHTTPRoute:
		ctx.HTTPRoute = route.(*HTTPRouteContext).HTTPRoute
	case resource.KindGRPCRoute:
		ctx.GRPCRoute = route.(*GRPCRouteContext).GRPCRoute
	case resource.KindTLSRoute:
		ctx.TLSRoute = route.(*TLSRouteContext).TLSRoute
	case resource.KindTCPRoute:
		ctx.TCPRoute = route.(*TCPRouteContext).TCPRoute
	case resource.KindUDPRoute:
		ctx.UDPRoute = route.(*UDPRouteContext).UDPRoute
	}

	route.SetRouteParentContext(forParentRef, ctx)
	return ctx
}

// IsParentRefEqual compares two ParentReference objects for equality without using reflection.
func IsParentRefEqual(ref1, ref2 gwapiv1.ParentReference, routeNS string) bool {
	// Compare Group with default handling
	defaultGroup := (*gwapiv1.Group)(&gwapiv1.GroupVersion.Group)
	group1 := ref1.Group
	if group1 == nil {
		group1 = defaultGroup
	}
	group2 := ref2.Group
	if group2 == nil {
		group2 = defaultGroup
	}
	if *group1 != *group2 {
		return false
	}

	// Compare Kind with default handling
	defaultKind := gwapiv1.Kind(resource.KindGateway)
	kind1 := ref1.Kind
	if kind1 == nil {
		kind1 = &defaultKind
	}
	kind2 := ref2.Kind
	if kind2 == nil {
		kind2 = &defaultKind
	}
	if *kind1 != *kind2 {
		return false
	}

	// Compare Name (required field)
	if ref1.Name != ref2.Name {
		return false
	}

	// Compare Namespace with default handling
	defaultNS := gwapiv1.Namespace(routeNS)
	namespace1 := ref1.Namespace
	if namespace1 == nil {
		namespace1 = &defaultNS
	}
	namespace2 := ref2.Namespace
	if namespace2 == nil {
		namespace2 = &defaultNS
	}
	if *namespace1 != *namespace2 {
		return false
	}

	// Compare SectionName (optional field)
	if ref1.SectionName == nil && ref2.SectionName == nil {
		return true
	}
	if ref1.SectionName == nil || ref2.SectionName == nil {
		return false
	}
	if *ref1.SectionName != *ref2.SectionName {
		return false
	}

	// Compare Port (optional field)
	if ref1.Port == nil && ref2.Port == nil {
		return true
	}
	if ref1.Port == nil || ref2.Port == nil {
		return false
	}
	return *ref1.Port == *ref2.Port
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

// BackendRefContext represents a generic BackendRef object
type BackendRefContext interface {
	GetBackendRef() *gwapiv1.BackendRef
	GetFilters() any
}

// BackendRefWithFilters wraps backend refs that have filters (HTTPBackendRef, GRPCBackendRef)
type BackendRefWithFilters struct {
	BackendRef *gwapiv1.BackendRef
	Filters    any // []gwapiv1.HTTPRouteFilter or []gwapiv1.GRPCRouteFilter
}

func (b BackendRefWithFilters) GetBackendRef() *gwapiv1.BackendRef {
	return b.BackendRef
}

func (b BackendRefWithFilters) GetFilters() any {
	return b.Filters
}

// DirectBackendRef wraps a BackendRef directly (used by TLS/TCP/UDP routes)
type DirectBackendRef struct {
	BackendRef *gwapiv1.BackendRef
}

func (d DirectBackendRef) GetBackendRef() *gwapiv1.BackendRef {
	return d.BackendRef
}

func (d DirectBackendRef) GetFilters() any {
	return nil
}
