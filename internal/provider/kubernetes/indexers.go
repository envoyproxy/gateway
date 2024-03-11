// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

const (
	classGatewayIndex          = "classGatewayIndex"
	gatewayTLSRouteIndex       = "gatewayTLSRouteIndex"
	gatewayHTTPRouteIndex      = "gatewayHTTPRouteIndex"
	gatewayGRPCRouteIndex      = "gatewayGRPCRouteIndex"
	gatewayTCPRouteIndex       = "gatewayTCPRouteIndex"
	gatewayUDPRouteIndex       = "gatewayUDPRouteIndex"
	secretGatewayIndex         = "secretGatewayIndex"
	targetRefGrantRouteIndex   = "targetRefGrantRouteIndex"
	backendHTTPRouteIndex      = "backendHTTPRouteIndex"
	backendGRPCRouteIndex      = "backendGRPCRouteIndex"
	backendTLSRouteIndex       = "backendTLSRouteIndex"
	backendTCPRouteIndex       = "backendTCPRouteIndex"
	backendUDPRouteIndex       = "backendUDPRouteIndex"
	secretSecurityPolicyIndex  = "secretSecurityPolicyIndex"
	backendSecurityPolicyIndex = "backendSecurityPolicyIndex"
	configMapCtpIndex          = "configMapCtpIndex"
	secretCtpIndex             = "secretCtpIndex"
	configMapBtlsIndex         = "configMapBtlsIndex"
)

func addReferenceGrantIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1b1.ReferenceGrant{}, targetRefGrantRouteIndex, func(rawObj client.Object) []string {
		refGrant := rawObj.(*gwapiv1b1.ReferenceGrant)
		var referredServices []string
		for _, target := range refGrant.Spec.To {
			referredServices = append(referredServices, string(target.Kind))
		}
		return referredServices
	}); err != nil {
		return err
	}
	return nil
}

// addHTTPRouteIndexers adds indexing on HTTPRoute.
//   - For Service, ServiceImports objects that are referenced in HTTPRoute objects via `.spec.rules.backendRefs`.
//     This helps in querying for HTTPRoutes that are affected by a particular Service CRUD.
func addHTTPRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1.HTTPRoute{}, gatewayHTTPRouteIndex, gatewayHTTPRouteIndexFunc); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1.HTTPRoute{}, backendHTTPRouteIndex, backendHTTPRouteIndexFunc); err != nil {
		return err
	}

	return nil
}

func gatewayHTTPRouteIndexFunc(rawObj client.Object) []string {
	httproute := rawObj.(*gwapiv1.HTTPRoute)
	var gateways []string
	for _, parent := range httproute.Spec.ParentRefs {
		if parent.Kind == nil || string(*parent.Kind) == gatewayapi.KindGateway {
			// If an explicit Gateway namespace is not provided, use the HTTPRoute namespace to
			// lookup the provided Gateway Name.
			gateways = append(gateways,
				types.NamespacedName{
					Namespace: gatewayapi.NamespaceDerefOr(parent.Namespace, httproute.Namespace),
					Name:      string(parent.Name),
				}.String(),
			)
		}
	}
	return gateways
}

func backendHTTPRouteIndexFunc(rawObj client.Object) []string {
	httproute := rawObj.(*gwapiv1.HTTPRoute)
	var backendRefs []string
	for _, rule := range httproute.Spec.Rules {
		for _, backend := range rule.BackendRefs {
			if backend.Kind == nil || string(*backend.Kind) == gatewayapi.KindService {
				// If an explicit Backend namespace is not provided, use the HTTPRoute namespace to
				// lookup the provided Gateway Name.
				backendRefs = append(backendRefs,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOr(backend.Namespace, httproute.Namespace),
						Name:      string(backend.Name),
					}.String(),
				)
			}
		}
	}
	return backendRefs
}

// addGRPCRouteIndexers adds indexing on GRPCRoute, for Service objects that are
// referenced in GRPCRoute objects via `.spec.rules.backendRefs`. This helps in
// querying for GRPCRoutes that are affected by a particular Service CRUD.
func addGRPCRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.GRPCRoute{}, gatewayGRPCRouteIndex, gatewayGRPCRouteIndexFunc); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.GRPCRoute{}, backendGRPCRouteIndex, backendGRPCRouteIndexFunc); err != nil {
		return err
	}

	return nil
}

func gatewayGRPCRouteIndexFunc(rawObj client.Object) []string {
	grpcroute := rawObj.(*gwapiv1a2.GRPCRoute)
	var gateways []string
	for _, parent := range grpcroute.Spec.ParentRefs {
		if parent.Kind == nil || string(*parent.Kind) == gatewayapi.KindGateway {
			// If an explicit Gateway namespace is not provided, use the GRPCRoute namespace to
			// lookup the provided Gateway Name.
			gateways = append(gateways,
				types.NamespacedName{
					Namespace: gatewayapi.NamespaceDerefOr(parent.Namespace, grpcroute.Namespace),
					Name:      string(parent.Name),
				}.String(),
			)
		}
	}
	return gateways
}

