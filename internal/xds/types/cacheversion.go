package types

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

// XdsResources represents all the xds resources
type XdsResources map[resource.Type][]types.Resource

// CacheVersion holds all the translated xds resources
type CacheVersion struct {
	XdsResources XdsResources
}

// GetXdsResources retrieves the translated xds resources saved in the translator context.
func (t *CacheVersion) GetXdsResources() XdsResources {
	return t.XdsResources
}

func (t *CacheVersion) AddXdsResource(rType resource.Type, xdsResource types.Resource) {
	if t.XdsResources == nil {
		t.XdsResources = make(XdsResources)
	}
	if t.XdsResources[rType] == nil {
		t.XdsResources[rType] = make([]types.Resource, 0, 1)
	}

	t.XdsResources[rType] = append(t.XdsResources[rType], xdsResource)
}
