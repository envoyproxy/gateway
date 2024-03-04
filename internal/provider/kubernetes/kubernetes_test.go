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
	"golang.org/x/exp/slices"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes/test"
	"github.com/envoyproxy/gateway/internal/utils"
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
	svr, err := config.New()
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
		"gatewayclass with parameters ref":     testGatewayClassWithParamRef,
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
	gwAPIs := filepath.Join("..", "..", "..", "charts", "gateway-helm", "crds", "gatewayapi-crds.yaml")
	egAPIs := filepath.Join("..", "..", "..", "charts", "gateway-helm", "crds", "generated")
	mcsAPIs := filepath.Join(".", "testdata", "crds", "multicluster-svc.yaml")

	env := &envtest.Environment{
		CRDDirectoryPaths: []string{gwAPIs, egAPIs, mcsAPIs},
	}
	cfg, err := env.Start()
	if err != nil {
		return env, nil, err
	}
	return env, cfg, nil
}

func testGatewayClassController(ctx context.Context, t *testing.T, provider *Provider, _ *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := test.GetGatewayClass("test-gc-controllername", egv1a1.GatewayControllerName, nil)
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

	gc := test.GetGatewayClass("test-gc-accepted-status", egv1a1.GatewayControllerName, nil)
	require.NoError(t, cli.Create(ctx, gc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc); err != nil {
			return false
		}

		for _, cond := range gc.Status.Conditions {
			if cond.Type == string(gwapiv1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
				return true
			}
		}

		return false
	}, defaultWait, defaultTick)

	require.Eventually(t, func() bool {
		return resources.GatewayAPIResources.Len() != 0
	}, defaultWait, defaultTick)

	// Even though no gateways exist, the controller loads the empty resource map
	// to support gateway deletions.
	require.Eventually(t, func() bool {
		_, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		return ok
	}, defaultWait, defaultTick)
}

func testGatewayClassWithParamRef(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	// Create the namespace for the test case.
	// Note: The namespace for the EnvoyProxy must match EG's configured namespace.
	testNs := config.DefaultNamespace
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testNs}}
	require.NoError(t, cli.Create(ctx, ns))

	defer func() {
		require.NoError(t, cli.Delete(ctx, ns))
	}()

	epName := "test-envoy-proxy"
	ep := test.GetEnvoyProxy(types.NamespacedName{Namespace: testNs, Name: epName}, false)
	require.NoError(t, cli.Create(ctx, ep))

	defer func() {
		require.NoError(t, cli.Delete(ctx, ep))
	}()

	gc := test.GetGatewayClass("gc-with-param-ref", egv1a1.GatewayControllerName, nil)
	gc.Spec.ParametersRef = &gwapiv1.ParametersReference{
		Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
		Kind:      egv1a1.KindEnvoyProxy,
		Name:      epName,
		Namespace: gatewayapi.NamespacePtr(testNs),
	}
	require.NoError(t, cli.Create(ctx, gc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	// Ensure the GatewayClass reports "Ready".
	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc); err != nil {
			return false
		}

		for _, cond := range gc.Status.Conditions {
			if cond.Type == string(gwapiv1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
				return true
			}
		}

		return false
	}, defaultWait, defaultTick)

	require.Eventually(t, func() bool {
		return resources.GatewayAPIResources.Len() != 0
	}, defaultWait, defaultTick)

	require.Eventually(t, func() bool {
		res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		if !ok {
			return false
		}

		if res.EnvoyProxy != nil {
			assert.Equal(t, res.EnvoyProxy.Spec, ep.Spec)
			return true
		}

		return false
	}, defaultWait, defaultTick)
}

