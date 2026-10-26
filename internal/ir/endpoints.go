// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

// EndpointBackend identifies a Kubernetes endpoint source, including its incarnation.
// +kubebuilder:object:generate=true
type EndpointBackend struct {
	Kind      string `json:"kind" yaml:"kind"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
	UID       string `json:"uid" yaml:"uid"`
}

// EndpointSource pairs a backend revision with the port selected by a destination.
// +kubebuilder:object:generate=true
type EndpointSource struct {
	Backend  EndpointBackend `json:"backend" yaml:"backend"`
	PortName string          `json:"portName" yaml:"portName"`
	Protocol string          `json:"protocol" yaml:"protocol"`
	Revision uint64          `json:"revision" yaml:"revision"`
}
