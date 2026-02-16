// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package net

import (
	"fmt"

	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func BackendHostAndPort(backendRef gwapiv1.BackendObjectReference, defaultNamespace string) (string, uint32) {
	ns := defaultNamespace
	if backendRef.Namespace != nil {
		ns = string(*backendRef.Namespace)
	}

	if ns == "" {
		return string(backendRef.Name), uint32(*backendRef.Port)
	}

	return fmt.Sprintf("%s.%s.svc", backendRef.Name, ns), uint32(*backendRef.Port)
}
