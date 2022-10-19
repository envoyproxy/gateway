// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=gatewayclasses;gateways;httproutes;tlsroutes;referencepolicies;referencegrants,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=gatewayclasses/status;gateways/status;httproutes/status;tlsroutes/status,verbs=update

// RBAC for watched resources of Gateway API controllers.
// +kubebuilder:rbac:groups="",resources=secrets;services;namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
