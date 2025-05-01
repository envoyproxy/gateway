// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, StatNameTest)
}

var StatNameTest = suite.ConformanceTest{
	ShortName:   "StatName",
	Description: "Make sure metric is working",
	Manifests:   []string{"testdata/envoyproxy-stat-name.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "stat-name-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: "stat-name-gtw", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		t.Run("prometheus", func(t *testing.T) {
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/foo",
				},
				Response: httputils.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			expectedResponse = httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/bar",
				},
				Response: httputils.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// make sure that a metrics for alt_stat_name exists in test gateway and they collapse stats from multiple clusters
			// expect to find 2 cluster members, since there are two routes with the same alt_stat_name and each cluster has a single member
			if err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, time.Minute, true, func(ctx context.Context) (done bool, err error) {
				v, err := prometheus.QueryPrometheus(suite.Client, `envoy_cluster_membership_healthy{envoy_cluster_name="gateway-conformance-infra/stat-name-route/gateway-conformance-infra/infra-backend-v1"}`)
				if err != nil {
					tlog.Logf(t, "failed to query prometheus: %v", err)
					return false, err
				}
				if v != nil && v.Type() == model.ValVector {
					vectorVal := v.(model.Vector)
					if len(vectorVal) == 1 && vectorVal[0].Value == 2 {
						tlog.Logf(t, "got expected value: %v", v)
						return true, nil
					}
				}
				return false, nil
			}); err != nil {
				t.Errorf("failed to get expected response for the last (fourth) request: %v", err)
			}
		})
	},
}
