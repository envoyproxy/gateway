// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"reflect"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1alpha1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

// GatewayContext wraps a Gateway and provides helper methods for
// setting conditions, accessing Listeners, etc.
type GatewayContext struct {
	*v1beta1.Gateway

	listeners []*ListenerContext
}

// ResetListeners resets the listener statuses and re-generates the GatewayContext
// ListenerContexts from the Gateway spec.
func (g *GatewayContext) ResetListeners() {
	numListeners := len(g.Spec.Listeners)
	g.Status.Listeners = make([]v1beta1.ListenerStatus, numListeners)
	g.listeners = make([]*ListenerContext, numListeners)
	for i := range g.Spec.Listeners {
		listener := &g.Spec.Listeners[i]
		g.Status.Listeners[i] = v1beta1.ListenerStatus{Name: listener.Name}
		g.listeners[i] = &ListenerContext{
			Listener:          listener,
			gateway:           g.Gateway,
			listenerStatusIdx: i,
		}
	}
}

// ListenerContext wraps a Listener and provides helper methods for
// setting conditions and other status information on the associated
// Gateway, etc.
type ListenerContext struct {
	*v1beta1.Listener

	gateway           *v1beta1.Gateway
	listenerStatusIdx int
	namespaceSelector labels.Selector
	tlsSecret         *v1.Secret
}

