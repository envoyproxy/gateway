// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, FileAccessLogTest)
}

var FileAccessLogTest = suite.ConformanceTest{
	ShortName:   "FileAccessLog",
	Description: "Make sure file access log is working",
	Manifests:   []string{"testdata/accesslog.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("default", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-infra-backend-v1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectOkResp := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/",
				},
				Response: httputils.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			_ = httputils.MakeRequest(t, &expectOkResp, gwAddr, "HTTP", "http")

			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(_ context.Context) (bool, error) {
					// query log count from loki
					count, err := QueryLogCountFromLoki(t, suite.Client, types.NamespacedName{
						Namespace: "envoy-gateway-system",
					})
					if err != nil {
						t.Logf("failed to get log count from loki: %v", err)
						return false, nil
					}

					if count != 1 {
						return true, nil
					}

					return false, nil
				}); err != nil {
				t.Errorf("failed to get log count from loki: %v", err)
			}

		})
	},
}

// QueryLogCountFromLoki queries log count from loki
// TODO: move to utils package if needed
func QueryLogCountFromLoki(t *testing.T, c client.Client, nn types.NamespacedName) (int, error) {
	svc := corev1.Service{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "monitoring",
		Name:      "loki",
	}, &svc); err != nil {
		return -1, err
	}
	lokiHost := ""
	for _, ing := range svc.Status.LoadBalancer.Ingress {
		if ing.IP != "" {
			lokiHost = ing.IP
			break
		}
	}

	q := fmt.Sprintf("{namespace=\"%s\",container=\"%s\"}", nn.Namespace, "envoy}")
	params := url.Values{}
	params.Add("query", q)
	lokiQueryURL := fmt.Sprintf("http://%s:3100/loki/api/v1/query_range?%s", lokiHost, params.Encode())
	res, err := http.DefaultClient.Get(lokiQueryURL)
	if err != nil {
		return -1, err
	}
	t.Logf("get response from loki: %v", res)

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return -1, err
	}

	lokiResponse := &LokiQueryResponse{}
	if err := json.Unmarshal(b, lokiResponse); err != nil {
		return -1, err
	}

	if len(lokiResponse.Data.Result) == 0 {
		return 0, nil
	}
	return len(lokiResponse.Data.Result[0].Values), nil
}

type LokiQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric interface{}
			Values []interface{} `json:"values"`
		}
	}
}
