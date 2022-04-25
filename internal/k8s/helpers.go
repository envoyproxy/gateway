// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapi_v1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// isStatusEqual checks that two objects of supported Kubernetes types
// have equivalent Status structs.
//
// Currently supports:
// gateway.networking.k8s.io/v1alpha2 (GatewayClass and Gateway only)
func isStatusEqual(objA, objB interface{}) bool {
	switch a := objA.(type) {
	case *gatewayapi_v1alpha2.GatewayClass:
		if b, ok := objB.(*gatewayapi_v1alpha2.GatewayClass); ok {
			if cmp.Equal(a.Status, b.Status,
				cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")) {
				return true
			}
		}
	case *gatewayapi_v1alpha2.Gateway:
		if b, ok := objB.(*gatewayapi_v1alpha2.Gateway); ok {
			if cmp.Equal(a.Status, b.Status,
				cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")) {
				return true
			}
		}
	}
	return false
}
