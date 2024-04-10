// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"

	"golang.org/x/exp/slices"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/api/v1alpha1"
)

const (
	DefaultProxyName = "default"
)

// Infra defines managed infrastructure.
// +k8s:deepcopy-gen=true
type Infra struct {
	// Proxy defines managed proxy infrastructure.
	Proxy *ProxyInfra `json:"proxy" yaml:"proxy"`
}

func (i Infra) YAMLString() string {
	y, _ := yaml.Marshal(&i)
	return string(y)
}

// ProxyInfra defines managed proxy infrastructure.
// +k8s:deepcopy-gen=true
type ProxyInfra struct {
	// Metadata defines metadata for the managed proxy infrastructure.
	Metadata *InfraMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// Name is the name used for managed proxy infrastructure.
	Name string `json:"name" yaml:"name"`
	// Config defines user-facing configuration of the managed proxy infrastructure.
	Config *v1alpha1.EnvoyProxy `json:"config,omitempty" yaml:"config,omitempty"`
	// Listeners define the listeners exposed by the proxy infrastructure.
	Listeners []*ProxyListener `json:"listeners,omitempty" yaml:"listeners,omitempty"`
	// Addresses contain the external addresses this gateway has been
	// requested to be available at.
	Addresses []string `json:"addresses,omitempty" yaml:"addresses,omitempty"`
}

