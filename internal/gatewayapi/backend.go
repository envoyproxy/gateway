// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

func (t *Translator) ProcessBackends(backends []*egv1a1.Backend) []*egv1a1.Backend {
	var res []*egv1a1.Backend
	for _, backend := range backends {
		backend := backend.DeepCopy()

		// Ensure Backends are enabled
		if !t.BackendEnabled {
			status.UpdateBackendStatusAcceptedCondition(backend, false)
		} else {
			status.UpdateBackendStatusAcceptedCondition(backend, true)
		}

		res = append(res, backend)
	}

	return res
}
