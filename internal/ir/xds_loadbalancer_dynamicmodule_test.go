// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"testing"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestLoadBalancerValidate_DynamicModuleLB(t *testing.T) {
	lb := LoadBalancer{
		DynamicModuleLB: &DynamicModuleLB{
			Name:       "my-module",
			PolicyName: "my-lb-policy",
		},
	}
	require.NoError(t, lb.Validate())
}

func TestLoadBalancerValidate_DynamicModuleLB_WithOther(t *testing.T) {
	lb := LoadBalancer{
		DynamicModuleLB: &DynamicModuleLB{
			Name:       "my-module",
			PolicyName: "my-lb-policy",
		},
		Random: &Random{},
	}
	require.EqualError(t, lb.Validate(), ErrLoadBalancerInvalid.Error())
}

func TestDynamicModuleLB_WithConfig(t *testing.T) {
	lb := LoadBalancer{
		DynamicModuleLB: &DynamicModuleLB{
			Name:       "my-module",
			PolicyName: "my-lb-policy",
			Config:     &apiextensionsv1.JSON{Raw: []byte(`{"key":"value"}`)},
		},
	}
	require.NoError(t, lb.Validate())
}

func TestDynamicModuleLB_WithRemote(t *testing.T) {
	lb := LoadBalancer{
		DynamicModuleLB: &DynamicModuleLB{
			Name:       "my-module",
			PolicyName: "my-lb-policy",
			Remote: &RemoteDynamicModuleSource{
				URL:    "https://example.com/module.so",
				SHA256: "abc123",
			},
		},
	}
	require.NoError(t, lb.Validate())
}
