// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/security/advancedtls"
	"google.golang.org/grpc/test/bufconn"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	k8scli "sigs.k8s.io/controller-runtime/pkg/client"
	k8sclicfg "sigs.k8s.io/controller-runtime/pkg/client/config"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	extTypes "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/kubernetes"
	"github.com/envoyproxy/gateway/proto/extension"
)

const grpcServiceConfigTemplate = `{
"methodConfig": [{
	"name": [{"service": "envoygateway.extension.EnvoyGatewayExtension"}],
	"waitForReady": true,
	"retryPolicy": {
		"MaxAttempts": %d,
		"InitialBackoff": "%fs",
		"MaxBackoff": "%fs",
		"BackoffMultiplier": %f,
		"RetryableStatusCodes": [ %s ]
	}
}]}`

var _ extTypes.Manager = (*Manager)(nil)

type Manager struct {
	k8sClient          k8scli.Client
	namespace          string
	extension          egv1a1.ExtensionManager
	extensionConnCache *grpc.ClientConn
}

// NewManager returns a new Manager
func NewManager(cfg *config.Server, inK8s bool) (extTypes.Manager, error) {
	var cli k8scli.Client
	var err error
	if inK8s {
		cli, err = k8scli.New(k8sclicfg.GetConfigOrDie(), k8scli.Options{Scheme: envoygateway.GetScheme()})
		if err != nil {
			return nil, err
		}
	}

	var extension *egv1a1.ExtensionManager
	if cfg.EnvoyGateway != nil {
		extension = cfg.EnvoyGateway.ExtensionManager
	}

	// Setup an empty default in the case that no config was provided
	if extension == nil {
		extension = &egv1a1.ExtensionManager{}
	}

	return &Manager{
		k8sClient: cli,
		namespace: cfg.ControllerNamespace,
		extension: *extension,
	}, nil
}

