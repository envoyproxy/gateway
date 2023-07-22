// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-logfmt/logfmt"
	"github.com/gogo/protobuf/jsonpb" // nolint: depguard // tempopb use gogo/protobuf
	"github.com/grafana/tempo/pkg/tempopb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/utils/naming"
)

func init() {
	ConformanceTests = append(ConformanceTests, OpenTelemetryTracingTest)
}

var OpenTelemetryTracingTest = suite.ConformanceTest{
	ShortName:   "OpenTelemetryTracing",
	Description: "Make sure OpenTelemetry tracing is working",
	Manifests:   []string{"testdata/tracing-otel.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("tempo", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "tracing-otel", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/tracing",
				},
				Response: httputils.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			tags := map[string]string{
				"component":    "proxy",
				"service.name": naming.ServiceName(gwNN),
			}
			// let's wait for the log to be sent to stdout
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(ctx context.Context) (bool, error) {
					count, err := QueryTraceFromTempo(t, suite.Client, tags)
					if err != nil {
						t.Logf("failed to get trace count from tempo: %v", err)
						return false, nil
					}

					if count > 0 {
						return true, nil
					}
					return false, nil
				}); err != nil {
				t.Errorf("failed to get trace from tempo: %v", err)
			}
		})
	},
}

// QueryTraceFromTempo queries span count from tempo
// TODO: move to utils package if needed
func QueryTraceFromTempo(t *testing.T, c client.Client, tags map[string]string) (int, error) {
	svc := corev1.Service{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "monitoring",
		Name:      "tempo",
	}, &svc); err != nil {
		return -1, err
	}
	host := ""
	for _, ing := range svc.Status.LoadBalancer.Ingress {
		if ing.IP != "" {
			host = ing.IP
			break
		}
	}

	tagsQueryParam, err := createTagsQueryParam(tags)
	if err != nil {
		return -1, err
	}

	tempoURL := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, "3100"),
		Path:   "/api/search",
	}
	query := tempoURL.Query()
	query.Add("start", fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix())) // query traces from last 10 minutes
	query.Add("end", fmt.Sprintf("%d", time.Now().Unix()))
	query.Add("tags", tagsQueryParam)
	tempoURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", tempoURL.String(), nil)
	if err != nil {

		return -1, err
	}

	t.Logf("send request to %s", tempoURL.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}

	if res.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("failed to query tempo, url=%s, status=%s", tempoURL.String(), res.Status)
	}

	tempoResponse := &tempopb.SearchResponse{}
	if err := jsonpb.Unmarshal(res.Body, tempoResponse); err != nil {
		return -1, err
	}

	total := len(tempoResponse.Traces)
	t.Logf("get response from tempo, url=%s, response=%v, total=%d", tempoURL.String(), tempoResponse, total)
	return total, nil
}

// copy from https://github.com/grafana/tempo/blob/c0127c78c368319433c7c67ca8967adbfed2259e/cmd/tempo-query/tempo/plugin.go#L361
func createTagsQueryParam(tags map[string]string) (string, error) {
	tagsBuilder := &strings.Builder{}
	tagsEncoder := logfmt.NewEncoder(tagsBuilder)
	for k, v := range tags {
		err := tagsEncoder.EncodeKeyval(k, v)
		if err != nil {
			return "", err
		}
	}
	return tagsBuilder.String(), nil
}
