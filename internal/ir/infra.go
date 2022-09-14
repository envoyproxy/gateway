package ir

import (
	"errors"
	"fmt"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
)

const (
	DefaultProxyName  = "default"
	DefaultProxyImage = "envoyproxy/envoy:v1.23-latest"
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
	// Metadata defines metadata for the managed proxy infrastructure.
	Metadata *InfraMetadata
	// Name is the name used for managed proxy infrastructure.
	Name string
	// Config defines user-facing configuration of the managed proxy infrastructure.
	Config *v1alpha1.EnvoyProxy
	// Image is the container image used for the managed proxy infrastructure.
	// If unset, defaults to "envoyproxy/envoy:v1.23-latest".
	Image string
	// Listeners define the listeners exposed by the proxy infrastructure.
	Listeners []ProxyListener
}

// InfraMetadata defines metadata for the managed proxy infrastructure.
// +k8s:deepcopy-gen=true
type InfraMetadata struct {
	// Labels define a map of string keys and values that can be used to organize
	// and categorize proxy infrastructure objects.
	Labels map[string]string
}

// ProxyListener defines the listener configuration of the proxy infrastructure.
// +k8s:deepcopy-gen=true
type ProxyListener struct {
	// Name is the name of the listener and must be unique within a list of listeners.
	// Required.
	Name string
	// Address is the address that the listener should listen on.
	Address string
	// Ports define network ports of the listener.
	Ports []ListenerPort
}

// ListenerPort defines a network port of a listener.
// +k8s:deepcopy-gen=true
type ListenerPort struct {
	// Name is the name of the listener port.
	Name string
	// Protocol is the protocol that the listener port will listener for.
	Protocol ProtocolType
	// ServicePort is the port number the proxy service is listening on.
	ServicePort int32
	// ContainerPort is the port number the proxy container is listening on.
	ContainerPort int32
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
		Metadata:  NewInfraMetadata(),
		Name:      DefaultProxyName,
		Image:     DefaultProxyImage,
		Listeners: NewProxyListeners(),
	}
}

// NewProxyListeners returns a new slice of ProxyListener.
func NewProxyListeners() []ProxyListener {
	return []ProxyListener{
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
	if len(i.Proxy.Image) == 0 {
		i.Proxy.Image = DefaultProxyImage
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

// GetProxyListeners returns a slice of ProxyListener.
func (p *ProxyInfra) GetProxyListeners() []ProxyListener {
	if len(p.Listeners) == 0 {
		p.Listeners = NewProxyListeners()
	}

	return p.Listeners
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

	if len(p.Image) == 0 {
		errs = append(errs, errors.New("image field required"))
	}

	if len(p.Listeners) > 0 {
		for i := range p.Listeners {
			listener := p.Listeners[i]
			if len(listener.Name) == 0 {
				// TODO: Validate name uniqueness across listeners.
				errs = append(errs, errors.New("listener name field required"))
			}
			if len(listener.Ports) == 0 {
				errs = append(errs, errors.New("listener ports field required"))
			}
			for j := range listener.Ports {
				if len(listener.Ports[j].Name) == 0 {
					errs = append(errs, errors.New("listener port name field required"))
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
