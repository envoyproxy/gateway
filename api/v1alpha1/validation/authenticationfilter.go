// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"errors"
	"fmt"
	"net/url"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"

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

	for i := range spec.JwtProviders {
		provider := spec.JwtProviders[i]
		if err := ValidateJwtProvider(&provider); err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}

// ValidateJwtProvider validates the provided JWT authentication filter provider.
func ValidateJwtProvider(jwt *egv1a1.JwtAuthenticationFilterProvider) error {
	var errs []error

	switch {
	case len(jwt.Name) == 0:
		errs = append(errs, errors.New("name must be set for jwt provider"))
	case len(jwt.Issuer) != 0:
		if _, err := url.ParseRequestURI(jwt.Issuer); err != nil {
			errs = append(errs, fmt.Errorf("invalid issuer URI: %v", err))
		}
	case len(jwt.RemoteJWKS.URI) == 0:
		errs = append(errs, fmt.Errorf("uri must be set for remote JWKS provider: %s", jwt.Name))
	}
	if _, err := url.ParseRequestURI(jwt.RemoteJWKS.URI); err != nil {
		errs = append(errs, fmt.Errorf("invalid remote JWKS URI: %v", err))
	}

	return utilerrors.NewAggregate(errs)
}
