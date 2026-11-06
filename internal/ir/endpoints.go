// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

// ServiceKey identifies the Service or ServiceImport that endpoints come from.
// +kubebuilder:object:generate=true
type ServiceKey struct {
	Kind      string `json:"kind" yaml:"kind"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
}

// EndpointSource records where a destination's endpoints came from: the Service,
// the port it uses, and the revision of the endpoints. The xDS publisher uses it
// to find the clusters to update when that Service's endpoints change, and the
// revision to skip updates that are not newer.
// +kubebuilder:object:generate=true
type EndpointSource struct {
	Service  ServiceKey `json:"service" yaml:"service"`
	PortName string     `json:"portName" yaml:"portName"`
	Protocol string     `json:"protocol" yaml:"protocol"`
	// Revision orders endpoint changes of one Service. The provider gives each
	// change a new, larger revision, so the publisher can tell whether an update
	// is newer than the endpoints it already published.
	Revision uint64 `json:"revision" yaml:"revision"`
}

// SamePort reports whether both sources name the same backend and port (name
// and protocol), ignoring Revision, so endpoints from one can stand in for the other.
func (s *EndpointSource) SamePort(other *EndpointSource) bool {
	return s.Service == other.Service && s.PortName == other.PortName && s.Protocol == other.Protocol
}
