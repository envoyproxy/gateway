package v1alpha1

const (
	// EnvoyGatewayServiceName is the Kubernetes service name of Envoy Gateway.
	EnvoyGatewayServiceName = "envoy-gateway"
)

// DefaultEnvoyGateway returns a new EnvoyGateway with default configuration parameters.
func DefaultEnvoyGateway() *EnvoyGateway {
	eg := new(EnvoyGateway)
	eg.SetDefaults()
	return eg
}

// SetDefaults sets default EnvoyGateway configuration parameters.
func (e *EnvoyGateway) SetDefaults() {
	if e.TypeMeta.Kind == "" {
		e.TypeMeta.Kind = KindEnvoyGateway
	}
	if e.TypeMeta.APIVersion == "" {
		e.TypeMeta.APIVersion = GroupVersion.String()
	}
	if e.Provider == nil {
		e.Provider = DefaultProvider()
	}
	if e.Gateway == nil {
		e.Gateway = DefaultGateway()
	}
}

// DefaultGateway returns a new Gateway with default configuration parameters.
func DefaultGateway() *Gateway {
	return &Gateway{
		ControllerName: GatewayControllerName,
	}
}

// DefaultProvider returns a new Provider with default configuration parameters.
func DefaultProvider() *Provider {
	return &Provider{
		Type: ProviderTypeKubernetes,
	}
}

// DefaultEnvoyProxy returns a new EnvoyProxy with default configuration parameters.
func DefaultEnvoyProxy() *EnvoyProxy {
	ep := new(EnvoyProxy)
	ep.SetDefaults()
	return ep
}

// SetDefaults sets default EnvoyProxy configuration parameters.
func (e *EnvoyProxy) SetDefaults() {
	if e.TypeMeta.Kind == "" {
		e.TypeMeta.Kind = KindEnvoyProxy
	}
	if e.TypeMeta.APIVersion == "" {
		e.TypeMeta.APIVersion = GroupVersion.String()
	}
	if e.Spec.XDSServer == nil {
		e.Spec.XDSServer = DefaultXDSServer()
	}
}

// DefaultXDSServer returns a new XDSServer with default configuration parameters.
func DefaultXDSServer() *XDSServer {
	return &XDSServer{
		Address: EnvoyGatewayServiceName,
	}
}
