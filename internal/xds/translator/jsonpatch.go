// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"strings"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/jsonpatch"
	_ "github.com/envoyproxy/gateway/internal/xds/extensions" // register the generated types to support protojson unmarshalling
	"github.com/envoyproxy/gateway/internal/xds/types"
)

type typedName struct {
	Type string
	Name string
}

func (t typedName) String() string {
	return fmt.Sprintf("%s/%s", t.Type, t.Name)
}

// processJSONPatches applies each JSONPatch to the Xds Resources for a specific type.
func processJSONPatches(tCtx *types.ResourceVersionTable, envoyPatchPolicies []*ir.EnvoyPatchPolicy) error {
	var errs error
	m := protojson.MarshalOptions{
		UseProtoNames: true,
	}

	for _, e := range envoyPatchPolicies {
		var (
			e                 = e
			tErrs             error
			notFoundResources []string
		)

		for _, p := range e.JSONPatches {
			var (
				listener     *listenerv3.Listener
				routeConfig  *routev3.RouteConfiguration
				cluster      *clusterv3.Cluster
				endpoint     *endpointv3.ClusterLoadAssignment
				secret       *tlsv3.Secret
				resourceJSON []byte
				err          error
			)

			if err := p.Operation.Validate(); err != nil {
				tErrs = errors.Join(tErrs, err)
				continue
			}

			// If Path and JSONPath is "" and op is "add", unmarshal and add the patch as a complete
			// resource
			if p.Operation.Op == ir.JSONPatchOpAdd && p.Operation.IsPathNilOrEmpty() && p.Operation.IsJSONPathNilOrEmpty() {
				// Convert patch to JSON
				// The patch library expects an array so convert it into one
				y, err := yaml.Marshal(p.Operation.Value)
				if err != nil {
					tErr := fmt.Errorf("unable to marshal patch %+v, err: %s", p.Operation.Value, err.Error())
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
				jsonBytes, err := yaml.YAMLToJSON(y)
				if err != nil {
					tErr := fmt.Errorf("unable to convert patch to json %s, err: %s", string(y), err.Error())
					tErrs = errors.Join(tErrs, tErr)
					continue
				}

				switch p.Type {
				case resourcev3.ListenerType:
					temp := &listenerv3.Listener{}
					if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
						tErr := errors.New(unmarshalErrorMessage(err, p.Operation.Value))
						tErrs = errors.Join(tErrs, tErr)
						continue
					}
					if err = tCtx.AddXdsResource(resourcev3.ListenerType, temp); err != nil {
						tErr := fmt.Errorf("validation failed for xds resource %+v, err:%s", p.Operation.Value, err.Error())
						tErrs = errors.Join(tErrs, tErr)
						continue
					}

				case resourcev3.RouteType:
					temp := &routev3.RouteConfiguration{}
					if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
						tErr := errors.New(unmarshalErrorMessage(err, p.Operation.Value))
						tErrs = errors.Join(tErrs, tErr)
						continue
					}
					if err = tCtx.AddXdsResource(resourcev3.RouteType, temp); err != nil {
						tErr := fmt.Errorf("validation failed for xds resource %+v, err:%s", p.Operation.Value, err.Error())
						tErrs = errors.Join(tErrs, tErr)
						continue
					}

				case resourcev3.ClusterType:
					temp := &clusterv3.Cluster{}
					if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
						tErr := errors.New(unmarshalErrorMessage(err, p.Operation.Value))
						tErrs = errors.Join(tErrs, tErr)
						continue
					}
					if err = tCtx.AddXdsResource(resourcev3.ClusterType, temp); err != nil {
						tErr := fmt.Errorf("validation failed for xds resource %+v, err:%s", p.Operation.Value, err.Error())
						tErrs = errors.Join(tErrs, tErr)
						continue
					}

				case resourcev3.EndpointType:
					temp := &endpointv3.ClusterLoadAssignment{}
					if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
						tErr := errors.New(unmarshalErrorMessage(err, p.Operation.Value))
						tErrs = errors.Join(tErrs, tErr)
						continue
					}
					if err = tCtx.AddXdsResource(resourcev3.EndpointType, temp); err != nil {
						tErr := fmt.Errorf("validation failed for xds resource %+v, err:%s", p.Operation.Value, err.Error())
						tErrs = errors.Join(tErrs, tErr)
						continue
					}

				case resourcev3.SecretType:
					temp := &tlsv3.Secret{}
					if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
						tErr := errors.New(unmarshalErrorMessage(err, p.Operation.Value))
						tErrs = errors.Join(tErrs, tErr)
						continue
					}
					if err = tCtx.AddXdsResource(resourcev3.SecretType, temp); err != nil {
						tErr := fmt.Errorf("validation failed for xds resource %+v, err:%s", p.Operation.Value, err.Error())
						tErrs = errors.Join(tErrs, tErr)
						continue
					}

				}

				// Skip further processing
				continue
			}

			// Find the resource to patch and convert it to JSON
			switch p.Type {
			case resourcev3.ListenerType:
				if listener = findXdsListener(tCtx, p.Name, corev3.SocketAddress_TCP); listener == nil {
					tn := typedName{p.Type, p.Name}
					notFoundResources = append(notFoundResources, tn.String())
					continue
				}

				if resourceJSON, err = m.Marshal(listener); err != nil {
					tErr := fmt.Errorf("unable to marshal xds resource %s: %s, err: %w", p.Type, p.Name, err)
					tErrs = errors.Join(tErrs, tErr)
					continue
				}

			case resourcev3.RouteType:
				if routeConfig = findXdsRouteConfig(tCtx, p.Name); routeConfig == nil {
					tn := typedName{p.Type, p.Name}
					notFoundResources = append(notFoundResources, tn.String())
					continue
				}

				if resourceJSON, err = m.Marshal(routeConfig); err != nil {
					tErr := fmt.Errorf("unable to marshal xds resource %s: %s, err: %w", p.Type, p.Name, err)
					tErrs = errors.Join(tErrs, tErr)
					continue
				}

			case resourcev3.ClusterType:
				if cluster = findXdsCluster(tCtx, p.Name); cluster == nil {
					tn := typedName{p.Type, p.Name}
					notFoundResources = append(notFoundResources, tn.String())
					continue
				}

				if resourceJSON, err = m.Marshal(cluster); err != nil {
					tErr := fmt.Errorf("unable to marshal xds resource %s: %s, err: %w", p.Type, p.Name, err)
					tErrs = errors.Join(tErrs, tErr)
					continue
				}

			case resourcev3.EndpointType:
				if endpoint = findXdsEndpoint(tCtx, p.Name); endpoint == nil {
					tn := typedName{p.Type, p.Name}
					notFoundResources = append(notFoundResources, tn.String())
					continue
				}
				if resourceJSON, err = m.Marshal(endpoint); err != nil {
					tErr := fmt.Errorf("unable to marshal xds resource %s: %s, err: %w", p.Type, p.Name, err)
					tErrs = errors.Join(tErrs, tErr)
					continue
				}

			case resourcev3.SecretType:
				if secret = findXdsSecret(tCtx, p.Name); secret == nil {
					tn := typedName{p.Type, p.Name}
					notFoundResources = append(notFoundResources, tn.String())
					continue
				}
				if resourceJSON, err = m.Marshal(secret); err != nil {
					tErr := fmt.Errorf("unable to marshal xds resource %s: %s, err: %w", p.Type, p.Name, err)
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
			}

			modifiedJSON, err := jsonpatch.ApplyJSONPatches(resourceJSON, p.Operation)
			if err != nil {
				tErrs = errors.Join(tErrs, err)
				continue
			}

			// Unmarshal back to typed resource
			// Use a temp staging variable that can be marshalled
			// into and validated before saving it into the xds output resource
			switch p.Type {
			case resourcev3.ListenerType:
				temp := &listenerv3.Listener{}
				if err = protojson.Unmarshal(modifiedJSON, temp); err != nil {
					tErr := errors.New(unmarshalErrorMessage(err, string(modifiedJSON)))
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
				if err = temp.Validate(); err != nil {
					tErr := fmt.Errorf("validation failed for xds resource %s, err:%s", string(modifiedJSON), err.Error())
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
				if err = deepCopyPtr(temp, listener); err != nil {
					tErr := fmt.Errorf("unable to copy xds resource %s, err: %w", string(modifiedJSON), err)
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
			case resourcev3.RouteType:
				temp := &routev3.RouteConfiguration{}
				if err = protojson.Unmarshal(modifiedJSON, temp); err != nil {
					tErr := errors.New(unmarshalErrorMessage(err, string(modifiedJSON)))
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
				if err = temp.Validate(); err != nil {
					tErr := fmt.Errorf("validation failed for xds resource %s, err:%s", string(modifiedJSON), err.Error())
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
				if err = deepCopyPtr(temp, routeConfig); err != nil {
					tErr := fmt.Errorf("unable to copy xds resource %s, err: %w", string(modifiedJSON), err)
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
			case resourcev3.ClusterType:
				temp := &clusterv3.Cluster{}
				if err = protojson.Unmarshal(modifiedJSON, temp); err != nil {
					tErr := errors.New(unmarshalErrorMessage(err, string(modifiedJSON)))
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
				if err = temp.Validate(); err != nil {
					tErr := fmt.Errorf("validation failed for xds resource %s, err:%s", string(modifiedJSON), err.Error())
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
				if err = deepCopyPtr(temp, cluster); err != nil {
					tErr := fmt.Errorf("unable to copy xds resource %s, err: %w", string(modifiedJSON), err)
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
			case resourcev3.EndpointType:
				temp := &endpointv3.ClusterLoadAssignment{}
				if err = protojson.Unmarshal(modifiedJSON, temp); err != nil {
					tErr := errors.New(unmarshalErrorMessage(err, string(modifiedJSON)))
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
				if err = temp.Validate(); err != nil {
					tErr := fmt.Errorf("validation failed for xds resource %s, err:%s", string(modifiedJSON), err.Error())
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
				if err = deepCopyPtr(temp, endpoint); err != nil {
					tErr := fmt.Errorf("unable to copy xds resource %s, err: %w", string(modifiedJSON), err)
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
			case resourcev3.SecretType:
				temp := &tlsv3.Secret{}
				if err = protojson.Unmarshal(modifiedJSON, temp); err != nil {
					tErr := errors.New(unmarshalErrorMessage(err, string(modifiedJSON)))
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
				if err = temp.Validate(); err != nil {
					tErr := fmt.Errorf("validation failed for xds resource %s, err:%s", string(modifiedJSON), err.Error())
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
				if err = deepCopyPtr(temp, secret); err != nil {
					tErr := fmt.Errorf("unable to copy xds resource %s, err: %w", string(modifiedJSON), err)
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
			}

		}

		// Set translation errors for every policy ancestor references
		if tErrs != nil {
			status.SetTranslationErrorForEnvoyPatchPolicy(e.Status, status.Error2ConditionMsg(tErrs))
			errs = errors.Join(errs, tErrs)
		}

		// Set resources not found errors for every policy ancestor references
		if len(notFoundResources) > 0 {
			status.SetResourceNotFoundErrorForEnvoyPatchPolicy(e.Status, notFoundResources)
		}

		// Set Programmed condition if not yet set
		status.SetProgrammedForEnvoyPatchPolicy(e.Status)

		// Set output context
		tCtx.EnvoyPatchPolicyStatuses = append(tCtx.EnvoyPatchPolicyStatuses, &e.EnvoyPatchPolicyStatus)
	}

	return errs
}

var unescaper = strings.NewReplacer(" ", " ")

func unmarshalErrorMessage(err error, xdsResource any) string {
	return fmt.Sprintf("unable to unmarshal xds resource %+v, err:%s", xdsResource, unescaper.Replace(err.Error()))
}
