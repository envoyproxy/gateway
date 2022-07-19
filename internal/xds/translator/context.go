package translator

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

// XdsResources represents all the xds resources
type XdsResources map[resource.Type][]types.Resource

// Context holds all the translated xds resources
type Context struct {
	xdsResources XdsResources
}

// GetXdsResources retrieves the translated xds resources saved in the translator context.
func (t *Context) GetXdsResources() XdsResources {
	return t.xdsResources
}

func (t *Context) addXdsResource(rType resource.Type, xdsResource types.Resource) {
	if t.xdsResources == nil {
		t.xdsResources = make(XdsResources)
	}
	if t.xdsResources[rType] == nil {
		t.xdsResources[rType] = make([]types.Resource, 0, 1)
	}

	t.xdsResources[rType] = append(t.xdsResources[rType], xdsResource)
}
