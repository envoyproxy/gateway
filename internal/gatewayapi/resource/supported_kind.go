// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

const (
	KindConfigMap            = "ConfigMap"
	KindClientTrafficPolicy  = "ClientTrafficPolicy"
	KindBackendTrafficPolicy = "BackendTrafficPolicy"
	KindBackendTLSPolicy     = "BackendTLSPolicy"
	KindBackend              = "Backend"
	KindEnvoyPatchPolicy     = "EnvoyPatchPolicy"
	KindEnvoyExtensionPolicy = "EnvoyExtensionPolicy"
	KindSecurityPolicy       = "SecurityPolicy"
	KindEnvoyProxy           = "EnvoyProxy"
	KindGateway              = "Gateway"
	KindGatewayClass         = "GatewayClass"
	KindGRPCRoute            = "GRPCRoute"
	KindHTTPRoute            = "HTTPRoute"
	KindNamespace            = "Namespace"
	KindTLSRoute             = "TLSRoute"
	KindTCPRoute             = "TCPRoute"
	KindUDPRoute             = "UDPRoute"
	KindService              = "Service"
	KindServiceImport        = "ServiceImport"
	KindSecret               = "Secret"
	KindHTTPRouteFilter      = "HTTPRouteFilter"
	KindReferenceGrant       = "ReferenceGrant"
)
