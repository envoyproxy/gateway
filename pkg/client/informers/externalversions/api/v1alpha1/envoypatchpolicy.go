// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	context "context"
	time "time"

	gatewayapiv1alpha1 "github.com/envoyproxy/gateway/api/v1alpha1"
	versioned "github.com/envoyproxy/gateway/pkg/client/clientset/versioned"
	internalinterfaces "github.com/envoyproxy/gateway/pkg/client/informers/externalversions/internalinterfaces"
	apiv1alpha1 "github.com/envoyproxy/gateway/pkg/client/listers/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// EnvoyPatchPolicyInformer provides access to a shared informer and lister for
// EnvoyPatchPolicies.
type EnvoyPatchPolicyInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() apiv1alpha1.EnvoyPatchPolicyLister
}

type envoyPatchPolicyInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewEnvoyPatchPolicyInformer constructs a new informer for EnvoyPatchPolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewEnvoyPatchPolicyInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredEnvoyPatchPolicyInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredEnvoyPatchPolicyInformer constructs a new informer for EnvoyPatchPolicy type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredEnvoyPatchPolicyInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EnvoyGatewayV1alpha1().EnvoyPatchPolicies(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.EnvoyGatewayV1alpha1().EnvoyPatchPolicies(namespace).Watch(context.TODO(), options)
			},
		},
		&gatewayapiv1alpha1.EnvoyPatchPolicy{},
		resyncPeriod,
		indexers,
	)
}

func (f *envoyPatchPolicyInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredEnvoyPatchPolicyInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *envoyPatchPolicyInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&gatewayapiv1alpha1.EnvoyPatchPolicy{}, f.defaultInformer)
}

func (f *envoyPatchPolicyInformer) Lister() apiv1alpha1.EnvoyPatchPolicyLister {
	return apiv1alpha1.NewEnvoyPatchPolicyLister(f.Informer().GetIndexer())
}
