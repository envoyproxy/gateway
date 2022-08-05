package types

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

// XdsResources represents all the xds resources
//
// This is the type that
// github.com/envoyproxy/go-control-plane/pkg/cache/v3.NewSnapshot
// takes; if we decide that we want to change this type, then we'd
// have to do conversion.
type XdsResources = map[resource.Type][]types.Resource

// ResourceVersionTable holds all the translated xds resources
type ResourceVersionTable struct {
	XdsResources XdsResources
}

// GetXdsResources retrieves the translated xds resources saved in the translator context.
func (t *ResourceVersionTable) GetXdsResources() XdsResources {
	return t.XdsResources
}

func (t *ResourceVersionTable) AddXdsResource(rType resource.Type, xdsResource types.Resource) {
	if t.XdsResources == nil {
		t.XdsResources = make(XdsResources)
	}
	if t.XdsResources[rType] == nil {
		t.XdsResources[rType] = make([]types.Resource, 0, 1)
	}

	t.XdsResources[rType] = append(t.XdsResources[rType], xdsResource)
}
