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
	oauth2credv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/injected_credentials/oauth2/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

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
	credential, err := buildInjectedCredential(credentialInjection)
	if err != nil {
		return nil, err
	}

	credentialInjector := &injectorv3.CredentialInjector{
		Credential: credential,
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

// buildInjectedCredential returns the credential extension of the credentialInjector filter:
// the generic credential read from an SDS secret, or the OAuth2 client credentials flow.
func buildInjectedCredential(credentialInjection *ir.CredentialInjection) (*corev3.TypedExtensionConfig, error) {
	if credentialInjection.OAuth2 != nil {
		return buildOAuth2InjectedCredential(credentialInjection)
	}

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

	return &corev3.TypedExtensionConfig{
		Name:        "envoy.http.injected_credentials.generic",
		TypedConfig: genericCredentialAny,
	}, nil
}

func buildOAuth2InjectedCredential(credentialInjection *ir.CredentialInjection) (*corev3.TypedExtensionConfig, error) {
	oauth2 := credentialInjection.OAuth2

	var tokenEndpointCluster string
	if oauth2.Destination != nil && len(oauth2.Destination.Settings) > 0 {
		tokenEndpointCluster = oauth2.Destination.Name
	} else {
		cluster, err := url2Cluster(oauth2.TokenEndpoint)
		if err != nil {
			return nil, err
		}
		if cluster.endpointType == EndpointTypeStatic {
			return nil, fmt.Errorf("static IP cluster is not allowed: %s", oauth2.TokenEndpoint)
		}
		tokenEndpointCluster = cluster.name
	}

	tokenEndpointTimeout := durationpb.New(defaultExtServiceRequestTimeout)
	if oauth2.Traffic != nil &&
		oauth2.Traffic.Timeout != nil &&
		oauth2.Traffic.Timeout.HTTP != nil &&
		oauth2.Traffic.Timeout.HTTP.RequestTimeout != nil {
		tokenEndpointTimeout = durationpb.New(oauth2.Traffic.Timeout.HTTP.RequestTimeout.Duration)
	}

	oauth2Credential := &oauth2credv3.OAuth2{
		TokenEndpoint: &corev3.HttpUri{
			Uri: oauth2.TokenEndpoint,
			HttpUpstreamType: &corev3.HttpUri_Cluster{
				Cluster: tokenEndpointCluster,
			},
			Timeout: tokenEndpointTimeout,
		},
		Scopes: oauth2.Scopes,
		FlowType: &oauth2credv3.OAuth2_ClientCredentials_{
			ClientCredentials: &oauth2credv3.OAuth2_ClientCredentials{
				ClientId: oauth2.ClientID,
				ClientSecret: &tlsv3.SdsSecretConfig{
					Name:      credentialOAuth2ClientSecretName(credentialInjection),
					SdsConfig: makeConfigSource(),
				},
			},
		},
	}
	oauth2CredentialAny, err := proto.ToAnyWithValidation(oauth2Credential)
	if err != nil {
		return nil, err
	}

	return &corev3.TypedExtensionConfig{
		Name:        "envoy.http.injected_credentials.oauth2",
		TypedConfig: oauth2CredentialAny,
	}, nil
}

func credentialOAuth2ClientSecretName(credentialInjection *ir.CredentialInjection) string {
	return fmt.Sprintf("credential_injector/oauth2_client_secret/%s", credentialInjection.Name)
}

func credentialSecretName(credentialInjection *ir.CredentialInjection) string {
	return fmt.Sprintf("credential_injector/credential/%s", credentialInjection.Name)
}

func (*credentialInjector) patchResources(resource *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	var errs error

	for _, route := range routes {
		if route.CredentialInjection == nil {
			continue
		}

		secret := buildCredentialInjectionSecret(route.CredentialInjection)
		if err := addXdsSecret(resource, secret); err != nil {
			errs = errors.Join(errs, err)
		}
		if err := addOAuth2TokenEndpointCluster(resource, route.CredentialInjection); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// buildCredentialInjectionSecret returns the SDS secret read by the credential injector filter:
// the OAuth2 client secret, or the credential to be injected.
func buildCredentialInjectionSecret(credentialInjection *ir.CredentialInjection) *tlsv3.Secret {
	if credentialInjection.OAuth2 != nil {
		return buildGenericSecret(credentialOAuth2ClientSecretName(credentialInjection), credentialInjection.OAuth2.ClientSecret)
	}
	return buildGenericSecret(credentialSecretName(credentialInjection), credentialInjection.Credential)
}

// addOAuth2TokenEndpointCluster adds the cluster for the OAuth2 token endpoint of the credential
// injector filter, if the filter injects an OAuth2 access token.
func addOAuth2TokenEndpointCluster(tCtx *types.ResourceVersionTable, credentialInjection *ir.CredentialInjection) error {
	oauth2 := credentialInjection.OAuth2
	if oauth2 == nil {
		return nil
	}

	if oauth2.Destination != nil && len(oauth2.Destination.Settings) > 0 {
		if err := createExtServiceXDSCluster(oauth2.Destination, oauth2.Traffic, tCtx); err != nil {
			return err
		}
		return processClientCertificates(tCtx, oauth2.Destination.Settings)
	}
	return addTokenEndpointCluster(tCtx, oauth2.TokenEndpoint, oauth2.Traffic)
}

func buildGenericSecret(name string, value []byte) *tlsv3.Secret {
	return &tlsv3.Secret{
		Name: name,
		Type: &tlsv3.Secret_GenericSecret{
			GenericSecret: &tlsv3.GenericSecret{
				Secret: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{
						InlineBytes: value,
					},
				},
			},
		},
	}
}

// patchRoute patches the provided route with the credential injector filter if applicable.
// Note: this method enables the corresponding credential injector filter for the provided route.
func (*credentialInjector) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
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
	if err := enableFilterOnRoute(route, filterName, &routev3.FilterConfig{
		Config: &anypb.Any{},
	}); err != nil {
		return err
	}
	return nil
}
