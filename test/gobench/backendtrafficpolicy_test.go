// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gobench

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/cmd/egctl"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func BenchmarkBackendTrafficPolicyToXDS(b *testing.B) {
	for _, count := range []int{1, 100, 1000} {
		b.Run(fmt.Sprintf("routes=%d", count), func(b *testing.B) {
			rs := backendTrafficPolicyResources(count)
			opts := &egctl.TranslationOptions{}
			result, err := egctl.TranslateGatewayAPIToXds("default", "cluster.local", "route", rs.DeepCopy(), opts)
			require.NoError(b, err)
			var configured int
			for _, config := range result {
				data, err := json.Marshal(config)
				require.NoError(b, err)
				wrapper := new(anypb.Any)
				require.NoError(b, protojson.Unmarshal(data, wrapper))
				dump := new(adminv3.RoutesConfigDump)
				require.NoError(b, wrapper.UnmarshalTo(dump))
				for _, entry := range dump.DynamicRouteConfigs {
					routes := new(routev3.RouteConfiguration)
					require.NoError(b, entry.RouteConfig.UnmarshalTo(routes))
					for _, host := range routes.VirtualHosts {
						for _, route := range host.Routes {
							require.NotNil(b, route.GetRoute(), route.Name)
							require.Equal(b, 5*time.Second, route.GetRoute().GetTimeout().AsDuration(), route.Name)
							configured++
						}
					}
				}
			}
			require.Equal(b, count*4*2, configured)

			b.ReportAllocs()
			for b.Loop() {
				// Include fresh inputs and the complete Gateway API -> IR -> xDS pipeline.
				_, err := egctl.TranslateGatewayAPIToXds("default", "cluster.local", "all", rs.DeepCopy(), opts)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func backendTrafficPolicyResources(count int) *resource.Resources {
	rs := resource.NewResources()
	rs.GatewayClass = &gwapiv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{Name: "eg"},
		Spec:       gwapiv1.GatewayClassSpec{ControllerName: egv1a1.GatewayControllerName},
	}
	rs.Gateways = []*gwapiv1.Gateway{{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "gateway"},
		Spec: gwapiv1.GatewaySpec{GatewayClassName: "eg", Listeners: []gwapiv1.Listener{
			{Name: "http", Protocol: gwapiv1.HTTPProtocolType, Port: 80},
			{Name: "http-alt", Protocol: gwapiv1.HTTPProtocolType, Port: 81},
		}},
	}}
	rs.Namespaces = []*corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "default"}}}
	rs.Services = []*corev1.Service{{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "backend"},
		Spec: corev1.ServiceSpec{ClusterIP: "10.96.0.1", Ports: []corev1.ServicePort{
			{Name: "http", Port: 8080, Protocol: corev1.ProtocolTCP},
		}},
	}}
	rs.EndpointSlices = []*discoveryv1.EndpointSlice{{
		ObjectMeta:  metav1.ObjectMeta{Namespace: "default", Name: "backend", Labels: map[string]string{discoveryv1.LabelServiceName: "backend"}},
		AddressType: discoveryv1.AddressTypeIPv4,
		Ports:       []discoveryv1.EndpointPort{{Name: new("http"), Port: new(int32(8080)), Protocol: new(corev1.ProtocolTCP)}},
		Endpoints:   []discoveryv1.Endpoint{{Addresses: []string{"10.0.0.1"}, Conditions: discoveryv1.EndpointConditions{Ready: new(true)}}},
	}}
	for i := range count {
		name := fmt.Sprintf("route-%04d", i)
		rule := gwapiv1.HTTPRouteRule{BackendRefs: []gwapiv1.HTTPBackendRef{{BackendRef: gwapiv1.BackendRef{
			BackendObjectReference: gwapiv1.BackendObjectReference{Name: "backend", Port: new(gwapiv1.PortNumber(8080))},
			Weight:                 new(int32(1)),
		}}}}
		for match := range 4 {
			rule.Matches = append(rule.Matches, gwapiv1.HTTPRouteMatch{Path: &gwapiv1.HTTPPathMatch{
				Type: new(gwapiv1.PathMatchPathPrefix), Value: new(fmt.Sprintf("/%s/%d", name, match)),
			}})
		}
		rs.HTTPRoutes = append(rs.HTTPRoutes, &gwapiv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: name},
			Spec: gwapiv1.HTTPRouteSpec{
				CommonRouteSpec: gwapiv1.CommonRouteSpec{ParentRefs: []gwapiv1.ParentReference{{Name: "gateway"}}},
				Rules:           []gwapiv1.HTTPRouteRule{rule},
			},
		})
		policy := &egv1a1.BackendTrafficPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: name}}
		policy.Spec.TargetRefs = []gwapiv1.LocalPolicyTargetReferenceWithSectionName{{LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
			Group: gwapiv1.GroupName, Kind: resource.KindHTTPRoute, Name: gwapiv1.ObjectName(name),
		}}}
		policy.Spec.Timeout = &egv1a1.Timeout{HTTP: &egv1a1.HTTPTimeout{RequestTimeout: new(gwapiv1.Duration("5s"))}}
		rs.BackendTrafficPolicies = append(rs.BackendTrafficPolicies, policy)
	}
	return rs
}
