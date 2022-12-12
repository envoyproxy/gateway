// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"errors"
	"fmt"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

// ValidateEnvoyProxy validates the provided EnvoyProxy.
func ValidateEnvoyProxy(ep *egcfgv1a1.EnvoyProxy) error {
	var errs []error
	if ep == nil {
		return errors.New("envoyproxy is nil")
	}
	if err := validateEnvoyProxySpec(&ep.Spec); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

// validateEnvoyProxySpec validates the provided EnvoyProxy spec.
func validateEnvoyProxySpec(spec *egcfgv1a1.EnvoyProxySpec) error {
	var errs []error

	switch {
	case spec == nil:
		errs = append(errs, errors.New("spec is nil"))
	case spec.Provider == nil:
		return utilerrors.NewAggregate(errs)
	case spec.Provider.Type != egcfgv1a1.ProviderTypeKubernetes:
		errs = append(errs, fmt.Errorf("unsupported provider type %v", spec.Provider.Type))
	}

	return utilerrors.NewAggregate(errs)
}
