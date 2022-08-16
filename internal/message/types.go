package message

import (
	"sync"

	"github.com/telepresenceio/watchable"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/ir"
	xdstypes "github.com/envoyproxy/gateway/internal/xds/types"
)

// ProviderResources message
type ProviderResources struct {
	GatewayClasses watchable.Map[string, *gwapiv1b1.GatewayClass]
	Gateways       watchable.Map[types.NamespacedName, *gwapiv1b1.Gateway]
	HTTPRoutes     watchable.Map[types.NamespacedName, *gwapiv1b1.HTTPRoute]
	// Initialized.Wait() will return once each of the maps in the
	// structure have been initialized at startup.
	Initialized sync.WaitGroup
}

func (p *ProviderResources) GetGatewayClasses() []*gwapiv1b1.GatewayClass {
	res := make([]*gwapiv1b1.GatewayClass, p.GatewayClasses.Len())
	for _, v := range p.GatewayClasses.LoadAll() {
		res = append(res, v)
	}
	return res
}

func (p *ProviderResources) GetGateways() []*gwapiv1b1.Gateway {
	res := make([]*gwapiv1b1.Gateway, p.Gateways.Len())
	for _, v := range p.Gateways.LoadAll() {
		res = append(res, v)
	}
	return res
}

// XdsIR message
type XdsIR struct {
	watchable.Map[string, *ir.Xds]
}

func (x *XdsIR) Get() *ir.Xds {
	// Iterate and return the first element
	for _, v := range x.LoadAll() {
		return v
	}
	return nil
}

// InfraIR message
type InfraIR struct {
	watchable.Map[string, *ir.Infra]
}

func (i *InfraIR) Get() *ir.Infra {
	// Iterate and return the first element
	for _, v := range i.LoadAll() {
		return v
	}
	return nil
}

// XdsResources message
type XdsResources struct {
	watchable.Map[string, *xdstypes.XdsResources]
}

func (x *XdsResources) Get() *xdstypes.XdsResources {
	// Itreate and return the first element
	for _, v := range x.LoadAll() {
		return v
	}
	return nil
}
