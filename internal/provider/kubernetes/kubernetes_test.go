//go:build integration
// +build integration

// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/utils"
)

const (
	defaultWait = time.Second * 60
	defaultTick = time.Millisecond * 20
)

func TestProvider(t *testing.T) {
	// Setup the test environment.
	testEnv, cliCfg, err := startEnv()
	require.NoError(t, err)

	// Setup and start the kube provider.
	svr, err := config.NewDefaultServer()
	require.NoError(t, err)
	resources := new(message.ProviderResources)
	provider, err := New(cliCfg, svr, resources)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(ctrl.SetupSignalHandler())
	go func() {
		require.NoError(t, provider.Start(ctx))
	}()

	// Stop the kube provider.
	defer func() {
		cancel()
		require.NoError(t, testEnv.Stop())
	}()

	testcases := map[string]func(context.Context, *testing.T, *Provider, *message.ProviderResources){
		"gatewayclass controller name":         testGatewayClassController,
		"gatewayclass accepted status":         testGatewayClassAcceptedStatus,
		"gateway scheduled status":             testGatewayScheduledStatus,
		"httproute":                            testHTTPRoute,
		"tlsroute":                             testTLSRoute,
		"stale service cleanup route deletion": testServiceCleanupForMultipleRoutes,
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			tc(ctx, t, provider, resources)
		})
	}
}

func startEnv() (*envtest.Environment, *rest.Config, error) {
	log.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))
	crd := filepath.Join(".", "testdata", "in")
	env := &envtest.Environment{
		CRDDirectoryPaths: []string{crd},
	}
	cfg, err := env.Start()
	if err != nil {
		return env, nil, err
	}
	return env, cfg, nil
}

func getGatewayClass(name string) *gwapiv1b1.GatewayClass {
	return &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
		},
	}
}

func getService(name, namespace string, ports map[string]int32) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{},
		},
	}
	for name, port := range ports {
		service.Spec.Ports = append(service.Spec.Ports, corev1.ServicePort{
			Name: name,
			Port: port,
		})
	}
	return service
}

func testGatewayClassController(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := getGatewayClass("test-gc-controllername")
	require.NoError(t, cli.Create(ctx, gc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	require.Eventually(t, func() bool {
		return cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc) == nil
	}, defaultWait, defaultTick)
	assert.Equal(t, gc.ObjectMeta.Generation, int64(1))
}

func testGatewayClassAcceptedStatus(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := getGatewayClass("test-gc-accepted-status")
	require.NoError(t, cli.Create(ctx, gc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc); err != nil {
			return false
		}

		for _, cond := range gc.Status.Conditions {
			if cond.Type == string(gwapiv1b1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
				return true
			}
		}

		return false
	}, defaultWait, defaultTick)

	// Ensure the GatewayClass resource map contains the gatewayclass under test.
	gcs, _ := resources.GatewayClasses.Load(gc.Name)
	assert.Equal(t, gc, gcs)
}

