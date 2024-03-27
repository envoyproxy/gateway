// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package probs

import (
	"errors"
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

const StartedXdsServer = "StartedXdsServer"
const NothingToReconcile = "NothingToReconcile"
const GeneratedNewXdsSnapshot = "GeneratedNewXdsSnapshot"

type HealthProb interface {
	// GetHealthCheckerProb returns a health checker.
	GetHealthCheckerProb() healthz.Checker
	// SetIndicator sets an indicator representing a state or status used to determine the readiness of a service.
	SetIndicator(name string)
}

func NewXdsReadyHealthProb() HealthProb {
	return &xdsHealthProb{
		isReady:    false,
		indicators: make(map[string]bool, 2),
	}
}

type xdsHealthProb struct {
	isReady    bool
	indicators map[string]bool
}

// SetIndicator sets an indicator representing a state or status used to determine the readiness of a service.
func (x *xdsHealthProb) SetIndicator(indicator string) {
	// Set the indicator in the map
	x.indicators[indicator] = true

	// If a snapshot is generated, consider the controller as healthy,
	// as it has passed at least one full reconcile and xDS is in sync
	if indicator == GeneratedNewXdsSnapshot {
		x.isReady = true
		return
	}

	// If we don't have any gateways to reconcile and the xDS server is ready,
	// we should be ready to start
	if x.indicators[StartedXdsServer] && indicator == NothingToReconcile {
		x.isReady = true
		return
	}

	// If none of the above conditions are met, expect the controller
	// to complete at least one reconcile flow and become healthy after xDS generate a snapshot
	// otherwise the instance is considered not healthy
}

func (x *xdsHealthProb) GetHealthCheckerProb() healthz.Checker {
	return func(req *http.Request) error {
		if x.isReady {
			return nil
		}
		return errors.New("the xds service is not ready")
	}
}