func NewInMemoryManager(cfg egv1a1.ExtensionManager, server extension.EnvoyGatewayExtensionServer) (extTypes.Manager, func(), error) {
	if server == nil {
		return nil, nil, fmt.Errorf("in-memory manager must be passed a server")
	}

	buffer := 10 * 1024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := grpc.NewServer()
	extension.RegisterEnvoyGatewayExtensionServer(baseServer, server)
	go func() {
		_ = baseServer.Serve(lis)
	}()

	inMemoryManagerOpts := []grpc.DialOption{
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if cfg.Service != nil {
		opts, err := setupGRPCOpts(context.Background(), nil, &cfg, "")
		if err != nil {
			return nil, nil, err
		}
		inMemoryManagerOpts = append(inMemoryManagerOpts, opts...)
	}

	conn, err := grpc.DialContext(context.Background(), "", inMemoryManagerOpts...)
	if err != nil {
		return nil, nil, err
	}
	c := func() {
		lis.Close()
		baseServer.Stop()
	}

	return &Manager{
		extensionConnCache: conn,
		extension:          cfg,
	}, c, nil
}

// FailOpen returns true if the extension manager is configured to fail open, and false otherwise.
func (m *Manager) FailOpen() bool {
	return m.extension.FailOpen
}

// GetTranslationHookConfig returns the translation hook configuration.
func (m *Manager) GetTranslationHookConfig() *egv1a1.TranslationConfig {
	if m.extension.Hooks == nil ||
		m.extension.Hooks.XDSTranslator == nil ||
		m.extension.Hooks.XDSTranslator.Translation == nil {
		return nil
	}
	return m.extension.Hooks.XDSTranslator.Translation
}

// HasExtension checks to see whether a given Group and Kind has an
// associated extension registered for it.
func (m *Manager) HasExtension(g gwapiv1.Group, k gwapiv1.Kind) bool {
	extension := m.extension
	// TODO: not currently checking the version since extensionRef only supports group and kind.
	for _, gvk := range extension.Resources {
		if g == gwapiv1.Group(gvk.Group) && k == gwapiv1.Kind(gvk.Kind) {
			return true
		}
	}
	// Also check backend resources for custom backend support
	for _, gvk := range extension.BackendResources {
		if g == gwapiv1.Group(gvk.Group) && k == gwapiv1.Kind(gvk.Kind) {
			return true
		}
	}
	return false
}

func getExtensionServerAddress(service *egv1a1.ExtensionService) string {
	var serverAddr string
	switch {
	case service.FQDN != nil:
		serverAddr = net.JoinHostPort(service.FQDN.Hostname, strconv.Itoa(int(service.FQDN.Port)))
	case service.IP != nil:
		serverAddr = net.JoinHostPort(service.IP.Address, strconv.Itoa(int(service.IP.Port)))
	case service.Unix != nil:
		serverAddr = fmt.Sprintf("unix://%s", service.Unix.Path)
	case service.Host != "":
		serverAddr = net.JoinHostPort(service.Host, strconv.Itoa(int(service.Port)))
	}
	return serverAddr
}

// GetPreXDSHookClient checks if the registered extension makes use of a particular hook type that modifies inputs
// that are used to generate an xDS resource.
// If the extension makes use of the hook then the XDS Hook Client is returned. If it does not support
// the hook type then nil is returned
func (m *Manager) GetPreXDSHookClient(xdsHookType egv1a1.XDSTranslatorHook) (extTypes.XDSHookClient, error) {
	ctx := context.Background()
	ext := m.extension

	if ext.Hooks == nil {
		return nil, nil
	}
	if ext.Hooks.XDSTranslator == nil {
		return nil, nil
	}

	hookUsed := false
	for _, hook := range ext.Hooks.XDSTranslator.Pre {
		if xdsHookType == hook {
			hookUsed = true
			break
		}
	}
	if !hookUsed {
		return nil, nil
	}

	if m.extensionConnCache == nil {
		serverAddr := getExtensionServerAddress(ext.Service)

		opts, err := setupGRPCOpts(ctx, m.k8sClient, &ext, m.namespace)
		if err != nil {
			return nil, err
		}

		conn, err := grpc.Dial(serverAddr, opts...)
		if err != nil {
			return nil, err
		}

		m.extensionConnCache = conn
	}

	client := extension.NewEnvoyGatewayExtensionClient(m.extensionConnCache)
	xdsHookClient := &XDSHook{
		grpcClient: client,
	}
	return xdsHookClient, nil
}

// GetPostXDSHookClient checks if the registered extension makes use of a particular hook type that modifies
// xDS resources after they are generated by Envoy Gateway.
// If the extension makes use of the hook then the XDS Hook Client is returned. If it does not support
// the hook type then nil is returned
func (m *Manager) GetPostXDSHookClient(xdsHookType egv1a1.XDSTranslatorHook) (extTypes.XDSHookClient, error) {
	ctx := context.Background()
	ext := m.extension

	if ext.Hooks == nil {
		return nil, nil
	}
	if ext.Hooks.XDSTranslator == nil {
		return nil, nil
	}

	hookUsed := false
	for _, hook := range ext.Hooks.XDSTranslator.Post {
		if xdsHookType == hook {
			hookUsed = true
			break
		}
	}
	if !hookUsed {
		return nil, nil
	}

	if m.extensionConnCache == nil {
		serverAddr := getExtensionServerAddress(ext.Service)

		opts, err := setupGRPCOpts(ctx, m.k8sClient, &ext, m.namespace)
		if err != nil {
			return nil, err
		}

		conn, err := grpc.Dial(serverAddr, opts...)
		if err != nil {
			return nil, err
		}

		m.extensionConnCache = conn
	}

	client := extension.NewEnvoyGatewayExtensionClient(m.extensionConnCache)
	xdsHookClient := &XDSHook{
		grpcClient: client,
	}
	return xdsHookClient, nil
}

func (m *Manager) CleanupHookConns() {
	if m.extensionConnCache != nil {
		m.extensionConnCache.Close()
	}
}

func setupGRPCOpts(ctx context.Context, client k8scli.Client, ext *egv1a1.ExtensionManager, namespace string) ([]grpc.DialOption, error) {
	// These two errors shouldn't happen since we check these conditions when loading the extension
	if ext == nil {
		return nil, errors.New("the registered extension's config is nil")
	}
	if ext.Service == nil {
		return nil, errors.New("the registered extension doesn't have a service config")
	}
	if ext.Service.TLS != nil && client == nil {
		return nil, errors.New("the registered extension's service config has TLS enabled but no k8s client was provided")
	}

	var opts []grpc.DialOption
	if ext.Service.TLS != nil {
		// Sanity check to ensure that the extension manager has a valid certificate reference
		_, err := getCertPoolFromSecret(ctx, client, ext, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get root CA certificates: %w", err)
		}
		creds, err := getGRPCCredentials(client, ext, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get gRPC TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	serviceConfig, err := buildServiceConfig(ext)
	if err != nil {
		return nil, err
	}
	opts = append(opts, grpc.WithDefaultServiceConfig(serviceConfig))

	if ext.Service.Retry != nil {
		maxAttempts := ptr.Deref(ext.Service.Retry.MaxAttempts, 4)
		opts = append(opts, grpc.WithMaxCallAttempts(maxAttempts))
	}

	if ext.MaxMessageSize != nil {
		maxMessageSize, ok := ext.MaxMessageSize.AsInt64()
		if !ok {
			return nil, fmt.Errorf("invalid Extension Manager MaxMessageSize value %s", ext.MaxMessageSize.String())
		}
		if maxMessageSize < 1 || maxMessageSize > math.MaxInt {
			return nil, fmt.Errorf("extension Manager MaxMessageSize value %s is out of range, must be between 1 and %d",
				ext.MaxMessageSize.String(), math.MaxInt)
		}
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(maxMessageSize)), grpc.MaxCallSendMsgSize(int(maxMessageSize))))
	}

	return opts, nil
}

