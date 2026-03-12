// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/encoding/protojson"

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

	for _, e := range envoyPatchPolicies {
		var (
			e                 = e
			tErrs             error
			notFoundResources []string
		)

		for _, p := range e.JSONPatches {
			var (
				resourceJSON []byte
				dest         cachetypes.Resource
				err          error
			)

			if err := p.Operation.Validate(); err != nil {
				tErrs = errors.Join(tErrs, err)
				continue
			}

			// If Path and JSONPath is "" and op is "add", unmarshal and add the patch as a complete
			// resource
			if p.Operation.Op == ir.JSONPatchOpAdd && p.Operation.IsPathNilOrEmpty() && p.Operation.IsJSONPathNilOrEmpty() {
				if p.Operation.Value == nil {
					tErr := fmt.Errorf("missing value for add operation with empty path")
					tErrs = errors.Join(tErrs, tErr)
					continue
				}

				jsonBytes := p.Operation.Value.Raw
				if len(jsonBytes) == 0 {
					tErr := fmt.Errorf("empty value for add operation with empty path")
					tErrs = errors.Join(tErrs, tErr)
					continue
				}

				temp, err := getXdsResourceType(p.Type)
				if err != nil {
					tErrs = errors.Join(tErrs, err)
					continue
				}

				if err = protojson.Unmarshal(jsonBytes, temp); err != nil {
					tErr := errors.New(unmarshalErrorMessage(err, string(jsonBytes)))
					tErrs = errors.Join(tErrs, tErr)
					continue
				}

				if err = tCtx.AddXdsResource(p.Type, temp); err != nil {
					tErr := fmt.Errorf("validation failed for xds resource %s, err:%s", p.Type, err.Error())
					tErrs = errors.Join(tErrs, tErr)
					continue
				}

				// Skip further processing
				continue
			}

			// find the resource to patch and convert it to JSON
			dest, err = findXdsResource(tCtx, p)
			if err != nil {
				if errors.Is(err, errResourceNotFound) {
					tn := typedName{p.Type, p.Name}
					notFoundResources = append(notFoundResources, tn.String())
					continue
				}

				tErrs = errors.Join(tErrs, err)
				continue
			}

			resourceJSON, err = jsonMarshalOpts.Marshal(dest)
			if err != nil {
				tErr := fmt.Errorf("unable to marshal xds resource %s, err: %w", p.Type, err)
				tErrs = errors.Join(tErrs, tErr)
				continue
			}

			modifiedJSON, err := jsonpatch.ApplyJSONPatches(resourceJSON, p.Operation)
			if err != nil {
				tErrs = errors.Join(tErrs, err)
				continue
			}

			// Unmarshal back to typed resource
			// Use a temp staging variable that can be marshalled
			// into and validated before saving it into the xds output resource
			temp, err := getXdsResourceType(p.Type)
			if err != nil {
				tErrs = errors.Join(tErrs, err)
				continue
			}

			if err = protojson.Unmarshal(modifiedJSON, temp); err != nil {
				tErr := errors.New(unmarshalErrorMessage(err, string(modifiedJSON)))
				tErrs = errors.Join(tErrs, tErr)
				continue
			}

			// Validate the patched resource
			validator, ok := temp.(interface{ Validate() error })
			if ok {
				if err = validator.Validate(); err != nil {
					tErr := fmt.Errorf("validation failed for xds resource %s, err:%s", p.Type, err.Error())
					tErrs = errors.Join(tErrs, tErr)
					continue
				}
			}

			if err = deepCopyPtr(temp, dest); err != nil {
				tErr := fmt.Errorf("unable to copy xds resource %s, err: %w", p.Type, err)
				tErrs = errors.Join(tErrs, tErr)
				continue
			}
		}

		// Set translation errors for every policy ancestor references
		if tErrs != nil {
			status.SetTranslationErrorForEnvoyPatchPolicy(e.Status, status.Error2ConditionMsg(tErrs), e.Generation)
			errs = errors.Join(errs, tErrs)
		}

		// Set resources not found errors for every policy ancestor references
		if len(notFoundResources) > 0 {
			status.SetResourceNotFoundErrorForEnvoyPatchPolicy(e.Status, notFoundResources, e.Generation)
		}

		// Set Programmed condition if not yet set
		status.SetProgrammedForEnvoyPatchPolicy(e.Status, e.Generation)

		// Set output context
		tCtx.EnvoyPatchPolicyStatuses = append(tCtx.EnvoyPatchPolicyStatuses, &e.EnvoyPatchPolicyStatus)
	}

	return errs
}

func getXdsResourceType(resourceType string) (cachetypes.Resource, error) {
	switch resourceType {
	case resourcev3.ListenerType:
		return &listenerv3.Listener{}, nil
	case resourcev3.RouteType:
		return &routev3.RouteConfiguration{}, nil
	case resourcev3.ClusterType:
		return &clusterv3.Cluster{}, nil
	case resourcev3.EndpointType:
		return &endpointv3.ClusterLoadAssignment{}, nil
	case resourcev3.SecretType:
		return &tlsv3.Secret{}, nil
	default:
		return nil, fmt.Errorf("unsupported patch type %s", resourceType)
	}
}

var (
	errResourceNotFound = errors.New("resource not found")
	jsonMarshalOpts     = protojson.MarshalOptions{
		UseProtoNames: true,
	}
)

// findXdsResource return the XDS resource to patch
// TODO: return multiple resources
func findXdsResource(tCtx *types.ResourceVersionTable, p *ir.JSONPatchConfig) (cachetypes.Resource, error) {
	var resource cachetypes.Resource
	switch p.Type {
	case resourcev3.ListenerType:
		resource = findXdsListener(tCtx, p.Name)
	case resourcev3.RouteType:
		resource = findXdsRouteConfig(tCtx, p.Name)
	case resourcev3.ClusterType:
		resource = findXdsCluster(tCtx, p.Name)
	case resourcev3.EndpointType:
		resource = findXdsEndpoint(tCtx, p.Name)
	case resourcev3.SecretType:
		resource = findXdsSecret(tCtx, p.Name)
	default:
		return nil, fmt.Errorf("unsupported patch type %s", p.Type)
	}

	// resource can be a non-nil interface with a nil pointer value
	if resource == nil || reflect.ValueOf(resource).IsNil() {
		return nil, errResourceNotFound
	}

	return resource, nil
}

var unescaper = strings.NewReplacer(" ", " ")

func unmarshalErrorMessage(err error, xdsResource any) string {
	return fmt.Sprintf("unable to unmarshal xds resource %+v, err:%s", xdsResource, unescaper.Replace(err.Error()))
}
