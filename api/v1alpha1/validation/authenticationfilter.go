// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/validation"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// ValidateAuthenticationFilter validates the provided filter. The only supported
// ValidateAuthenticationFilter type is "JWT".
func ValidateAuthenticationFilter(filter *egv1a1.AuthenticationFilter) error {
	var errs []error
	if filter == nil {
		return errors.New("filter is nil")
	}
	if err := validateAuthenticationFilterSpec(&filter.Spec); err != nil {
		errs = append(errs, errors.New("filter is nil"))
	}

	return utilerrors.NewAggregate(errs)
}

// validateAuthenticationFilterSpec validates the provided spec. The only supported
// ValidateAuthenticationFilter type is "JWT".
func validateAuthenticationFilterSpec(spec *egv1a1.AuthenticationFilterSpec) error {
	var errs []error

	switch {
	case spec == nil:
		errs = append(errs, errors.New("spec is nil"))
	case spec.Type != egv1a1.JwtAuthenticationFilterProviderType:
		errs = append(errs, fmt.Errorf("unsupported authenticationfilter type: %v", spec.Type))
	case len(spec.JwtProviders) == 0:
		errs = append(errs, fmt.Errorf("at least one provider must be specified for type %v", spec.Type))
	}

	// Return early if any errors exist.
	if len(errs) != 0 {
		return utilerrors.NewAggregate(errs)
	}

	if err := ValidateJwtProviders(spec.JwtProviders); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

// ValidateJwtProviders validates the provided JWT authentication filter providers.
func ValidateJwtProviders(providers []egv1a1.JwtAuthenticationFilterProvider) error {
	var errs []error

	var names []string
	for _, provider := range providers {
		switch {
		case len(provider.Name) == 0:
			errs = append(errs, errors.New("jwt provider cannot be an empty string"))
		case len(provider.Issuer) != 0:
			// Issuer can take the format of a URL or an email address.
			if _, err := url.ParseRequestURI(provider.Issuer); err != nil {
				_, err := mail.ParseAddress(provider.Issuer)
				if err != nil {
					errs = append(errs, fmt.Errorf("invalid issuer; must be a URL or email address: %v", err))
				}
			}
		case len(provider.RemoteJWKS.URI) == 0:
			errs = append(errs, fmt.Errorf("uri must be set for remote JWKS provider: %s", provider.Name))
		}
		if _, err := url.ParseRequestURI(provider.RemoteJWKS.URI); err != nil {
			errs = append(errs, fmt.Errorf("invalid remote JWKS URI: %v", err))
		}

		if len(errs) == 0 {
			if strErrs := validation.IsQualifiedName(provider.Name); len(strErrs) != 0 {
				for _, strErr := range strErrs {
					errs = append(errs, errors.New(strErr))
				}
			}
			// Ensure uniqueness among provider names.
			if names == nil {
				names = append(names, provider.Name)
			} else {
				for _, name := range names {
					if name == provider.Name {
						errs = append(errs, fmt.Errorf("provider name %s must be unique", provider.Name))
					} else {
						names = append(names, provider.Name)
					}
				}
			}
		}
	}

	return utilerrors.NewAggregate(errs)
}
