// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	certificatesv1b1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
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
func (g *GatewayContext) ResetListeners() {
	numListeners := len(g.Spec.Listeners)
	g.Status.AttachedListenerSets = nil
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
}

func (g *GatewayContext) attachEnvoyProxy(resources *resource.Resources, epMap map[types.NamespacedName]*egv1a1.EnvoyProxy) {
	// Priority order (highest to lowest):
	// 1. Gateway-level EnvoyProxy (via parametersRef)
	// 2. GatewayClass-level EnvoyProxy
	// 3. Default EnvoyProxySpec from EnvoyGateway configuration

	if g.Spec.Infrastructure != nil && g.Spec.Infrastructure.ParametersRef != nil && !IsMergeGatewaysEnabled(resources) {
		ref := g.Spec.Infrastructure.ParametersRef
		if string(ref.Group) == egv1a1.GroupVersion.Group && ref.Kind == egv1a1.KindEnvoyProxy {
			ep, exists := epMap[types.NamespacedName{Namespace: g.Namespace, Name: ref.Name}]
			if exists {
				g.envoyProxy = ep
				return
			}
		}
		// not found, fallthrough to use envoyProxy attached to gatewayclass
	}

	// Use GatewayClass-level EnvoyProxy if available
	if resources.EnvoyProxyForGatewayClass != nil {
		g.envoyProxy = resources.EnvoyProxyForGatewayClass
		return
	}

	// Fall back to default EnvoyProxySpec from EnvoyGateway configuration
	if resources.EnvoyProxyDefaultSpec != nil {
		// Create a synthetic EnvoyProxy object from the default spec
		g.envoyProxy = &egv1a1.EnvoyProxy{
			Spec: *resources.EnvoyProxyDefaultSpec,
		}
	}
}

func (g *GatewayContext) IncreaseAttachedListenerSets() {
	if g.Status.AttachedListenerSets == nil {
		g.Status.AttachedListenerSets = ptr.To[int32](1)
	} else {
		*g.Status.AttachedListenerSets++
	}
}

// ListenerContext wraps a Listener and provides helper methods for
// setting conditions and other status information on the associated
// Gateway, etc.
type ListenerContext struct {
	*gwapiv1.Listener

	gateway           *GatewayContext
	listenerStatusIdx int

	// listenerSet is the ListenerSet that this listener belongs to.
	// If nil, this listener belongs to the Gateway.
	listenerSet          *gwapiv1.ListenerSet
	listenerSetStatusIdx int

	namespaceSelector labels.Selector
	tlsSecrets        []*corev1.Secret
	certDNSNames      []string

	httpIR *ir.HTTPListener
}

// isFromListenerSet returns true if the listener belongs to a ListenerSet instead of a Gateway.
func (l *ListenerContext) isFromListenerSet() bool {
	return l.listenerSet != nil
}

func (l *ListenerContext) SetSupportedKinds(kinds ...gwapiv1.RouteGroupKind) {
	if l.isFromListenerSet() {
		if len(kinds) > 0 {
			l.listenerSet.Status.Listeners[l.listenerSetStatusIdx].SupportedKinds = make([]gwapiv1.RouteGroupKind, len(kinds))
			copy(l.listenerSet.Status.Listeners[l.listenerSetStatusIdx].SupportedKinds, kinds)
		} else {
			l.listenerSet.Status.Listeners[l.listenerSetStatusIdx].SupportedKinds = []gwapiv1.RouteGroupKind{}
		}
	} else {
		if len(kinds) > 0 {
			l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds = make([]gwapiv1.RouteGroupKind, len(kinds))
			copy(l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds, kinds)
		} else {
			l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds = []gwapiv1.RouteGroupKind{}
		}
	}
}

// IncrementAttachedRoutes increments the number of attached routes for the listener in the status.
//
// xref: https://github.com/kubernetes-sigs/gateway-api/issues/2402
// Namely:
// - AttachedRoutes should be set on Listeners that are valid or invalid
// - The count of AttachedRoutes should include Routes that are valid or invalid
func (l *ListenerContext) IncrementAttachedRoutes() {
	if l.isFromListenerSet() {
		l.listenerSet.Status.Listeners[l.listenerSetStatusIdx].AttachedRoutes++
	} else {
		l.gateway.Status.Listeners[l.listenerStatusIdx].AttachedRoutes++
	}
}

func (l *ListenerContext) AttachedRoutes() int32 {
	if l.isFromListenerSet() {
		return l.listenerSet.Status.Listeners[l.listenerSetStatusIdx].AttachedRoutes
	}
	return l.gateway.Status.Listeners[l.listenerStatusIdx].AttachedRoutes
}

