// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"
	"strings"

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
	"github.com/envoyproxy/gateway/internal/status"
	_ "github.com/envoyproxy/gateway/internal/xds/extensions" // register the generated types to support protojson unmarshalling
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	AddOperation = "add"
	EmptyPath    = ""
)

// processJSONPatches applies each JSONPatch to the Xds Resources for a specific type.
func processJSONPatches(tCtx *types.ResourceVersionTable, envoyPatchPolicies []*ir.EnvoyPatchPolicy) error {
	var errs error
	m := protojson.MarshalOptions{
		UseProtoNames: true,
	}

	for _, e := range envoyPatchPolicies {
		for _, p := range e.JSONPatches {
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
				y, err := yaml.Marshal(p.Operation.Value)
				if err != nil {
					msg := fmt.Sprintf("unable to marshal patch %+v, err: %s", p.Operation.Value, err.Error())
					status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
					continue
				}
				jsonBytes, err := yaml.YAMLToJSON(y)
				if err != nil {
					msg := fmt.Sprintf("unable to convert patch to json %s, err: %s", string(y), err.Error())
					status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
					continue
				}
				switch p.Type {
				case string(resourcev3.ListenerType):
					temp := &listenerv3.Listener{}
					if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
						msg := unmarshalErrorMessage(err, p.Operation.Value)
						status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
						continue
					}
					if err = tCtx.AddXdsResource(resourcev3.ListenerType, temp); err != nil {
						msg := fmt.Sprintf("validation failed for xds resource %+v, err:%s", p.Operation.Value, err.Error())
						status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
						continue
					}

				case string(resourcev3.RouteType):
					temp := &routev3.RouteConfiguration{}
					if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
						msg := unmarshalErrorMessage(err, p.Operation.Value)
						status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
						continue
					}
					if err = tCtx.AddXdsResource(resourcev3.RouteType, temp); err != nil {
						msg := fmt.Sprintf("validation failed for xds resource %+v, err:%s", p.Operation.Value, err.Error())
						status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
						continue
					}

				case string(resourcev3.ClusterType):
					temp := &clusterv3.Cluster{}
					if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
						msg := unmarshalErrorMessage(err, p.Operation.Value)
						status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
						continue
					}
					if err = tCtx.AddXdsResource(resourcev3.ClusterType, temp); err != nil {
						msg := fmt.Sprintf("validation failed for xds resource %+v, err:%s", p.Operation.Value, err.Error())
						status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
						continue
					}

				case string(resourcev3.EndpointType):
					temp := &endpointv3.ClusterLoadAssignment{}
					if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
						msg := unmarshalErrorMessage(err, p.Operation.Value)
						status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
						continue
					}
					if err = tCtx.AddXdsResource(resourcev3.EndpointType, temp); err != nil {
						msg := fmt.Sprintf("validation failed for xds resource %+v, err:%s", p.Operation.Value, err.Error())
						status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
						continue
					}

				}

				// Skip further processing
				continue
			}
			// Find the resource to patch and convert it to JSON
			switch p.Type {
			case string(resourcev3.ListenerType):
				if listener = findXdsListener(tCtx, p.Name); listener == nil {
					msg := fmt.Sprintf("unable to find xds resource %s: %s", p.Type, p.Name)
					status.SetEnvoyPatchPolicyResourceNotFound(e.Status, msg)
					continue
				}

				if resourceJSON, err = m.Marshal(listener); err != nil {
					err := fmt.Errorf("unable to marshal xds resource %s: %s, err:%v", p.Type, p.Name, err)
					errs = multierror.Append(errs, err)
					continue
				}

			case string(resourcev3.RouteType):
				if routeConfig = findXdsRouteConfig(tCtx, p.Name); routeConfig == nil {
					msg := fmt.Sprintf("unable to find xds resource %s: %s", p.Type, p.Name)
					status.SetEnvoyPatchPolicyResourceNotFound(e.Status, msg)
					continue
				}

				if resourceJSON, err = m.Marshal(routeConfig); err != nil {
					err = fmt.Errorf("unable to marshal xds resource %s: %s, err:%v", p.Type, p.Name, err)
					errs = multierror.Append(errs, err)
					continue
				}

			case string(resourcev3.ClusterType):
				if cluster = findXdsCluster(tCtx, p.Name); cluster == nil {
					msg := fmt.Sprintf("unable to find xds resource %s: %s", p.Type, p.Name)
					status.SetEnvoyPatchPolicyResourceNotFound(e.Status, msg)
					continue
				}

				if resourceJSON, err = m.Marshal(cluster); err != nil {
					err = fmt.Errorf("unable to marshal xds resource %s: %s, err:%v", p.Type, p.Name, err)
					errs = multierror.Append(errs, err)
					continue
				}
			case string(resourcev3.EndpointType):
				if endpoint = findXdsEndpoint(tCtx, p.Name); endpoint == nil {
					msg := fmt.Sprintf("unable to find xds resource %s: %s", p.Type, p.Name)
					status.SetEnvoyPatchPolicyResourceNotFound(e.Status, msg)
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
				msg := fmt.Sprintf("unable to marshal patch %+v, err: %s", p.Operation, err.Error())
				status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
				continue
			}
			jsonBytes, err := yaml.YAMLToJSON(y)
			if err != nil {
				msg := fmt.Sprintf("unable to convert patch to json %s, err: %s", string(y), err.Error())
				status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
				continue
			}

			patchObj, err := jsonpatchv5.DecodePatch(jsonBytes)
			if err != nil {
				msg := fmt.Sprintf("unable to decode patch %s, err: %s", string(jsonBytes), err.Error())
				status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
				continue
			}

			// Apply patch
			opts := jsonpatchv5.NewApplyOptions()
			opts.EnsurePathExistsOnAdd = true
			modifiedJSON, err := patchObj.ApplyWithOptions(resourceJSON, opts)
			if err != nil {
				msg := fmt.Sprintf("unable to apply patch:\n%s on resource:\n%s, err: %s", string(jsonBytes), string(resourceJSON), err.Error())
				status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
				continue
			}

			// Unmarshal back to typed resource
			// Use a temp staging variable that can be marshalled
			// into and validated before saving it into the xds output resource
			switch p.Type {
			case string(resourcev3.ListenerType):
				temp := &listenerv3.Listener{}
				if err = protojson.Unmarshal(modifiedJSON, temp); err != nil {
					msg := fmt.Sprintf("unable to unmarshal xds resource %s, err:%s", string(modifiedJSON), err.Error())
					status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
					continue
				}
				if err = temp.Validate(); err != nil {
					msg := fmt.Sprintf("validation failed for xds resource %s, err:%s", string(modifiedJSON), err.Error())
					status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
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
					msg := fmt.Sprintf("unable to unmarshal xds resource %s, err:%s", string(modifiedJSON), err.Error())
					status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
					continue
				}
				if err = temp.Validate(); err != nil {
					msg := fmt.Sprintf("validation failed for xds resource %s, err:%s", string(modifiedJSON), err.Error())
					status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
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
					msg := fmt.Sprintf("unable to unmarshal xds resource %s, err:%s", string(modifiedJSON), err.Error())
					status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
					continue
				}
				if err = temp.Validate(); err != nil {
					msg := fmt.Sprintf("validation failed for xds resource %s, err:%s", string(modifiedJSON), err.Error())
					status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
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
					msg := fmt.Sprintf("unable to unmarshal xds resource %s, err:%s", string(modifiedJSON), err.Error())
					status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
					continue
				}
				if err = temp.Validate(); err != nil {
					msg := fmt.Sprintf("validation failed for xds resource %s, err:%s", string(modifiedJSON), err.Error())
					status.SetEnvoyPatchPolicyInvalid(e.Status, msg)
					continue
				}
				if err = deepCopyPtr(temp, endpoint); err != nil {
					err := fmt.Errorf("unable to copy xds resource %s, err:%v", string(modifiedJSON), err)
					errs = multierror.Append(errs, err)
					continue
				}
			}
		}

		// Set Programmed condition if not yet set
		status.SetEnvoyPatchPolicyProgrammedIfUnset(e.Status, "successfully applied patches.")

		// Set output context

		tCtx.EnvoyPatchPolicyStatuses = append(tCtx.EnvoyPatchPolicyStatuses, &e.EnvoyPatchPolicyStatus)
	}
	return errs
}

var unescaper = strings.NewReplacer(`\u00a0`, ` `)

func unmarshalErrorMessage(err error, xdsResource any) string {
	return fmt.Sprintf("unable to unmarshal xds resource %+v, err:%s", xdsResource, unescaper.Replace(err.Error()))
}
