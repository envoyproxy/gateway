// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// PrivateKeyProviderType defines the types of private key providers supported by Envoy Gateway.
//
// +kubebuilder:validation:Enum=Default;CryptoMB;QAT
type PrivateKeyProviderType string

const (
	// PrivateKeyProviderTypeDefault defines the "default" private key provider.
	PrivateKeyProviderTypeDefault PrivateKeyProviderType = "Default"
	PrivateKeyProviderTypeCryptoMB PrivateKeyProviderType = "CryptoMB"
	PrivateKeyProviderTypeQAT PrivateKeyProviderType = "QAT"
)

// DefaultEnvoyPrivateKeyProvider returns a new EnvoyPrivateKeyProvider with default configuration parameters.
func DefaultEnvoyPrivateKeyProvider() *EnvoyPrivateKeyProvider {
	return &EnvoyPrivateKeyProvider{
		Type: PrivateKeyProviderTypeDefault,
	}
}

// DefaultEnvoyDefaultPrivateKeyProvider returns a new EnvoyDefaultPrivateKeyProvider with default settings.
func DefaultEnvoyDefaultPrivateKeyProvider() *EnvoyDefaultPrivateKeyProvider {
	return &EnvoyDefaultPrivateKeyProvider{}
}