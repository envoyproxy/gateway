// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"

	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

// collectExtensionRefMatches runs once per reference, before actions can short-circuit.
func collectExtensionRefMatches(ref *gwapiv1.LocalObjectReference, context *HTTPFiltersContext, resources *resource.Resources) status.Error {
	if ref == nil || string(ref.Group) != egv1a1.GroupName || string(ref.Kind) != egv1a1.KindHTTPRouteFilter {
		return nil
	}
	for _, hrf := range resources.HTTPRouteFilters {
		if hrf.Namespace != context.Route.GetNamespace() || hrf.Name != string(ref.Name) || len(hrf.Spec.Matches) == 0 {
			continue
		}
		if len(context.Matches) > 0 {
			return status.NewRouteStatusError(
				errors.New("only one HTTPRouteFilter with matches is supported per HTTPRouteRule"),
				gwapiv1.RouteReasonUnsupportedValue,
			).WithType(gwapiv1.RouteConditionAccepted)
		}
		context.Matches = hrf.Spec.Matches
		break
	}
	return nil
}

// validateExtensionRefHTTPFilter checks references without executing actions or
// registering extension resources. Callers validate supported group/kinds first.
func (t *Translator) validateExtensionRefHTTPFilter(ref *gwapiv1.LocalObjectReference, context *HTTPFiltersContext, resources *resource.Resources) status.Error {
	if ref == nil {
		return nil
	}
	namespace := context.Route.GetNamespace()
	if string(ref.Group) == egv1a1.GroupName && string(ref.Kind) == egv1a1.KindHTTPRouteFilter {
		for _, hrf := range resources.HTTPRouteFilters {
			if hrf.Namespace == namespace && hrf.Name == string(ref.Name) {
				if hrf.Spec.DirectResponse != nil && hrf.Spec.DirectResponse.Body != nil {
					if _, err := t.getCustomResponseBody(hrf.Spec.DirectResponse.Body, namespace); err != nil {
						return t.processInvalidHTTPFilter(egv1a1.KindHTTPRouteFilter, context, err)
					}
				}
				if hrf.Spec.CredentialInjection != nil {
					return t.validateCredentialInjectionFilter(hrf, namespace, resources, context)
				}
				return nil
			}
		}
		return status.NewRouteStatusError(
			fmt.Errorf("Unable to translate HTTPRouteFilter: %s/%s", namespace, ref.Name),
			gwapiv1.RouteReasonUnsupportedValue,
		).WithType(gwapiv1.RouteConditionResolvedRefs)
	}
	res, malformed := t.getExtensionRef(resources, string(ref.Group), string(ref.Kind), namespace, string(ref.Name))
	if malformed {
		return status.NewRouteStatusError(
			fmt.Errorf("Unable to translate APIVersion for Extension Filter: kind: %s, %s/%s", ref.Kind, namespace, ref.Name),
			gwapiv1.RouteReasonUnsupportedValue,
		).WithType(gwapiv1.RouteConditionResolvedRefs)
	}
	if res != nil {
		return nil
	}
	return status.NewRouteStatusError(
		fmt.Errorf("Reference %s/%s not found for filter type: %v", namespace, ref.Name, ref.Kind),
		gwapiv1.RouteReasonUnsupportedValue,
	).WithType(gwapiv1.RouteConditionResolvedRefs)
}

func (t *Translator) validateCredentialInjectionFilter(
	hrf *egv1a1.HTTPRouteFilter,
	filterNs string,
	resources *resource.Resources,
	filterContext *HTTPFiltersContext,
) status.Error {
	secret, err := t.validateSecretRef(
		true,
		crossNamespaceFrom{
			group: egv1a1.GroupName, kind: resource.KindHTTPRouteFilter, namespace: filterNs,
		},
		hrf.Spec.CredentialInjection.Credential.ValueRef, resources)
	if err != nil {
		return t.processInvalidHTTPFilter(egv1a1.KindHTTPRouteFilter, filterContext, err)
	}
	if secretBytes, ok := secret.Data[egv1a1.InjectedCredentialKey]; !ok || len(secretBytes) == 0 {
		err := fmt.Errorf("credential key %s not found in secret %s/%s", egv1a1.InjectedCredentialKey, secret.Namespace, secret.Name)
		return t.processInvalidHTTPFilter(egv1a1.KindHTTPRouteFilter, filterContext, err)
	}
	return nil
}

// validateRequestMirrorFilter checks references without creating mirror destinations.
func (t *Translator) validateRequestMirrorFilter(mirrorFilter *gwapiv1.HTTPRequestMirrorFilter, route RouteContext, resources *resource.Resources) status.Error {
	if mirrorFilter == nil {
		return nil
	}
	backendRef := DirectBackendRef{BackendRef: &gwapiv1.BackendRef{BackendObjectReference: mirrorFilter.BackendRef}}
	backendNamespace := NamespaceDerefOr(mirrorFilter.BackendRef.Namespace, route.GetNamespace())
	if err := t.validateBackendRef(backendRef, route, resources, backendNamespace, route.GetRouteType()); err != nil {
		return status.NewRouteStatusError(
			fmt.Errorf("failed to validate the RequestMirror filter: %w", err), err.Reason()).WithType(gwapiv1.RouteConditionResolvedRefs)
	}
	return nil
}