func backendGRPCRouteIndexFunc(rawObj client.Object) []string {
	grpcroute := rawObj.(*gwapiv1a2.GRPCRoute)
	var backendRefs []string
	for _, rule := range grpcroute.Spec.Rules {
		for _, backend := range rule.BackendRefs {
			if backend.Kind == nil || string(*backend.Kind) == gatewayapi.KindService {
				// If an explicit Backend namespace is not provided, use the GRPCRoute namespace to
				// lookup the provided Gateway Name.
				backendRefs = append(backendRefs,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOr(backend.Namespace, grpcroute.Namespace),
						Name:      string(backend.Name),
					}.String(),
				)
			}
		}
	}
	return backendRefs
}

// addTLSRouteIndexers adds indexing on TLSRoute, for Service objects that are
// referenced in TLSRoute objects via `.spec.rules.backendRefs`. This helps in
// querying for TLSRoutes that are affected by a particular Service CRUD.
func addTLSRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.TLSRoute{}, gatewayTLSRouteIndex, func(rawObj client.Object) []string {
		tlsRoute := rawObj.(*gwapiv1a2.TLSRoute)
		var gateways []string
		for _, parent := range tlsRoute.Spec.ParentRefs {
			if string(*parent.Kind) == gatewayapi.KindGateway {
				// If an explicit Gateway namespace is not provided, use the TLSRoute namespace to
				// lookup the provided Gateway Name.
				gateways = append(gateways,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOrAlpha(parent.Namespace, tlsRoute.Namespace),
						Name:      string(parent.Name),
					}.String(),
				)
			}
		}
		return gateways
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.TLSRoute{}, backendTLSRouteIndex, backendTLSRouteIndexFunc); err != nil {
		return err
	}
	return nil
}

func backendTLSRouteIndexFunc(rawObj client.Object) []string {
	tlsroute := rawObj.(*gwapiv1a2.TLSRoute)
	var backendRefs []string
	for _, rule := range tlsroute.Spec.Rules {
		for _, backend := range rule.BackendRefs {
			if backend.Kind == nil || string(*backend.Kind) == gatewayapi.KindService {
				// If an explicit Backend namespace is not provided, use the TLSRoute namespace to
				// lookup the provided Gateway Name.
				backendRefs = append(backendRefs,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOrAlpha(backend.Namespace, tlsroute.Namespace),
						Name:      string(backend.Name),
					}.String(),
				)
			}
		}
	}
	return backendRefs
}

// addTCPRouteIndexers adds indexing on TCPRoute, for Service objects that are
// referenced in TCPRoute objects via `.spec.rules.backendRefs`. This helps in
// querying for TCPRoutes that are affected by a particular Service CRUD.
func addTCPRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.TCPRoute{}, gatewayTCPRouteIndex, func(rawObj client.Object) []string {
		tcpRoute := rawObj.(*gwapiv1a2.TCPRoute)
		var gateways []string
		for _, parent := range tcpRoute.Spec.ParentRefs {
			if string(*parent.Kind) == gatewayapi.KindGateway {
				// If an explicit Gateway namespace is not provided, use the TCPRoute namespace to
				// lookup the provided Gateway Name.
				gateways = append(gateways,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOrAlpha(parent.Namespace, tcpRoute.Namespace),
						Name:      string(parent.Name),
					}.String(),
				)
			}
		}
		return gateways
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.TCPRoute{}, backendTCPRouteIndex, backendTCPRouteIndexFunc); err != nil {
		return err
	}
	return nil
}

func backendTCPRouteIndexFunc(rawObj client.Object) []string {
	tcpRoute := rawObj.(*gwapiv1a2.TCPRoute)
	var backendRefs []string
	for _, rule := range tcpRoute.Spec.Rules {
		for _, backend := range rule.BackendRefs {
			if backend.Kind == nil || string(*backend.Kind) == gatewayapi.KindService {
				// If an explicit Backend namespace is not provided, use the TCPRoute namespace to
				// lookup the provided Gateway Name.
				backendRefs = append(backendRefs,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOrAlpha(backend.Namespace, tcpRoute.Namespace),
						Name:      string(backend.Name),
					}.String(),
				)
			}
		}
	}
	return backendRefs
}