func getGRPCCredentials(client k8scli.Client, ext *egv1a1.ExtensionManager, namespace string) (credentials.TransportCredentials, error) {
	return advancedtls.NewClientCreds(&advancedtls.Options{
		RootOptions: advancedtls.RootCertificateOptions{
			// A callback function that dynamically loads root CA certificates from secret
			GetRootCertificates: createGetRootCertificatesHandler(client, ext, namespace),
		},
	})
}

func createGetRootCertificatesHandler(client k8scli.Client, ext *egv1a1.ExtensionManager, namespace string) func(*advancedtls.ConnectionInfo) (*advancedtls.RootCertificates, error) {
	return func(params *advancedtls.ConnectionInfo) (*advancedtls.RootCertificates, error) {
		ctx := context.Background()
		cp, err := getCertPoolFromSecret(ctx, client, ext, namespace)
		if err != nil {
			return nil, err
		}

		return &advancedtls.RootCertificates{TrustCerts: cp}, nil
	}
}

func getCertPoolFromSecret(ctx context.Context, client k8scli.Client, ext *egv1a1.ExtensionManager, namespace string) (*x509.CertPool, error) {
	certRef := ext.Service.TLS.CertificateRef
	secret, _, err := kubernetes.ValidateSecretObjectReference(ctx, client, &certRef, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to validate TLS certificate reference: %w", err)
	}

	caCertPEMBytes, ok := secret.Data[corev1.TLSCertKey]
	if !ok {
		return nil, fmt.Errorf("no CA certificate found in Kubernetes Secret %s in namespace %s", secret.GetName(), secret.GetNamespace())
	}
	cp := x509.NewCertPool()
	if ok := cp.AppendCertsFromPEM(caCertPEMBytes); !ok {
		return nil, errors.New("failed to append certificates from CA secret")
	}
	return cp, nil
}

func fractionOrDefault(fraction *gwapiv1.Fraction, defaultValue float64) float64 {
	if fraction != nil {
		numerator := float64(fraction.Numerator)
		denominator := float64(100)
		if fraction.Denominator != nil {
			denominator = float64(*fraction.Denominator)
		}
		return numerator / denominator
	}
	return defaultValue
}

