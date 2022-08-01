package ir

import (
	cfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

// Infra defines managed infrastructure.
type Infra struct {
	// Proxy defines managed proxy infrastructure.
	Proxy *ProxyInfra
}

// ProxyInfra defines managed proxy infrastructure.
type ProxyInfra struct {
	// TODO: Figure out how to represent metadata in the IR.
	// xref: https://github.com/envoyproxy/gateway/issues/173
	//
	// Name is the name used for managed proxy infrastructure.
	Name string
	// Namespace is the namespace used for managed proxy infrastructure.
	Namespace string
	// Config defines user-facing configuration of the managed proxy infrastructure.
	Config *cfgv1a1.EnvoyProxy
	// Image is the container image used for the managed proxy infrastructure.
	Image string
	// Listeners define the listeners exposed by the proxy infrastructure.
	Listeners []ProxyListener
}

// ProxyListener defines the listener configuration of the proxy infrastructure.
type ProxyListener struct {
	// Address is the address that the listener should listen on.
	Address string
	// Ports define network ports of the listener.
	Ports []ListenerPort
}

// ListenerPort defines a network port of a listener.
type ListenerPort struct {
	// Name is the name of the listener port.
	Name string
	// Port is the port number to listen on.
	Port int32
}
