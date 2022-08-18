package ir

import (
	"errors"
	"fmt"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
)

const (
	DefaultProxyName  = "default"
	DefaultProxyImage = "envoyproxy/envoy-dev:latest"
)

// Infra defines managed infrastructure.
// +k8s:deepcopy-gen=true
type Infra struct {
	// Proxy defines managed proxy infrastructure.
	Proxy *ProxyInfra
}

// ProxyInfra defines managed proxy infrastructure.
// +k8s:deepcopy-gen=true
type ProxyInfra struct {
	// TODO: Figure out how to represent metadata in the IR.
	// xref: https://github.com/envoyproxy/gateway/issues/173
	//
	// Name is the name used for managed proxy infrastructure.
	Name string
	// Config defines user-facing configuration of the managed proxy infrastructure.
	Config *v1alpha1.EnvoyProxy
	// Image is the container image used for the managed proxy infrastructure.
	// If unset, defaults to "envoyproxy/envoy-dev:latest".
	Image string
	// Listeners define the listeners exposed by the proxy infrastructure.
	Listeners []ProxyListener
}

// ProxyListener defines the listener configuration of the proxy infrastructure.
// +k8s:deepcopy-gen=true
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
	// Protocol is the protocol that the listener port will listener for.
	Protocol ProtocolType
	// Port is the port number to listen on.
	Port int32
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
		Name:      DefaultProxyName,
		Image:     DefaultProxyImage,
		Listeners: NewProxyListeners(),
	}
}

// NewProxyListeners returns a new slice of ProxyListener with default parameters.
func NewProxyListeners() []ProxyListener {
	return []ProxyListener{
		{
			Ports: nil,
		},
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
	if len(i.Proxy.Image) == 0 {
		i.Proxy.Image = DefaultProxyImage
	}
	if len(i.Proxy.Listeners) == 0 {
		i.Proxy.Listeners = NewProxyListeners()
	}

	return i.Proxy
}

// ValidateInfra validates the provided Infra.
func ValidateInfra(infra *Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	var errs []error

	if infra.Proxy != nil {
		if err := ValidateProxyInfra(infra.Proxy); err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}

// ValidateProxyInfra validates the provided ProxyInfra.
func ValidateProxyInfra(pInfra *ProxyInfra) error {
	var errs []error

	if len(pInfra.Name) == 0 {
		errs = append(errs, errors.New("name field required"))
	}

	if len(pInfra.Image) == 0 {
		errs = append(errs, errors.New("image field required"))
	}

	if len(pInfra.Listeners) > 1 {
		errs = append(errs, errors.New("no more than 1 listener is supported"))
	}

	if len(pInfra.Listeners) > 0 {
		for i := range pInfra.Listeners {
			listener := pInfra.Listeners[i]
			if len(listener.Ports) == 0 {
				errs = append(errs, errors.New("listener ports field required"))
			}
			for j := range listener.Ports {
				if len(listener.Ports[j].Name) == 0 {
					errs = append(errs, errors.New("listener name field required"))
				}
				if listener.Ports[j].Port < 1 || listener.Ports[j].Port > 65353 {
					errs = append(errs, errors.New("listener port must be a valid port number"))
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
