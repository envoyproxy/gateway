// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kubectl-validate/pkg/openapiclient"
)

func TestNewOpenAPIClient(t *testing.T) {
	apiClient := openapiclient.NewLocalCRDFiles(gatewayCRDsFS)
	gvs, err := apiClient.Paths()
	require.NoError(t, err)

	groups := make([]string, 0, len(gvs))
	for g := range gvs {
		groups = append(groups, g)
	}
	require.ElementsMatch(t, groups, []string{
		"apis/gateway.networking.k8s.io/v1alpha2",
		"apis/gateway.networking.x-k8s.io/v1alpha1",
		"apis/gateway.envoyproxy.io/v1alpha1",
		"apis/gateway.networking.k8s.io/v1alpha3",
		"apis/gateway.networking.k8s.io/v1",
		"apis/gateway.networking.k8s.io/v1beta1",
	})
}