func testGatewayScheduledStatus(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := getGatewayClass("gc-scheduled-status-test")
	require.NoError(t, cli.Create(ctx, gc))

	// Ensure the GatewayClass reports "Ready".
	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc); err != nil {
			return false
		}

		for _, cond := range gc.Status.Conditions {
			if cond.Type == string(gwapiv1b1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
				return true
			}
		}

		return false
	}, defaultWait, defaultTick)

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	// Create the namespace for the Gateway under test.
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-gw-of-class"}}
	require.NoError(t, cli.Create(ctx, ns))

	gw := &gwapiv1b1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "scheduled-status-test",
			Namespace: ns.Name,
		},
		Spec: gwapiv1b1.GatewaySpec{
			GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
			Listeners: []gwapiv1b1.Listener{
				{
					Name:     "test",
					Port:     gwapiv1b1.PortNumber(int32(8080)),
					Protocol: gwapiv1b1.HTTPProtocolType,
				},
			},
		},
	}
	require.NoError(t, cli.Create(ctx, gw))

	// Ensure the Gateway reports "Scheduled".
	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, types.NamespacedName{Namespace: gw.Namespace, Name: gw.Name}, gw); err != nil {
			return false
		}

		for _, cond := range gw.Status.Conditions {
			fmt.Printf("Condition: %v", cond)
			if cond.Type == string(gwapiv1b1.GatewayConditionScheduled) && cond.Status == metav1.ConditionTrue {
				return true
			}
		}

		// Scheduled=True condition not found.
		return false
	}, defaultWait, defaultTick)

	defer func() {
		require.NoError(t, cli.Delete(ctx, gw))
	}()

	// Ensure the gatewayclass has been finalized.
	require.NoError(t, cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc))
	require.Contains(t, gc.Finalizers, gatewayClassFinalizer)

	// Ensure the number of Gateways in the Gateway resource table is as expected.
	require.Eventually(t, func() bool {
		return resources.Gateways.Len() == 1
	}, defaultWait, defaultTick)

	// Ensure the test Gateway in the Gateway resources is as expected.
	key := types.NamespacedName{
		Namespace: gw.Namespace,
		Name:      gw.Name,
	}
	require.Eventually(t, func() bool {
		return cli.Get(ctx, key, gw) == nil
	}, defaultWait, defaultTick)
	gws, _ := resources.Gateways.Load(key)
	// Only check if the spec is equal
	// The watchable map will not store a resource
	// with an updated status if the spec has not changed
	// to eliminate this endless loop:
	// reconcile->store->translate->update-status->reconcile
	assert.Equal(t, gw.Spec, gws.Spec)
}

// Test that even when resources such as the Service/Deployment get hashed names (because of a gateway with a very long name)
func testLongNameHashedResources(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := getGatewayClass("envoy-gateway-class")
	require.NoError(t, cli.Create(ctx, gc))

	// Ensure the GatewayClass reports "Ready".
	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc); err != nil {
			return false
		}

		for _, cond := range gc.Status.Conditions {
			if cond.Type == string(gwapiv1b1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
				return true
			}
		}

		return false
	}, defaultWait, defaultTick)

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	// Create the namespace for the Gateway under test.
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "envoy-gateway"}}
	require.NoError(t, cli.Create(ctx, ns))

	gw := &gwapiv1b1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gatewaywithaverylongnamethatwillresultinhashedresources",
			Namespace: ns.Name,
		},
		Spec: gwapiv1b1.GatewaySpec{
			GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
			Listeners: []gwapiv1b1.Listener{
				{
					Name:     "test",
					Port:     gwapiv1b1.PortNumber(int32(8080)),
					Protocol: gwapiv1b1.HTTPProtocolType,
				},
			},
		},
	}
	require.NoError(t, cli.Create(ctx, gw))

	// Ensure the Gateway is ready and gets an address.
	ready := false
	hasAddress := false
	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, types.NamespacedName{Namespace: gw.Namespace, Name: gw.Name}, gw); err != nil {
			return false
		}

		for _, cond := range gw.Status.Conditions {
			fmt.Printf("Condition: %v", cond)
			if cond.Type == string(gwapiv1b1.GatewayConditionReady) && cond.Status == metav1.ConditionTrue {
				ready = true
			}
		}

		if gw.Status.Addresses != nil {
			hasAddress = len(gw.Status.Addresses) >= 1
		}

		return ready && hasAddress
	}, defaultWait, defaultTick)

	defer func() {
		require.NoError(t, cli.Delete(ctx, gw))
	}()

	// Ensure the gatewayclass has been finalized.
	require.NoError(t, cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc))
	require.Contains(t, gc.Finalizers, gatewayClassFinalizer)

	// Ensure the number of Gateways in the Gateway resource table is as expected.
	require.Eventually(t, func() bool {
		return resources.Gateways.Len() == 1
	}, defaultWait, defaultTick)

	// Ensure the test Gateway in the Gateway resources is as expected.
	key := types.NamespacedName{
		Namespace: gw.Namespace,
		Name:      gw.Name,
	}
	require.Eventually(t, func() bool {
		return cli.Get(ctx, key, gw) == nil
	}, defaultWait, defaultTick)
	gws, _ := resources.Gateways.Load(key)
	// Only check if the spec is equal
	// The watchable map will not store a resource
	// with an updated status if the spec has not changed
	// to eliminate this endless loop:
	// reconcile->store->translate->update-status->reconcile
	assert.Equal(t, gw.Spec, gws.Spec)
}

