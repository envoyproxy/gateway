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

// fakeHTTPRouteFilters implements HTTPRouteFilterInterface
type fakeHTTPRouteFilters struct {
	*gentype.FakeClientWithList[*v1alpha1.HTTPRouteFilter, *v1alpha1.HTTPRouteFilterList]
	Fake *FakeEnvoyGatewayV1alpha1
}

func newFakeHTTPRouteFilters(fake *FakeEnvoyGatewayV1alpha1, namespace string) apiv1alpha1.HTTPRouteFilterInterface {
	return &fakeHTTPRouteFilters{
		gentype.NewFakeClientWithList[*v1alpha1.HTTPRouteFilter, *v1alpha1.HTTPRouteFilterList](
			fake.Fake,
			namespace,
			v1alpha1.SchemeGroupVersion.WithResource("httproutefilters"),
			v1alpha1.SchemeGroupVersion.WithKind("HTTPRouteFilter"),
			func() *v1alpha1.HTTPRouteFilter { return &v1alpha1.HTTPRouteFilter{} },
			func() *v1alpha1.HTTPRouteFilterList { return &v1alpha1.HTTPRouteFilterList{} },
			func(dst, src *v1alpha1.HTTPRouteFilterList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.HTTPRouteFilterList) []*v1alpha1.HTTPRouteFilter {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1alpha1.HTTPRouteFilterList, items []*v1alpha1.HTTPRouteFilter) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
