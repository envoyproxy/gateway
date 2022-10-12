package kubernetes

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

// providerCache maintains additional mappings related to Kubernetes provider
// resources. The mappings are regularly updated from the reconcilers based
// on the existence of the object in the Kubernetes datastore.
type providerCache struct {
	mu sync.Mutex

	// routeToServicesMappings maintains a mapping of a Route object,
	// and the Services it references. For instance
	// HTTPRoute/ns1/route1	-> { ns1/svc1, ns1/svc2, ns2/svc1 }
	// TLSRoute/ns1/route1	-> { ns1/svc1, ns2/svc2 }
	routeToServicesMappings map[string]sets.String
}

func newProviderCache() *providerCache {
	return &providerCache{
		routeToServicesMappings: make(map[string]sets.String),
	}
}

func (p *providerCache) getRouteToServicesMapping(route string) []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.routeToServicesMappings[route].List()
}

func (p *providerCache) updateRouteToServicesMapping(route, service string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.routeToServicesMappings[route].Len() == 0 {
		p.routeToServicesMappings[route] = sets.NewString(service)
	} else {
		p.routeToServicesMappings[route].Insert(service)
	}
}

func (p *providerCache) removeRouteToServicesMapping(route, service string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.routeToServicesMappings[route].Delete(service)
	if p.routeToServicesMappings[route].Len() == 0 {
		delete(p.routeToServicesMappings, route)
	}
}

func (p *providerCache) isServiceReferredByRoutes(service string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, svcs := range p.routeToServicesMappings {
		if svcs.Has(service) {
			return true
		}
	}
	return false
}
