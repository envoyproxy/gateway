// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoygateway

import (
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	mcsapi "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var (
	// scheme contains all the API types necessary for the provider's dynamic
	// clients to work. Any new non-core types must be added here.
	//
	// NOTE: The discovery mechanism used by the client doesn't automatically
	// refresh, so only add types here that are guaranteed to exist before the
	// provider starts.
	scheme = runtime.NewScheme()
)

func init() {
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		panic(err)
	}
	// Add Envoy Gateway types.
	if err := egv1a1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	// Add Gateway API types.
	if err := gwapiv1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := gwapiv1b1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := gwapiv1a2.AddToScheme(scheme); err != nil {
		panic(err)
	}
	// Add mcs api types.
	if err := mcsapi.AddToScheme(scheme); err != nil {
		panic(err)
	}
}

// GetScheme returns a scheme with types supported by the Kubernetes provider.
func GetScheme() *runtime.Scheme {
	return scheme
}
