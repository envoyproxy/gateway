// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"path/filepath"
	"strconv"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimev3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretv3 "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/telepresenceio/watchable"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	ktypes "k8s.io/apimachinery/pkg/types"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	extension "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/infrastructure/host"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	"github.com/envoyproxy/gateway/internal/xds/server/kubejwt"
	"github.com/envoyproxy/gateway/internal/xds/translator"
)

const (
	// XdsServerAddress is the listening address of the xds-server.
	XdsServerAddress = "0.0.0.0"

	// Default certificates path for envoy-gateway with Kubernetes provider.
	// xdsTLSCertFilepath is the fully qualified path of the file containing the
	// xDS server TLS certificate.
	xdsTLSCertFilepath = "/certs/tls.crt"
	// xdsTLSKeyFilepath is the fully qualified path of the file containing the
	// xDS server TLS key.
	xdsTLSKeyFilepath = "/certs/tls.key"
	// xdsTLSCaFilepath is the fully qualified path of the file containing the
	// xDS server trusted CA certificate.
	xdsTLSCaFilepath = "/certs/ca.crt"

	// defaultKubernetesIssuer is the default issuer URL for Kubernetes.
	// This is used for validating Service Account JWT tokens.
	defaultKubernetesIssuer = "https://kubernetes.default.svc.cluster.local"

	defaultMaxConnectionAgeGrace = 2 * time.Minute
)

var tracer = otel.Tracer("envoy-gateway/xds")

var maxConnectionAgeValues = []time.Duration{
	10 * time.Hour,
	11 * time.Hour,
	12 * time.Hour,
}

type Config struct {
	config.Server
	grpc              *grpc.Server
	cache             cache.SnapshotCacheWithCallbacks
	XdsIR             *message.XdsIR
	ExtensionManager  extension.Manager
	ProviderResources *message.ProviderResources
	RunnerErrors      *message.RunnerErrors
	// Test-configurable TLS paths
	TLSCertPath string
	TLSKeyPath  string
	TLSCaPath   string
}

type Runner struct {
	Config
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

func (r *Runner) Name() string {
	return string(egv1a1.LogComponentXdsRunner)
}

func (r *Runner) serverKeepaliveParams() (keepalive.ServerParameters, error) {
	params := keepalive.ServerParameters{
		MaxConnectionAge:      getRandomMaxConnectionAge(),
		MaxConnectionAgeGrace: defaultMaxConnectionAgeGrace,
	}

	if r.EnvoyGateway == nil || r.EnvoyGateway.XDSServer == nil {
		return params, nil
	}

	cfg := r.EnvoyGateway.XDSServer

	if cfg.MaxConnectionAge != nil {
		d, err := time.ParseDuration(string(*cfg.MaxConnectionAge))
		if err != nil {
			return keepalive.ServerParameters{}, fmt.Errorf("invalid xdsServer.maxConnectionAge: %w", err)
		}
		if d <= 0 {
			return keepalive.ServerParameters{}, fmt.Errorf("xdsServer.maxConnectionAge must be greater than zero")
		}
		params.MaxConnectionAge = d
	}

	if cfg.MaxConnectionAgeGrace != nil {
		d, err := time.ParseDuration(string(*cfg.MaxConnectionAgeGrace))
		if err != nil {
			return keepalive.ServerParameters{}, fmt.Errorf("invalid xdsServer.maxConnectionAgeGrace: %w", err)
		}
		if d <= 0 {
			return keepalive.ServerParameters{}, fmt.Errorf("xdsServer.maxConnectionAgeGrace must be greater than zero")
		}
		params.MaxConnectionAgeGrace = d
	}

	return params, nil
}

// getRandomMaxConnectionAge picks a random maxConnectionAge value
// to spread out envoy proxy connections over multiple envoy gateway replicas
func getRandomMaxConnectionAge() time.Duration {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	return maxConnectionAgeValues[rnd.Intn(len(maxConnectionAgeValues))]
}

// Close implements Runner interface.
func (r *Runner) Close() error { return nil }

// Start starts the xds-server runner
func (r *Runner) Start(ctx context.Context) error {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())
	r.cache = cache.NewSnapshotCache(true, r.Logger)

	// Set up the gRPC server and register the xDS handler.
	// Create SnapshotCache before start subscribeAndTranslate,
	// prevent panics in case cache is nil.
	tlsConfig, err := r.loadTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to load TLS config: %w", err)
	}
	r.Logger.Info("loaded TLS certificate and key")

	keepaliveParams, err := r.serverKeepaliveParams()
	if err != nil {
		return err
	}
	r.Logger.Info("configured gRPC keepalive", "maxConnectionAge", keepaliveParams.MaxConnectionAge, "maxConnectionAgeGrace", keepaliveParams.MaxConnectionAgeGrace)

	enforcementPolicy := keepalive.EnforcementPolicy{
		MinTime:             15 * time.Second,
		PermitWithoutStream: true,
	}

	baseKeepaliveOptions := []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(enforcementPolicy),
		grpc.KeepaliveParams(keepaliveParams),
	}

	grpcOpts := append([]grpc.ServerOption{}, baseKeepaliveOptions...)
	grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(tlsConfig)))

	// When GatewayNamespaceMode is enabled, we will use sTLS and Service Account JWT tokens to authenticate envoy proxy infra and xds server.
	if r.EnvoyGateway.GatewayNamespaceMode() {
		r.Logger.Info("gatewayNamespaceMode is enabled, setting up JWTAuthInterceptor and sTLS server")
		clientset, err := kubejwt.GetKubernetesClient()
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes client: %w", err)
		}
		saAudience := fmt.Sprintf("%s.%s.svc.%s", config.EnvoyGatewayServiceName, r.ControllerNamespace, r.DNSDomain)
		jwtInterceptor := kubejwt.NewJWTAuthInterceptor(
			r.Logger,
			clientset,
			defaultKubernetesIssuer,
			saAudience,
		)

		creds, err := credentials.NewServerTLSFromFile(xdsTLSCertFilepath, xdsTLSKeyFilepath)
		if err != nil {
			return fmt.Errorf("failed to create TLS credentials: %w", err)
		}

		grpcOpts = append([]grpc.ServerOption{}, baseKeepaliveOptions...)
		grpcOpts = append(grpcOpts,
			grpc.Creds(creds),
			grpc.StreamInterceptor(jwtInterceptor.Stream()),
		)
	}

	r.grpc = grpc.NewServer(grpcOpts...)
	registerServer(serverv3.NewServer(ctx, r.cache, r.cache), r.grpc)

	// Start and listen xDS gRPC Server.
	go r.serveXdsServer(ctx)

	// Do not call .Subscribe() inside Goroutine since it is supposed to be called from the same
	// Goroutine where Close() is called.
	sub := r.XdsIR.Subscribe(ctx)
	go r.translateFromSubscription(sub)
	r.Logger.Info("started")
	return err
}

