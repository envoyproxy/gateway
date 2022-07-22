package ir

import (
	"errors"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	cfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

const (
	DefaultProxyNamespace = "default"
	DefaultProxyImage     = "envoyproxy/envoy-dev:latest"
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
	// If unset, defaults to "default".
	Namespace string
	// Config defines user-facing configuration of the managed proxy infrastructure.
	Config *cfgv1a1.EnvoyProxy
	// Image is the container image used for the managed proxy infrastructure.
	// If unset, defaults to "envoyproxy/envoy-dev:latest".
	Image string
}

// NewInfra returns a new Infra with default parameters.
func NewInfra() *Infra {
	return &Infra{
		Proxy: NewProxyInfra(),
	}
}

// NewProxyInfra returns a new ProxyInfra with default parameters.
func NewProxyInfra() *ProxyInfra {
	return &ProxyInfra{
		Namespace: DefaultProxyNamespace,
		Image:     DefaultProxyImage,
	}
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

	if len(pInfra.Namespace) == 0 {
		errs = append(errs, errors.New("namespace field required"))
	}

	if len(pInfra.Image) == 0 {
		errs = append(errs, errors.New("image field required"))
	}

	return utilerrors.NewAggregate(errs)
}
