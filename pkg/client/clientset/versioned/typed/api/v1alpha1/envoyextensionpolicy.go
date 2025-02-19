// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	context "context"

	apiv1alpha1 "github.com/envoyproxy/gateway/api/v1alpha1"
	scheme "github.com/envoyproxy/gateway/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// EnvoyExtensionPoliciesGetter has a method to return a EnvoyExtensionPolicyInterface.
// A group's client should implement this interface.
type EnvoyExtensionPoliciesGetter interface {
	EnvoyExtensionPolicies(namespace string) EnvoyExtensionPolicyInterface
}

// EnvoyExtensionPolicyInterface has methods to work with EnvoyExtensionPolicy resources.
type EnvoyExtensionPolicyInterface interface {
	Create(ctx context.Context, envoyExtensionPolicy *apiv1alpha1.EnvoyExtensionPolicy, opts v1.CreateOptions) (*apiv1alpha1.EnvoyExtensionPolicy, error)
	Update(ctx context.Context, envoyExtensionPolicy *apiv1alpha1.EnvoyExtensionPolicy, opts v1.UpdateOptions) (*apiv1alpha1.EnvoyExtensionPolicy, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, envoyExtensionPolicy *apiv1alpha1.EnvoyExtensionPolicy, opts v1.UpdateOptions) (*apiv1alpha1.EnvoyExtensionPolicy, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*apiv1alpha1.EnvoyExtensionPolicy, error)
	List(ctx context.Context, opts v1.ListOptions) (*apiv1alpha1.EnvoyExtensionPolicyList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *apiv1alpha1.EnvoyExtensionPolicy, err error)
	EnvoyExtensionPolicyExpansion
}

// envoyExtensionPolicies implements EnvoyExtensionPolicyInterface
type envoyExtensionPolicies struct {
	*gentype.ClientWithList[*apiv1alpha1.EnvoyExtensionPolicy, *apiv1alpha1.EnvoyExtensionPolicyList]
}

// newEnvoyExtensionPolicies returns a EnvoyExtensionPolicies
func newEnvoyExtensionPolicies(c *EnvoyGatewayV1alpha1Client, namespace string) *envoyExtensionPolicies {
	return &envoyExtensionPolicies{
		gentype.NewClientWithList[*apiv1alpha1.EnvoyExtensionPolicy, *apiv1alpha1.EnvoyExtensionPolicyList](
			"envoyextensionpolicies",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *apiv1alpha1.EnvoyExtensionPolicy { return &apiv1alpha1.EnvoyExtensionPolicy{} },
			func() *apiv1alpha1.EnvoyExtensionPolicyList { return &apiv1alpha1.EnvoyExtensionPolicyList{} },
		),
	}
}