func (r *Runner) serveXdsServer(ctx context.Context) {
	addr := net.JoinHostPort(XdsServerAddress, strconv.Itoa(bootstrap.DefaultXdsServerPort))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		r.Logger.Error(err, "failed to listen on address", "address", addr)
		return
	}

	go func() {
		<-ctx.Done()
		r.Logger.Info("grpc server shutting down")
		// We don't use GracefulStop here because envoy
		// has long-lived hanging xDS requests. There's no
		// mechanism to make those pending requests fail,
		// so we forcibly terminate the TCP sessions.
		r.grpc.Stop()
	}()

	if err = r.grpc.Serve(l); err != nil {
		r.Logger.Error(err, "failed to start grpc based xds server")
	}
}

// registerServer registers the given xDS protocol Server with the gRPC
// runtime.
func registerServer(srv serverv3.Server, g *grpc.Server) {
	// register services
	discoveryv3.RegisterAggregatedDiscoveryServiceServer(g, srv)
	secretv3.RegisterSecretDiscoveryServiceServer(g, srv)
	clusterv3.RegisterClusterDiscoveryServiceServer(g, srv)
	endpointv3.RegisterEndpointDiscoveryServiceServer(g, srv)
	listenerv3.RegisterListenerDiscoveryServiceServer(g, srv)
	routev3.RegisterRouteDiscoveryServiceServer(g, srv)
	runtimev3.RegisterRuntimeDiscoveryServiceServer(g, srv)
}

