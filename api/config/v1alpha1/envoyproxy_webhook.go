// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (e *EnvoyProxy) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(e).
		Complete()
}

//+kubebuilder:webhook:path=/validate-config-gateway-envoyproxy-io-v1alpha1-envoyproxy,mutating=false,failurePolicy=fail,sideEffects=None,groups=config.gateway.envoyproxy.io,resources=envoyproxies,verbs=create;update,versions=v1alpha1,name=vgateway.kb.io,admissionReviewVersions=v1alpha1

var _ webhook.Validator = &EnvoyProxy{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (e *EnvoyProxy) ValidateCreate() error {
	return e.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (e *EnvoyProxy) ValidateUpdate(old runtime.Object) error {
	return e.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (e *EnvoyProxy) ValidateDelete() error {
	return nil
}
