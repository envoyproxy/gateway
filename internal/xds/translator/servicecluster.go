// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func processServiceCluster(tCtx *types.ResourceVersionTable, xdsIR *ir.Xds) error {
	if ptr.Deref(xdsIR, ir.Xds{}).ProxyServiceCluster == nil {
		return nil
	}
	return addXdsCluster(tCtx, &xdsClusterArgs{
		name:         xdsIR.ProxyServiceCluster.Name,
		settings:     []*ir.DestinationSetting{xdsIR.ProxyServiceCluster.Destination},
		endpointType: EndpointTypeStatic,
	})
}