func (l *ListenerContext) SetCondition(conditionType v1beta1.ListenerConditionType, status metav1.ConditionStatus, reason v1beta1.ListenerConditionReason, message string) {
	cond := metav1.Condition{
		Type:               string(conditionType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		ObservedGeneration: l.gateway.Generation,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	idx := -1
	for i, existing := range l.gateway.Status.Listeners[l.listenerStatusIdx].Conditions {
		if existing.Type == cond.Type {
			// return early if the condition is unchanged
			if existing.Status == cond.Status &&
				existing.Reason == cond.Reason &&
				existing.Message == cond.Message &&
				existing.ObservedGeneration == cond.ObservedGeneration {
				return
			}
			idx = i
			break
		}
	}

	if idx > -1 {
		l.gateway.Status.Listeners[l.listenerStatusIdx].Conditions[idx] = cond
	} else {
		l.gateway.Status.Listeners[l.listenerStatusIdx].Conditions = append(l.gateway.Status.Listeners[l.listenerStatusIdx].Conditions, cond)
	}
}

func (l *ListenerContext) SetSupportedKinds(kinds ...v1beta1.RouteGroupKind) {
	l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds = kinds
}

func (l *ListenerContext) IncrementAttachedRoutes() {
	l.gateway.Status.Listeners[l.listenerStatusIdx].AttachedRoutes++
}

func (l *ListenerContext) AttachedRoutes() int32 {
	return l.gateway.Status.Listeners[l.listenerStatusIdx].AttachedRoutes
}

func (l *ListenerContext) AllowsKind(kind v1beta1.RouteGroupKind) bool {
	for _, allowed := range l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds {
		if GroupDerefOr(allowed.Group, "") == GroupDerefOr(kind.Group, "") &&
			allowed.Kind == kind.Kind {
			return true
		}
	}

	return false
}

func (l *ListenerContext) AllowsNamespace(namespace *v1.Namespace) bool {
	if namespace == nil {
		return false
	}

	if l.AllowedRoutes == nil || l.AllowedRoutes.Namespaces == nil || l.AllowedRoutes.Namespaces.From == nil {
		return l.gateway.Namespace == namespace.Name
	}

	switch *l.AllowedRoutes.Namespaces.From {
	case v1beta1.NamespacesFromAll:
		return true
	case v1beta1.NamespacesFromSelector:
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
		if cond.Type == string(v1beta1.ListenerConditionProgrammed) && cond.Status == metav1.ConditionTrue {
			return true
		}
	}

	return false
}

func (l *ListenerContext) GetConditions() []metav1.Condition {
	return l.gateway.Status.Listeners[l.listenerStatusIdx].Conditions
}

func (l *ListenerContext) SetTLSSecret(tlsSecret *v1.Secret) {
	l.tlsSecret = tlsSecret
}

// RouteContext represents a generic Route object (HTTPRoute, TLSRoute, etc.)
// that can reference Gateway objects.
type RouteContext interface {
	client.Object

	// GetRouteType returns the Kind of the Route object, HTTPRoute,
	// TLSRoute, TCPRoute, UDPRoute etc.
	GetRouteType() string

	// GetRouteStatus returns the RouteStatus object associated with the Route.
	GetRouteStatus() *v1beta1.RouteStatus

	// TODO: [v1alpha2-v1beta1] This should not be required once all Route
	// objects being implemented are of type v1beta1.
	// GetParentReferences returns the ParentReference of the Route object.
	GetParentReferences() []v1beta1.ParentReference

	// GetRouteParentContext returns RouteParentContext by using the Route
	// objects' ParentReference.
	GetRouteParentContext(forParentRef v1beta1.ParentReference) *RouteParentContext

	// TODO: [v1alpha2-v1beta1] This should not be required once all Route
	// objects being implemented are of type v1beta1.
	// GetHostnames returns the hosts targeted by the Route object.
	GetHostnames() []string
}

// HTTPRouteContext wraps an HTTPRoute and provides helper methods for
// accessing the route's parents.
type HTTPRouteContext struct {
	*v1beta1.HTTPRoute

	parentRefs map[v1beta1.ParentReference]*RouteParentContext
}

func (h *HTTPRouteContext) GetRouteType() string {
	return KindHTTPRoute
}

func (h *HTTPRouteContext) GetHostnames() []string {
	hostnames := make([]string, len(h.Spec.Hostnames))
	for idx, s := range h.Spec.Hostnames {
		hostnames[idx] = string(s)
	}
	return hostnames
}

func (h *HTTPRouteContext) GetParentReferences() []v1beta1.ParentReference {
	return h.Spec.ParentRefs
}

func (h *HTTPRouteContext) GetRouteStatus() *v1beta1.RouteStatus {
	return &h.Status.RouteStatus
}

func (h *HTTPRouteContext) GetRouteParentContext(forParentRef v1beta1.ParentReference) *RouteParentContext {
	if h.parentRefs == nil {
		h.parentRefs = make(map[v1beta1.ParentReference]*RouteParentContext)
	}

	if ctx := h.parentRefs[forParentRef]; ctx != nil {
		return ctx
	}

	var parentRef *v1beta1.ParentReference
	for i, p := range h.Spec.ParentRefs {
		if reflect.DeepEqual(p, forParentRef) {
			parentRef = &h.Spec.ParentRefs[i]
			break
		}
	}
	if parentRef == nil {
		panic("parentRef not found")
	}

	routeParentStatusIdx := -1
	for i := range h.Status.Parents {
		if reflect.DeepEqual(h.Status.Parents[i].ParentRef, forParentRef) {
			routeParentStatusIdx = i
			break
		}
	}
	if routeParentStatusIdx == -1 {
		rParentStatus := v1beta1.RouteParentStatus{
			// TODO: get this value from the config
			ControllerName: v1beta1.GatewayController(egv1alpha1.GatewayControllerName),
			ParentRef:      forParentRef,
		}
		h.Status.Parents = append(h.Status.Parents, rParentStatus)
		routeParentStatusIdx = len(h.Status.Parents) - 1
	}

	ctx := &RouteParentContext{
		ParentReference: parentRef,

		httpRoute:            h.HTTPRoute,
		routeParentStatusIdx: routeParentStatusIdx,
	}
	h.parentRefs[forParentRef] = ctx
	return ctx
}

// GRPCRouteContext wraps a GRPCRoute and provides helper methods for
// accessing the route's parents.
type GRPCRouteContext struct {
	*v1alpha2.GRPCRoute

	parentRefs map[v1beta1.ParentReference]*RouteParentContext
}

func (g *GRPCRouteContext) GetRouteType() string {
	return KindGRPCRoute
}

func (g *GRPCRouteContext) GetHostnames() []string {
	hostnames := make([]string, len(g.Spec.Hostnames))
	for idx, s := range g.Spec.Hostnames {
		hostnames[idx] = string(s)
	}
	return hostnames
}

func (g *GRPCRouteContext) GetParentReferences() []v1beta1.ParentReference {
	return g.Spec.ParentRefs
}

func (g *GRPCRouteContext) GetRouteStatus() *v1beta1.RouteStatus {
	return &g.Status.RouteStatus
}

func (g *GRPCRouteContext) GetRouteParentContext(forParentRef v1beta1.ParentReference) *RouteParentContext {
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

		grpcRoute:            g.GRPCRoute,
		routeParentStatusIdx: routeParentStatusIdx,
	}
	g.parentRefs[forParentRef] = ctx
	return ctx
}

// TLSRouteContext wraps a TLSRoute and provides helper methods for
// accessing the route's parents.
type TLSRouteContext struct {
	*v1alpha2.TLSRoute

	parentRefs map[v1beta1.ParentReference]*RouteParentContext
}

func (t *TLSRouteContext) GetRouteType() string {
	return KindTLSRoute
}

func (t *TLSRouteContext) GetHostnames() []string {
	hostnames := make([]string, len(t.Spec.Hostnames))
	for idx, s := range t.Spec.Hostnames {
		hostnames[idx] = string(s)
	}
	return hostnames
}

func (t *TLSRouteContext) GetRouteStatus() *v1beta1.RouteStatus {
	return &t.Status.RouteStatus
}

func (t *TLSRouteContext) GetParentReferences() []v1beta1.ParentReference {
	parentReferences := make([]v1beta1.ParentReference, len(t.Spec.ParentRefs))
	for idx, p := range t.Spec.ParentRefs {
		parentReferences[idx] = UpgradeParentReference(p)
	}
	return parentReferences
}

func (t *TLSRouteContext) GetRouteParentContext(forParentRef v1beta1.ParentReference) *RouteParentContext {
	if t.parentRefs == nil {
		t.parentRefs = make(map[v1beta1.ParentReference]*RouteParentContext)
	}

	if ctx := t.parentRefs[forParentRef]; ctx != nil {
		return ctx
	}

	var parentRef *v1beta1.ParentReference
	for i, p := range t.Spec.ParentRefs {
		p := UpgradeParentReference(p)
		if reflect.DeepEqual(p, forParentRef) {
			upgraded := UpgradeParentReference(t.Spec.ParentRefs[i])
			parentRef = &upgraded
			break
		}
	}
	if parentRef == nil {
		panic("parentRef not found")
	}

	routeParentStatusIdx := -1
	for i := range t.Status.Parents {
		p := UpgradeParentReference(t.Status.Parents[i].ParentRef)
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
		t.Status.Parents = append(t.Status.Parents, rParentStatus)
		routeParentStatusIdx = len(t.Status.Parents) - 1
	}

	ctx := &RouteParentContext{
		ParentReference: parentRef,

		tlsRoute:             t.TLSRoute,
		routeParentStatusIdx: routeParentStatusIdx,
	}
	t.parentRefs[forParentRef] = ctx
	return ctx
}

// UDPRouteContext wraps a UDPRoute and provides helper methods for
// accessing the route's parents.
type UDPRouteContext struct {
	*v1alpha2.UDPRoute

	parentRefs map[v1beta1.ParentReference]*RouteParentContext
}

func (u *UDPRouteContext) GetRouteType() string {
	return KindUDPRoute
}

func (u *UDPRouteContext) GetParentReferences() []v1beta1.ParentReference {
	parentReferences := make([]v1beta1.ParentReference, len(u.Spec.ParentRefs))
	for idx, p := range u.Spec.ParentRefs {
		parentReferences[idx] = UpgradeParentReference(p)
	}
	return parentReferences
}

func (u *UDPRouteContext) GetRouteStatus() *v1beta1.RouteStatus {
	return &u.Status.RouteStatus
}

func (u *UDPRouteContext) GetRouteParentContext(forParentRef v1beta1.ParentReference) *RouteParentContext {
	if u.parentRefs == nil {
		u.parentRefs = make(map[v1beta1.ParentReference]*RouteParentContext)
	}

	if ctx := u.parentRefs[forParentRef]; ctx != nil {
		return ctx
	}

	var parentRef *v1beta1.ParentReference
	for i, p := range u.Spec.ParentRefs {
		p := UpgradeParentReference(p)
		if reflect.DeepEqual(p, forParentRef) {
			upgraded := UpgradeParentReference(u.Spec.ParentRefs[i])
			parentRef = &upgraded
			break
		}
	}
	if parentRef == nil {
		panic("parentRef not found")
	}

	routeParentStatusIdx := -1
	for i := range u.Status.Parents {
		p := UpgradeParentReference(u.Status.Parents[i].ParentRef)
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
		u.Status.Parents = append(u.Status.Parents, rParentStatus)
		routeParentStatusIdx = len(u.Status.Parents) - 1
	}

	ctx := &RouteParentContext{
		ParentReference: parentRef,

		udpRoute:             u.UDPRoute,
		routeParentStatusIdx: routeParentStatusIdx,
	}
	u.parentRefs[forParentRef] = ctx
	return ctx
}

func (u *UDPRouteContext) GetHostnames() []string {
	return nil
}

// TCPRouteContext wraps a TCPRoute and provides helper methods for
// accessing the route's parents.
type TCPRouteContext struct {
	*v1alpha2.TCPRoute

	parentRefs map[v1beta1.ParentReference]*RouteParentContext
}

func (t *TCPRouteContext) GetRouteType() string {
	return KindTCPRoute
}

func (t *TCPRouteContext) GetParentReferences() []v1beta1.ParentReference {
	parentReferences := make([]v1beta1.ParentReference, len(t.Spec.ParentRefs))
	for idx, p := range t.Spec.ParentRefs {
		parentReferences[idx] = UpgradeParentReference(p)
	}
	return parentReferences
}

func (t *TCPRouteContext) GetRouteStatus() *v1beta1.RouteStatus {
	return &t.Status.RouteStatus
}

func (t *TCPRouteContext) GetRouteParentContext(forParentRef v1beta1.ParentReference) *RouteParentContext {
	if t.parentRefs == nil {
		t.parentRefs = make(map[v1beta1.ParentReference]*RouteParentContext)
	}

	if ctx := t.parentRefs[forParentRef]; ctx != nil {
		return ctx
	}

	var parentRef *v1beta1.ParentReference
	for i, p := range t.Spec.ParentRefs {
		p := UpgradeParentReference(p)
		if reflect.DeepEqual(p, forParentRef) {
			upgraded := UpgradeParentReference(t.Spec.ParentRefs[i])
			parentRef = &upgraded
			break
		}
	}
	if parentRef == nil {
		panic("parentRef not found")
	}

	routeParentStatusIdx := -1
	for i := range t.Status.Parents {
		p := UpgradeParentReference(t.Status.Parents[i].ParentRef)
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
		t.Status.Parents = append(t.Status.Parents, rParentStatus)
		routeParentStatusIdx = len(t.Status.Parents) - 1
	}

	ctx := &RouteParentContext{
		ParentReference: parentRef,

		tcpRoute:             t.TCPRoute,
		routeParentStatusIdx: routeParentStatusIdx,
	}
	t.parentRefs[forParentRef] = ctx
	return ctx
}

func (t *TCPRouteContext) GetHostnames() []string {
	return nil
}

// RouteParentContext wraps a ParentReference and provides helper methods for
// setting conditions and other status information on the associated
// HTTPRoute, TLSRoute etc.
type RouteParentContext struct {
	*v1beta1.ParentReference

	// TODO: [v1alpha2-v1beta1] This can probably be replaced with
	// a single field pointing to *v1beta1.RouteStatus.
	httpRoute *v1beta1.HTTPRoute
	grpcRoute *v1alpha2.GRPCRoute
	tlsRoute  *v1alpha2.TLSRoute
	tcpRoute  *v1alpha2.TCPRoute
	udpRoute  *v1alpha2.UDPRoute

	routeParentStatusIdx int
	listeners            []*ListenerContext
}

func (r *RouteParentContext) SetListeners(listeners ...*ListenerContext) {
	r.listeners = append(r.listeners, listeners...)
}

func (r *RouteParentContext) SetCondition(route RouteContext, conditionType v1beta1.RouteConditionType, status metav1.ConditionStatus, reason v1beta1.RouteConditionReason, message string) {
	cond := metav1.Condition{
		Type:               string(conditionType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		ObservedGeneration: route.GetGeneration(),
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	idx := -1
	routeStatus := route.GetRouteStatus()
	for i, existing := range routeStatus.Parents[r.routeParentStatusIdx].Conditions {
		if existing.Type == cond.Type {
			// return early if the condition is unchanged
			if existing.Status == cond.Status &&
				existing.Reason == cond.Reason &&
				existing.Message == cond.Message &&
				existing.ObservedGeneration == cond.ObservedGeneration {
				return
			}
			idx = i
			break
		}
	}

	if idx > -1 {
		routeStatus.Parents[r.routeParentStatusIdx].Conditions[idx] = cond
	} else {
		routeStatus.Parents[r.routeParentStatusIdx].Conditions = append(routeStatus.Parents[r.routeParentStatusIdx].Conditions, cond)
	}
}

func (r *RouteParentContext) ResetConditions(route RouteContext) {
	routeStatus := route.GetRouteStatus()
	routeStatus.Parents[r.routeParentStatusIdx].Conditions = make([]metav1.Condition, 0)
}

func (r *RouteParentContext) HasCondition(route RouteContext, condType v1beta1.RouteConditionType, status metav1.ConditionStatus) bool {
	var conditions []metav1.Condition
	routeStatus := route.GetRouteStatus()
	conditions = routeStatus.Parents[r.routeParentStatusIdx].Conditions
	for _, c := range conditions {
		if c.Type == string(condType) && c.Status == status {
			return true
		}
	}
	return false
}
