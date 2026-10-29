// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"errors"
	"fmt"
	"io"
	"time"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

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
	// EnvoyGatewayServiceAccountName is the well-known name of the Envoy Gateway service account.
	EnvoyGatewayServiceAccountName = "envoy-gateway"
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
	// ControllerName is the well-known name of the Envoy Gateway controller's own resources
	// (Deployment, Service, ServiceAccount). Used to detect naming collision attacks.
	ControllerName string
	// ControllerFullName is the Helm release fullname (eg.fullname) used for certgen and other
	// release-scoped resources. Defaults to EnvoyGatewayServiceName.
	ControllerFullName string
	// ControllerServiceAccountName is the service account name of the Envoy Gateway controller pod.
	ControllerServiceAccountName string
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
		EnvoyGateway:                 egv1a1.DefaultEnvoyGateway(),
		ControllerNamespace:          env.Lookup("ENVOY_GATEWAY_NAMESPACE", DefaultNamespace),
		ControllerName:               EnvoyGatewayServiceName,
		ControllerFullName:           env.Lookup("ENVOY_GATEWAY_FULLNAME", EnvoyGatewayServiceName),
		ControllerServiceAccountName: env.Lookup("ENVOY_GATEWAY_SERVICE_ACCOUNT", EnvoyGatewayServiceAccountName),
		DNSDomain:                    env.Lookup("KUBERNETES_CLUSTER_DOMAIN", DefaultDNSDomain),
		Logger:                       logging.DefaultLogger(stdout, egv1a1.LogLevelInfo),
		Stdout:                       stdout,
		Stderr:                       stderr,
		Elected:                      make(chan struct{}),
		ProviderReady:                make(chan struct{}),
		KubernetesClient:             NewKubernetesClientHolder(),
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

	// The CRD schema, including its CEL rules, is the single source of truth for
	// ClusterSettings, so validate against it rather than duplicating each rule in Go.
	if eg.RateLimit != nil {
		fldPath := field.NewPath("rateLimit", "serviceSettings")
		if err := ValidateClusterSettings(fldPath, eg.RateLimit.ServiceSettings); err != nil {
			return fmt.Errorf("invalid rateLimit.serviceSettings: %w", err)
		}
		if err := validateRateLimitServiceSettings(fldPath, eg.RateLimit.ServiceSettings); err != nil {
			return fmt.Errorf("invalid rateLimit.serviceSettings: %w", err)
		}
	}

	return nil
}

// ApplyEnvoyGatewaySchemaDefaults applies the defaults declared in the CRD schemas to the parts
// of eg that are borrowed from CRD types, as the API server would on admission. EnvoyGateway is
// loaded as static configuration, so this runs once when it's loaded or reloaded, after
// ValidateEnvoyGateway, rather than on every translation.
func ApplyEnvoyGatewaySchemaDefaults(eg *egv1a1.EnvoyGateway) error {
	if eg == nil || eg.RateLimit == nil || eg.RateLimit.ServiceSettings == nil {
		return nil
	}
	cs, err := DefaultClusterSettings(eg.RateLimit.ServiceSettings)
	if err != nil {
		return fmt.Errorf("failed to default rateLimit.serviceSettings: %w", err)
	}
	eg.RateLimit.ServiceSettings = cs
	return nil
}

// validateRateLimitServiceSettings rejects the ClusterSettings fields that can't be used for the
// rate limit service cluster: it is a gRPC (HTTP/2) service reached over mTLS, with no associated route.
func validateRateLimitServiceSettings(fldPath *field.Path, cs *egv1a1.ClusterSettings) error {
	if cs == nil {
		return nil
	}

	var errs field.ErrorList
	if cs.HTTP1 != nil {
		errs = append(errs, field.Forbidden(fldPath.Child("http1"), "the rate limit service only accepts gRPC over mTLS"))
	}
	if cs.HTTP3 != nil {
		errs = append(errs, field.Forbidden(fldPath.Child("http3"), "the rate limit service only accepts gRPC over mTLS"))
	}
	if cs.ProxyProtocol != nil {
		errs = append(errs, field.Forbidden(fldPath.Child("proxyProtocol"), "the rate limit service only accepts gRPC over mTLS"))
	}
	if cs.HealthCheck != nil && cs.HealthCheck.Active != nil && cs.HealthCheck.Active.Type == egv1a1.ActiveHealthCheckerTypeHTTP {
		errs = append(errs, field.NotSupported(fldPath.Child("healthCheck", "active", "type"), cs.HealthCheck.Active.Type,
			[]egv1a1.ActiveHealthCheckerType{egv1a1.ActiveHealthCheckerTypeGRPC, egv1a1.ActiveHealthCheckerTypeTCP}))
	}
	if cs.Timeout != nil && cs.Timeout.HTTP != nil && cs.Timeout.HTTP.RequestTimeout != nil {
		errs = append(errs, field.Forbidden(fldPath.Child("timeout", "http", "requestTimeout"), "use rateLimit.timeout instead"))
	}
	if cs.LoadBalancer != nil && cs.LoadBalancer.Type == egv1a1.ConsistentHashLoadBalancerType {
		errs = append(errs, field.Forbidden(fldPath.Child("loadBalancer", "type"), "there is no route to compute the hash from"))
	}
	// The schema admits zero durations that the IR rejects, which would only surface when the
	// xDS translator builds the cluster and stall every xDS update.
	if hc := cs.HealthCheck; hc != nil {
		hcPath := fldPath.Child("healthCheck")
		if hc.Active != nil {
			errs = append(errs, validateNonZeroDuration(hcPath.Child("active", "timeout"), hc.Active.Timeout)...)
			errs = append(errs, validateNonZeroDuration(hcPath.Child("active", "interval"), hc.Active.Interval)...)
		}
		if hc.Passive != nil {
			errs = append(errs, validateNonZeroDuration(hcPath.Child("passive", "interval"), hc.Passive.Interval)...)
			errs = append(errs, validateNonZeroDuration(hcPath.Child("passive", "baseEjectionTime"), hc.Passive.BaseEjectionTime)...)
		}
	}
	return errs.ToAggregate()
}

func validateNonZeroDuration(fldPath *field.Path, d *gwapiv1.Duration) field.ErrorList {
	if d == nil {
		return nil
	}
	if parsed, err := time.ParseDuration(string(*d)); err == nil && parsed == 0 {
		return field.ErrorList{field.Invalid(fldPath, *d, "must be greater than zero")}
	}
	return nil
}
