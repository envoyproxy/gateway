package kubernetes

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
)

// providerReferenceStore maintains additional mappings related to Kubernetes provider
// resources. The mappings are regularly updated from the reconcilers based
// on the existence of the object in the Kubernetes datastore.
type providerReferenceStore struct {
	mu sync.Mutex

	// routeToServicesMappings maintains a mapping of a Route object,
	// and the Services it references. For instance
	// HTTPRoute/ns1/route1	-> { ns1/svc1, ns1/svc2, ns2/svc1 }
	// TLSRoute/ns1/route1	-> { ns1/svc1, ns2/svc2 }
	routeToServicesMappings map[ObjectKindNamespacedName]map[types.NamespacedName]struct{}
}

type ObjectKindNamespacedName struct {
	kind      string
	namespace string
	name      string
}

func newProviderReferenceStore() *providerReferenceStore {
	return &providerReferenceStore{
		routeToServicesMappings: map[ObjectKindNamespacedName]map[types.NamespacedName]struct{}{},
	}
}

func (p *providerReferenceStore) getRouteToServicesMapping(route ObjectKindNamespacedName) map[types.NamespacedName]struct{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.routeToServicesMappings[route]
}

func (p *providerReferenceStore) updateRouteToServicesMapping(route ObjectKindNamespacedName, service types.NamespacedName) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.routeToServicesMappings[route]) == 0 {
		p.routeToServicesMappings[route] = map[types.NamespacedName]struct{}{service: {}}
	} else {
		p.routeToServicesMappings[route][service] = struct{}{}
	}
}

func (p *providerReferenceStore) removeRouteToServicesMapping(route ObjectKindNamespacedName, service types.NamespacedName) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.routeToServicesMappings[route], service)
	if len(p.routeToServicesMappings[route]) == 0 {
		delete(p.routeToServicesMappings, route)
	}
}

func (p *providerReferenceStore) isServiceReferredByRoutes(service types.NamespacedName) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, svcs := range p.routeToServicesMappings {
		if _, ok := svcs[service]; ok {
			return true
		}
	}
	return false
}
