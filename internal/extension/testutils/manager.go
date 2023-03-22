// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package testutils

import (
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	extType "github.com/envoyproxy/gateway/internal/extension/types"
)

var _ extType.Manager = (*Manager)(nil)

type Manager struct {
	extension v1alpha1.Extension
}

func NewManager(ext v1alpha1.Extension) extType.Manager {
	return &Manager{
		extension: ext,
	}
}

func (m *Manager) HasExtension(g v1beta1.Group, k v1beta1.Kind) bool {
	extension := m.extension
	// TODO: not currently checking the version since extensionRef only supports group and kind.
	for _, gvk := range extension.Resources {
		if g == v1beta1.Group(gvk.Group) && k == v1beta1.Kind(gvk.Kind) {
			return true
		}
	}
	return false
}

func (m *Manager) GetXDSHookClient(xdsHookType extType.ExtensionXDSHookType) extType.XDSHookClient {
	if m.extension.Hooks == nil {
		return nil
	}

	if m.extension.Hooks.XDSTranslator == nil {
		return nil
	}

	for _, hook := range m.extension.Hooks.XDSTranslator.Post {
		if xdsHookType == extType.ExtensionXDSHookType("Post"+hook) {
			return &XDSHookClient{}
		}
	}
	for _, hook := range m.extension.Hooks.XDSTranslator.Pre {
		if xdsHookType == extType.ExtensionXDSHookType("Pre"+hook) {
			return &XDSHookClient{}
		}
	}
	return nil
}
