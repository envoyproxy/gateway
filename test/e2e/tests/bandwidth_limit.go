// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, BandwidthLimitTest)
}

var BandwidthLimitTest = suite.ConformanceTest{
	ShortName:   "BandwidthLimit",
	Description: "Verify that bandwidth limit enforces per-direction transfer rate limits",
	Manifests:   []string{"testdata/bandwidth-limit.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "bandwidth-limit-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(
			t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), routeNN,
		)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client,
			types.NamespacedName{Name: "bandwidth-limit-policy", Namespace: ns},
			suite.ControllerName, ancestorRef)

		// Warm up: confirm the route is reachable with a plain GET.
		httputils.MakeRequestAndExpectEventuallyConsistentResponse(
			t, suite.RoundTripper, suite.TimeoutConfig, gwAddr,
			httputils.ExpectedResponse{
				Request:   httputils.Request{Path: "/bandwidth-limit"},
				Response:  httputils.Response{StatusCodes: []int{200}},
				Namespace: ns,
			})

		t.Run("RequestThrottled", func(t *testing.T) {
			// POST a 5 KiB body — well above the 1 KiB/s token bucket capacity —
			// to deterministically trigger request-side bandwidth throttling.
			body := strings.Repeat("x", 5*1024)
			resp, err := http.Post(
				fmt.Sprintf("http://%s/bandwidth-limit", gwAddr),
				"application/octet-stream",
				bytes.NewBufferString(body),
			)
			if err != nil {
				t.Logf("POST error (may be expected during throttle): %v", err)
			} else {
				_ = resp.Body.Close()
			}

			// Verify the bandwidth limit filter throttled the request.
			// The stat_prefix "http_bandwidth_limiter" exposes the Prometheus counter
			// envoy_http_bandwidth_limiter_http_bandwidth_limit_request_enforced.
			if err := wait.PollUntilContextTimeout(t.Context(), 3*time.Second, time.Minute, true,
				func(_ context.Context) (bool, error) {
					v, err := prometheus.QueryPrometheus(suite.Client,
						`sum(envoy_http_bandwidth_limiter_http_bandwidth_limit_request_enforced)`)
					if err != nil {
						tlog.Logf(t, "failed to query Prometheus: %v", err)
						return false, err
					}
					if v != nil && v.Type() == model.ValVector {
						vec := v.(model.Vector)
						if len(vec) > 0 && vec[0].Value > 0 {
							tlog.Logf(t, "request_enforced = %v", vec[0].Value)
							return true, nil
						}
					}
					return false, nil
				}); err != nil {
				t.Errorf("envoy_http_bandwidth_limiter_http_bandwidth_limit_request_enforced never became positive: %v", err)
			}
		})
	},
}
