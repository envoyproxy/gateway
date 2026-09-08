// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"errors"
	"fmt"
	"io"

	"sigs.k8s.io/controller-runtime/pkg/client"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils/env"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

const (
	// DefaultNamespace is the default namespace of Envoy Gateway.
	DefaultNamespace = "envoy-gateway-system"
	// DefaultDNSDomain is the default DNS domain used by k8s services.
	DefaultDNSDomain = "cluster.local"
	// EnvoyGatewayServiceName is the name of the Envoy Gateway service.
	EnvoyGatewayServiceName = "envoy-gateway"
	// EnvoyPrefix is the prefix applied to the Envoy ConfigMap, Service, Deployment, and ServiceAccount.
	EnvoyPrefix = "envoy"
)

// Server wraps the EnvoyGateway configuration and additional parameters
// used by Envoy Gateway server.
type Server struct {
	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *egv1a1.EnvoyGateway
	// ControllerNamespace is the namespace that Envoy Gateway runs in.
	ControllerNamespace string
	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string
	// Logger is the logr implementation used by Envoy Gateway.
	Logger logging.Logger
	// Elected chan is used to signal when an EG instance is elected as leader.
	Elected chan struct{}
	// ProviderReady is closed once the Kubernetes provider cache is synced and the cached client is ready for consumers.
	ProviderReady chan struct{}
	// Stdout is the writer for standard output.
	Stdout io.Writer
	// Stderr is the writer for error output.
	Stderr io.Writer
	// KubernetesClient holds the controller-runtime client created by the Kubernetes provider.
	// This is used by the infrastructure runner to create the envoy proxy and rate limit infra resources.
	KubernetesClient *KubernetesClientHolder
}

type KubernetesClientHolder struct {
	client client.Client
}

func NewKubernetesClientHolder() *KubernetesClientHolder {
	return &KubernetesClientHolder{}
}

func (h *KubernetesClientHolder) Set(cli client.Client) {
	if h != nil {
		h.client = cli
	}
}

func (h *KubernetesClientHolder) Get() client.Client {
	if h == nil {
		return nil
	}
	return h.client
}

// New returns a Server with default parameters.
func New(stdout, stderr io.Writer) (*Server, error) {
	return &Server{
		EnvoyGateway:        egv1a1.DefaultEnvoyGateway(),
		ControllerNamespace: env.Lookup("ENVOY_GATEWAY_NAMESPACE", DefaultNamespace),
		DNSDomain:           env.Lookup("KUBERNETES_CLUSTER_DOMAIN", DefaultDNSDomain),
		Logger:              logging.DefaultLogger(stdout, egv1a1.LogLevelInfo),
		Stdout:              stdout,
		Stderr:              stderr,
		Elected:             make(chan struct{}),
		ProviderReady:       make(chan struct{}),
		KubernetesClient:    NewKubernetesClientHolder(),
	}, nil
}

// Validate validates a Server config and returns any warnings.
func (s *Server) Validate() ([]string, error) {
	switch {
	case s == nil:
		return nil, errors.New("server config is unspecified")
	case len(s.ControllerNamespace) == 0:
		return nil, errors.New("namespace is empty string")
	}
	if err := ValidateEnvoyGateway(s.EnvoyGateway); err != nil {
		return nil, err
	}

	warnings := validation.WarnEnvoyGateway(s.EnvoyGateway)
	return warnings, nil
}

// validateEnvoyGateway validates the provided EnvoyGateway config, including
// the bootstrap override under the embedded default EnvoyProxy spec.
//
// api/v1alpha1/validation.ValidateEnvoyGateway intentionally skips that check:
// validating a bootstrap override means patching it onto the internal xDS
// bootstrap template and diffing the result, which the api package cannot do
// without depending on internal packages (see validateEnvoyProxySpec's doc
// comment). Standalone EnvoyProxy resources get the same extra check in
// internal/gatewayapi/translator.go's validateEnvoyProxy; this is the
// equivalent for the merged default spec used by the config loader, so that
// an override which breaks dynamic_resources or the xDS cluster is rejected
// here too, rather than leaving Envoy unable to reach the control plane.
func ValidateEnvoyGateway(eg *egv1a1.EnvoyGateway) error {
	if err := validation.ValidateEnvoyGateway(eg); err != nil {
		return err
	}

	if eg.EnvoyProxy != nil && eg.EnvoyProxy.Bootstrap != nil {
		if err := bootstrap.Validate(eg.EnvoyProxy.Bootstrap); err != nil {
			return fmt.Errorf("invalid EnvoyProxy template: %w", err)
		}
	}

	return nil
}
