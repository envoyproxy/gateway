// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package types

import (
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

// Manager handles and maintains registered extensions and returns clients for
// different Hook types.
type Manager interface {
	// HasExtension checks to see whether a given Group and Kind has an
	// associated extension registered for it.
	//
	// If a Group and Kind is registered with an extension, then it should
	// return true, otherwise return false.
	HasExtension(g v1beta1.Group, k v1beta1.Kind) bool

	// GetXDSHookClient returns the XDS hook client for an extension.
	//
	// If the extension does not support this hook, then it should return
	// (nil, error)
	GetXDSHookClient(xdsHookType ExtensionXDSHookType) XDSHookClient
}
