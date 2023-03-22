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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	corev1 "k8s.io/api/core/v1"
	k8smachinery "k8s.io/apimachinery/pkg/types"
	k8scli "sigs.k8s.io/controller-runtime/pkg/client"
	k8sclicfg "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	extTypes "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/proto/extension"
)

const grpcServiceConfig = `{
"methodConfig": [{
	"name": [{"service": "envoygateway.extension.EnvoyGatewayExtension"}],
	"waitForReady": true,
	"retryPolicy": {
		"MaxAttempts": 4,
		"InitialBackoff": "0.1s",
		"MaxBackoff": "1s",
		"BackoffMultiplier": 2.0,
		"RetryableStatusCodes": [ "UNAVAILABLE" ]
	}
}]}`

var _ extTypes.Manager = (*Manager)(nil)

type Manager struct {
	k8sClient     k8scli.Client
	namespace     string
	extension     v1alpha1.Extension
	hookConnCache map[extTypes.ExtensionXDSHookType]*grpc.ClientConn
}

// NewManager returns a new Manager
func NewManager(cfg *config.Server) (extTypes.Manager, error) {
	cli, err := k8scli.New(k8sclicfg.GetConfigOrDie(), k8scli.Options{Scheme: envoygateway.GetScheme()})
	if err != nil {
		return nil, err
	}

	var extension *v1alpha1.Extension
	if cfg.EnvoyGateway == nil {
		extension = cfg.EnvoyGateway.Extension
	}

	// Setup an empty default in the case that no config was provided
	if extension == nil {
		extension = &v1alpha1.Extension{}
	}

	hookConnCache := make(map[extTypes.ExtensionXDSHookType]*grpc.ClientConn)

	return &Manager{
		k8sClient:     cli,
		namespace:     cfg.Namespace,
		extension:     *extension,
		hookConnCache: hookConnCache,
	}, nil
}

// HasExtension checks to see whether a given Group and Kind has an
// associated extension registered for it.
func (m *Manager) HasExtension(g v1beta1.Group, k v1beta1.Kind) bool {
	extension := m.extension
	// TODO: not currently checking the version since extensionRef only supports group and kind.
	for _, gvk := range extension.Resources {
		if g == v1beta1.Group(gvk.Group) && k == v1beta1.Kind(gvk.Kind) {
			return true
		}
	}
	return false
}

// GetXDSHookClient checks if the registered extension makes use of a particular hook type.
// If the extension makes use of the hook then the XDS Hook Client is returned. If it does not support
// the hook type then nil is returned
func (m *Manager) GetXDSHookClient(xdsHookType extTypes.ExtensionXDSHookType) extTypes.XDSHookClient {
	ctx := context.Background()
	ext := m.extension

	if ext.Hooks == nil {
		return nil
	}
	if ext.Hooks.XDSTranslator == nil {
		return nil
	}

	hookUsed := false
	for _, hook := range ext.Hooks.XDSTranslator.Post {
		if xdsHookType == extTypes.ExtensionXDSHookType("Post"+hook) {
			hookUsed = true
			break
		}
	}
	for _, hook := range ext.Hooks.XDSTranslator.Pre {
		if xdsHookType == extTypes.ExtensionXDSHookType("Pre"+hook) {
			hookUsed = true
			break
		}
	}
	if !hookUsed {
		return nil
	}

	conn, cached := m.hookConnCache[xdsHookType]
	if !cached {
		serverAddr := fmt.Sprintf("%s:%d", ext.Service.Host, ext.Service.Port)

		opts, err := setupGRPCOpts(ctx, m.k8sClient, &ext, m.namespace)
		if err != nil {
			return nil
		}

		conn, err = grpc.Dial(serverAddr, opts...)
		if err != nil {
			return nil
		}

		m.hookConnCache[xdsHookType] = conn
	}

	client := extension.NewEnvoyGatewayExtensionClient(conn)
	xdsHookClient := &XDSHook{
		grpcClient: client,
	}
	return xdsHookClient
}

func (m *Manager) CleanupHookConns() {
	for _, conn := range m.hookConnCache {
		conn.Close()
	}
	m.hookConnCache = make(map[extTypes.ExtensionXDSHookType]*grpc.ClientConn)
}

func parseCA(caSecret *corev1.Secret) (*x509.CertPool, error) {
	caCertPEMBytes, ok := caSecret.Data[corev1.TLSCertKey]
	if !ok {
		return nil, errors.New("no cert found in CA secret")
	}
	cp := x509.NewCertPool()
	if ok := cp.AppendCertsFromPEM(caCertPEMBytes); !ok {
		return nil, errors.New("failed to append certificates")
	}
	return cp, nil
}

func setupGRPCOpts(ctx context.Context, client k8scli.Client, ext *v1alpha1.Extension, namespace string) ([]grpc.DialOption, error) {
	// These two errors shouldn't happen since we check these conditions when loading the extension
	if ext == nil {
		return nil, errors.New("the registered extension's config is nil")
	}
	if ext.Service == nil {
		return nil, errors.New("the registered extension doesn't have a service config")
	}

	var opts []grpc.DialOption
	var creds credentials.TransportCredentials
	if ext.Service.TLS != nil {
		certRef := ext.Service.TLS.CertificateRef
		if (certRef.Group == nil || *certRef.Group == corev1.GroupName) &&
			(certRef.Kind == nil || *certRef.Kind == gatewayapi.KindSecret) {
			secret := &corev1.Secret{}
			secretNamespace := namespace
			if certRef.Namespace != nil && string(*certRef.Namespace) != "" {
				secretNamespace = string(*certRef.Namespace)
			}
			key := k8smachinery.NamespacedName{
				Namespace: secretNamespace,
				Name:      string(certRef.Name),
			}
			if err := client.Get(ctx, key, secret); err != nil {
				return nil, fmt.Errorf("cannot find TLS Secret %s in namespace %s", string(certRef.Name), secretNamespace)
			}
			cp, err := parseCA(secret)
			if err != nil {
				return nil, fmt.Errorf("error parsing cert in Secret %s in namespace %s", string(certRef.Name), secretNamespace)
			}
			creds = credentials.NewClientTLSFromCert(cp, "")

		} else {
			return nil, errors.New("unsupported Extension TLS certificateRef group/kind")
		}

		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	opts = append(opts, grpc.WithDefaultServiceConfig(grpcServiceConfig))
	return opts, nil
}