func (l *ListenerContext) AllowsKind(kind gwapiv1.RouteGroupKind) bool {
	var supportedKinds []gwapiv1.RouteGroupKind
	if l.listenerSet != nil {
		// Convert ListenerSet supported kinds to Gateway API kinds
		// Since they are alias types, we can cast them
		supportedKinds = append(supportedKinds, l.listenerSet.Status.Listeners[l.listenerSetStatusIdx].SupportedKinds...)
	} else {
		supportedKinds = l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds
	}

	for _, allowed := range supportedKinds {
		// The default group is "gateway.networking.k8s.io"
		if GroupDerefOr(allowed.Group, "gateway.networking.k8s.io") == GroupDerefOr(kind.Group, "gateway.networking.k8s.io") &&
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
		return l.GetNamespace() == namespace.Name
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
		return l.GetNamespace() == namespace.Name
	}
}

func (l *ListenerContext) IsReady() bool {
	var conditions []metav1.Condition
	if l.isFromListenerSet() {
		conditions = l.listenerSet.Status.Listeners[l.listenerSetStatusIdx].Conditions
	} else {
		conditions = l.gateway.Status.Listeners[l.listenerStatusIdx].Conditions
	}

	for _, cond := range conditions {
		if cond.Type == string(gwapiv1.ListenerConditionProgrammed) && cond.Status == metav1.ConditionTrue {
			return true
		}
	}

	return false
}

func (l *ListenerContext) GetNamespace() string {
	if l.isFromListenerSet() {
		return l.listenerSet.Namespace
	}
	return l.gateway.Namespace
}

func (l *ListenerContext) GetConditions() []metav1.Condition {
	if l.isFromListenerSet() {
		return l.listenerSet.Status.Listeners[l.listenerSetStatusIdx].Conditions
	}
	return l.gateway.Status.Listeners[l.listenerStatusIdx].Conditions
}