func testHTTPRoute(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := getGatewayClass("httproute-test")
	require.NoError(t, cli.Create(ctx, gc))

	// Ensure the GatewayClass reports ready.
	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc); err != nil {
			return false
		}

		for _, cond := range gc.Status.Conditions {
			if cond.Type == string(gwapiv1b1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
				return true
			}
		}

		return false
	}, defaultWait, defaultTick)

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	// Create the namespace for the Gateway under test.
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "httproute-test"}}
	require.NoError(t, cli.Create(ctx, ns))

	gw := &gwapiv1b1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httproute-test",
			Namespace: ns.Name,
		},
		Spec: gwapiv1b1.GatewaySpec{
			GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
			Listeners: []gwapiv1b1.Listener{
				{
					Name:     "test",
					Port:     gwapiv1b1.PortNumber(int32(8080)),
					Protocol: gwapiv1b1.HTTPProtocolType,
				},
			},
		},
	}
	require.NoError(t, cli.Create(ctx, gw))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gw))
	}()

	svc := getService("test", ns.Name, map[string]int32{
		"http":  80,
		"https": 443,
	})

	require.NoError(t, cli.Create(ctx, svc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, svc))
	}()

	redirectHostname := gwapiv1b1.PreciseHostname("redirect.hostname.local")
	redirectPort := gwapiv1b1.PortNumber(8443)
	redirectStatus := 301

	var testCases = []struct {
		name  string
		route gwapiv1b1.HTTPRoute
	}{
		{
			name: "destination-httproute",
			route: gwapiv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1b1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
						ParentRefs: []gwapiv1b1.ParentReference{
							{
								Name: gwapiv1b1.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1b1.Hostname{"test.hostname.local"},
					Rules: []gwapiv1b1.HTTPRouteRule{
						{
							Matches: []gwapiv1b1.HTTPRouteMatch{
								{
									Path: &gwapiv1b1.HTTPPathMatch{
										Type:  gatewayapi.PathMatchTypePtr(gwapiv1b1.PathMatchPathPrefix),
										Value: gatewayapi.StringPtr("/"),
									},
								},
							},
							BackendRefs: []gwapiv1b1.HTTPBackendRef{
								{
									BackendRef: gwapiv1b1.BackendRef{
										BackendObjectReference: gwapiv1b1.BackendObjectReference{
											Name: "test",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "redirect-httproute",
			route: gwapiv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-redirect-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1b1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
						ParentRefs: []gwapiv1b1.ParentReference{
							{
								Name: gwapiv1b1.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1b1.Hostname{"test.hostname.local"},
					Rules: []gwapiv1b1.HTTPRouteRule{
						{
							Matches: []gwapiv1b1.HTTPRouteMatch{
								{
									Path: &gwapiv1b1.HTTPPathMatch{
										Type:  gatewayapi.PathMatchTypePtr(gwapiv1b1.PathMatchPathPrefix),
										Value: gatewayapi.StringPtr("/redirect/"),
									},
								},
							},
							BackendRefs: []gwapiv1b1.HTTPBackendRef{
								{
									BackendRef: gwapiv1b1.BackendRef{
										BackendObjectReference: gwapiv1b1.BackendObjectReference{
											Name: "test",
										},
									},
								},
							},
							Filters: []gwapiv1b1.HTTPRouteFilter{
								{
									Type: gwapiv1b1.HTTPRouteFilterType("RequestRedirect"),
									RequestRedirect: &gwapiv1b1.HTTPRequestRedirectFilter{
										Scheme:   gatewayapi.StringPtr("https"),
										Hostname: &redirectHostname,
										Path: &gwapiv1b1.HTTPPathModifier{
											Type:            gwapiv1b1.HTTPPathModifierType("ReplaceFullPath"),
											ReplaceFullPath: gatewayapi.StringPtr("/newpath"),
										},
										Port:       &redirectPort,
										StatusCode: &redirectStatus,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "addheader-httproute",
			route: gwapiv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-addheader-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1b1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
						ParentRefs: []gwapiv1b1.ParentReference{
							{
								Name: gwapiv1b1.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1b1.Hostname{"test.hostname.local"},
					Rules: []gwapiv1b1.HTTPRouteRule{
						{
							Matches: []gwapiv1b1.HTTPRouteMatch{
								{
									Path: &gwapiv1b1.HTTPPathMatch{
										Type:  gatewayapi.PathMatchTypePtr(gwapiv1b1.PathMatchPathPrefix),
										Value: gatewayapi.StringPtr("/addheader/"),
									},
								},
							},
							BackendRefs: []gwapiv1b1.HTTPBackendRef{
								{
									BackendRef: gwapiv1b1.BackendRef{
										BackendObjectReference: gwapiv1b1.BackendObjectReference{
											Name: "test",
										},
									},
								},
							},
							Filters: []gwapiv1b1.HTTPRouteFilter{
								{
									Type: gwapiv1b1.HTTPRouteFilterType("RequestHeaderModifier"),
									RequestHeaderModifier: &gwapiv1b1.HTTPRequestHeaderFilter{
										Add: []gwapiv1b1.HTTPHeader{
											{
												Name:  gwapiv1b1.HTTPHeaderName("header-1"),
												Value: "value-1",
											},
											{
												Name:  gwapiv1b1.HTTPHeaderName("header-2"),
												Value: "value-2",
											},
										},
										Set: []gwapiv1b1.HTTPHeader{
											{
												Name:  gwapiv1b1.HTTPHeaderName("header-3"),
												Value: "value-3",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "remheader-httproute",
			route: gwapiv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-remheader-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1b1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
						ParentRefs: []gwapiv1b1.ParentReference{
							{
								Name: gwapiv1b1.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1b1.Hostname{"test.hostname.local"},
					Rules: []gwapiv1b1.HTTPRouteRule{
						{
							Matches: []gwapiv1b1.HTTPRouteMatch{
								{
									Path: &gwapiv1b1.HTTPPathMatch{
										Type:  gatewayapi.PathMatchTypePtr(gwapiv1b1.PathMatchPathPrefix),
										Value: gatewayapi.StringPtr("/remheader/"),
									},
								},
							},
							BackendRefs: []gwapiv1b1.HTTPBackendRef{
								{
									BackendRef: gwapiv1b1.BackendRef{
										BackendObjectReference: gwapiv1b1.BackendObjectReference{
											Name: "test",
										},
									},
								},
							},
							Filters: []gwapiv1b1.HTTPRouteFilter{
								{
									Type: gwapiv1b1.HTTPRouteFilterType("RequestHeaderModifier"),
									RequestHeaderModifier: &gwapiv1b1.HTTPRequestHeaderFilter{
										Remove: []string{
											"example-header-1",
											"test-header",
											"example",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.NoError(t, cli.Create(ctx, &testCase.route))
			defer func() {
				require.NoError(t, cli.Delete(ctx, &testCase.route))
			}()

			require.Eventually(t, func() bool {
				return resources.HTTPRoutes.Len() == 1
			}, defaultWait, defaultTick)

			// Ensure the test HTTPRoute in the HTTPRoute resources is as expected.
			key := types.NamespacedName{
				Namespace: testCase.route.Namespace,
				Name:      testCase.route.Name,
			}
			require.Eventually(t, func() bool {
				return cli.Get(ctx, key, &testCase.route) == nil
			}, defaultWait, defaultTick)
			hroutes, _ := resources.HTTPRoutes.Load(key)
			assert.Equal(t, &testCase.route, hroutes)

			// Ensure the HTTPRoute Namespace is in the Namespace resource map.
			require.Eventually(t, func() bool {
				_, ok := resources.Namespaces.Load(testCase.route.Namespace)
				return ok
			}, defaultWait, defaultTick)

			// Ensure the Service is in the resource map.
			svcKey := utils.NamespacedName(svc)
			require.Eventually(t, func() bool {
				_, ok := resources.Services.Load(svcKey)
				return ok
			}, defaultWait, defaultTick)

		})
	}
}

func testTLSRoute(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := getGatewayClass("tlsroute-test")
	require.NoError(t, cli.Create(ctx, gc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	// Create the namespace for the Gateway under test.
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "tlsroute-test"}}
	require.NoError(t, cli.Create(ctx, ns))

	gw := &gwapiv1b1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tlsroute-test",
			Namespace: ns.Name,
		},
		Spec: gwapiv1b1.GatewaySpec{
			GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
			Listeners: []gwapiv1b1.Listener{
				{
					Name:     "test",
					Port:     gwapiv1b1.PortNumber(int32(8080)),
					Protocol: gwapiv1b1.TLSProtocolType,
				},
			},
		},
	}
	require.NoError(t, cli.Create(ctx, gw))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gw))
	}()

	svc := getService("test", ns.Name, map[string]int32{
		"tls": 90,
	})
	require.NoError(t, cli.Create(ctx, svc))
	defer func() {
		require.NoError(t, cli.Delete(ctx, svc))
	}()

	var testCases = []struct {
		name  string
		route gwapiv1a2.TLSRoute
	}{
		{
			name: "tlsroute",
			route: gwapiv1a2.TLSRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tlsroute-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1a2.TLSRouteSpec{
					CommonRouteSpec: gwapiv1a2.CommonRouteSpec{
						ParentRefs: []gwapiv1a2.ParentReference{
							{
								Name: gwapiv1a2.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1a2.Hostname{"test.hostname.local"},
					Rules: []gwapiv1a2.TLSRouteRule{
						{
							BackendRefs: []gwapiv1a2.BackendRef{
								{
									BackendObjectReference: gwapiv1a2.BackendObjectReference{
										Name: "test",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.NoError(t, cli.Create(ctx, &testCase.route))
			defer func() {
				require.NoError(t, cli.Delete(ctx, &testCase.route))
			}()

			require.Eventually(t, func() bool {
				return resources.TLSRoutes.Len() == 1
			}, defaultWait, defaultTick)

			// Ensure the test TLSRoute in the TLSRoute resources is as expected.
			key := types.NamespacedName{
				Namespace: testCase.route.Namespace,
				Name:      testCase.route.Name,
			}
			require.Eventually(t, func() bool {
				return cli.Get(ctx, key, &testCase.route) == nil
			}, defaultWait, defaultTick)
			troutes, _ := resources.TLSRoutes.Load(key)
			assert.Equal(t, &testCase.route, troutes)

			// Ensure the TLSRoute Namespace is in the Namespace resource map.
			require.Eventually(t, func() bool {
				_, ok := resources.Namespaces.Load(testCase.route.Namespace)
				return ok
			}, defaultWait, defaultTick)

			// Ensure the Service is in the resource map.
			svcKey := utils.NamespacedName(svc)
			require.Eventually(t, func() bool {
				_, ok := resources.Services.Load(svcKey)
				return ok
			}, defaultWait, defaultTick)
		})
	}
}

// testServiceCleanupForMultipleRoutes creates multiple Routes pointing to the
// same backend Service, and checks whether the Service is properly removed
// from the resource map after Route deletion.
func testServiceCleanupForMultipleRoutes(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := getGatewayClass("service-cleanup-test")
	require.NoError(t, cli.Create(ctx, gc))
	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	// Create the namespace for the Gateway under test.
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "service-cleanup-test"}}
	require.NoError(t, cli.Create(ctx, ns))

	gw := &gwapiv1b1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-cleanup-test",
			Namespace: ns.Name,
		},
		Spec: gwapiv1b1.GatewaySpec{
			GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
			Listeners: []gwapiv1b1.Listener{
				{
					Name:     "httptest",
					Port:     gwapiv1b1.PortNumber(int32(8080)),
					Protocol: gwapiv1b1.HTTPProtocolType,
				},
				{
					Name:     "tlstest",
					Port:     gwapiv1b1.PortNumber(int32(8043)),
					Protocol: gwapiv1b1.TLSProtocolType,
				},
			},
		},
	}
	require.NoError(t, cli.Create(ctx, gw))
	defer func() {
		require.NoError(t, cli.Delete(ctx, gw))
	}()

	svc := getService("test-common-svc", ns.Name, map[string]int32{
		"http": 80,
		"tls":  90,
	})
	require.NoError(t, cli.Create(ctx, svc))
	defer func() {
		require.NoError(t, cli.Delete(ctx, svc))
	}()

	tlsRoute := gwapiv1a2.TLSRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tlsroute-test",
			Namespace: ns.Name,
		},
		Spec: gwapiv1a2.TLSRouteSpec{
			CommonRouteSpec: gwapiv1a2.CommonRouteSpec{
				ParentRefs: []gwapiv1a2.ParentReference{{
					Name: gwapiv1a2.ObjectName(gw.Name),
				}},
			},
			Hostnames: []gwapiv1a2.Hostname{"test-tls.hostname.local"},
			Rules: []gwapiv1a2.TLSRouteRule{{
				BackendRefs: []gwapiv1a2.BackendRef{{
					BackendObjectReference: gwapiv1a2.BackendObjectReference{
						Name: "test-common-svc",
					}},
				}},
			},
		},
	}

	httpRoute := gwapiv1b1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httproute-test",
			Namespace: ns.Name,
		},
		Spec: gwapiv1b1.HTTPRouteSpec{
			CommonRouteSpec: gwapiv1b1.CommonRouteSpec{
				ParentRefs: []gwapiv1b1.ParentReference{{
					Name: gwapiv1b1.ObjectName(gw.Name),
				}},
			},
			Hostnames: []gwapiv1b1.Hostname{"test-http.hostname.local"},
			Rules: []gwapiv1b1.HTTPRouteRule{{
				Matches: []gwapiv1b1.HTTPRouteMatch{{
					Path: &gwapiv1b1.HTTPPathMatch{
						Type:  gatewayapi.PathMatchTypePtr(gwapiv1b1.PathMatchPathPrefix),
						Value: gatewayapi.StringPtr("/"),
					},
				}},
				BackendRefs: []gwapiv1b1.HTTPBackendRef{{
					BackendRef: gwapiv1b1.BackendRef{
						BackendObjectReference: gwapiv1b1.BackendObjectReference{
							Name: "test-common-svc",
						},
					},
				}},
			}},
		},
	}

	// Create the TLSRoute and HTTPRoute
	require.NoError(t, cli.Create(ctx, &tlsRoute))
	require.NoError(t, cli.Create(ctx, &httpRoute))

	// Check that the Service is present in the resource map
	key := types.NamespacedName{
		Namespace: svc.Namespace,
		Name:      svc.Name,
	}

	require.Eventually(t, func() bool {
		rSvc, _ := resources.Services.Load(key)
		return rSvc != nil
	}, defaultWait, defaultTick)

	// Delete the TLSRoute, and check if the Service is still present
	require.NoError(t, cli.Delete(ctx, &tlsRoute))
	require.Eventually(t, func() bool {
		rSvc, _ := resources.Services.Load(key)
		return rSvc != nil
	}, defaultWait, defaultTick)

	// Delete the HTTPRoute, and check if the Service is also removed
	require.NoError(t, cli.Delete(ctx, &httpRoute))
	require.Eventually(t, func() bool {
		rSvc, _ := resources.Services.Load(key)
		return rSvc == nil
	}, defaultWait, defaultTick)
}