// addUDPRouteIndexers adds indexing on UDPRoute.
//   - For Gateway objects that are referenced in UDPRoute objects via `.spec.parentRefs`. This helps in
//     querying for UDPRoutes that are affected by a particular Gateway CRUD.
//   - For Service objects that are referenced in UDPRoute objects via `.spec.rules.backendRefs`. This helps in
//     querying for UDPRoutes that are affected by a particular Service CRUD.
func addUDPRouteIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.UDPRoute{}, gatewayUDPRouteIndex, func(rawObj client.Object) []string {
		udpRoute := rawObj.(*gwapiv1a2.UDPRoute)
		var gateways []string
		for _, parent := range udpRoute.Spec.ParentRefs {
			if string(*parent.Kind) == gatewayapi.KindGateway {
				// If an explicit Gateway namespace is not provided, use the UDPRoute namespace to
				// lookup the provided Gateway Name.
				gateways = append(gateways,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOrAlpha(parent.Namespace, udpRoute.Namespace),
						Name:      string(parent.Name),
					}.String(),
				)
			}
		}
		return gateways
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.UDPRoute{}, backendUDPRouteIndex, backendUDPRouteIndexFunc); err != nil {
		return err
	}
	return nil
}

func backendUDPRouteIndexFunc(rawObj client.Object) []string {
	udproute := rawObj.(*gwapiv1a2.UDPRoute)
	var backendRefs []string
	for _, rule := range udproute.Spec.Rules {
		for _, backend := range rule.BackendRefs {
			if backend.Kind == nil || string(*backend.Kind) == gatewayapi.KindService {
				// If an explicit Backend namespace is not provided, use the UDPRoute namespace to
				// lookup the provided Gateway Name.
				backendRefs = append(backendRefs,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOrAlpha(backend.Namespace, udproute.Namespace),
						Name:      string(backend.Name),
					}.String(),
				)
			}
		}
	}
	return backendRefs
}

// addGatewayIndexers adds indexing on Gateway, for Secret objects that are
// referenced in Gateway objects. This helps in querying for Gateways that are
// affected by a particular Secret CRUD.
func addGatewayIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1.Gateway{}, secretGatewayIndex, secretGatewayIndexFunc); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1.Gateway{}, classGatewayIndex, func(rawObj client.Object) []string {
		gateway := rawObj.(*gwapiv1.Gateway)
		return []string{string(gateway.Spec.GatewayClassName)}
	}); err != nil {
		return err
	}
	return nil
}

func secretGatewayIndexFunc(rawObj client.Object) []string {
	gateway := rawObj.(*gwapiv1.Gateway)
	var secretReferences []string
	for _, listener := range gateway.Spec.Listeners {
		if listener.TLS == nil || *listener.TLS.Mode != gwapiv1.TLSModeTerminate {
			continue
		}
		for _, cert := range listener.TLS.CertificateRefs {
			if *cert.Kind == gatewayapi.KindSecret {
				// If an explicit Secret namespace is not provided, use the Gateway namespace to
				// lookup the provided Secret Name.
				secretReferences = append(secretReferences,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOr(cert.Namespace, gateway.Namespace),
						Name:      string(cert.Name),
					}.String(),
				)
			}
		}
	}
	return secretReferences
}

// addSecurityPolicyIndexers adds indexing on SecurityPolicy.
//   - For Secret objects that are referenced in SecurityPolicy objects via
//     `.spec.OIDC.clientSecret` and `.spec.basicAuth.users`. This helps in
//     querying for SecurityPolicies that are affected by a particular Secret CRUD.
//   - For Service objects that are referenced in SecurityPolicy objects via
//     `.spec.extAuth.http.backendObjectReference`. This helps in querying for
//     SecurityPolicies that are affected by a particular Service CRUD.
func addSecurityPolicyIndexers(ctx context.Context, mgr manager.Manager) error {
	var err error

	if err = mgr.GetFieldIndexer().IndexField(
		ctx, &v1alpha1.SecurityPolicy{}, secretSecurityPolicyIndex,
		secretSecurityPolicyIndexFunc); err != nil {
		return err
	}

	if err = mgr.GetFieldIndexer().IndexField(
		ctx, &v1alpha1.SecurityPolicy{}, backendSecurityPolicyIndex,
		backendSecurityPolicyIndexFunc); err != nil {
		return err
	}

	return nil
}

func secretSecurityPolicyIndexFunc(rawObj client.Object) []string {
	securityPolicy := rawObj.(*v1alpha1.SecurityPolicy)

	var (
		secretReferences []gwapiv1b1.SecretObjectReference
		values           []string
	)

	if securityPolicy.Spec.OIDC != nil {
		secretReferences = append(secretReferences, securityPolicy.Spec.OIDC.ClientSecret)
	}
	if securityPolicy.Spec.BasicAuth != nil {
		secretReferences = append(secretReferences, securityPolicy.Spec.BasicAuth.Users)
	}

	for _, reference := range secretReferences {
		values = append(values,
			types.NamespacedName{
				Namespace: gatewayapi.NamespaceDerefOr(reference.Namespace, securityPolicy.Namespace),
				Name:      string(reference.Name),
			}.String(),
		)
	}
	return values
}

