// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoygateway

import (
	"bytes"
	_ "embed"
)

var (
	//go:embed gatewayapi-crds.yaml
	gatewayAPICRDs []byte

	//go:embed charts/gateway-helm/crds/generated/gateway.envoyproxy.io_backends.yaml
	backendCRD []byte

	//go:embed charts/gateway-helm/crds/generated/gateway.envoyproxy.io_backendtrafficpolicies.yaml
	backendTrafficPolicyCRD []byte

	//go:embed charts/gateway-helm/crds/generated/gateway.envoyproxy.io_clienttrafficpolicies.yaml
	clientTrafficPolicyCRD []byte

	//go:embed charts/gateway-helm/crds/generated/gateway.envoyproxy.io_envoyextensionpolicies.yaml
	envoyExtensionPolicyCRD []byte

	//go:embed charts/gateway-helm/crds/generated/gateway.envoyproxy.io_envoypatchpolicies.yaml
	envoyPatchPolicyCRD []byte

	//go:embed charts/gateway-helm/crds/generated/gateway.envoyproxy.io_envoyproxies.yaml
	envoyProxyCRD []byte

	//go:embed charts/gateway-helm/crds/generated/gateway.envoyproxy.io_httproutefilters.yaml
	httpRouteFilterCRD []byte

	//go:embed charts/gateway-helm/crds/generated/gateway.envoyproxy.io_securitypolicies.yaml
	securityPolicyCRD []byte
)

var GatewayCRDs = bytes.Join([][]byte{
	gatewayAPICRDs,
	backendCRD,
	backendTrafficPolicyCRD,
	clientTrafficPolicyCRD,
	envoyExtensionPolicyCRD,
	envoyPatchPolicyCRD,
	envoyProxyCRD,
	httpRouteFilterCRD,
	securityPolicyCRD,
}, []byte(""))