func testGatewayScheduledStatus(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := test.GetGatewayClass("gc-scheduled-status-test", egv1a1.GatewayControllerName, nil)
	require.NoError(t, cli.Create(ctx, gc))

	// Ensure the GatewayClass reports "Ready".
	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc); err != nil {
			return false
		}

		for _, cond := range gc.Status.Conditions {
			if cond.Type == string(gwapiv1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
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

	gw := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "scheduled-status-test",
			Namespace: ns.Name,
		},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: gwapiv1.ObjectName(gc.Name),
			Listeners: []gwapiv1.Listener{
				{
					Name:     "test",
					Port:     gwapiv1.PortNumber(int32(8080)),
					Protocol: gwapiv1.HTTPProtocolType,
				},
			},
		},
	}
	require.NoError(t, cli.Create(ctx, gw))

	labels := map[string]string{
		gatewayapi.OwningGatewayNameLabel:      gw.Name,
		gatewayapi.OwningGatewayNamespaceLabel: gw.Namespace,
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: gw.Namespace,
			Name:      gw.Name + "-deployment",
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "dummy",
						Image: "dummy",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
						}},
					}},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 1,
		},
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: gw.Namespace,
			Name:      gw.Name + "-svc",
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Port: 80,
			}},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{{IP: "1.1.1.1"}},
			},
		},
	}

	require.NoError(t, cli.Create(ctx, deploy))
	require.NoError(t, cli.Create(ctx, svc))

	// Ensure the Gateway reports "Scheduled".
	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, utils.NamespacedName(gw), gw); err != nil {
			return false
		}

		for _, cond := range gw.Status.Conditions {
			fmt.Printf("Condition: %v\n", cond)
			if cond.Type == string(gwapiv1.GatewayConditionAccepted) && cond.Status == metav1.ConditionTrue {
				return true
			}
		}

		// Scheduled=True condition not found.
		return false
	}, defaultWait, defaultTick)

	defer func() {
		require.NoError(t, cli.Delete(ctx, gw))
	}()

	require.Eventually(t, func() bool {
		return resources.GatewayAPIResources.Len() != 0
	}, defaultWait, defaultTick)

	// Ensure the number of Gateways in the Gateway resource table is as expected.
	require.Eventually(t, func() bool {
		res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		if !ok {
			return false
		}
		return res != nil && len(res.Gateways) == 1
	}, defaultWait, defaultTick)

	// Ensure the gatewayclass has been finalized.
	require.Eventually(t, func() bool {
		err := cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc)
		return err == nil && slices.Contains(gc.Finalizers, gatewayClassFinalizer)
	}, defaultWait, defaultTick)

	// Ensure the test Gateway in the Gateway resources is as expected.
	key := utils.NamespacedName(gw)
	require.Eventually(t, func() bool {
		return cli.Get(ctx, key, gw) == nil
	}, defaultWait, defaultTick)

	res := resources.GetResourcesByGatewayClass(gc.Name)
	assert.NotNil(t, res)
	// Only check if the spec is equal
	// The watchable map will not store a resource
	// with an updated status if the spec has not changed
	// to eliminate this endless loop:
	// reconcile->store->translate->update-status->reconcile
	assert.Equal(t, gw.Spec, res.Gateways[0].Spec)
}

