package gatewayapi

import (
	"reflect"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1alpha1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

// GatewayContext wraps a Gateway and provides helper methods for
// setting conditions, accessing Listeners, etc.
type GatewayContext struct {
	*v1beta1.Gateway

	listeners map[v1beta1.SectionName]*ListenerContext
}

func (g *GatewayContext) GetListenerContext(listenerName v1beta1.SectionName) *ListenerContext {
	if g.listeners == nil {
		g.listeners = make(map[v1beta1.SectionName]*ListenerContext)
	}

	if ctx := g.listeners[listenerName]; ctx != nil {
		return ctx
	}

	var listener *v1beta1.Listener
	for i, l := range g.Spec.Listeners {
		if l.Name == listenerName {
			listener = &g.Spec.Listeners[i]
			break
		}
	}
	if listener == nil {
		panic("listener not found")
	}

	listenerStatusIdx := -1
	for i := range g.Status.Listeners {
		if g.Status.Listeners[i].Name == listenerName {
			listenerStatusIdx = i
			break
		}
	}
	if listenerStatusIdx == -1 {
		g.Status.Listeners = append(g.Status.Listeners, v1beta1.ListenerStatus{Name: listenerName})
		listenerStatusIdx = len(g.Status.Listeners) - 1
	}

	ctx := &ListenerContext{
		Listener:          listener,
		gateway:           g.Gateway,
		listenerStatusIdx: listenerStatusIdx,
	}
	g.listeners[listenerName] = ctx
	return ctx
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
				existing.Message == cond.Message {
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

func (l *ListenerContext) ResetConditions() {
	l.gateway.Status.Listeners[l.listenerStatusIdx].Conditions = make([]metav1.Condition, 0)
}

func (l *ListenerContext) SetSupportedKinds(kinds ...v1beta1.RouteGroupKind) {
	l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds = kinds
}

func (l *ListenerContext) ResetAttachedRoutes() {
	// Reset attached route count since it will be recomputed during translation.
	l.gateway.Status.Listeners[l.listenerStatusIdx].AttachedRoutes = 0
}

func (l *ListenerContext) IncrementAttachedRoutes() {
	l.gateway.Status.Listeners[l.listenerStatusIdx].AttachedRoutes++
}

func (l *ListenerContext) AllowsKind(kind v1beta1.RouteGroupKind) bool {
	for _, allowed := range l.gateway.Status.Listeners[l.listenerStatusIdx].SupportedKinds {
		if GroupDerefOr(allowed.Group, "") == GroupDerefOr(kind.Group, "") && allowed.Kind == kind.Kind {
			return true
		}
	}

	return false
}

func (l *ListenerContext) AllowsNamespace(namespace *v1.Namespace) bool {
	if namespace == nil {
		return false
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
		if cond.Type == string(v1beta1.ListenerConditionReady) && cond.Status == metav1.ConditionTrue {
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

// HTTPRouteContext wraps an HTTPRoute and provides helper methods for
// accessing the route's parents.
type HTTPRouteContext struct {
	*v1beta1.HTTPRoute

	parentRefs map[v1beta1.ParentReference]*RouteParentContext
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
		if p == forParentRef {
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

		route:                h.HTTPRoute,
		routeParentStatusIdx: routeParentStatusIdx,
	}
	h.parentRefs[forParentRef] = ctx
	return ctx
}

// RouteParentContext wraps a ParentReference and provides helper methods for
// setting conditions and other status information on the associated
// HTTPRoute, etc.
type RouteParentContext struct {
	*v1beta1.ParentReference

	route                *v1beta1.HTTPRoute
	routeParentStatusIdx int
	listeners            []*ListenerContext
}

func (r *RouteParentContext) SetListeners(listeners ...*ListenerContext) {
	r.listeners = append(r.listeners, listeners...)
}

func (r *RouteParentContext) SetCondition(conditionType v1beta1.RouteConditionType, status metav1.ConditionStatus, reason v1beta1.RouteConditionReason, message string) {
	cond := metav1.Condition{
		Type:               string(conditionType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		ObservedGeneration: r.route.Generation,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	idx := -1
	for i, existing := range r.route.Status.Parents[r.routeParentStatusIdx].Conditions {
		if existing.Type == cond.Type {
			// return early if the condition is unchanged
			if existing.Status == cond.Status &&
				existing.Reason == cond.Reason &&
				existing.Message == cond.Message {
				return
			}
			idx = i
			break
		}
	}

	if idx > -1 {
		r.route.Status.Parents[r.routeParentStatusIdx].Conditions[idx] = cond
	} else {
		r.route.Status.Parents[r.routeParentStatusIdx].Conditions = append(r.route.Status.Parents[r.routeParentStatusIdx].Conditions, cond)
	}
}

func (r *RouteParentContext) ResetConditions() {
	r.route.Status.Parents[r.routeParentStatusIdx].Conditions = make([]metav1.Condition, 0)
}

func (r *RouteParentContext) IsAccepted() bool {
	for _, cond := range r.route.Status.Parents[r.routeParentStatusIdx].Conditions {
		if cond.Type == string(v1beta1.RouteConditionAccepted) && cond.Status == metav1.ConditionTrue {
			return true
		}
	}

	return false
}
