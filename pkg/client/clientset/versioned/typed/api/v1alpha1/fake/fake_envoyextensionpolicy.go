// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/envoyproxy/gateway/api/v1alpha1"
	apiv1alpha1 "github.com/envoyproxy/gateway/pkg/client/clientset/versioned/typed/api/v1alpha1"
	gentype "k8s.io/client-go/gentype"
)

// fakeEnvoyExtensionPolicies implements EnvoyExtensionPolicyInterface
type fakeEnvoyExtensionPolicies struct {
	*gentype.FakeClientWithList[*v1alpha1.EnvoyExtensionPolicy, *v1alpha1.EnvoyExtensionPolicyList]
	Fake *FakeEnvoyGatewayV1alpha1
}

func newFakeEnvoyExtensionPolicies(fake *FakeEnvoyGatewayV1alpha1, namespace string) apiv1alpha1.EnvoyExtensionPolicyInterface {
	return &fakeEnvoyExtensionPolicies{
		gentype.NewFakeClientWithList[*v1alpha1.EnvoyExtensionPolicy, *v1alpha1.EnvoyExtensionPolicyList](
			fake.Fake,
			namespace,
			v1alpha1.SchemeGroupVersion.WithResource("envoyextensionpolicies"),
			v1alpha1.SchemeGroupVersion.WithKind("EnvoyExtensionPolicy"),
			func() *v1alpha1.EnvoyExtensionPolicy { return &v1alpha1.EnvoyExtensionPolicy{} },
			func() *v1alpha1.EnvoyExtensionPolicyList { return &v1alpha1.EnvoyExtensionPolicyList{} },
			func(dst, src *v1alpha1.EnvoyExtensionPolicyList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.EnvoyExtensionPolicyList) []*v1alpha1.EnvoyExtensionPolicy {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1alpha1.EnvoyExtensionPolicyList, items []*v1alpha1.EnvoyExtensionPolicy) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
