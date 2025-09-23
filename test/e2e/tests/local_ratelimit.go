// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/test/e2e/utils"
)

func init() {
	ConformanceTests = append(ConformanceTests, LocalRateLimitTest)
}

const (
	RatelimitLimitHeaderName     = "x-ratelimit-limit"
	RatelimitRemainingHeaderName = "x-ratelimit-remaining"
	RatelimitResetHeaderName     = "x-ratelimit-reset"
)

var allRateLimitHeaders = []string{
	RatelimitLimitHeaderName,
	RatelimitRemainingHeaderName,
	RatelimitResetHeaderName,
}

var LocalRateLimitTest = suite.ConformanceTest{
	ShortName:   "LocalRateLimit",
	Description: "Make sure local rate limit work",
	Manifests:   []string{"testdata/local-ratelimit.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		for _, disableHeader := range []bool{true, false} {
			runNoRateLimitTest(t, suite, disableHeader)
			caseSuffix := "disableHeader"
			if !disableHeader {
				caseSuffix = "withHeader"
			}
			t.Run(fmt.Sprintf("SpecificUser-%s", caseSuffix), func(t *testing.T) {
				runSpecificUserRateLimitTest(t, suite, disableHeader)
			})

			t.Run(fmt.Sprintf("AllTraffic-%s", caseSuffix), func(t *testing.T) {
				runAllTrafficRateLimitTest(t, suite, disableHeader)
			})

			t.Run(fmt.Sprintf("HeaderInvertMatch-%s", caseSuffix), func(t *testing.T) {
				runHeaderInvertMatchRateLimitTest(t, suite, disableHeader)
			})

			t.Run(fmt.Sprintf("PathMatch-%s", caseSuffix), func(t *testing.T) {
				runPathMatchRateLimitTest(t, suite, disableHeader)
			})

			t.Run(fmt.Sprintf("MethodMatch-%s", caseSuffix), func(t *testing.T) {
				runMethodMatchRateLimitTest(t, suite, disableHeader)
			})
		}
	},
}

// gatewayNN return the gateway namespace name when disabled header or not
// All the HTTPRoute attached to the two gateways, the different is that we
// disabled rate limit headers on all-namespace gateway
func gatewayNN(disableHeader bool) types.NamespacedName {
	if disableHeader {
		return types.NamespacedName{Name: "all-namespaces", Namespace: "gateway-conformance-infra"}
	}
	return types.NamespacedName{Name: "same-namespace", Namespace: "gateway-conformance-infra"}
}

func gatewayAndHTTPRoutesMustBeAccepted(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName) string {
	gwRefs := []kubernetes.GatewayRef{
		kubernetes.NewGatewayRef(gatewayNN(true)),
		kubernetes.NewGatewayRef(gatewayNN(false)),
	}
	gwAddrMap := utils.GatewaysMustBeAccepted(t, suite, gwRefs)
	return gwAddrMap[gwNN]
}

func runNoRateLimitTest(t *testing.T, suite *suite.ConformanceTestSuite, disableHeader bool) {
	// let make sure the gateway and http route are accepted
	// and there's no rate limit on this route
	ns := "gateway-conformance-infra"
	gwNN := gatewayNN(disableHeader)
	gwAddr := gatewayAndHTTPRoutesMustBeAccepted(t, suite, gwNN)

	expectOkResp := http.ExpectedResponse{
		Request: http.Request{
			Path: "/no-ratelimit",
		},
		Response: http.Response{
			StatusCode:    200,
			AbsentHeaders: allRateLimitHeaders,
		},
		Namespace: ns,
	}

	// keep sending requests till get 200 first, that will cost one 200
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)

	// send 10+ more
	total := 10
	for total > 0 {
		// keep sending requests till get 200 first, that will cost one 200
		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)
		total--
	}
}

