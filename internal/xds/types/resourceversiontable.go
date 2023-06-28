// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package types

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/proto"
)

// XdsResources represents all the xds resources
type XdsResources = map[resourcev3.Type][]types.Resource

// ResourceVersionTable holds all the translated xds resources
type ResourceVersionTable struct {
	XdsResources
}

// DeepCopyInto copies the contents into the output object
// This was generated by controller-gen, moved from
// zz_generated.deepcopy.go and updated to use proto.Clone
// to deep copy the proto.Message
func (t *ResourceVersionTable) DeepCopyInto(out *ResourceVersionTable) {
	*out = *t
	if t.XdsResources != nil {
		in, out := &t.XdsResources, &out.XdsResources
		*out = make(map[string][]types.Resource, len(*in))
		for key, val := range *in {
			var outVal []types.Resource
			if val == nil {
				(*out)[key] = nil
			} else {
				// Snippet was generated by controller-gen
				//G601: Implicit memory aliasing in for loop.
				in, out := &val, &outVal //nolint:gosec,scopelint
				*out = make([]types.Resource, len(*in))
				for i := range *in {
					(*out)[i] = proto.Clone((*in)[i])
				}
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopy generates a deep copy of the ResourceVersionTable object.
// This was generated by controller-gen and moved over from
// zz_generated.deepcopy.go to this file.
func (t *ResourceVersionTable) DeepCopy() *ResourceVersionTable {
	if t == nil {
		return nil
	}
	out := new(ResourceVersionTable)
	t.DeepCopyInto(out)
	return out
}

// GetXdsResources retrieves the translated xds resources saved in the translator context.
func (t *ResourceVersionTable) GetXdsResources() XdsResources {
	return t.XdsResources
}

func (t *ResourceVersionTable) AddXdsResource(rType resourcev3.Type, xdsResource types.Resource) {
	if t.XdsResources == nil {
		t.XdsResources = make(XdsResources)
	}
	if t.XdsResources[rType] == nil {
		t.XdsResources[rType] = make([]types.Resource, 0, 1)
	}

	t.XdsResources[rType] = append(t.XdsResources[rType], xdsResource)
}

// AddOrReplaceXdsResource will update an existing resource of rType according to matchFunc or add as a new resource
// if none satisfy the match criteria. It will only update the first match it finds, regardless
// if multiple resources satisfy the match criteria.
func (t *ResourceVersionTable) AddOrReplaceXdsResource(rType resourcev3.Type, resource types.Resource, matchFunc func(existing types.Resource, new types.Resource) bool) {
	if t.XdsResources == nil || t.XdsResources[rType] == nil {
		t.AddXdsResource(rType, resource)
		return
	}

	var found bool
	for i, r := range t.XdsResources[rType] {
		if matchFunc(r, resource) {
			t.XdsResources[rType][i] = resource
			found = true
			break
		}
	}
	if !found {
		t.AddXdsResource(rType, resource)
	}
}

// SetResources will update an entire entry of the XdsResources for a certain type to the provided resources
func (t *ResourceVersionTable) SetResources(rType resourcev3.Type, xdsResources []types.Resource) {
	if t.XdsResources == nil {
		t.XdsResources = make(XdsResources)
	}

	t.XdsResources[rType] = xdsResources
}