func (r *Runner) translateFromSubscription(sub <-chan watchable.Snapshot[string, *message.XdsIRWithContext]) {
	// Subscribe to resources
	message.HandleSubscription(
		r.Logger,
		message.Metadata{Runner: r.Name(), Message: message.XDSIRMessageName}, sub,
		func(update message.Update[string, *message.XdsIRWithContext], errChan chan error) {
			parentCtx := context.Background()
			if update.Value != nil && update.Value.Context != nil {
				parentCtx = update.Value.Context
			}

			traceCtx, span := tracer.Start(parentCtx, "XdsRunner.subscribeAndTranslate")
			defer span.End()
			traceLogger := r.Logger.WithTrace(traceCtx)
			traceLogger.Info("received an update")

			key := update.Key
			val := update.Value

			// Add span attributes for observability
			span.SetAttributes(
				attribute.String("xds-ir.key", update.Key),
				attribute.Bool("update.delete", update.Delete),
			)

			if update.Delete {
				if err := r.cache.GenerateNewSnapshot(key, nil, traceCtx); err != nil {
					traceLogger.Error(err, "failed to delete the snapshot")
					errChan <- err
				}
			} else {
				// Translate to xds resources
				t := &translator.Translator{
					ControllerNamespace: r.ControllerNamespace,
					FilterOrder:         val.XdsIR.FilterOrder,
					RuntimeFlags:        r.EnvoyGateway.RuntimeFlags,
					Logger:              traceLogger,
				}

				// Set the extension manager if an extension is loaded
				if r.ExtensionManager != nil {
					t.ExtensionManager = &r.ExtensionManager
				}

				// Set the rate limit service URL if global rate limiting is enabled.
				if r.EnvoyGateway.RateLimit != nil {
					t.GlobalRateLimit = &translator.GlobalRateLimitSettings{
						ServiceURL: ratelimit.GetServiceURL(r.ControllerNamespace, r.DNSDomain),
						FailClosed: r.EnvoyGateway.RateLimit.FailClosed,
					}
					if r.EnvoyGateway.RateLimit.Timeout != nil {
						d, err := time.ParseDuration(string(*r.EnvoyGateway.RateLimit.Timeout))
						if err != nil {
							traceLogger.Error(err, "invalid rateLimit timeout")
							errChan <- err
						} else {
							t.GlobalRateLimit.Timeout = d
						}
					}
				}

				_, translateSpan := tracer.Start(traceCtx, "Translator.Translate")
				result, err := t.Translate(val.XdsIR)
				translateSpan.End()
				if err != nil {
					traceLogger.Error(err, "skipped publishing xds resources: failed to translate xds ir")
					errChan <- err
				}

				// xDS translation is done in a best-effort manner, so the result
				// may contain partial resources even if there are errors.
				if result == nil {
					traceLogger.Info("no xds resources to publish")
					return
				}

				// Only update the snapshot cache when there are no errors, to avoid publishing partial resources.
				// This allows Envoy to continue using the previous known-good snapshot until the next successful translation.
				if err == nil {
					if result.XdsResources != nil {
						if r.cache == nil {
							r.Logger.Error(err, "failed to init snapshot cache")
							errChan <- err
						} else {
							// Update snapshot cache
							if err := r.cache.GenerateNewSnapshot(key, result.XdsResources, traceCtx); err != nil {
								r.Logger.Error(err, "failed to generate a snapshot")
								errChan <- err
							}
						}
					} else {
						r.Logger.Error(err, "skipped publishing xds resources")
					}
				}

				// Get all status keys from watchable and save them in the map statusesToDelete.
				// Iterating through result.EnvoyPatchPolicyStatuses, any valid keys will be removed from statusesToDelete.
				// Remaining keys will be deleted from watchable before we exit this function.
				statusesToDelete := make(map[ktypes.NamespacedName]bool)
				for key := range r.ProviderResources.EnvoyPatchPolicyStatuses.LoadAll() {
					statusesToDelete[key] = true
				}

				// Publish EnvoyPatchPolicyStatus
				for _, e := range result.EnvoyPatchPolicyStatuses {
					key := ktypes.NamespacedName{
						Name:      e.Name,
						Namespace: e.Namespace,
					}
					// Skip updating status for policies with empty status
					// They may have been skipped in this translation because
					// their target is not found (not relevant)
					if len(e.Status.Ancestors) > 0 {
						r.ProviderResources.EnvoyPatchPolicyStatuses.Store(key, e.Status)
					}
					delete(statusesToDelete, key)
				}
				// Discard the EnvoyPatchPolicyStatuses to reduce memory footprint
				result.EnvoyPatchPolicyStatuses = nil

				// Delete all the deletable status keys
				for key := range statusesToDelete {
					r.ProviderResources.EnvoyPatchPolicyStatuses.Delete(key)
				}
			}
		},
	)
	r.Logger.Info("subscriber shutting down")
}

func (r *Runner) loadTLSConfig() (*tls.Config, error) {
	var certPath, keyPath, caPath string

	// Use test-configurable paths if provided
	if r.TLSCertPath != "" && r.TLSKeyPath != "" && r.TLSCaPath != "" {
		certPath = r.TLSCertPath
		keyPath = r.TLSKeyPath
		caPath = r.TLSCaPath
	} else {
		// Use default paths based on provider type
		switch {
		case r.EnvoyGateway.Provider.IsRunningOnKubernetes():
			certPath = xdsTLSCertFilepath
			keyPath = xdsTLSKeyFilepath
			caPath = xdsTLSCaFilepath
		case r.EnvoyGateway.Provider.IsRunningOnHost():
			// Get config
			var hostCfg *egv1a1.EnvoyGatewayHostInfrastructureProvider
			if p := r.EnvoyGateway.Provider; p != nil && p.Custom != nil &&
				p.Custom.Infrastructure != nil && p.Custom.Infrastructure.Host != nil {
				hostCfg = p.Custom.Infrastructure.Host
			}

			paths, err := host.GetPaths(hostCfg)
			if err != nil {
				return nil, fmt.Errorf("failed to determine paths: %w", err)
			}

			certDir := paths.CertDir("envoy-gateway")
			certPath = filepath.Join(certDir, "tls.crt")
			keyPath = filepath.Join(certDir, "tls.key")
			caPath = filepath.Join(certDir, "ca.crt")
		default:
			return nil, fmt.Errorf("no valid tls certificates")
		}
	}

	tlsConfig, err := crypto.LoadTLSConfig(certPath, keyPath, caPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create tls config: %w", err)
	}
	return tlsConfig, err
}
