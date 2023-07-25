// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

// RBAC for Gateway API resources.
// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=gatewayclasses;gateways;httproutes;grpcroutes;tlsroutes;tcproutes;udproutes;referencepolicies;referencegrants,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=gatewayclasses/status;gateways/status;httproutes/status;grpcroutes/status;tlsroutes/status;tcproutes/status;udproutes/status,verbs=update

// RBAC for Envoy Gateway custom resources.
// +kubebuilder:rbac:groups="config.gateway.envoyproxy.io",resources=envoyproxies,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="gateway.envoyproxy.io",resources=authenticationfilters;ratelimitfilters;envoypatchpolicies,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="gateway.envoyproxy.io",resources=envoypatchpolicies/status,verbs=update

// RBAC for watched resources of Gateway API controllers.
// +kubebuilder:rbac:groups="",resources=secrets;services;namespaces;nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=discovery.k8s.io,resources=endpointslices,verbs=get;list;watch
