// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
)

// OfflineGatewayAPIReconciler can be used for non-kuberetes provider.
// It can let other providers to have the same reconcile logic without rely on apiserver.
type OfflineGatewayAPIReconciler struct {
	gatewayAPIReconciler

	Client client.Client
}

func NewOfflineGatewayAPIController(
	ctx context.Context, cfg *config.Server, su Updater, resources *message.ProviderResources,
) (*OfflineGatewayAPIReconciler, error) {
	if cfg == nil || resources == nil {
		return nil, fmt.Errorf("missing config or resources that offline controller requires")
	}

	// Check provider type.
	if cfg.EnvoyGateway.Provider.Type == egv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("offline controller cannot work with kubernetes provider")
	}

	// Gather additional resources to watch from registered extensions.
	var (
		extGVKs               []schema.GroupVersionKind
		extServerPoliciesGVKs []schema.GroupVersionKind
	)

	if cfg.EnvoyGateway.ExtensionManager != nil {
		for _, rsrc := range cfg.EnvoyGateway.ExtensionManager.Resources {
			gvk := schema.GroupVersionKind(rsrc)
			extGVKs = append(extGVKs, gvk)
		}
		for _, rsrc := range cfg.EnvoyGateway.ExtensionManager.PolicyResources {
			gvk := schema.GroupVersionKind(rsrc)
			extServerPoliciesGVKs = append(extServerPoliciesGVKs, gvk)
		}
	}

	cli := newOfflineGatewayAPIClient()
	r := gatewayAPIReconciler{
		client:            cli,
		log:               cfg.Logger,
		classController:   gwapiv1.GatewayController(cfg.EnvoyGateway.Gateway.ControllerName),
		namespace:         cfg.ControllerNamespace,
		statusUpdater:     su,
		resources:         resources,
		extGVKs:           extGVKs,
		store:             newProviderStore(),
		envoyGateway:      cfg.EnvoyGateway,
		mergeGateways:     sets.New[string](),
		extServerPolicies: extServerPoliciesGVKs,
		// We assume all CRDs are available in offline mode.
		bTLSPolicyCRDExists:    true,
		btpCRDExists:           true,
		ctpCRDExists:           true,
		eepCRDExists:           true,
		epCRDExists:            true,
		eppCRDExists:           true,
		hrfCRDExists:           true,
		grpcRouteCRDExists:     true,
		serviceImportCRDExists: true,
		spCRDExists:            true,
		tcpRouteCRDExists:      true,
		tlsRouteCRDExists:      true,
		udpRouteCRDExists:      true,
		backendCRDExists:       true,
	}

	r.log.Info("created offline gatewayapi controller")
	if su != nil {
		r.subscribeAndUpdateStatus(ctx, cfg.EnvoyGateway.ExtensionManager != nil)
	}

	return &OfflineGatewayAPIReconciler{
		gatewayAPIReconciler: r,
		Client:               cli,
	}, nil
}

// Reconcile calls reconcile method in gateway-api controller, this method
// should be called manually.
func (r *OfflineGatewayAPIReconciler) Reconcile(ctx context.Context) error {
	_, err := r.gatewayAPIReconciler.Reconcile(ctx, reconcile.Request{})
	return err
}

// newOfflineGatewayAPIClient returns a offline client with gateway-api schemas and indexes.
func newOfflineGatewayAPIClient() client.Client {
	return fake.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithIndex(&gwapiv1.Gateway{}, classGatewayIndex, gatewayIndexFunc).
		WithIndex(&gwapiv1.Gateway{}, secretGatewayIndex, secretGatewayIndexFunc).
		WithIndex(&gwapiv1.HTTPRoute{}, gatewayHTTPRouteIndex, gatewayHTTPRouteIndexFunc).
		WithIndex(&gwapiv1.HTTPRoute{}, backendHTTPRouteIndex, backendHTTPRouteIndexFunc).
		WithIndex(&gwapiv1.HTTPRoute{}, httpRouteFilterHTTPRouteIndex, httpRouteFilterHTTPRouteIndexFunc).
		WithIndex(&gwapiv1.GRPCRoute{}, gatewayGRPCRouteIndex, gatewayGRPCRouteIndexFunc).
		WithIndex(&gwapiv1.GRPCRoute{}, backendGRPCRouteIndex, backendGRPCRouteIndexFunc).
		WithIndex(&gwapiv1a2.TCPRoute{}, gatewayTCPRouteIndex, gatewayTCPRouteIndexFunc).
		WithIndex(&gwapiv1a2.TCPRoute{}, backendTCPRouteIndex, backendTCPRouteIndexFunc).
		WithIndex(&gwapiv1a2.UDPRoute{}, gatewayUDPRouteIndex, gatewayUDPRouteIndexFunc).
		WithIndex(&gwapiv1a2.UDPRoute{}, backendUDPRouteIndex, backendUDPRouteIndexFunc).
		WithIndex(&gwapiv1a2.TLSRoute{}, gatewayTLSRouteIndex, gatewayTLSRouteIndexFunc).
		WithIndex(&gwapiv1a2.TLSRoute{}, backendTLSRouteIndex, backendTLSRouteIndexFunc).
		WithIndex(&egv1a1.EnvoyProxy{}, backendEnvoyProxyTelemetryIndex, backendEnvoyProxyTelemetryIndexFunc).
		WithIndex(&egv1a1.EnvoyProxy{}, secretEnvoyProxyIndex, secretEnvoyProxyIndexFunc).
		WithIndex(&egv1a1.BackendTrafficPolicy{}, configMapBtpIndex, configMapBtpIndexFunc).
		WithIndex(&egv1a1.ClientTrafficPolicy{}, configMapCtpIndex, configMapCtpIndexFunc).
		WithIndex(&egv1a1.ClientTrafficPolicy{}, secretCtpIndex, secretCtpIndexFunc).
		WithIndex(&egv1a1.SecurityPolicy{}, secretSecurityPolicyIndex, secretSecurityPolicyIndexFunc).
		WithIndex(&egv1a1.SecurityPolicy{}, backendSecurityPolicyIndex, backendSecurityPolicyIndexFunc).
		WithIndex(&egv1a1.SecurityPolicy{}, configMapSecurityPolicyIndex, configMapSecurityPolicyIndexFunc).
		WithIndex(&egv1a1.EnvoyExtensionPolicy{}, backendEnvoyExtensionPolicyIndex, backendEnvoyExtensionPolicyIndexFunc).
		WithIndex(&egv1a1.EnvoyExtensionPolicy{}, secretEnvoyExtensionPolicyIndex, secretEnvoyExtensionPolicyIndexFunc).
		WithIndex(&egv1a1.EnvoyExtensionPolicy{}, configMapEepIndex, configMapEepIndexFunc).
		WithIndex(&gwapiv1a3.BackendTLSPolicy{}, configMapBtlsIndex, configMapBtlsIndexFunc).
		WithIndex(&gwapiv1a3.BackendTLSPolicy{}, secretBtlsIndex, secretBtlsIndexFunc).
		WithIndex(&egv1a1.HTTPRouteFilter{}, configMapHTTPRouteFilterIndex, configMapRouteFilterIndexFunc).
		WithIndex(&egv1a1.HTTPRouteFilter{}, secretHTTPRouteFilterIndex, secretRouteFilterIndexFunc).
		WithIndex(&gwapiv1b1.ReferenceGrant{}, targetRefGrantRouteIndex, getReferenceGrantIndexerFunc).
		Build()
}
