// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package utils

import (
	"fmt"
	"os"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/yaml"
)

type BenchmarkTestSuite struct {
	Client client.Client

	// Resources template for supported benchmark targets.
	GatewayTemplate   *gwapiv1.Gateway
	HTTPRouteTemplate *gwapiv1.HTTPRoute
}

func NewBenchmarkTestSuite(client client.Client, gatewayManifest, httpRouteManifest string) (*BenchmarkTestSuite, error) {
	var (
		gateway   *gwapiv1.Gateway
		httproute *gwapiv1.HTTPRoute
	)

	gm, err := os.ReadFile(gatewayManifest)
	if err != nil {
		return nil, err
	}

	hm, err := os.ReadFile(httpRouteManifest)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(gm, gateway); err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(hm, httproute); err != nil {
		return nil, err
	}

	return &BenchmarkTestSuite{
		Client:            client,
		GatewayTemplate:   gateway,
		HTTPRouteTemplate: httproute,
	}, nil
}

func (b *BenchmarkTestSuite) Run(t *testing.T, tests []BenchmarkTest) {
	t.Logf("Running benchmark test")

	for _, test := range tests {
		t.Logf("Running benchmark test: %s", test.ShortName)

		b.runTarget(t, test)
	}
}

func (b *BenchmarkTestSuite) runTarget(t *testing.T, test BenchmarkTest) {
	for gatewayName, gatewayTarget := range test.GatewayTargets {
		if gatewayTarget == nil {
			continue
		}

		matched := false
		for httprouteName, httprouteTarget := range test.HTTPRoutTargets {
			if httprouteTarget != nil &&
				gatewayName == httprouteTarget.TargetGateway {
				t.Logf("Running benchmark test: %s for '%s' gateway and '%s' httproutes",
					test.ShortName, gatewayName, httprouteName)

				matched = true
				b.runBounds(t, test, gatewayTarget, httprouteTarget)
				break
			}
		}

		if !matched {
			t.Errorf("Unable to find any matched httproutes for gateway '%s'", gatewayName)
		}
	}
}

func (b *BenchmarkTestSuite) runBounds(t *testing.T, test BenchmarkTest, gatewayTarget *GatewayTarget, httprouteTarget *HTTPRouteTarget) {
	for _, gatewayBound := range gatewayTarget.Bounds {
		for _, httprouteBound := range httprouteTarget.Bounds {
			name := fmt.Sprintf("gateway scale = %d and httproute scale = %d", gatewayBound, httprouteBound)

			t.Logf("Running benchmark test: %s with %s", test.ShortName, name)

			t.Run(name, func(t *testing.T) {
				// TODO: implement benchmark run
			})
		}
	}
}