func backendSecurityPolicyIndexFunc(rawObj client.Object) []string {
	securityPolicy := rawObj.(*v1alpha1.SecurityPolicy)

	var backendRef *gwapiv1.BackendObjectReference

	if securityPolicy.Spec.ExtAuth != nil {
		if securityPolicy.Spec.ExtAuth.HTTP != nil {
			backendRef = &securityPolicy.Spec.ExtAuth.HTTP.BackendRef
		} else if securityPolicy.Spec.ExtAuth.GRPC != nil {
			backendRef = &securityPolicy.Spec.ExtAuth.GRPC.BackendRef
		}
	}

	if backendRef != nil {
		return []string{
			types.NamespacedName{
				Namespace: gatewayapi.NamespaceDerefOr(backendRef.Namespace, securityPolicy.Namespace),
				Name:      string(backendRef.Name),
			}.String(),
		}
	}

	// This should not happen because the CEL validation should catch it.
	return []string{}
}

// addCtpIndexers adds indexing on ClientTrafficPolicy, for ConfigMap or Secret objects that are
// referenced in ClientTrafficPolicy objects. This helps in querying for ClientTrafficPolicies that are
// affected by a particular ConfigMap or Secret CRUD.
func addCtpIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &v1alpha1.ClientTrafficPolicy{}, configMapCtpIndex, configMapCtpIndexFunc); err != nil {
		return err
	}
	if err := mgr.GetFieldIndexer().IndexField(ctx, &v1alpha1.ClientTrafficPolicy{}, secretCtpIndex, secretCtpIndexFunc); err != nil {
		return err
	}

	return nil
}

func configMapCtpIndexFunc(rawObj client.Object) []string {
	ctp := rawObj.(*v1alpha1.ClientTrafficPolicy)
	var configMapReferences []string
	if ctp.Spec.TLS != nil && ctp.Spec.TLS.ClientValidation != nil {
		for _, caCertRef := range ctp.Spec.TLS.ClientValidation.CACertificateRefs {
			if caCertRef.Kind != nil && string(*caCertRef.Kind) == gatewayapi.KindConfigMap {
				// If an explicit configmap namespace is not provided, use the ctp namespace to
				// lookup the provided config map Name.
				configMapReferences = append(configMapReferences,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOr(caCertRef.Namespace, ctp.Namespace),
						Name:      string(caCertRef.Name),
					}.String(),
				)
			}
		}
	}
	return configMapReferences
}

func secretCtpIndexFunc(rawObj client.Object) []string {
	ctp := rawObj.(*v1alpha1.ClientTrafficPolicy)
	var secretReferences []string
	if ctp.Spec.TLS != nil && ctp.Spec.TLS.ClientValidation != nil {
		for _, caCertRef := range ctp.Spec.TLS.ClientValidation.CACertificateRefs {
			if caCertRef.Kind == nil || (string(*caCertRef.Kind) == gatewayapi.KindSecret) {
				// If an explicit namespace is not provided, use the ctp namespace to
				// lookup the provided secrent Name.
				secretReferences = append(secretReferences,
					types.NamespacedName{
						Namespace: gatewayapi.NamespaceDerefOr(caCertRef.Namespace, ctp.Namespace),
						Name:      string(caCertRef.Name),
					}.String(),
				)
			}
		}
	}
	return secretReferences
}

// addBtlsIndexers adds indexing on BackendTLSPolicy, for ConfigMap objects that are
// referenced in BackendTLSPolicy objects. This helps in querying for BackendTLSPolicies that are
// affected by a particular ConfigMap CRUD.
func addBtlsIndexers(ctx context.Context, mgr manager.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(ctx, &gwapiv1a2.BackendTLSPolicy{}, configMapBtlsIndex, configMapBtlsIndexFunc); err != nil {
		return err
	}

	return nil
}

func configMapBtlsIndexFunc(rawObj client.Object) []string {
	btls := rawObj.(*gwapiv1a2.BackendTLSPolicy)
	var configMapReferences []string
	if btls.Spec.TLS.CACertRefs != nil {
		for _, caCertRef := range btls.Spec.TLS.CACertRefs {
			if string(caCertRef.Kind) == gatewayapi.KindConfigMap {
				configMapReferences = append(configMapReferences,
					types.NamespacedName{
						Namespace: btls.Namespace,
						Name:      string(caCertRef.Name),
					}.String(),
				)
			}
		}
	}
	return configMapReferences
}
