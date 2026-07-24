// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	k8scli "sigs.k8s.io/controller-runtime/pkg/client"
	k8sclicfg "sigs.k8s.io/controller-runtime/pkg/client/config"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	grpcExtension "github.com/envoyproxy/gateway/internal/extension"
	extTypes "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/proto/extension"
)

const serviceName = "envoygateway.extension.EnvoyGatewayExtension"

var _ extTypes.Manager = (*Manager)(nil)

type Manager struct {
	k8sClient          k8scli.Client
	namespace          string
	extension          egv1a1.ExtensionManager
	extensionConnCache *grpc.ClientConn
}

// newK8sClient creates a Kubernetes client if running in-cluster.
func newK8sClient(inK8s bool) (k8scli.Client, error) {
	if !inK8s {
		return nil, nil
	}
	return k8scli.New(k8sclicfg.GetConfigOrDie(), k8scli.Options{Scheme: envoygateway.GetScheme()})
}

// NewManager creates a Manager (or CompositeManager) from the server configuration.
// It uses GetExtensionManagers() to normalize the singular/plural extension manager fields.
//   - 0 extensions → returns a Manager with empty config (no-op)
//   - 1 extension → returns a plain Manager
//   - 2+ extensions → creates individual Managers per extension, wraps in CompositeManager
func NewManager(cfg *config.Server, inK8s bool) (extTypes.Manager, error) {
	cli, err := newK8sClient(inK8s)
	if err != nil {
		return nil, err
	}

	extensions := cfg.EnvoyGateway.GetExtensionManagers()

	switch len(extensions) {
	case 0:
		return &Manager{
			k8sClient: cli,
			namespace: cfg.ControllerNamespace,
			extension: egv1a1.ExtensionManager{},
		}, nil
	case 1:
		return &Manager{
			k8sClient: cli,
			namespace: cfg.ControllerNamespace,
			extension: extensions[0],
		}, nil
	default:
		named := make([]namedManager, 0, len(extensions))
		for i := range extensions {
			ext := &extensions[i]
			mgr := &Manager{
				k8sClient: cli,
				namespace: cfg.ControllerNamespace,
				extension: *ext,
			}

			resourceGKSet, policyGKSet := buildManagerGKSets(ext)

			named = append(named, namedManager{
				name:            ext.Name,
				manager:         mgr,
				resourceGKSet:   resourceGKSet,
				policyGKSet:     policyGKSet,
				cleanupHookConn: mgr.CleanupHookConns,
			})
		}

		return NewCompositeManager(named), nil
	}
}

// buildManagerGKSets returns (resourceGKSet, policyGKSet) for an ExtensionManager.
// resourceGKSet covers Resources + BackendResources (used for per-extension filtering
// in PostRouteModifyHook / PostClusterModifyHook). policyGKSet covers PolicyResources
// (used in PostHTTPListenerModifyHook / PostTranslateModifyHook).
// Version is intentionally dropped so matching aligns with runner.ExtensionGroupKinds
// and Manager.HasExtension, which also compare by group+kind only.
func buildManagerGKSets(ext *egv1a1.ExtensionManager) (sets.Set[schema.GroupKind], sets.Set[schema.GroupKind]) {
	resourceGKSet := sets.New[schema.GroupKind]()
	for _, gvk := range ext.Resources {
		resourceGKSet.Insert(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind})
	}
	for _, gvk := range ext.BackendResources {
		resourceGKSet.Insert(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind})
	}

	policyGKSet := sets.New[schema.GroupKind]()
	for _, gvk := range ext.PolicyResources {
		policyGKSet.Insert(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind})
	}
	return resourceGKSet, policyGKSet
}

func NewInMemoryManager(cfg *egv1a1.ExtensionManager, server extension.EnvoyGatewayExtensionServer) (extTypes.Manager, func(), error) {
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
		opts, err := grpcExtension.GenerateGRPCOptions(context.Background(), nil, cfg.Service, cfg.MaxMessageSize, serviceName, "")
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
		extension:          *cfg,
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

		opts, err := grpcExtension.GenerateGRPCOptions(ctx, m.k8sClient, ext.Service, ext.MaxMessageSize, serviceName, m.namespace)
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

		opts, err := grpcExtension.GenerateGRPCOptions(ctx, m.k8sClient, ext.Service, ext.MaxMessageSize, serviceName, m.namespace)
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
