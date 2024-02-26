// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"errors"
	"os"
	"time"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils/env"
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
	EnvoyGateway *v1alpha1.EnvoyGateway
	// Namespace is the namespace that Envoy Gateway runs in.
	Namespace string
	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string
	// Logger is the logr implementation used by Envoy Gateway.
	Logger logging.Logger
	// Leader election settings
	LeaderElection *LeaderElection
}

type LeaderElection struct {
	// LeaseDuration defines the time non-leader contenders will wait before attempting to claim leadership. It's based on the timestamp of the last acknowledged signal. The default setting is 15 seconds.
	LeaseDuration *time.Duration
	// RenewDeadline represents the time frame within which the current leader will attempt to renew its leadership status before relinquishing its position. The default setting is 10 seconds.
	RenewDeadline *time.Duration
	// RetryPeriod denotes the interval at which LeaderElector clients should perform action retries. The default setting is 2 seconds.
	RetryPeriod *time.Duration
	// Enables or disables the leader election feature. It is enabled by default.
	Enabled bool
}

// New returns a Server with default parameters.
func New() (*Server, error) {
	settings, err := withLeaderElectionSettings()
	if err != nil {
		return nil, err
	}

	return &Server{
		EnvoyGateway: v1alpha1.DefaultEnvoyGateway(),
		Namespace:    env.Lookup("ENVOY_GATEWAY_NAMESPACE", DefaultNamespace),
		DNSDomain:    env.Lookup("KUBERNETES_CLUSTER_DOMAIN", DefaultDNSDomain),
		// the default logger
		Logger:         logging.DefaultLogger(v1alpha1.LogLevelInfo),
		LeaderElection: settings,
	}, nil
}

func parseEnvDuration(envVar string, duration **time.Duration) error {
	if value := os.Getenv(envVar); value != "" {
		if parsedDuration, err := time.ParseDuration(value); err == nil {
			*duration = &parsedDuration
		} else {
			return err
		}
	}
	return nil
}

func withLeaderElectionSettings() (*LeaderElection, error) {
	le := &LeaderElection{
		Enabled: true,
	}

	if err := parseEnvDuration("ENVOY_GATEWAY_LEADER_ELECTION_RENEW_DEADLINE", &le.RenewDeadline); err != nil {
		return nil, err
	}

	if err := parseEnvDuration("ENVOY_GATEWAY_LEADER_ELECTION_LEASE_DURATION", &le.LeaseDuration); err != nil {
		return nil, err
	}

	if err := parseEnvDuration("ENVOY_GATEWAY_LEADER_ELECTION_RETRY_PERIOD", &le.RetryPeriod); err != nil {
		return nil, err
	}

	if os.Getenv("ENVOY_GATEWAY_LEADER_ELECTION_ENABLED") == "false" {
		le.Enabled = false
	}
	return le, nil
}

// Validate validates a Server config.
func (s *Server) Validate() error {
	switch {
	case s == nil:
		return errors.New("server config is unspecified")
	case len(s.Namespace) == 0:
		return errors.New("namespace is empty string")
	}
	if err := validation.ValidateEnvoyGateway(s.EnvoyGateway); err != nil {
		return err
	}

	return nil
}
