// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"io"
	nethttp "net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTP1ProperCaseWordsTest)
}

var HTTP1ProperCaseWordsTest = suite.ConformanceTest{
	ShortName:   "HTTP1ProperCaseWords",
	Description: "HTTP/1 header names use proper case words when ClientTrafficPolicy sets headerKeyFormat",
	Manifests:   []string{"testdata/http1-proper-case-words.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := ConformanceInfraNamespace

		t.Run("proper case words", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "http1-proper-case-gateway", Namespace: ns}
			routeNN := types.NamespacedName{Name: "http1-proper-case-words", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), routeNN)
			ClientTrafficPolicyMustBeAccepted(t, suite.Client,
				types.NamespacedName{Name: "http1-proper-case-ctp", Namespace: ns},
				suite.ControllerName, gatewayPolicyAncestor(ns, gwNN.Name))

			requireHeaderNames(t, suite, gwAddr, "/http1-proper-case", []string{
				"Location",
				"X-Added-By-Route",
			}, []string{
				"location",
				"x-added-by-route",
			})
		})

		t.Run("default lowercase header names", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "http1-default-header-case-gateway", Namespace: ns}
			routeNN := types.NamespacedName{Name: "http1-default-header-case", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), routeNN)

			requireHeaderNames(t, suite, gwAddr, "/http1-default-case", []string{
				"location",
				"x-added-by-route",
			}, []string{
				"Location",
				"X-Added-By-Route",
			})
		})
	},
}

func requireHeaderNames(
	t *testing.T, suite *suite.ConformanceTestSuite,
	gwAddr, path string, want, reject []string,
) {
	t.Helper()
	var last string
	err := wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
		reqCtx, cancel := context.WithTimeout(ctx, suite.TimeoutConfig.RequestTimeout)
		defer cancel()
		status, names, raw, reqErr := readHTTP1HeaderNames(reqCtx, gwAddr, path)
		last = raw
		if reqErr != nil {
			tlog.Logf(t, "request failed: %v", reqErr)
			return false, nil
		}
		if status != nethttp.StatusFound {
			tlog.Logf(t, "status %d\n%s", status, raw)
			return false, nil
		}
		for _, name := range want {
			if !slices.Contains(names, name) {
				tlog.Logf(t, "missing header %q in %v\n%s", name, names, raw)
				return false, nil
			}
		}
		for _, name := range reject {
			if slices.Contains(names, name) {
				tlog.Logf(t, "unexpected header %q in %v\n%s", name, names, raw)
				return false, nil
			}
		}
		return true, nil
	})
	require.NoError(t, err, "header names did not match\n%s", last)
}

// readHTTP1HeaderNames returns the status and the header names as Envoy wrote
// them. net/http canonicalizes header names, so the names come from the raw
// response bytes. RoundTrip does not follow the redirect.
func readHTTP1HeaderNames(ctx context.Context, gwAddr, path string) (int, []string, string, error) {
	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodGet, "http://"+gwAddr+path, nil)
	if err != nil {
		return 0, nil, "", err
	}
	req.Host = "www.example.com"
	req.Close = true

	tt := newTeeTransport(nil)
	resp, err := tt.RoundTrip(req)
	if err != nil {
		return 0, nil, tt.RawResponse(), err
	}
	defer resp.Body.Close()
	if _, err = io.Copy(io.Discard, resp.Body); err != nil {
		return 0, nil, tt.RawResponse(), err
	}

	raw := tt.RawResponse()
	head, _, _ := strings.Cut(raw, "\r\n\r\n")
	lines := strings.Split(head, "\r\n")
	names := make([]string, 0, len(lines))
	for _, line := range lines[1:] {
		if name, _, ok := strings.Cut(line, ":"); ok {
			names = append(names, name)
		}
	}
	return resp.StatusCode, names, raw, nil
}
