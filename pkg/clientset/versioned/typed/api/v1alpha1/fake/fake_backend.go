// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/envoyproxy/gateway/api/v1alpha1"
	apiv1alpha1 "github.com/envoyproxy/gateway/pkg/clientset/versioned/typed/api/v1alpha1"
	gentype "k8s.io/client-go/gentype"
)

// fakeBackends implements BackendInterface
type fakeBackends struct {
	*gentype.FakeClientWithList[*v1alpha1.Backend, *v1alpha1.BackendList]
	Fake *FakeEnvoyGatewayV1alpha1
}

func newFakeBackends(fake *FakeEnvoyGatewayV1alpha1, namespace string) apiv1alpha1.BackendInterface {
	return &fakeBackends{
		gentype.NewFakeClientWithList[*v1alpha1.Backend, *v1alpha1.BackendList](
			fake.Fake,
			namespace,
			v1alpha1.SchemeGroupVersion.WithResource("backends"),
			v1alpha1.SchemeGroupVersion.WithKind("Backend"),
			func() *v1alpha1.Backend { return &v1alpha1.Backend{} },
			func() *v1alpha1.BackendList { return &v1alpha1.BackendList{} },
			func(dst, src *v1alpha1.BackendList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.BackendList) []*v1alpha1.Backend { return gentype.ToPointerSlice(list.Items) },
			func(list *v1alpha1.BackendList, items []*v1alpha1.Backend) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
