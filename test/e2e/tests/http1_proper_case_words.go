// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
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
			gwHost := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
			acceptClientTrafficPolicy(t, suite, types.NamespacedName{Name: "http1-proper-case-ctp", Namespace: ns}, gwNN)

			requireHeaderNames(t, suite, net.JoinHostPort(gwHost, "80"), "/http1-proper-case", []string{
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
			gwHost := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			requireHeaderNames(t, suite, net.JoinHostPort(gwHost, "80"), "/http1-default-case", []string{
				"location",
				"x-added-by-route",
			}, []string{
				"Location",
				"X-Added-By-Route",
			})
		})
	},
}

func acceptClientTrafficPolicy(t *testing.T, suite *suite.ConformanceTestSuite, policy types.NamespacedName, gwNN types.NamespacedName) {
	t.Helper()
	ancestorRef := gwapiv1.ParentReference{
		Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
		Kind:      gatewayapi.KindPtr(resource.KindGateway),
		Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
		Name:      gwapiv1.ObjectName(gwNN.Name),
	}
	ClientTrafficPolicyMustBeAccepted(t, suite.Client, policy, suite.ControllerName, ancestorRef)
}

func requireHeaderNames(t *testing.T, suite *suite.ConformanceTestSuite, addr, path string, want, reject []string) {
	t.Helper()
	var last string
	err := wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(context.Context) (bool, error) {
		status, names, raw, reqErr := readHTTP1HeaderNames(addr, path, "www.example.com")
		last = raw
		if reqErr != nil {
			tlog.Logf(t, "request failed: %v", reqErr)
			return false, nil
		}
		if status != 302 {
			tlog.Logf(t, "status %d\n%s", status, raw)
			return false, nil
		}
		for _, name := range want {
			if !headerNamePresent(names, name) {
				tlog.Logf(t, "missing header %q in %v\n%s", name, names, raw)
				return false, nil
			}
		}
		for _, name := range reject {
			if headerNamePresent(names, name) {
				tlog.Logf(t, "unexpected header %q in %v\n%s", name, names, raw)
				return false, nil
			}
		}
		return true, nil
	})
	require.NoError(t, err, "header names did not match\n%s", last)
}

func headerNamePresent(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}
	return false
}

// readHTTP1HeaderNames reads the status and header names from the raw HTTP/1.1
// response. net/http canonicalizes header names, so it cannot show wire case.
func readHTTP1HeaderNames(addr, path, host string) (int, []string, string, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return 0, nil, "", err
	}
	defer conn.Close()
	if err = conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return 0, nil, "", err
	}
	if _, err = io.WriteString(conn, fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", path, host)); err != nil {
		return 0, nil, "", err
	}

	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return 0, nil, statusLine, err
	}
	var status int
	if _, err = fmt.Sscanf(statusLine, "HTTP/1.%d %d", new(int), &status); err != nil {
		return 0, nil, statusLine, fmt.Errorf("status line %q: %w", strings.TrimSpace(statusLine), err)
	}

	var raw strings.Builder
	raw.WriteString(statusLine)
	var names []string
	for {
		line, readErr := reader.ReadString('\n')
		raw.WriteString(line)
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" || readErr != nil && line == "" {
			if readErr != nil && readErr != io.EOF {
				return status, names, raw.String(), readErr
			}
			break
		}
		name, _, ok := strings.Cut(trimmed, ":")
		if ok {
			names = append(names, name)
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return status, names, raw.String(), readErr
		}
	}
	return status, names, raw.String(), nil
}
