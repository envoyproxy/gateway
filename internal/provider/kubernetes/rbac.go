// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=gatewayclasses;gateways;httproutes;grpcroutes;tlsroutes;tcproutes;udproutes;referencepolicies;referencegrants,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=gatewayclasses/status;gateways/status;httproutes/status;grpcroutes/status;tlsroutes/status;tcproutes/status;udproutes/status,verbs=update
// +kubebuilder:rbac:groups="gateway.envoyproxy.io",resources=authenticationfilters,verbs=get;list;watch;update

// RBAC for watched resources of Gateway API controllers.
// +kubebuilder:rbac:groups="",resources=secrets;services;namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
