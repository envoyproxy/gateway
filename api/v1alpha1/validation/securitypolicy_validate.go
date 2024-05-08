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

// ValidateSecurityPolicy validates the provided SecurityPolicy.
func ValidateSecurityPolicy(policy *egv1a1.SecurityPolicy) error {
	var errs []error
	if policy == nil {
		return errors.New("policy is nil")
	}
	if err := validateSecurityPolicySpec(&policy.Spec); err != nil {
		errs = append(errs, errors.New("policy is nil"))
	}

	return utilerrors.NewAggregate(errs)
}

// validateSecurityPolicySpec validates the provided spec.
func validateSecurityPolicySpec(spec *egv1a1.SecurityPolicySpec) error {
	var errs []error

	sum := 0
	switch {
	case spec == nil:
		errs = append(errs, errors.New("spec is nil"))
	case spec.CORS != nil:
		sum++
	case spec.JWT != nil:
		sum++
	}
	if sum == 0 {
		errs = append(errs, errors.New("no security policy is specified"))
	}

	// Return early if any errors exist.
	if len(errs) != 0 {
		return utilerrors.NewAggregate(errs)
	}

	if err := ValidateJWTProvider(spec.JWT.Providers); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

// ValidateJWTProvider validates the provided JWT authentication configuration.
func ValidateJWTProvider(providers []egv1a1.JWTProvider) error {
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
					errs = append(errs, fmt.Errorf("invalid issuer; must be a URL or email address: %w", err))
				}
			}
		case len(provider.RemoteJWKS.URI) == 0:
			errs = append(errs, fmt.Errorf("uri must be set for remote JWKS provider: %s", provider.Name))
		}
		if _, err := url.ParseRequestURI(provider.RemoteJWKS.URI); err != nil {
			errs = append(errs, fmt.Errorf("invalid remote JWKS URI: %w", err))
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

		for _, claimToHeader := range provider.ClaimToHeaders {
			switch {
			case len(claimToHeader.Header) == 0:
				errs = append(errs, fmt.Errorf("header must be set for claimToHeader provider: %s", claimToHeader.Header))
			case len(claimToHeader.Claim) == 0:
				errs = append(errs, fmt.Errorf("claim must be set for claimToHeader provider: %s", claimToHeader.Claim))
			}
		}
	}

	return utilerrors.NewAggregate(errs)
}