var retryStrToCode = map[string]codes.Code{
	`"CANCELLED"`:           codes.Canceled,
	`"UNKNOWN"`:             codes.Unknown,
	`"INVALID_ARGUMENT"`:    codes.InvalidArgument,
	`"DEADLINE_EXCEEDED"`:   codes.DeadlineExceeded,
	`"NOT_FOUND"`:           codes.NotFound,
	`"ALREADY_EXISTS"`:      codes.AlreadyExists,
	`"PERMISSION_DENIED"`:   codes.PermissionDenied,
	`"RESOURCE_EXHAUSTED"`:  codes.ResourceExhausted,
	`"FAILED_PRECONDITION"`: codes.FailedPrecondition,
	`"ABORTED"`:             codes.Aborted,
	`"OUT_OF_RANGE"`:        codes.OutOfRange,
	`"UNIMPLEMENTED"`:       codes.Unimplemented,
	`"INTERNAL"`:            codes.Internal,
	`"UNAVAILABLE"`:         codes.Unavailable,
	`"DATA_LOSS"`:           codes.DataLoss,
	`"UNAUTHENTICATED"`:     codes.Unauthenticated,
}

func getRetryableGRPCCode(retryableCodes []egv1a1.RetryableGRPCStatusCode) (string, error) {
	var quotedCodes []string
	for _, statusCode := range retryableCodes {
		quotedCode := strconv.Quote(string(statusCode))
		_, found := retryStrToCode[quotedCode]
		if !found {
			return "", fmt.Errorf("invalid Extension Manager GRPC Retry Status code value %s", statusCode)
		} else {
			quotedCodes = append(quotedCodes, quotedCode)
		}
	}

	return strings.Join(quotedCodes, ","), nil
}

func buildServiceConfig(ext *egv1a1.ExtensionManager) (string, error) {
	const defaultMaxAttempts = 4
	const defaultBackoffMultiplier = 2.0
	const defaultRetryableCodes = "UNAVAILABLE"

	defaultInitialBackoff := gwapiv1.Duration("100ms")
	defaultMaxBackoff := gwapiv1.Duration("1s")

	maxAttempts := defaultMaxAttempts
	initialBackoff := defaultInitialBackoff
	maxBackoff := defaultMaxBackoff
	backoffMultiplier := defaultBackoffMultiplier
	grpcRetryableStatusCodes := strconv.Quote(defaultRetryableCodes)

	if ext.Service.Retry != nil {
		maxAttempts = ptr.Deref(ext.Service.Retry.MaxAttempts, defaultMaxAttempts)
		initialBackoff = ptr.Deref(ext.Service.Retry.InitialBackoff, defaultInitialBackoff)
		maxBackoff = ptr.Deref(ext.Service.Retry.MaxBackoff, defaultMaxBackoff)
		backoffMultiplier = fractionOrDefault(ext.Service.Retry.BackoffMultiplier, defaultBackoffMultiplier)

		if len(ext.Service.Retry.RetryableStatusCodes) > 0 {
			var err error
			grpcRetryableStatusCodes, err = getRetryableGRPCCode(ext.Service.Retry.RetryableStatusCodes)
			if err != nil {
				return "", err
			}
		}
	}

	initialBackoffDuration, err := time.ParseDuration(string(initialBackoff))
	if err != nil {
		return "", fmt.Errorf("invalid Extension Manager GRPC Retry Initial Backoff %s", initialBackoff)
	}

	maxBackoffDuration, err := time.ParseDuration(string(maxBackoff))
	if err != nil {
		return "", fmt.Errorf("invalid Extension Manager GRPC Retry Max Backoff %s", maxBackoff)
	}

	return fmt.Sprintf(grpcServiceConfigTemplate, maxAttempts, initialBackoffDuration.Seconds(), maxBackoffDuration.Seconds(),
		backoffMultiplier, grpcRetryableStatusCodes), nil
}
