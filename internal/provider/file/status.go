// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/message"
)

func (p *Provider) subscribeAndUpdateStatus(ctx context.Context) {
	// TODO: trigger gatewayclass status update in file-provider
	// GatewayClass object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "gatewayclass-status"},
			p.resourcesStore.resources.GatewayClassStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1.GatewayClassStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindGateway)
			},
		)
		p.logger.Info("gatewayClass status subscriber shutting down")
	}()

	// Gateway object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "gateway-status"},
			p.resourcesStore.resources.GatewayStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1.GatewayStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				// Update Gateway conditions, ignore addresses
				gtw := new(gwapiv1.Gateway)
				gtw.Status = *update.Value
				status.UpdateGatewayStatusAccepted(gtw)

				p.logStatus(gtw.Status, resource.KindGateway)
			},
		)
		p.logger.Info("gateway status subscriber shutting down")
	}()

	// HTTPRoute object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "httproute-status"},
			p.resourcesStore.resources.HTTPRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1.HTTPRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindHTTPRoute)
			},
		)
		p.logger.Info("httpRoute status subscriber shutting down")
	}()

	// GRPCRoute object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "grpcroute-status"},
			p.resourcesStore.resources.GRPCRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1.GRPCRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindGRPCRoute)
			},
		)
		p.logger.Info("grpcRoute status subscriber shutting down")
	}()

	// TLSRoute object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "tlsroute-status"},
			p.resourcesStore.resources.TLSRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.TLSRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindTLSRoute)
			},
		)
		p.logger.Info("tlsRoute status subscriber shutting down")
	}()

	// TCPRoute object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "tcproute-status"},
			p.resourcesStore.resources.TCPRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.TCPRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindTCPRoute)
			},
		)
		p.logger.Info("tcpRoute status subscriber shutting down")
	}()

	// UDPRoute object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "udproute-status"},
			p.resourcesStore.resources.UDPRouteStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.UDPRouteStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindUDPRoute)
			},
		)
		p.logger.Info("udpRoute status subscriber shutting down")
	}()

	// EnvoyPatchPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "envoypatchpolicy-status"},
			p.resourcesStore.resources.EnvoyPatchPolicyStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindEnvoyPatchPolicy)
			},
		)
		p.logger.Info("envoyPatchPolicy status subscriber shutting down")
	}()

	// ClientTrafficPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "clienttrafficpolicy-status"},
			p.resourcesStore.resources.ClientTrafficPolicyStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindClientTrafficPolicy)
			},
		)
		p.logger.Info("clientTrafficPolicy status subscriber shutting down")
	}()

	// BackendTrafficPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "backendtrafficpolicy-status"},
			p.resourcesStore.resources.BackendTrafficPolicyStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindBackendTrafficPolicy)
			},
		)
		p.logger.Info("backendTrafficPolicy status subscriber shutting down")
	}()

	// SecurityPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "securitypolicy-status"},
			p.resourcesStore.resources.SecurityPolicyStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindSecurityPolicy)
			},
		)
		p.logger.Info("securityPolicy status subscriber shutting down")
	}()

	// BackendTLSPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "backendtlspolicy-status"},
			p.resourcesStore.resources.BackendTLSPolicyStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindBackendTLSPolicy)
			},
		)
		p.logger.Info("backendTLSPolicy status subscriber shutting down")
	}()

	// EnvoyExtensionPolicy object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "envoyextensionpolicy-status"},
			p.resourcesStore.resources.EnvoyExtensionPolicyStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *gwapiv1a2.PolicyStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindEnvoyExtensionPolicy)
			},
		)
		p.logger.Info("envoyExtensionPolicy status subscriber shutting down")
	}()

	// Backend object status updater
	go func() {
		message.HandleSubscription(
			message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "backend-status"},
			p.resourcesStore.resources.BackendStatuses.Subscribe(ctx),
			func(update message.Update[types.NamespacedName, *egv1a1.BackendStatus], errChan chan error) {
				// skip delete updates.
				if update.Delete {
					return
				}

				p.logStatus(*update.Value, resource.KindBackend)
			},
		)
		p.logger.Info("backend status subscriber shutting down")
	}()

	if p.extensionManagerEnabled {
		// ExtensionServerPolicy object status updater
		go func() {
			message.HandleSubscription(
				message.Metadata{Runner: string(egv1a1.LogComponentProviderRunner), Message: "extensionserverpolicies-status"},
				p.resourcesStore.resources.ExtensionPolicyStatuses.Subscribe(ctx),
				func(update message.Update[message.NamespacedNameAndGVK, *gwapiv1a2.PolicyStatus], errChan chan error) {
					// skip delete updates.
					if update.Delete {
						return
					}

					p.logStatus(*update.Value, "ExtensionServerPolicy")
				},
			)
			p.logger.Info("extensionServerPolicies status subscriber shutting down")
		}()
	}
}

func (p *Provider) logStatus(obj interface{}, statusType string) {
	if status, err := yaml.Marshal(obj); err == nil {
		p.logger.Info(fmt.Sprintf("Got new status for %s \n%s", statusType, string(status)))
	} else {
		p.logger.Error(err, "failed to log status", "type", statusType)
	}
}
