// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package provider

import (
	"context"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

type Provider interface {
	// Start starts the resource provider.
	Start(ctx context.Context) error

	// Type returns the type of resource provider.
	Type() egv1a1.ProviderType

	// Stop stops the resource provider.
	Stop()
}