// InfraMetadata defines metadata for the managed proxy infrastructure.
// +k8s:deepcopy-gen=true
type InfraMetadata struct {
	// Annotations define a map of string keys and values that can be used to
	// organize and categorize proxy infrastructure objects.
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	// Labels define a map of string keys and values that can be used to organize
	// and categorize proxy infrastructure objects.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// ProxyListener defines the listener configuration of the proxy infrastructure.
// +k8s:deepcopy-gen=true
type ProxyListener struct {
	// Name of the ProxyListener
	Name string `json:"name" yaml:"name"`
	// Address is the address that the listener should listen on.
	Address *string `json:"address" yaml:"address"`
	// Ports define network ports of the listener.
	Ports []ListenerPort `json:"ports,omitempty" yaml:"ports,omitempty"`
	// HTTP3 provides HTTP/3 configuration on the listener.
	// +optional
	HTTP3 *HTTP3Settings `json:"http3,omitempty"`
}

// HTTP3Settings provides HTTP/3 configuration on the listener.
type HTTP3Settings struct {
	QUICPort int32 `json:"quicPort" yaml:"quicPort"`
}

// ListenerPort defines a network port of a listener.
// +k8s:deepcopy-gen=true
type ListenerPort struct {
	// Name is the name of the listener port.
	Name string `json:"name" yaml:"name"`
	// Protocol is the protocol that the listener port will listener for.
	Protocol ProtocolType `json:"protocol" yaml:"protocol"`
	// ServicePort is the port number the proxy service is listening on.
	ServicePort int32 `json:"servicePort" yaml:"servicePort"`
	// ContainerPort is the port number the proxy container is listening on.
	ContainerPort int32 `json:"containerPort" yaml:"containerPort"`
}

// ProtocolType defines the application protocol accepted by a ListenerPort.
//
// Valid values include "HTTP" and "HTTPS".
type ProtocolType string

const (
	// HTTPProtocolType accepts cleartext HTTP/1.1 sessions over TCP or HTTP/2
	// over cleartext.
	HTTPProtocolType ProtocolType = "HTTP"

	// HTTPSProtocolType accepts HTTP/1.1 or HTTP/2 sessions over TLS.
	HTTPSProtocolType ProtocolType = "HTTPS"

	// TLSProtocolType accepts TLS sessions over TCP.
	TLSProtocolType ProtocolType = "TLS"

	// TCPProtocolType accepts TCP connection.
	TCPProtocolType ProtocolType = "TCP"

	// UDPProtocolType accepts UDP connection.
	UDPProtocolType ProtocolType = "UDP"
)

// NewInfra returns a new Infra with default parameters.
func NewInfra() *Infra {
	return &Infra{
		Proxy: NewProxyInfra(),
	}
}

// NewProxyInfra returns a new ProxyInfra with default parameters.
func NewProxyInfra() *ProxyInfra {
	return &ProxyInfra{
		Metadata: NewInfraMetadata(),
		Name:     DefaultProxyName,
	}
}

// NewProxyListeners returns a new slice of ProxyListener with default parameters.
func NewProxyListeners() []*ProxyListener {
	return []*ProxyListener{
		{
			Ports: nil,
		},
	}
}

// NewInfraMetadata returns a new InfraMetadata.
func NewInfraMetadata() *InfraMetadata {
	return &InfraMetadata{
		Labels: map[string]string{},
	}
}

// GetProxyInfra returns the ProxyInfra.
func (i *Infra) GetProxyInfra() *ProxyInfra {
	if i.Proxy == nil {
		i.Proxy = NewProxyInfra()
		return i.Proxy
	}
	if len(i.Proxy.Name) == 0 {
		i.Proxy.Name = DefaultProxyName
	}
	if len(i.Proxy.Listeners) == 0 {
		i.Proxy.Listeners = NewProxyListeners()
	}
	if i.Proxy.Metadata == nil {
		i.Proxy.Metadata = NewInfraMetadata()
	}

	return i.Proxy
}

// GetProxyMetadata returns the InfraMetadata.
func (p *ProxyInfra) GetProxyMetadata() *InfraMetadata {
	if p.Metadata == nil {
		p.Metadata = NewInfraMetadata()
	}

	return p.Metadata
}

// GetProxyConfig returns the ProxyInfra config.
func (p *ProxyInfra) GetProxyConfig() *v1alpha1.EnvoyProxy {
	if p.Config == nil {
		p.Config = new(v1alpha1.EnvoyProxy)
	}

	return p.Config
}

// Validate validates the provided Infra.
func (i *Infra) Validate() error {
	if i == nil {
		return errors.New("infra ir is nil")
	}

	var errs []error

	if i.Proxy != nil {
		if err := i.Proxy.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}

// Validate validates the provided ProxyInfra.
func (p *ProxyInfra) Validate() error {
	var errs []error

	if len(p.Name) == 0 {
		errs = append(errs, errors.New("name field required"))
	}

	if len(p.Listeners) > 0 {
		for i := range p.Listeners {
			listener := p.Listeners[i]
			if len(listener.Ports) == 0 {
				errs = append(errs, errors.New("listener ports field required"))
			}
			for j := range listener.Ports {
				if len(listener.Ports[j].Name) == 0 {
					errs = append(errs, errors.New("listener name field required"))
				}
				if listener.Ports[j].ServicePort < 1 || listener.Ports[j].ServicePort > 65353 {
					errs = append(errs, errors.New("listener service port must be a valid port number"))
				}
				if listener.Ports[j].ContainerPort < 1024 || listener.Ports[j].ContainerPort > 65353 {
					errs = append(errs, errors.New("listener container port must be a valid ephemeral port number"))
				}
			}
		}
	}

	return utilerrors.NewAggregate(errs)
}

// ObjectName returns the name of the proxy infrastructure object.
func (p *ProxyInfra) ObjectName() string {
	if len(p.Name) == 0 {
		return fmt.Sprintf("envoy-%s", DefaultProxyName)
	}
	return "envoy-" + p.Name
}

// Equal implements the Comparable interface used by watchable.DeepEqual to skip unnecessary updates.
func (p *ProxyInfra) Equal(y *ProxyInfra) bool {
	// Deep copy to avoid modifying the original ordering.
	p = p.DeepCopy()
	p.sort()
	y = y.DeepCopy()
	y.sort()
	return reflect.DeepEqual(p, y)
}

// sort ensures the listeners are in a consistent order.
func (p *ProxyInfra) sort() {
	slices.SortFunc(p.Listeners, func(l1, l2 *ProxyListener) int {
		return cmp.Compare(l1.Name, l2.Name)
	})
}
