// Copyright The Envoy Project Authors

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/pkg/log"
)

func TestHasMatchingController(t *testing.T) {
	testCases := []struct {
		name   string
		obj    client.Object
		expect bool
	}{
		{
			name: "configured controllerName",
			obj: &gwapiv1a2.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1a2.GatewayClassSpec{
					ControllerName: gwapiv1a2.GatewayController(v1alpha1.GatewayControllerName),
				},
			},
			expect: true,
		},
		{
			name: "not configured controllerName",
			obj: &gwapiv1a2.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1a2.GatewayClassSpec{
					ControllerName: gwapiv1a2.GatewayController("not.configured/controller"),
				},
			},
			expect: false,
		},
	}

	// Create the reconciler.
	logger, err := log.NewLogger()
	require.NoError(t, err)
	r := gatewayClassReconciler{
		controller: v1alpha1.GatewayControllerName,
		log:        logger,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := r.hasMatchingController(tc.obj)
			require.Equal(t, tc.expect, res)
		})
	}
}
