// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"sync"

	kvalidate "sigs.k8s.io/kubectl-validate/pkg/cmd"
	"sigs.k8s.io/kubectl-validate/pkg/openapiclient"
	"sigs.k8s.io/kubectl-validate/pkg/validator"
)

var (
	defaultValidator     *Validator
	defaultValidatorOnce sync.Once
)

func GetDefaultValidator() *Validator {
	defaultValidatorOnce.Do(func() {
		defaultValidator = newDefaultValidator()
	})
	return defaultValidator
}

// Validator is a local/offline Kubernetes resources validator.
type Validator struct {
	resolver *validator.Validator
}

// newDefaultValidator init a default validator for internal usage.
func newDefaultValidator() *Validator {
	factory, _ := validator.New(openapiclient.NewLocalCRDFiles(gatewayCRDsFS))
	return &Validator{resolver: factory}
}

// Validate validates one Kubernetes resource.
func (v Validator) Validate(content []byte) error {
	return kvalidate.ValidateDocument(content, v.resolver)
}