func testHTTPRoute(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := test.GetGatewayClass("httproute-test", egv1a1.GatewayControllerName, nil)
	require.NoError(t, cli.Create(ctx, gc))

	// Ensure the GatewayClass reports ready.
	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc); err != nil {
			return false
		}

		for _, cond := range gc.Status.Conditions {
			if cond.Type == string(gwapiv1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
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

	gw := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httproute-test",
			Namespace: ns.Name,
		},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: gwapiv1.ObjectName(gc.Name),
			Listeners: []gwapiv1.Listener{
				{
					Name:     "test",
					Port:     gwapiv1.PortNumber(int32(8080)),
					Protocol: gwapiv1.HTTPProtocolType,
				},
			},
		},
	}
	require.NoError(t, cli.Create(ctx, gw))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gw))
	}()

	svc := test.GetService(types.NamespacedName{Namespace: ns.Name, Name: "test"}, nil, map[string]int32{
		"http":  80,
		"https": 443,
	})

	require.NoError(t, cli.Create(ctx, svc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, svc))
	}()

	redirectHostname := gwapiv1.PreciseHostname("redirect.hostname.local")
	redirectPort := gwapiv1.PortNumber(8443)
	redirectStatus := 301

	rewriteHostname := gwapiv1.PreciseHostname("rewrite.hostname.local")

	var testCases = []struct {
		name  string
		route gwapiv1.HTTPRoute
	}{
		{
			name: "destination-httproute",
			route: gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Name: gwapiv1.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1.Hostname{"test.hostname.local"},
					Rules: []gwapiv1.HTTPRouteRule{
						{
							Matches: []gwapiv1.HTTPRouteMatch{
								{
									Path: &gwapiv1.HTTPPathMatch{
										Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
										Value: ptr.To("/"),
									},
								},
							},
							BackendRefs: []gwapiv1.HTTPBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "test",
											Port: ptr.To(gwapiv1.PortNumber(80)),
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
			route: gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-redirect-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Name: gwapiv1.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1.Hostname{"test.hostname.local"},
					Rules: []gwapiv1.HTTPRouteRule{
						{
							Matches: []gwapiv1.HTTPRouteMatch{
								{
									Path: &gwapiv1.HTTPPathMatch{
										Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
										Value: ptr.To("/redirect/"),
									},
								},
							},
							Filters: []gwapiv1.HTTPRouteFilter{
								{
									Type: gwapiv1.HTTPRouteFilterType("RequestRedirect"),
									RequestRedirect: &gwapiv1.HTTPRequestRedirectFilter{
										Scheme:   ptr.To("https"),
										Hostname: &redirectHostname,
										Path: &gwapiv1.HTTPPathModifier{
											Type:            gwapiv1.HTTPPathModifierType("ReplaceFullPath"),
											ReplaceFullPath: ptr.To("/newpath"),
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
			name: "rewrite-httproute",
			route: gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-rewrite-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Name: gwapiv1.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1.Hostname{"test.hostname.local"},
					Rules: []gwapiv1.HTTPRouteRule{
						{
							Matches: []gwapiv1.HTTPRouteMatch{
								{
									Path: &gwapiv1.HTTPPathMatch{
										Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
										Value: ptr.To("/rewrite/"),
									},
								},
							},
							BackendRefs: []gwapiv1.HTTPBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "test",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							Filters: []gwapiv1.HTTPRouteFilter{
								{
									Type: gwapiv1.HTTPRouteFilterType("URLRewrite"),
									URLRewrite: &gwapiv1.HTTPURLRewriteFilter{
										Hostname: &rewriteHostname,
										Path: &gwapiv1.HTTPPathModifier{
											Type:            gwapiv1.HTTPPathModifierType("ReplaceFullPath"),
											ReplaceFullPath: ptr.To("/newpath"),
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
			name: "add-request-header-httproute",
			route: gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-add-request-header-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Name: gwapiv1.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1.Hostname{"test.hostname.local"},
					Rules: []gwapiv1.HTTPRouteRule{
						{
							Matches: []gwapiv1.HTTPRouteMatch{
								{
									Path: &gwapiv1.HTTPPathMatch{
										Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
										Value: ptr.To("/addheader/"),
									},
								},
							},
							BackendRefs: []gwapiv1.HTTPBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "test",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							Filters: []gwapiv1.HTTPRouteFilter{
								{
									Type: gwapiv1.HTTPRouteFilterType("RequestHeaderModifier"),
									RequestHeaderModifier: &gwapiv1.HTTPHeaderFilter{
										Add: []gwapiv1.HTTPHeader{
											{
												Name:  gwapiv1.HTTPHeaderName("header-1"),
												Value: "value-1",
											},
											{
												Name:  gwapiv1.HTTPHeaderName("header-2"),
												Value: "value-2",
											},
										},
										Set: []gwapiv1.HTTPHeader{
											{
												Name:  gwapiv1.HTTPHeaderName("header-3"),
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
			name: "remove-request-header-httproute",
			route: gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-remove-request-header-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Name: gwapiv1.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1.Hostname{"test.hostname.local"},
					Rules: []gwapiv1.HTTPRouteRule{
						{
							Matches: []gwapiv1.HTTPRouteMatch{
								{
									Path: &gwapiv1.HTTPPathMatch{
										Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
										Value: ptr.To("/remheader/"),
									},
								},
							},
							BackendRefs: []gwapiv1.HTTPBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "test",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							Filters: []gwapiv1.HTTPRouteFilter{
								{
									Type: gwapiv1.HTTPRouteFilterType("RequestHeaderModifier"),
									RequestHeaderModifier: &gwapiv1.HTTPHeaderFilter{
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
		{
			name: "add-response-header-httproute",
			route: gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-add-response-header-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Name: gwapiv1.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1.Hostname{"test.hostname.local"},
					Rules: []gwapiv1.HTTPRouteRule{
						{
							Matches: []gwapiv1.HTTPRouteMatch{
								{
									Path: &gwapiv1.HTTPPathMatch{
										Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
										Value: ptr.To("/addheader/"),
									},
								},
							},
							BackendRefs: []gwapiv1.HTTPBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "test",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							Filters: []gwapiv1.HTTPRouteFilter{
								{
									Type: gwapiv1.HTTPRouteFilterType("ResponseHeaderModifier"),
									ResponseHeaderModifier: &gwapiv1.HTTPHeaderFilter{
										Add: []gwapiv1.HTTPHeader{
											{
												Name:  gwapiv1.HTTPHeaderName("header-1"),
												Value: "value-1",
											},
											{
												Name:  gwapiv1.HTTPHeaderName("header-2"),
												Value: "value-2",
											},
										},
										Set: []gwapiv1.HTTPHeader{
											{
												Name:  gwapiv1.HTTPHeaderName("header-3"),
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
			name: "remove-response-header-httproute",
			route: gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-remove-response-header-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Name: gwapiv1.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1.Hostname{"test.hostname.local"},
					Rules: []gwapiv1.HTTPRouteRule{
						{
							Matches: []gwapiv1.HTTPRouteMatch{
								{
									Path: &gwapiv1.HTTPPathMatch{
										Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
										Value: ptr.To("/remheader/"),
									},
								},
							},
							BackendRefs: []gwapiv1.HTTPBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "test",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							Filters: []gwapiv1.HTTPRouteFilter{
								{
									Type: gwapiv1.HTTPRouteFilterType("ResponseHeaderModifier"),
									ResponseHeaderModifier: &gwapiv1.HTTPHeaderFilter{
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
		{
			name: "mirror-httproute",
			route: gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "httproute-mirror-test",
					Namespace: ns.Name,
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Name: gwapiv1.ObjectName(gw.Name),
							},
						},
					},
					Hostnames: []gwapiv1.Hostname{"test.hostname.local"},
					Rules: []gwapiv1.HTTPRouteRule{
						{
							Matches: []gwapiv1.HTTPRouteMatch{
								{
									Path: &gwapiv1.HTTPPathMatch{
										Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
										Value: ptr.To("/mirror/"),
									},
								},
							},
							BackendRefs: []gwapiv1.HTTPBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "test",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							Filters: []gwapiv1.HTTPRouteFilter{
								{
									Type: gwapiv1.HTTPRouteFilterType("RequestMirror"),
									RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
										BackendRef: gwapiv1.BackendObjectReference{
											Name: "test",
											Port: ptr.To(gwapiv1.PortNumber(80)),
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
				return resources.GatewayAPIResources.Len() != 0
			}, defaultWait, defaultTick)

			// Ensure the test HTTPRoute in the HTTPRoute resources is as expected.
			key := utils.NamespacedName(&testCase.route)
			require.Eventually(t, func() bool {
				return cli.Get(ctx, key, &testCase.route) == nil
			}, defaultWait, defaultTick)

			require.Eventually(t, func() bool {
				res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
				if !ok {
					return false
				}
				return ok && len(res.HTTPRoutes) != 0
			}, defaultWait, defaultTick)

			res := resources.GetResourcesByGatewayClass(gc.Name)
			assert.NotNil(t, res)
			assert.Equal(t, &testCase.route, res.HTTPRoutes[0])

			// Ensure the HTTPRoute Namespace is in the Namespace resource map.
			require.Eventually(t, func() bool {
				res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
				if !ok {
					return false
				}
				for _, ns := range res.Namespaces {
					if ns.Name == testCase.route.Namespace {
						return true
					}
				}
				return false
			}, defaultWait, defaultTick)

			// Ensure the Service is in the resource map.
			require.Eventually(t, func() bool {
				// The redirect test case does not have a service.
				if testCase.name == "redirect-httproute" {
					return true
				}

				res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
				if !ok {
					return false
				}
				for _, s := range res.Services {
					if s.Name == svc.Name && s.Namespace == svc.Namespace {
						return true
					}
				}
				return false
			}, defaultWait, defaultTick)
		})
	}
}

func testTLSRoute(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := test.GetGatewayClass("tlsroute-test", egv1a1.GatewayControllerName, nil)
	require.NoError(t, cli.Create(ctx, gc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	// Create the namespace for the Gateway under test.
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "tlsroute-test"}}
	require.NoError(t, cli.Create(ctx, ns))

	gw := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tlsroute-test",
			Namespace: ns.Name,
		},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: gwapiv1.ObjectName(gc.Name),
			Listeners: []gwapiv1.Listener{
				{
					Name:     "test",
					Port:     gwapiv1.PortNumber(int32(8080)),
					Protocol: gwapiv1.TLSProtocolType,
					TLS: &gwapiv1.GatewayTLSConfig{
						Mode: ptr.To(gwapiv1.TLSModePassthrough),
					},
				},
			},
		},
	}
	require.NoError(t, cli.Create(ctx, gw))

	defer func() {
		require.NoError(t, cli.Delete(ctx, gw))
	}()

	svc := test.GetService(types.NamespacedName{Namespace: ns.Name, Name: "test"}, nil, map[string]int32{
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
										Port: ptr.To(gwapiv1.PortNumber(90)),
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
				return resources.GatewayAPIResources.Len() != 0
			}, defaultWait, defaultTick)

			// Ensure the test TLSRoute in the TLSRoute resources is as expected.
			key := utils.NamespacedName(&testCase.route)
			require.Eventually(t, func() bool {
				return cli.Get(ctx, key, &testCase.route) == nil
			}, defaultWait, defaultTick)

			require.Eventually(t, func() bool {
				res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
				if !ok {
					return false
				}
				return ok && len(res.TLSRoutes) != 0
			}, defaultWait, defaultTick)

			res := resources.GetResourcesByGatewayClass(gc.Name)
			assert.NotNil(t, res)
			assert.Equal(t, &testCase.route, res.TLSRoutes[0])

			// Ensure the HTTPRoute Namespace is in the Namespace resource map.
			require.Eventually(t, func() bool {
				res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
				if !ok {
					return false
				}
				for _, ns := range res.Namespaces {
					if ns.Name == testCase.route.Namespace {
						return true
					}
				}
				return false
			}, defaultWait, defaultTick)

			// Ensure the Service is in the resource map.
			require.Eventually(t, func() bool {
				res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
				if !ok {
					return false
				}
				for _, s := range res.Services {
					if s.Name == svc.Name && s.Namespace == svc.Namespace {
						return true
					}
				}
				return false
			}, defaultWait, defaultTick)
		})
	}
}

// testServiceCleanupForMultipleRoutes creates multiple Routes pointing to the
// same backend Service, and checks whether the Service is properly removed
// from the resource map after Route deletion.
func testServiceCleanupForMultipleRoutes(ctx context.Context, t *testing.T, provider *Provider, resources *message.ProviderResources) {
	cli := provider.manager.GetClient()

	gc := test.GetGatewayClass("service-cleanup-test", egv1a1.GatewayControllerName, nil)
	require.NoError(t, cli.Create(ctx, gc))
	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	// Create the namespace for the Gateway under test.
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "service-cleanup-test"}}
	require.NoError(t, cli.Create(ctx, ns))

	gw := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-cleanup-test",
			Namespace: ns.Name,
		},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: gwapiv1.ObjectName(gc.Name),
			Listeners: []gwapiv1.Listener{
				{
					Name:     "httptest",
					Port:     gwapiv1.PortNumber(int32(8080)),
					Protocol: gwapiv1.HTTPProtocolType,
				},
				{
					Name:     "tlstest",
					Port:     gwapiv1.PortNumber(int32(8043)),
					Protocol: gwapiv1.TLSProtocolType,
					TLS: &gwapiv1.GatewayTLSConfig{
						Mode: ptr.To(gwapiv1.TLSModePassthrough),
					},
				},
			},
		},
	}
	require.NoError(t, cli.Create(ctx, gw))
	defer func() {
		require.NoError(t, cli.Delete(ctx, gw))
	}()

	svc := test.GetService(types.NamespacedName{Namespace: ns.Name, Name: "test-common-svc"}, nil, map[string]int32{
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
						Port: ptr.To(gwapiv1.PortNumber(90)),
					}},
				}},
			},
		},
	}

	httpRoute := gwapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "httproute-test",
			Namespace: ns.Name,
		},
		Spec: gwapiv1.HTTPRouteSpec{
			CommonRouteSpec: gwapiv1.CommonRouteSpec{
				ParentRefs: []gwapiv1.ParentReference{{
					Name: gwapiv1.ObjectName(gw.Name),
				}},
			},
			Hostnames: []gwapiv1.Hostname{"test-http.hostname.local"},
			Rules: []gwapiv1.HTTPRouteRule{{
				Matches: []gwapiv1.HTTPRouteMatch{{
					Path: &gwapiv1.HTTPPathMatch{
						Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
						Value: ptr.To("/"),
					},
				}},
				BackendRefs: []gwapiv1.HTTPBackendRef{{
					BackendRef: gwapiv1.BackendRef{
						BackendObjectReference: gwapiv1.BackendObjectReference{
							Name: "test-common-svc",
							Port: ptr.To(gwapiv1.PortNumber(80)),
						},
					},
				}},
			}},
		},
	}

	// Create the TLSRoute and HTTPRoute
	require.NoError(t, cli.Create(ctx, &tlsRoute))
	require.NoError(t, cli.Create(ctx, &httpRoute))

	require.Eventually(t, func() bool {
		return resources.GatewayAPIResources.Len() != 0
	}, defaultWait, defaultTick)

	// Check that the Service is present in the resource map
	require.Eventually(t, func() bool {
		res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		if !ok {
			return false
		}
		for _, s := range res.Services {
			if s.Namespace == svc.Namespace && s.Name == svc.Name {
				return true
			}
		}
		return false
	}, defaultWait, defaultTick)

	// Delete the TLSRoute, and check if the Service is still present
	require.NoError(t, cli.Delete(ctx, &tlsRoute))
	require.Eventually(t, func() bool {
		res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		if !ok {
			return false
		}
		for _, s := range res.Services {
			if s.Namespace == svc.Namespace && s.Name == svc.Name {
				return true
			}
		}
		return false
	}, defaultWait, defaultTick)

	// Delete the HTTPRoute, and check if the Service is also removed
	require.NoError(t, cli.Delete(ctx, &httpRoute))
	require.Eventually(t, func() bool {
		res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		if !ok {
			return false
		}
		for _, s := range res.Services {
			if s.Namespace == svc.Namespace && s.Name == svc.Name {
				return false
			}
		}
		return true
	}, defaultWait, defaultTick)
}

func TestNamespacedProvider(t *testing.T) {
	// Setup the test environment.
	testEnv, cliCfg, err := startEnv()
	require.NoError(t, err)

	// Setup and start the kube provider.
	svr, err := config.New()
	require.NoError(t, err)
	// config to watch a subset of namespaces
	svr.EnvoyGateway.Provider.Kubernetes = &egv1a1.EnvoyGatewayKubernetesProvider{
		Watch: &egv1a1.KubernetesWatchMode{
			Type:       egv1a1.KubernetesWatchModeTypeNamespaces,
			Namespaces: []string{"ns1", "ns2"},
		},
	}

	resources := new(message.ProviderResources)
	provider, err := New(cliCfg, svr, resources)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		require.NoError(t, provider.Start(ctx))
	}()

	// Make sure a cluster scoped gatewayclass can be reconciled
	testGatewayClassController(ctx, t, provider, resources)

	cli := provider.manager.GetClient()
	gcName := "gc-watch-ns"
	gc := test.GetGatewayClass(gcName, egv1a1.GatewayControllerName, nil)
	require.NoError(t, cli.Create(ctx, gc))

	// Create the namespaces.
	ns1 := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}}
	require.NoError(t, cli.Create(ctx, ns1))
	ns2 := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}}
	require.NoError(t, cli.Create(ctx, ns2))
	ns3 := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns3"}}
	require.NoError(t, cli.Create(ctx, ns3))

	// Create the gateways
	gw1 := test.GetGateway(types.NamespacedName{Name: "gw-ns1", Namespace: "ns1"}, gcName, 8080)
	require.NoError(t, cli.Create(ctx, gw1))
	gw2 := test.GetGateway(types.NamespacedName{Name: "gw-ns2", Namespace: "ns2"}, gcName, 8080)
	require.NoError(t, cli.Create(ctx, gw2))
	gw3 := test.GetGateway(types.NamespacedName{Name: "gw-ns3", Namespace: "ns3"}, gcName, 8080)
	require.NoError(t, cli.Create(ctx, gw3))

	// Ensure only 2 gateways are reconciled
	gatewayList := &gwapiv1.GatewayList{}
	require.NoError(t, cli.List(ctx, gatewayList))
	assert.Equal(t, len(gatewayList.Items), 2)

	// Stop the kube provider.
	defer func() {
		cancel()
		require.NoError(t, testEnv.Stop())
	}()
}

func TestNamespaceSelectorProvider(t *testing.T) {
	// Setup the test environment.
	testEnv, cliCfg, err := startEnv()
	require.NoError(t, err)

	// Setup and start the kube provider.
	svr, err := config.New()
	require.NoError(t, err)
	// config to watch a subset of namespaces
	svr.EnvoyGateway.Provider.Kubernetes = &egv1a1.EnvoyGatewayKubernetesProvider{
		Watch: &egv1a1.KubernetesWatchMode{
			Type:              egv1a1.KubernetesWatchModeTypeNamespaceSelector,
			NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"label-1": "true", "label-2": "true"}},
		},
	}

	resources := new(message.ProviderResources)
	provider, err := New(cliCfg, svr, resources)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		require.NoError(t, provider.Start(ctx))
	}()

	defer func() {
		cancel()
		require.NoError(t, testEnv.Stop())
	}()

	cli := provider.manager.GetClient()
	watchedNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
		Name:   "watched-ns",
		Labels: map[string]string{"label-1": "true", "label-2": "true"},
	}}
	require.NoError(t, cli.Create(ctx, watchedNS))
	nonWatchedNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "non-watched-ns"}}
	require.NoError(t, cli.Create(ctx, nonWatchedNS))

	gcName := "gc-name"
	gc := test.GetGatewayClass(gcName, egv1a1.GatewayControllerName, nil)
	require.NoError(t, cli.Create(ctx, gc))

	require.Eventually(t, func() bool {
		if err := cli.Get(ctx, types.NamespacedName{Name: gc.Name}, gc); err != nil {
			return false
		}

		for _, cond := range gc.Status.Conditions {
			if cond.Type == string(gwapiv1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
				return true
			}
		}

		return false
	}, defaultWait, defaultTick)

	defer func() {
		require.NoError(t, cli.Delete(ctx, gc))
	}()

	// Create the gateways
	watchedGateway := test.GetGateway(types.NamespacedName{Name: "watched-gateway", Namespace: watchedNS.Name}, gcName, 8080)
	require.NoError(t, cli.Create(ctx, watchedGateway))
	defer func() {
		require.NoError(t, cli.Delete(ctx, watchedGateway))
	}()

	nonWatchedGateway := test.GetGateway(types.NamespacedName{Name: "non-watched-gateway", Namespace: nonWatchedNS.Name}, gcName, 8080)
	require.NoError(t, cli.Create(ctx, nonWatchedGateway))
	defer func() {
		require.NoError(t, cli.Delete(ctx, nonWatchedGateway))
	}()

	require.Eventually(t, func() bool {
		return resources.GatewayAPIResources.Len() != 0
	}, defaultWait, defaultTick)

	require.Eventually(t, func() bool {
		res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		if !ok {
			return false
		}
		return res != nil && len(res.Gateways) == 1
	}, defaultWait, defaultTick)

	_, ok := resources.GatewayStatuses.Load(types.NamespacedName{Name: "non-watched-gateway", Namespace: nonWatchedNS.Name})
	require.Equal(t, false, ok)

	watchedSvc := test.GetService(types.NamespacedName{Namespace: watchedNS.Name, Name: "watched-service"}, nil, map[string]int32{
		"http":  80,
		"https": 443,
	})
	require.NoError(t, cli.Create(ctx, watchedSvc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, watchedSvc))
	}()

	nonWatchedSvc := test.GetService(types.NamespacedName{Namespace: nonWatchedNS.Name, Name: "non-watched-service"}, nil, map[string]int32{
		"http":  8001,
		"https": 44300,
	})
	require.NoError(t, cli.Create(ctx, nonWatchedSvc))

	defer func() {
		require.NoError(t, cli.Delete(ctx, nonWatchedSvc))
	}()

	watchedHTTPRoute := test.GetHTTPRoute(
		types.NamespacedName{
			Namespace: watchedNS.Name,
			Name:      "watched-http-route",
		},
		watchedGateway.Name,
		types.NamespacedName{Name: watchedSvc.Name}, 80)

	require.NoError(t, cli.Create(ctx, watchedHTTPRoute))
	defer func() {
		require.NoError(t, cli.Delete(ctx, watchedHTTPRoute))
	}()

	nonWatchedHTTPRoute := test.GetHTTPRoute(
		types.NamespacedName{
			Namespace: nonWatchedNS.Name,
			Name:      "non-watched-http-route",
		},
		nonWatchedGateway.Name,
		types.NamespacedName{Name: nonWatchedSvc.Name}, 8001)
	require.NoError(t, cli.Create(ctx, nonWatchedHTTPRoute))
	defer func() {
		require.NoError(t, cli.Delete(ctx, nonWatchedHTTPRoute))
	}()

	watchedGRPCRoute := test.GetGRPCRoute(
		types.NamespacedName{
			Namespace: watchedNS.Name,
			Name:      "watched-grpc-route",
		},
		watchedGateway.Name,
		types.NamespacedName{Name: watchedSvc.Name}, 80)
	require.NoError(t, cli.Create(ctx, watchedGRPCRoute))
	defer func() {
		require.NoError(t, cli.Delete(ctx, watchedGRPCRoute))
	}()

	nonWatchedGRPCRoute := test.GetGRPCRoute(
		types.NamespacedName{
			Namespace: nonWatchedNS.Name,
			Name:      "non-watched-grpc-route",
		},
		nonWatchedGateway.Name,
		types.NamespacedName{Name: nonWatchedNS.Name}, 8001)
	require.NoError(t, cli.Create(ctx, nonWatchedGRPCRoute))
	defer func() {
		require.NoError(t, cli.Delete(ctx, nonWatchedGRPCRoute))
	}()

	watchedTCPRoute := test.GetTCPRoute(
		types.NamespacedName{
			Namespace: watchedNS.Name,
			Name:      "watched-tcp-route",
		},
		watchedGateway.Name,
		types.NamespacedName{Name: watchedSvc.Name}, 80)
	require.NoError(t, cli.Create(ctx, watchedTCPRoute))
	defer func() {
		require.NoError(t, cli.Delete(ctx, watchedTCPRoute))
	}()

	nonWatchedTCPRoute := test.GetTCPRoute(
		types.NamespacedName{
			Namespace: nonWatchedNS.Name,
			Name:      "non-watched-tcp-route",
		},
		nonWatchedGateway.Name,
		types.NamespacedName{Name: nonWatchedNS.Name}, 80)
	require.NoError(t, cli.Create(ctx, nonWatchedTCPRoute))
	defer func() {
		require.NoError(t, cli.Delete(ctx, nonWatchedTCPRoute))
	}()

	watchedTLSRoute := test.GetTLSRoute(
		types.NamespacedName{
			Namespace: watchedNS.Name,
			Name:      "watched-tls-route",
		},
		watchedGateway.Name,
		types.NamespacedName{Name: watchedSvc.Name}, 443)
	require.NoError(t, cli.Create(ctx, watchedTLSRoute))
	defer func() {
		require.NoError(t, cli.Delete(ctx, watchedTLSRoute))
	}()

	nonWatchedTLSRoute := test.GetTLSRoute(
		types.NamespacedName{
			Namespace: nonWatchedNS.Name,
			Name:      "non-watched-tls-route",
		},
		nonWatchedGateway.Name,
		types.NamespacedName{Name: nonWatchedNS.Name}, 443)
	require.NoError(t, cli.Create(ctx, nonWatchedTLSRoute))
	defer func() {
		require.NoError(t, cli.Delete(ctx, nonWatchedTLSRoute))
	}()

	watchedUDPRoute := test.GetUDPRoute(
		types.NamespacedName{
			Namespace: watchedNS.Name,
			Name:      "watched-udp-route",
		},
		watchedGateway.Name,
		types.NamespacedName{Name: watchedSvc.Name}, 80)
	require.NoError(t, cli.Create(ctx, watchedUDPRoute))
	defer func() {
		require.NoError(t, cli.Delete(ctx, watchedUDPRoute))
	}()

	nonWatchedUDPRoute := test.GetUDPRoute(
		types.NamespacedName{
			Namespace: nonWatchedNS.Name,
			Name:      "non-watched-udp-route",
		},
		nonWatchedGateway.Name,
		types.NamespacedName{Name: nonWatchedNS.Name}, 80)
	require.NoError(t, cli.Create(ctx, nonWatchedUDPRoute))
	defer func() {
		require.NoError(t, cli.Delete(ctx, nonWatchedUDPRoute))
	}()

	require.Eventually(t, func() bool {
		res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		if !ok {
			return false
		}
		// The service number dependes on the service created and the backendRef
		return res != nil && len(res.Services) == 5
	}, defaultWait, defaultTick)

	require.Eventually(t, func() bool {
		res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		if !ok {
			return false
		}
		return res != nil && len(res.HTTPRoutes) == 1
	}, defaultWait, defaultTick)

	require.Eventually(t, func() bool {
		res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		if !ok {
			return false
		}
		return res != nil && len(res.TCPRoutes) == 1
	}, defaultWait, defaultTick)

	require.Eventually(t, func() bool {
		res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		if !ok {
			return false
		}
		return res != nil && len(res.TLSRoutes) == 1
	}, defaultWait, defaultTick)

	require.Eventually(t, func() bool {
		res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		if !ok {
			return false
		}
		return res != nil && len(res.UDPRoutes) == 1
	}, defaultWait, defaultTick)

	require.Eventually(t, func() bool {
		res, ok := waitUntilGatewayClassResourcesAreReady(resources, gc.Name)
		if !ok {
			return false
		}
		return res != nil && len(res.GRPCRoutes) == 1
	}, defaultWait, defaultTick)

}

func waitUntilGatewayClassResourcesAreReady(resources *message.ProviderResources, gatewayClassName string) (*gatewayapi.Resources, bool) {
	res := resources.GetResourcesByGatewayClass(gatewayClassName)
	if res == nil {
		return nil, false
	}

	return res, true
}
