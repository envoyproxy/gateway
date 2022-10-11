package message

import (
	"github.com/telepresenceio/watchable"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/ir"
	xdstypes "github.com/envoyproxy/gateway/internal/xds/types"
)

// ProviderResources message
type ProviderResources struct {
	GatewayClasses watchable.Map[string, *gwapiv1b1.GatewayClass]
	Gateways       watchable.Map[types.NamespacedName, *gwapiv1b1.Gateway]
	HTTPRoutes     watchable.Map[types.NamespacedName, *gwapiv1b1.HTTPRoute]
	TLSRoutes      watchable.Map[types.NamespacedName, *gwapiv1a2.TLSRoute]
	Namespaces     watchable.Map[string, *corev1.Namespace]
	Services       watchable.Map[types.NamespacedName, *corev1.Service]
	Secrets        watchable.Map[types.NamespacedName, *corev1.Secret]

	ReferenceGrants watchable.Map[types.NamespacedName, *gwapiv1a2.ReferenceGrant]

	GatewayStatuses   watchable.Map[types.NamespacedName, *gwapiv1b1.Gateway]
	HTTPRouteStatuses watchable.Map[types.NamespacedName, *gwapiv1b1.HTTPRoute]
	TLSRouteStatuses  watchable.Map[types.NamespacedName, *gwapiv1a2.TLSRoute]
}

func (p *ProviderResources) GetGatewayClasses() []*gwapiv1b1.GatewayClass {
	if p.GatewayClasses.Len() == 0 {
		return nil
	}

	res := make([]*gwapiv1b1.GatewayClass, 0, p.GatewayClasses.Len())
	for _, v := range p.GatewayClasses.LoadAll() {
		res = append(res, v)
	}
	return res
}

func (p *ProviderResources) GetGateways() []*gwapiv1b1.Gateway {
	if p.Gateways.Len() == 0 {
		return nil
	}
	res := make([]*gwapiv1b1.Gateway, 0, p.Gateways.Len())
	for _, v := range p.Gateways.LoadAll() {
		res = append(res, v)
	}
	return res
}

func (p *ProviderResources) GetHTTPRoutes() []*gwapiv1b1.HTTPRoute {
	if p.HTTPRoutes.Len() == 0 {
		return nil
	}
	res := make([]*gwapiv1b1.HTTPRoute, 0, p.HTTPRoutes.Len())
	for _, v := range p.HTTPRoutes.LoadAll() {
		res = append(res, v)
	}
	return res
}

func (p *ProviderResources) GetTLSRoutes() []*gwapiv1a2.TLSRoute {
	if p.TLSRoutes.Len() == 0 {
		return nil
	}
	res := make([]*gwapiv1a2.TLSRoute, 0, p.TLSRoutes.Len())
	for _, v := range p.TLSRoutes.LoadAll() {
		res = append(res, v)
	}
	return res
}

func (p *ProviderResources) GetNamespaces() []*corev1.Namespace {
	if p.Namespaces.Len() == 0 {
		return nil
	}

	res := make([]*corev1.Namespace, 0, p.Namespaces.Len())
	for _, v := range p.Namespaces.LoadAll() {
		res = append(res, v)
	}
	return res
}

func (p *ProviderResources) GetServices() []*corev1.Service {
	if p.Services.Len() == 0 {
		return nil
	}
	res := make([]*corev1.Service, 0, p.Services.Len())
	for _, v := range p.Services.LoadAll() {
		res = append(res, v)
	}
	return res
}

func (p *ProviderResources) GetSecrets() []*corev1.Secret {
	if p.Secrets.Len() == 0 {
		return nil
	}
	res := make([]*corev1.Secret, 0, p.Secrets.Len())
	for _, v := range p.Secrets.LoadAll() {
		res = append(res, v)
	}
	return res
}

func (p *ProviderResources) GetReferenceGrants() []*gwapiv1a2.ReferenceGrant {
	if p.ReferenceGrants.Len() == 0 {
		return nil
	}
	res := make([]*gwapiv1a2.ReferenceGrant, 0, p.ReferenceGrants.Len())
	for _, v := range p.ReferenceGrants.LoadAll() {
		res = append(res, v)
	}
	return res
}

// XdsIR message
type XdsIR struct {
	watchable.Map[string, *ir.Xds]
}

// InfraIR message
type InfraIR struct {
	watchable.Map[string, *ir.Infra]
}

// Xds message
type Xds struct {
	watchable.Map[string, *xdstypes.ResourceVersionTable]
}
