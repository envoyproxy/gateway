//go:build integration
// +build integration

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
		"gatewayclass controller name": testGatewayClassController,
		"gatewayclass accepted status": testGatewayClassAcceptedStatus,
		"gateway scheduled status":     testGatewayScheduledStatus,
		"httproute":                    testHTTPRoute,
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

func testGatewayClassController(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gc-controllername",
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: v1alpha1.GatewayControllerName,
		},
	}
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

	gc := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gc-accepted-status",
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: v1alpha1.GatewayControllerName,
		},
	}
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

	gc := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gc-scheduled-status-test",
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
		},
	}
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

func testHTTPRoute(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "httproute-test",
		},
		Spec: gwapiv1b1.GatewayClassSpec{
			ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
		},
	}
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

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns.Name,
			Name:      "test",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 80,
				},
				{
					Name: "https",
					Port: 443,
				},
			},
		},
	}

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
