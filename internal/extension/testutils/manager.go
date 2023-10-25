// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package testutils

import (
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	extType "github.com/envoyproxy/gateway/internal/extension/types"
)

var _ extType.Manager = (*Manager)(nil)

type Manager struct {
	extension v1alpha1.ExtensionManager
}

func NewManager(ext v1alpha1.ExtensionManager) extType.Manager {
	return &Manager{
		extension: ext,
	}
}

func (m *Manager) HasExtension(g v1.Group, k v1.Kind) bool {
	extension := m.extension
	// TODO: not currently checking the version since extensionRef only supports group and kind.
	for _, gvk := range extension.Resources {
		if g == v1.Group(gvk.Group) && k == v1.Kind(gvk.Kind) {
			return true
		}
	}
	return false
}

func (m *Manager) GetPreXDSHookClient(xdsHookType v1alpha1.XDSTranslatorHook) extType.XDSHookClient {
	if m.extension.Hooks == nil {
		return nil
	}

	if m.extension.Hooks.XDSTranslator == nil {
		return nil
	}

	for _, hook := range m.extension.Hooks.XDSTranslator.Pre {
		if xdsHookType == hook {
			return &XDSHookClient{}
		}
	}
	return nil
}

func (m *Manager) GetPostXDSHookClient(xdsHookType v1alpha1.XDSTranslatorHook) extType.XDSHookClient {
	if m.extension.Hooks == nil {
		return nil
	}

	if m.extension.Hooks.XDSTranslator == nil {
		return nil
	}

	for _, hook := range m.extension.Hooks.XDSTranslator.Post {
		if xdsHookType == hook {
			return &XDSHookClient{}
		}
	}
	return nil
}
