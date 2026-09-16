// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package synthesizer

/*
All structs come from internal/ir/infra.go. They are inaccessible from this package, due to the internal/ root directory.
These structs have been pared down to demonstrate how to use the IR to construct a custom Envoy fleet.
*/

// ProxyInfra - a pared down version of the full Proxy IR that Envoy Gateway sends to the remote server.
// Use this metadata to construct the Envoy fleet.

type Infra struct {
	// Proxy defines managed proxy infrastructure.
	Proxy *ProxyInfra `json:"proxy" yaml:"proxy"`
}

type ProxyInfra struct {
	// Metadata defines metadata for the managed proxy infrastructure.
	Metadata *InfraMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// Name is the name used for managed proxy infrastructure.
	Name string `json:"name" yaml:"name"`
	// Namespace is the namespace used for managed proxy infrastructure.
	Namespace string `json:"namespace" yaml:"namespace"`
	// Listeners define the listeners exposed by the proxy infrastructure.
	Listeners []*ProxyListener `json:"listeners,omitempty" yaml:"listeners,omitempty"`
}

// InfraMetadata defines metadata for the managed proxy infrastructure.
type InfraMetadata struct {
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

type ProxyListener struct {
	// Name of the ProxyListener
	Name string `json:"name" yaml:"name"`
	// Ports define network ports of the listener.
	Ports []ListenerPort `json:"ports,omitempty" yaml:"ports,omitempty"`
}

type ListenerPort struct {
	// Name is the name of the listener port.
	Name string `json:"name" yaml:"name"`
	// Protocol is the protocol that the listener port will listener for.
	Protocol string `json:"protocol" yaml:"protocol"`
	// ServicePort is the port number the proxy service is listening on.
	ServicePort int32 `json:"servicePort" yaml:"servicePort"`
	// ContainerPort is the port number the proxy container is listening on.
	ContainerPort int32 `json:"containerPort" yaml:"containerPort"`
}
