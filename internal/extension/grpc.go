package extension

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/kubernetes"
	"github.com/envoyproxy/gateway/internal/utils/fraction"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/security/advancedtls"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
	k8scli "sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const grpcServiceConfigTemplate = `{
"methodConfig": [{
	"name": [{"service": "%s"}],
	"waitForReady": true,
	"retryPolicy": {
		"MaxAttempts": %d,
		"InitialBackoff": "%fs",
		"MaxBackoff": "%fs",
		"BackoffMultiplier": %f,
		"RetryableStatusCodes": [ %s ]
	}
}]}`

// GetExtensionServerAddress parses the address from an ExtensionService.
func GetExtensionServerAddress(service *egv1a1.ExtensionService) string {
	var serverAddr string
	switch {
	case service.FQDN != nil:
		serverAddr = net.JoinHostPort(service.FQDN.Hostname, strconv.Itoa(int(service.FQDN.Port)))
	case service.IP != nil:
		serverAddr = net.JoinHostPort(service.IP.Address, strconv.Itoa(int(service.IP.Port)))
	case service.Unix != nil:
		serverAddr = fmt.Sprintf("unix://%s", service.Unix.Path)
	case service.Hostname != nil:
		serverAddr = net.JoinHostPort(*service.Hostname, strconv.Itoa(int(service.Port)))
	}
	return serverAddr
}

// GenerateGRPCOptions generates the necessary connection descriptors for a GRPC client.
func GenerateGRPCOptions(ctx context.Context, client k8scli.Client, ext *egv1a1.ExtensionService, maxMessageSize *resource.Quantity, svcName string, namespace string) ([]grpc.DialOption, error) {
	// These two errors shouldn't happen since we check these conditions when loading the extension
	if ext == nil {
		return nil, errors.New("the registered extension's config is nil")
	}
	if ext.TLS != nil && client == nil {
		return nil, errors.New("the registered extension's service config has TLS enabled but no k8s client was provided")
	}

	var opts []grpc.DialOption
	if ext.TLS != nil {
		// Sanity check to ensure that the extension manager has a valid certificate reference
		_, err := getCertPoolFromSecret(ctx, client, ext, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get root CA certificates: %w", err)
		}

		// Sanity check to ensure that the client certificate reference is valid if mTLS is configured
		if ext.TLS.ClientCertificateRef != nil {
			_, clientCertErr := getClientCertificateFromSecret(ctx, client, ext, namespace)
			if clientCertErr != nil {
				return nil, fmt.Errorf("failed to get client certificate for mTLS: %w", clientCertErr)
			}
		}

		creds, err := getGRPCCredentials(client, ext, namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get gRPC TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	serviceConfig, err := buildServiceConfig(svcName, ext)
	if err != nil {
		return nil, err
	}
	opts = append(opts, grpc.WithDefaultServiceConfig(serviceConfig))

	if ext.Retry != nil {
		maxAttempts := ptr.Deref(ext.Retry.MaxAttempts, 4)
		opts = append(opts, grpc.WithMaxCallAttempts(maxAttempts))
	}

	if maxMessageSize != nil {
		maxMessageSizeInt64, ok := maxMessageSize.AsInt64()
		if !ok {
			return nil, fmt.Errorf("invalid Extension Manager MaxMessageSize value %s", maxMessageSize.String())
		}
		if maxMessageSizeInt64 < 1 || maxMessageSizeInt64 > math.MaxInt {
			return nil, fmt.Errorf("extension Manager MaxMessageSize value %s is out of range, must be between 1 and %d",
				maxMessageSize.String(), math.MaxInt)
		}
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(maxMessageSizeInt64)), grpc.MaxCallSendMsgSize(int(maxMessageSizeInt64))))
	}

	return opts, nil
}

func getGRPCCredentials(client k8scli.Client, ext *egv1a1.ExtensionService, namespace string) (credentials.TransportCredentials, error) {
	options := &advancedtls.Options{
		RootOptions: advancedtls.RootCertificateOptions{
			// A callback function that dynamically loads root CA certificates from secret
			GetRootCertificates: createGetRootCertificatesHandler(client, ext, namespace),
		},
	}

	// Add client certificate options for mTLS if configured
	if ext.TLS.ClientCertificateRef != nil {
		options.IdentityOptions = advancedtls.IdentityCertificateOptions{
			GetIdentityCertificatesForClient: createGetClientCertificatesHandler(client, ext, namespace),
		}
	}

	return advancedtls.NewClientCreds(options)
}

func createGetRootCertificatesHandler(client k8scli.Client, ext *egv1a1.ExtensionService, namespace string) func(*advancedtls.ConnectionInfo) (*advancedtls.RootCertificates, error) {
	return func(_ *advancedtls.ConnectionInfo) (*advancedtls.RootCertificates, error) {
		ctx := context.Background()
		cp, err := getCertPoolFromSecret(ctx, client, ext, namespace)
		if err != nil {
			return nil, err
		}

		return &advancedtls.RootCertificates{TrustCerts: cp}, nil
	}
}

func getCertPoolFromSecret(ctx context.Context, client k8scli.Client, ext *egv1a1.ExtensionService, namespace string) (*x509.CertPool, error) {
	certRef := ext.TLS.CertificateRef
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

func createGetClientCertificatesHandler(client k8scli.Client, ext *egv1a1.ExtensionService, namespace string) func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
	return func(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		ctx := context.Background()
		cert, err := getClientCertificateFromSecret(ctx, client, ext, namespace)
		if err != nil {
			return nil, err
		}
		return cert, nil
	}
}

func getClientCertificateFromSecret(ctx context.Context, client k8scli.Client, ext *egv1a1.ExtensionService, namespace string) (*tls.Certificate, error) {
	if ext.TLS.ClientCertificateRef == nil {
		return nil, errors.New("client certificate reference is nil")
	}

	certRef := *ext.TLS.ClientCertificateRef
	secret, _, err := kubernetes.ValidateSecretObjectReference(ctx, client, &certRef, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to validate client certificate reference: %w", err)
	}

	certPEMBytes, ok := secret.Data[corev1.TLSCertKey]
	if !ok {
		return nil, fmt.Errorf("no client certificate found in Kubernetes Secret %s in namespace %s", secret.GetName(), secret.GetNamespace())
	}

	keyPEMBytes, ok := secret.Data[corev1.TLSPrivateKeyKey]
	if !ok {
		return nil, fmt.Errorf("no client private key found in Kubernetes Secret %s in namespace %s", secret.GetName(), secret.GetNamespace())
	}

	cert, err := tls.X509KeyPair(certPEMBytes, keyPEMBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate and key: %w", err)
	}

	return &cert, nil
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

func buildServiceConfig(svcName string, ext *egv1a1.ExtensionService) (string, error) {
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

	if ext.Retry != nil {
		maxAttempts = ptr.Deref(ext.Retry.MaxAttempts, defaultMaxAttempts)
		initialBackoff = ptr.Deref(ext.Retry.InitialBackoff, defaultInitialBackoff)
		maxBackoff = ptr.Deref(ext.Retry.MaxBackoff, defaultMaxBackoff)
		backoffMultiplier = fraction.Deref(ext.Retry.BackoffMultiplier, defaultBackoffMultiplier)

		if len(ext.Retry.RetryableStatusCodes) > 0 {
			var err error
			grpcRetryableStatusCodes, err = getRetryableGRPCCode(ext.Retry.RetryableStatusCodes)
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

	return fmt.Sprintf(grpcServiceConfigTemplate, svcName, maxAttempts, initialBackoffDuration.Seconds(), maxBackoffDuration.Seconds(),
		backoffMultiplier, grpcRetryableStatusCodes), nil
}
