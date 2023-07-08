// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	jsonpatchv5 "github.com/evanphx/json-patch/v5"
	"github.com/tetratelabs/multierror"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/ir"
	_ "github.com/envoyproxy/gateway/internal/xds/extensions" // register the generated types to support protojson unmarshalling
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	AddOperation = "add"
	EmptyPath    = ""
)

// processJSONPatches applies each JSONPatch to the Xds Resources for a specific type.
func processJSONPatches(tCtx *types.ResourceVersionTable, jsonPatches []*ir.JSONPatchConfig) error {
	var errs error
	m := protojson.MarshalOptions{
		UseProtoNames: true,
	}

	for _, p := range jsonPatches {
		var (
			listener     *listenerv3.Listener
			routeConfig  *routev3.RouteConfiguration
			cluster      *clusterv3.Cluster
			endpoint     *endpointv3.ClusterLoadAssignment
			resourceJSON []byte
			err          error
		)

		// If Path is "" and op is "add", unmarshal and add the patch as a complete
		// resource
		if p.Operation.Op == AddOperation && p.Operation.Path == EmptyPath {
			// Convert patch to JSON
			// The patch library expects an array so convert it into one
			jsonBytes, err := yaml.YAMLToJSON([]byte(p.Operation.Value))
			if err != nil {
				err := fmt.Errorf("unable to convert patch to json %s, err: %v", string(jsonBytes), err)
				errs = multierror.Append(errs, err)
				continue
			}
			switch p.Type {
			case string(resourcev3.ListenerType):
				temp := &listenerv3.Listener{}
				if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
					err := fmt.Errorf("unable to unmarshal xds resource %+v, err:%v", p.Operation.Value, err)
					errs = multierror.Append(errs, err)
					continue
				}
				if err = temp.Validate(); err != nil {
					err := fmt.Errorf("validation failed for xds resource %+v, err:%v", p.Operation.Value, err)
					errs = multierror.Append(errs, err)
					continue
				}

				tCtx.AddXdsResource(resourcev3.ListenerType, temp)
			case string(resourcev3.RouteType):
				temp := &routev3.RouteConfiguration{}
				if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
					err := fmt.Errorf("unable to unmarshal xds resource %+v, err:%v", p.Operation.Value, err)
					errs = multierror.Append(errs, err)
					continue
				}
				if err = temp.Validate(); err != nil {
					err := fmt.Errorf("validation failed for xds resource %+v, err:%v", p.Operation.Value, err)
					errs = multierror.Append(errs, err)
					continue
				}
				tCtx.AddXdsResource(resourcev3.RouteType, temp)
			case string(resourcev3.ClusterType):
				temp := &clusterv3.Cluster{}
				if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
					err := fmt.Errorf("unable to unmarshal xds resource %+v, err:%v", p.Operation.Value, err)
					errs = multierror.Append(errs, err)
					continue
				}
				if err = temp.Validate(); err != nil {
					err := fmt.Errorf("validation failed for xds resource %+v, err:%v", p.Operation.Value, err)
					errs = multierror.Append(errs, err)
					continue
				}
				tCtx.AddXdsResource(resourcev3.ClusterType, temp)
			case string(resourcev3.EndpointType):
				temp := &endpointv3.ClusterLoadAssignment{}
				if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
					err := fmt.Errorf("unable to unmarshal xds resource %+v, err:%v", p.Operation.Value, err)
					errs = multierror.Append(errs, err)
					continue
				}
				if err = temp.Validate(); err != nil {
					err := fmt.Errorf("validation failed for xds resource %+v, err:%v", p.Operation.Value, err)
					errs = multierror.Append(errs, err)
					continue
				}
				tCtx.AddXdsResource(resourcev3.EndpointType, temp)
			}

			// Skip further processing
			continue
		}
		// Find the resource to patch and convert it to JSON
		switch p.Type {
		case string(resourcev3.ListenerType):
			if listener = findXdsListener(tCtx, p.Name); listener == nil {
				err = fmt.Errorf("unable to find xds resource %s: %s", p.Type, p.Name)
				errs = multierror.Append(errs, err)
				continue
			}

			if resourceJSON, err = m.Marshal(listener); err != nil {
				err = fmt.Errorf("unable to marshal xds resource %s: %s, err:%v", p.Type, p.Name, err)
				errs = multierror.Append(errs, err)
				continue
			}

		case string(resourcev3.RouteType):
			if routeConfig = findXdsRouteConfig(tCtx, p.Name); routeConfig == nil {
				err := fmt.Errorf("unable to find xds resource %s: %s", p.Type, p.Name)
				errs = multierror.Append(errs, err)
				continue
			}

			if resourceJSON, err = m.Marshal(routeConfig); err != nil {
				err = fmt.Errorf("unable to marshal xds resource %s: %s, err:%v", p.Type, p.Name, err)
				errs = multierror.Append(errs, err)
				continue
			}

		case string(resourcev3.ClusterType):
			if cluster := findXdsCluster(tCtx, p.Name); cluster == nil {
				err := fmt.Errorf("unable to find xds resource %s: %s", p.Type, p.Name)
				errs = multierror.Append(errs, err)
				continue
			}

			if resourceJSON, err = m.Marshal(cluster); err != nil {
				err = fmt.Errorf("unable to marshal xds resource %s: %s, err:%v", p.Type, p.Name, err)
				errs = multierror.Append(errs, err)
				continue
			}
		case string(resourcev3.EndpointType):
			endpoint = findXdsEndpoint(tCtx, p.Name)
			if endpoint == nil {
				err = fmt.Errorf("unable to marshal xds resource %s: %s, err:%v", p.Type, p.Name, err)
				errs = multierror.Append(errs, err)
				continue
			}
			if resourceJSON, err = m.Marshal(endpoint); err != nil {
				err = fmt.Errorf("unable to marshal xds resource %s: %s, err:%v", p.Type, p.Name, err)
				errs = multierror.Append(errs, err)
				continue
			}
		}

		// Convert patch to JSON
		jsonBytes, err := yaml.YAMLToJSON([]byte(p.Operation.Value))
		if err != nil {
			err := fmt.Errorf("unable to convert patch to JSON, err: %v", err)
			errs = multierror.Append(errs, err)
			continue
		}

		// see https://jsonpatch.com/ to understand the format of patchJSON
		patchJSON := []byte(`[{ "op": "` + p.Operation.Op + `", "path": "` + p.Operation.Path + `", "value": ` + string(jsonBytes) + `}]`)

		patchObj, err := jsonpatchv5.DecodePatch(patchJSON)
		if err != nil {
			err := fmt.Errorf("unable to decode patch %s, err: %v", string(patchJSON), err)
			errs = multierror.Append(errs, err)
			continue
		}

		// Apply patch
		opts := jsonpatchv5.NewApplyOptions()
		opts.EnsurePathExistsOnAdd = true
		modifiedJSON, err := patchObj.ApplyWithOptions(resourceJSON, opts)
		if err != nil {
			err := fmt.Errorf("unable to apply patch:\n%s on resource:\n%s, err: %v", string(jsonBytes), string(resourceJSON), err)
			errs = multierror.Append(errs, err)
			continue
		}

		// Unmarshal back to typed resource
		// Use a temp staging variable that can be marshalled
		// into and validated before saving it into the xds output resource
		switch p.Type {
		case string(resourcev3.ListenerType):
			temp := &listenerv3.Listener{}
			if err = protojson.Unmarshal(modifiedJSON, temp); err != nil {
				err := fmt.Errorf("unable to unmarshal xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
			if err = temp.Validate(); err != nil {
				err := fmt.Errorf("validation failed for xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
			if err = deepCopyPtr(temp, listener); err != nil {
				err := fmt.Errorf("unable to copy xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
		case string(resourcev3.RouteType):
			temp := &routev3.RouteConfiguration{}
			if err = protojson.Unmarshal(modifiedJSON, temp); err != nil {
				err := fmt.Errorf("unable to unmarshal xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
			if err = temp.Validate(); err != nil {
				err := fmt.Errorf("validation failed for xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
			if err = deepCopyPtr(temp, routeConfig); err != nil {
				err := fmt.Errorf("unable to copy xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
		case string(resourcev3.ClusterType):
			temp := &clusterv3.Cluster{}
			if err = protojson.Unmarshal(modifiedJSON, temp); err != nil {
				err := fmt.Errorf("unable to unmarshal xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
			if err = temp.Validate(); err != nil {
				err := fmt.Errorf("validation failed for xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
			if err = deepCopyPtr(temp, cluster); err != nil {
				err := fmt.Errorf("unable to copy xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
		case string(resourcev3.EndpointType):
			temp := &endpointv3.ClusterLoadAssignment{}
			if err = protojson.Unmarshal(modifiedJSON, temp); err != nil {
				err := fmt.Errorf("unable to unmarshal xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
			if err = temp.Validate(); err != nil {
				err := fmt.Errorf("validation failed for xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
			if err = deepCopyPtr(temp, endpoint); err != nil {
				err := fmt.Errorf("unable to copy xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
		}
	}
	return errs
}