func runSpecificUserRateLimitTest(t *testing.T, suite *suite.ConformanceTestSuite, disableHeader bool) {
	ns := "gateway-conformance-infra"
	gwNN := gatewayNN(disableHeader)
	gwAddr := gatewayAndHTTPRoutesMustBeAccepted(t, suite, gwNN)

	ancestorRef := gwapiv1a2.ParentReference{
		Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
		Kind:      gatewayapi.KindPtr(resource.KindGateway),
		Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
		Name:      gwapiv1.ObjectName(gwNN.Name),
	}
	BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-specific-user", Namespace: ns}, suite.ControllerName, ancestorRef)

	// keep sending requests till get 200 first, that will cost one 200
	// use EG forked function to check the existence of the header
	okResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-specific-user",
			Headers: map[string]string{
				"x-user-id": "john",
			},
		},
		Response: http.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}
	if !disableHeader {
		okResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName:     "3",
			RatelimitRemainingHeaderName: "1",
			RatelimitResetHeaderName:     "0",
		}
	} else {
		okResponse.Response.AbsentHeaders = allRateLimitHeaders
	}
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, okResponse)

	// this request should be limited because the user is john
	limitResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-specific-user",
			Headers: map[string]string{
				"x-user-id": "john",
			},
		},
		Response: http.Response{
			StatusCode: 429,
		},
		Namespace: ns,
	}
	if !disableHeader {
		limitResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName:     "3",
			RatelimitRemainingHeaderName: "0",
		}
	} else {
		limitResponse.Response.AbsentHeaders = allRateLimitHeaders
	}
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, limitResponse)

	// this request should not be limited because the user is not john hit default bucket
	notJohnResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-specific-user",
			Headers: map[string]string{
				"x-user-id": "mike",
			},
		},
		Response: http.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}
	if !disableHeader {
		notJohnResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName:     "10",
			RatelimitRemainingHeaderName: "2", // there almost 8 requests before reach this
			RatelimitResetHeaderName:     "0",
		}
	} else {
		notJohnResponse.Response.AbsentHeaders = allRateLimitHeaders
	}
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, notJohnResponse)

	// In the end it will hit the limit
	notJohnLimitResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-specific-user",
			Headers: map[string]string{
				"x-user-id": "mike",
			},
		},
		Response: http.Response{
			StatusCode: 429,
		},
		Namespace: ns,
	}
	if !disableHeader {
		notJohnLimitResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName:     "10",
			RatelimitRemainingHeaderName: "0", // it will be limited at the end
		}
	} else {
		notJohnLimitResponse.Response.AbsentHeaders = allRateLimitHeaders
	}
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, notJohnLimitResponse)
}

func runAllTrafficRateLimitTest(t *testing.T, suite *suite.ConformanceTestSuite, disableHeader bool) {
	ns := "gateway-conformance-infra"
	gwNN := gatewayNN(disableHeader)
	gwAddr := gatewayAndHTTPRoutesMustBeAccepted(t, suite, gwNN)

	ancestorRef := gwapiv1a2.ParentReference{
		Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
		Kind:      gatewayapi.KindPtr(resource.KindGateway),
		Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
		Name:      gwapiv1.ObjectName(gwNN.Name),
	}
	BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-all-traffic", Namespace: ns}, suite.ControllerName, ancestorRef)

	okResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-all-traffic",
		},
		Response: http.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}
	if !disableHeader {
		okResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName:     "3",
			RatelimitRemainingHeaderName: "1",
			RatelimitResetHeaderName:     "0",
		}
	} else {
		okResponse.Response.AbsentHeaders = allRateLimitHeaders
	}
	// keep sending requests till get 200 first, that will cost one 200
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, okResponse)

	limitResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-all-traffic",
		},
		Response: http.Response{
			StatusCode: 429,
		},
		Namespace: ns,
	}
	if !disableHeader {
		limitResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName:     "3",
			RatelimitRemainingHeaderName: "0", // at the end the remaining should be 0
		}
	} else {
		limitResponse.Response.AbsentHeaders = allRateLimitHeaders
	}
	// this request should be limited at the end
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, limitResponse)
}

