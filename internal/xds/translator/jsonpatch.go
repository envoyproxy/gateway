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

// processJSONPatches applies each JSONPatch to the Xds Resources for a specific type.
func processJSONPatches(tCtx *types.ResourceVersionTable, JSONPatches []*ir.JSONPatchConfig) error {
	var errs error
	for _, p := range JSONPatches {
		var (
			listener     *listenerv3.Listener
			routeConfig  *routev3.RouteConfiguration
			cluster      *clusterv3.Cluster
			endpoint     *endpointv3.ClusterLoadAssignment
			resourceJSON []byte
			err          error
		)
		m := protojson.MarshalOptions{
			UseProtoNames: true,
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
		// The patch library expects an array so convert it into one
		y, err := yaml.Marshal([]ir.JSONPatchOperation{p.Operation})
		if err != nil {
			err := fmt.Errorf("unable to marshal patch %+v, err: %v", p.Operation, err)
			errs = multierror.Append(errs, err)
			continue
		}
		jsonBytes, err := yaml.YAMLToJSON(y)

		patchObj, err := jsonpatchv5.DecodePatch(jsonBytes)
		if err != nil {
			err := fmt.Errorf("unable to decode patch %s, err: %v", string(jsonBytes), err)
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
		my, _ := yaml.JSONToYAML(modifiedJSON)
		fmt.Println(string(my))
		// Unmarshal back to typed resource
		switch p.Type {
		case string(resourcev3.ListenerType):
			if err = protojson.Unmarshal(modifiedJSON, listener); err != nil {
				err := fmt.Errorf("unable to unmarshal xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
		case string(resourcev3.RouteType):
			if err = protojson.Unmarshal(modifiedJSON, routeConfig); err != nil {
				err := fmt.Errorf("unable to unmarshal xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
		case string(resourcev3.ClusterType):
			if err = protojson.Unmarshal(modifiedJSON, cluster); err != nil {
				err := fmt.Errorf("unable to unmarshal xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
		case string(resourcev3.EndpointType):
			if err = protojson.Unmarshal(modifiedJSON, endpoint); err != nil {
				err := fmt.Errorf("unable to unmarshal xds resource %s, err:%v", string(modifiedJSON), err)
				errs = multierror.Append(errs, err)
				continue
			}
		}
	}
	return errs
}
