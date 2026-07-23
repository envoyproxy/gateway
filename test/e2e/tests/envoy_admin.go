// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	"github.com/envoyproxy/gateway/internal/troubleshoot/collect"
)

// fetchEnvoyClustersOutput returns the raw text served by the admin /clusters endpoint of the
// Envoy proxy pod matching selector, fetched via port-forward (see
// collect.RequestWithPortForwarder, used the same way in test/benchmark/suite/report.go for
// pprof).
func fetchEnvoyClustersOutput(t *testing.T, suite *suite.ConformanceTestSuite, selector ...string) (string, error) {
	t.Helper()

	cli, err := kube.NewForRestConfig(suite.RestConfig)
	if err != nil {
		return "", err
	}

	pods, err := cli.PodsForSelector(GetGatewayResourceNamespace(), selector...)
	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		t.Fatalf("no Envoy pod found for selector %v", selector)
	}

	body, err := collect.RequestWithPortForwarder(cli, types.NamespacedName{
		Namespace: pods.Items[0].Namespace,
		Name:      pods.Items[0].Name,
	}, 19000, "/clusters")
	if err != nil {
		return "", err
	}
	return string(body), nil
}
