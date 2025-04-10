// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	injectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/credential_injector/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	genericv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/injected_credentials/generic/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&credentialInjector{})
}

type credentialInjector struct{}

var _ httpFilter = &credentialInjector{}

// patchHCM updates the HTTPConnectionManager with a Credential Injector HTTP filter for routes requiring credential injection.
func (*credentialInjector) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	var errs error

	for _, route := range irListener.Routes {
		if route.CredentialInjection == nil {
			continue
		}

		if hcmContainsFilter(mgr, credentialInjectorFilterName(route.CredentialInjection)) {
			continue
		}

		filter, err := buildHCMCredentialInjectorFilter(route.CredentialInjection)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}
	return errs
}

func credentialInjectorFilterName(credentialInjection *ir.CredentialInjection) string {
	return perRouteFilterName(egv1a1.EnvoyFilterCredentialInjector, credentialInjection.Name)
}

// buildHCMCredentialInjectorFilter returns a credentialInjector HTTP filter from the provided IR HTTPRoute.
func buildHCMCredentialInjectorFilter(credentialInjection *ir.CredentialInjection) (*hcmv3.HttpFilter, error) {
	genericCredential := &genericv3.Generic{
		Credential: &tlsv3.SdsSecretConfig{
			Name:      credentialSecretName(credentialInjection),
			SdsConfig: makeConfigSource(),
		},
	}
	if credentialInjection.Header != nil && *credentialInjection.Header != "" {
		genericCredential.Header = *credentialInjection.Header
	}
	genericCredentialAny, err := proto.ToAnyWithValidation(genericCredential)
	if err != nil {
		return nil, err
	}

	credentialInjector := &injectorv3.CredentialInjector{
		Credential: &corev3.TypedExtensionConfig{
			Name:        "envoy.http.injected_credentials.generic",
			TypedConfig: genericCredentialAny,
		},
	}
	if credentialInjection.Overwrite != nil {
		credentialInjector.Overwrite = *credentialInjection.Overwrite
	}

	credentialInjectorAny, err := proto.ToAnyWithValidation(credentialInjector)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: credentialInjectorFilterName(credentialInjection),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: credentialInjectorAny,
		},
		Disabled: true,
	}, nil
}

func credentialSecretName(credentialInjection *ir.CredentialInjection) string {
	return fmt.Sprintf("credential_injector/credential/%s", credentialInjection.Name)
}


func (*credentialInjector) patchResources(resource *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	var errs error

	for _, route := range routes {
		if route.CredentialInjection != nil {
			secret := buildCredentialSecret(route.CredentialInjection)
			if err := addXdsSecret(resource, secret); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	return errs
}

func buildCredentialSecret(credentialInjection *ir.CredentialInjection) *tlsv3.Secret {
	return &tlsv3.Secret{
		Name: credentialSecretName(credentialInjection),
		Type: &tlsv3.Secret_GenericSecret{
			GenericSecret: &tlsv3.GenericSecret{
				Secret: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{
						InlineBytes: credentialInjection.Credential,
					},
				},
			},
		},
	}
}

// patchRoute patches the provided route with the credential injector filter if applicable.
// Note: this method enables the corresponding credential injector filter for the provided route.
func (*credentialInjector) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.CredentialInjection == nil {
		return nil
	}
	filterName := credentialInjectorFilterName(irRoute.CredentialInjection)
	if err := enableFilterOnRoute(route, filterName); err != nil {
		return err
	}
	return nil
}