func (l *ListenerContext) SetCondition(conditionType gwapiv1.ListenerConditionType, conditionStatus metav1.ConditionStatus, reason gwapiv1.ListenerConditionReason, message string) {
	if l.isFromListenerSet() {
		r := string(reason)
		if reason == gwapiv1.ListenerReasonInvalid {
			r = string(gwapiv1.ListenerSetReasonListenersNotValid)
		}
		// Convert Gateway API types to ListenerSet types
		// Note: The string values are expected to match between the APIs
		status.SetListenerSetListenerStatusCondition(l.listenerSet, l.listenerSetStatusIdx,
			gwapiv1.ListenerEntryConditionType(conditionType),
			conditionStatus, r,
			message)
	} else {
		status.SetGatewayListenerStatusCondition(l.gateway.Gateway, l.listenerStatusIdx,
			conditionType,
			conditionStatus,
			reason,
			message)
	}
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
	*gwapiv1.TLSRoute

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

// GetManagedParentReferences returns route parentRefs that are managed by this controller.
func GetManagedParentReferences(route RouteContext) []gwapiv1.ParentReference {
	parentRefs := GetParentReferences(route)
	managed := make([]gwapiv1.ParentReference, 0, len(parentRefs))
	for _, parentRef := range parentRefs {
		// RouteParentContext is only created for parentRefs handled by this
		// translator run. If absent, the parentRef points to a Gateway that is
		// not managed by this controller.
		if route.GetRouteParentContext(parentRef) == nil {
			continue
		}
		managed = append(managed, parentRef)
	}
	return managed
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
	TLSRoute  *gwapiv1.TLSRoute
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

type backendServiceKey struct {
	kind      string
	namespace string
	name      string
}

type TranslatorContext struct {
	NamespaceMap          map[types.NamespacedName]*corev1.Namespace
	ServiceMap            map[types.NamespacedName]*corev1.Service
	ServiceImportMap      map[types.NamespacedName]*mcsapiv1a1.ServiceImport
	BackendMap            map[types.NamespacedName]*egv1a1.Backend
	SecretMap             map[types.NamespacedName]*corev1.Secret
	ConfigMapMap          map[types.NamespacedName]*corev1.ConfigMap
	ClusterTrustBundleMap map[types.NamespacedName]*certificatesv1b1.ClusterTrustBundle
	EndpointSliceMap      map[backendServiceKey][]*discoveryv1.EndpointSlice
	BTPRoutingTypeIndex   *BTPRoutingTypeIndex
}

func (t *TranslatorContext) GetNamespace(name string) *corev1.Namespace {
	if ns, ok := t.NamespaceMap[types.NamespacedName{Name: name}]; ok {
		return ns
	}
	return nil
}

func (t *TranslatorContext) SetNamespaces(namespaces []*corev1.Namespace) {
	namespaceMap := make(map[types.NamespacedName]*corev1.Namespace, len(namespaces))
	for _, ns := range namespaces {
		namespaceMap[types.NamespacedName{Name: ns.Name}] = ns
	}
	t.NamespaceMap = namespaceMap
}

func (t *TranslatorContext) GetService(namespace, name string) *corev1.Service {
	if svc, ok := t.ServiceMap[types.NamespacedName{Namespace: namespace, Name: name}]; ok {
		return svc
	}
	return nil
}

func (t *TranslatorContext) SetServices(svcs []*corev1.Service) {
	serviceMap := make(map[types.NamespacedName]*corev1.Service, len(svcs))
	for _, svc := range svcs {
		serviceMap[utils.NamespacedName(svc)] = svc
	}
	t.ServiceMap = serviceMap
}

func (t *TranslatorContext) GetServiceImport(namespace, name string) *mcsapiv1a1.ServiceImport {
	if svcImp, ok := t.ServiceImportMap[types.NamespacedName{Namespace: namespace, Name: name}]; ok {
		return svcImp
	}
	return nil
}

func (t *TranslatorContext) SetServiceImports(svcImps []*mcsapiv1a1.ServiceImport) {
	serviceImportMap := make(map[types.NamespacedName]*mcsapiv1a1.ServiceImport, len(svcImps))
	for _, svcImp := range svcImps {
		serviceImportMap[utils.NamespacedName(svcImp)] = svcImp
	}
	t.ServiceImportMap = serviceImportMap
}

func (t *TranslatorContext) GetBackend(namespace, name string) *egv1a1.Backend {
	if backend, ok := t.BackendMap[types.NamespacedName{Namespace: namespace, Name: name}]; ok {
		return backend
	}
	return nil
}

func (t *TranslatorContext) SetBackends(backends []*egv1a1.Backend) {
	backendMap := make(map[types.NamespacedName]*egv1a1.Backend, len(backends))
	for _, backend := range backends {
		backendMap[utils.NamespacedName(backend)] = backend
	}
	t.BackendMap = backendMap
}

func (t *TranslatorContext) GetSecret(namespace, name string) *corev1.Secret {
	if secret, ok := t.SecretMap[types.NamespacedName{Namespace: namespace, Name: name}]; ok {
		return secret
	}
	return nil
}

func (t *TranslatorContext) SetSecrets(secrets []*corev1.Secret) {
	secretMap := make(map[types.NamespacedName]*corev1.Secret, len(secrets))
	for _, secret := range secrets {
		secretMap[utils.NamespacedName(secret)] = secret
	}
	t.SecretMap = secretMap
}

func (t *TranslatorContext) GetConfigMap(namespace, name string) *corev1.ConfigMap {
	if configMap, ok := t.ConfigMapMap[types.NamespacedName{Namespace: namespace, Name: name}]; ok {
		return configMap
	}
	return nil
}

func (t *TranslatorContext) SetConfigMaps(configMaps []*corev1.ConfigMap) {
	configMapMap := make(map[types.NamespacedName]*corev1.ConfigMap, len(configMaps))
	for _, configMap := range configMaps {
		configMapMap[utils.NamespacedName(configMap)] = configMap
	}
	t.ConfigMapMap = configMapMap
}

func (t *TranslatorContext) GetClusterTrustBundle(name string) *certificatesv1b1.ClusterTrustBundle {
	if ctb, ok := t.ClusterTrustBundleMap[types.NamespacedName{Name: name}]; ok {
		return ctb
	}
	return nil
}

func (t *TranslatorContext) SetClusterTrustBundles(ctbs []*certificatesv1b1.ClusterTrustBundle) {
	ctbMap := make(map[types.NamespacedName]*certificatesv1b1.ClusterTrustBundle, len(ctbs))
	for _, ctb := range ctbs {
		ctbMap[types.NamespacedName{Name: ctb.Name}] = ctb
	}
	t.ClusterTrustBundleMap = ctbMap
}

func (t *TranslatorContext) GetEndpointSlicesForBackend(svcNamespace, svcName, backendKind string) []*discoveryv1.EndpointSlice {
	key := backendServiceKey{
		kind:      backendKind,
		namespace: svcNamespace,
		name:      svcName,
	}
	if slices, ok := t.EndpointSliceMap[key]; ok {
		return slices
	}
	return nil
}

func (t *TranslatorContext) SetEndpointSlicesForBackend(slices []*discoveryv1.EndpointSlice) {
	t.EndpointSliceMap = make(map[backendServiceKey][]*discoveryv1.EndpointSlice)

	var kind, svcName string
	for _, slice := range slices {
		if name, ok := slice.Labels[discoveryv1.LabelServiceName]; ok {
			kind = resource.KindService
			svcName = name
		} else if name, ok := slice.Labels[mcsapiv1a1.LabelServiceName]; ok {
			kind = resource.KindServiceImport
			svcName = name
		} else {
			continue
		}

		key := backendServiceKey{
			kind:      kind,
			namespace: slice.Namespace,
			name:      svcName,
		}
		t.EndpointSliceMap[key] = append(t.EndpointSliceMap[key], slice)
	}
}