func runHeaderInvertMatchRateLimitTest(t *testing.T, suite *suite.ConformanceTestSuite, disableHeader bool) {
	ns := "gateway-conformance-infra"
	gwNN := gatewayNN(disableHeader)
	gwAddr := gatewayAndHTTPRoutesMustBeAccepted(t, suite, gwNN)

	ancestorRef := gwapiv1a2.ParentReference{
		Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
		Kind:      gatewayapi.KindPtr(resource.KindGateway),
		Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
		Name:      gwapiv1.ObjectName(gwNN.Name),
	}
	BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-invert-match", Namespace: ns}, suite.ControllerName, ancestorRef)

	// keep sending requests till get 200 first, that will cost one 200
	okResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-invert-match",
			Headers: map[string]string{
				"x-user-id": "one",
				"x-org-id":  "org1",
			},
		},
		Response: http.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}
	if !disableHeader {
		okResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName:     "3",
			RatelimitRemainingHeaderName: "1",
			RatelimitResetHeaderName:     "0",
		}
	}
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, okResponse)

	// this request should be limited because the user is one and org is not test and the limit is 3
	limitResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-invert-match",
			Headers: map[string]string{
				"x-user-id": "one",
				"x-org-id":  "org1",
			},
		},
		Response: http.Response{
			StatusCode: 429,
		},
		Namespace: ns,
	}
	if !disableHeader {
		limitResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName:     "3",
			RatelimitRemainingHeaderName: "0",
		}
	} else {
		limitResponse.Response.AbsentHeaders = allRateLimitHeaders
	}
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, limitResponse)

	// with test org
	testOrgResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-invert-match",
			Headers: map[string]string{
				"x-user-id": "one",
				"x-org-id":  "test",
			},
		},
		Response: http.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}
	if !disableHeader {
		testOrgResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName: "4294967295",
			RatelimitResetHeaderName: "0",
		}
	} else {
		testOrgResponse.Response.AbsentHeaders = allRateLimitHeaders
	}
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, testOrgResponse)
}

func runPathMatchRateLimitTest(t *testing.T, suite *suite.ConformanceTestSuite, disableHeader bool) {
	ns := "gateway-conformance-infra"
	gwNN := gatewayNN(disableHeader)
	gwAddr := gatewayAndHTTPRoutesMustBeAccepted(t, suite, gwNN)

	ancestorRef := gwapiv1a2.ParentReference{
		Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
		Kind:      gatewayapi.KindPtr(resource.KindGateway),
		Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
		Name:      gwapiv1.ObjectName(gwNN.Name),
	}
	BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-path-match", Namespace: ns}, suite.ControllerName, ancestorRef)

	okResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-path-match/foo",
		},
		Response: http.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}
	if !disableHeader {
		okResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName:     "3",
			RatelimitRemainingHeaderName: "1",
			RatelimitResetHeaderName:     "0",
		}
	} else {
		okResponse.Response.AbsentHeaders = allRateLimitHeaders
	}
	// keep sending requests till get 200 first, that will cost one 200
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, okResponse)

	limitResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-path-match/foo",
		},
		Response: http.Response{
			StatusCode: 429,
		},
		Namespace: ns,
	}
	if !disableHeader {
		limitResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName:     "3",
			RatelimitRemainingHeaderName: "0", // at the end the remaining should be 0
		}
	} else {
		limitResponse.Response.AbsentHeaders = allRateLimitHeaders
	}
	// this request should be limited at the end
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, limitResponse)

	okResponse = http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-path-match/bar",
		},
		Response: http.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}
	// this request should not be limited
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, okResponse)
}

func runMethodMatchRateLimitTest(t *testing.T, suite *suite.ConformanceTestSuite, disableHeader bool) {
	ns := "gateway-conformance-infra"
	gwNN := gatewayNN(disableHeader)
	gwAddr := gatewayAndHTTPRoutesMustBeAccepted(t, suite, gwNN)

	ancestorRef := gwapiv1a2.ParentReference{
		Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
		Kind:      gatewayapi.KindPtr(resource.KindGateway),
		Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
		Name:      gwapiv1.ObjectName(gwNN.Name),
	}
	BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-method-match", Namespace: ns}, suite.ControllerName, ancestorRef)

	okResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-method-match",
		},
		Response: http.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}
	if !disableHeader {
		okResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName:     "3",
			RatelimitRemainingHeaderName: "1",
			RatelimitResetHeaderName:     "0",
		}
	} else {
		okResponse.Response.AbsentHeaders = allRateLimitHeaders
	}
	// keep sending requests till get 200 first, that will cost one 200
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, okResponse)

	limitResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: "/ratelimit-method-match",
		},
		Response: http.Response{
			StatusCode: 429,
		},
		Namespace: ns,
	}
	if !disableHeader {
		limitResponse.Response.Headers = map[string]string{
			RatelimitLimitHeaderName:     "3",
			RatelimitRemainingHeaderName: "0", // at the end the remaining should be 0
		}
	} else {
		limitResponse.Response.AbsentHeaders = allRateLimitHeaders
	}
	// this request should be limited at the end
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, limitResponse)

	okResponse = http.ExpectedResponse{
		Request: http.Request{
			Path:   "/ratelimit-method-match",
			Method: "POST",
		},
		Response: http.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}
	// this request should not be limited
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, okResponse)
}
