// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"

	"github.com/envoyproxy/gateway/internal/envoygateway"
)

func ProtobufConfig(restCfg *rest.Config) *rest.Config {
	return setRestDefaults(restCfg)
}

func setRestDefaults(config *rest.Config) *rest.Config {
	config = metadata.ConfigFor(config)
	if config.GroupVersion == nil || config.GroupVersion.Empty() {
		config.GroupVersion = &corev1.SchemeGroupVersion
	}
	if len(config.APIPath) == 0 {
		if len(config.GroupVersion.Group) == 0 {
			config.APIPath = "/api"
		} else {
			config.APIPath = "/apis"
		}
	}

	// This codec factory ensures the resources are not converted. Therefore, resources
	// will not be round-tripped through internal versions. Defaulting does not happen
	// on the client.
	config.NegotiatedSerializer = serializer.NewCodecFactory(envoygateway.GetScheme()).WithoutConversion()

	return config
}
