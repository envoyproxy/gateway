// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"
	"sort"

	runtimev3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// RuntimeLayerName is the name of the RTDS layer that Envoy Gateway serves. It must match
// the rtds_layer name in the bootstrap, since Envoy subscribes to the Runtime resource by
// that name.
//
// The name is deliberately Envoy Gateway specific. Envoy rejects a bootstrap containing two
// runtime layers with the same name, so a generic name risks colliding with a layer a user
// already added through spec.bootstrap and leaving the proxy unable to start.
const RuntimeLayerName = "envoy-gateway-runtime"

// processRuntime builds the RTDS Runtime resource from the runtime values in the IR.
//
// The resource is emitted unconditionally, even when no runtime values are configured. The
// bootstrap subscribes to this layer, so a proxy that never receives the resource logs an
// initial fetch timeout for the Runtime type and keeps an uninitialized layer. Always
// sending it, empty or not, keeps the subscription satisfied.
func processRuntime(tCtx *types.ResourceVersionTable, xdsIR *ir.Xds) error {
	layer, err := buildRuntimeLayer(xdsIR)
	if err != nil {
		return err
	}

	return tCtx.AddXdsResource(resourcev3.RuntimeType, &runtimev3.Runtime{
		Name:  RuntimeLayerName,
		Layer: layer,
	})
}

// buildRuntimeLayer converts the IR runtime values into the Struct that Envoy loads as a
// runtime layer. Values keep the JSON type they were written with: Envoy resolves boolean
// runtime guards only from real booleans, and numeric limits from numbers or from strings
// that parse as numbers, so collapsing everything to a string would silently disable
// boolean keys.
//
// An unparsable value fails translation, which holds back the whole snapshot. That is
// deliberate: the API server only admits valid JSON here, so reaching this is close to
// impossible, and a runtime key that silently fails to apply is worse than a stalled
// update. A dropped circuit breaker limit, for instance, would leave the proxy running on
// Envoy's default with nothing to indicate it.
func buildRuntimeLayer(xdsIR *ir.Xds) (*structpb.Struct, error) {
	if len(xdsIR.Runtime) == 0 {
		return nil, nil
	}

	// The layer is a proto map, so insertion order does not affect the resulting resource.
	// The keys are sorted only so that a config with more than one bad value always reports
	// the same one, rather than a different key on each translation.
	keys := make([]string, 0, len(xdsIR.Runtime))
	for k := range xdsIR.Runtime {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fields := make(map[string]*structpb.Value, len(keys))
	for _, k := range keys {
		raw := xdsIR.Runtime[k]
		value := &structpb.Value{}
		if err := value.UnmarshalJSON(raw.Raw); err != nil {
			return nil, fmt.Errorf("invalid value for runtime key %q: %w", k, err)
		}
		fields[k] = value
	}

	return &structpb.Struct{Fields: fields}, nil
}
