// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dag

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	gatewayapi_v1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	// LoadBalancerPolicyWeightedLeastRequest specifies the backend with least
	// active requests will be chosen by the load balancer.
	LoadBalancerPolicyWeightedLeastRequest = "WeightedLeastRequest"

	// LoadBalancerPolicyRandom denotes the load balancer will choose a random
	// backend when routing a request.
	LoadBalancerPolicyRandom = "Random"

	// LoadBalancerPolicyRoundRobin denotes the load balancer will route
	// requests in a round-robin fashion among backend instances.
	LoadBalancerPolicyRoundRobin = "RoundRobin"

	// LoadBalancerPolicyCookie denotes load balancing will be performed via a
	// Contour specified cookie.
	LoadBalancerPolicyCookie = "Cookie"

	// LoadBalancerPolicyRequestHash denotes request attribute hashing is used
	// to make load balancing decisions.
	LoadBalancerPolicyRequestHash = "RequestHash"
)

// headersPolicyGatewayAPI builds a *HeaderPolicy for the supplied HTTPRequestHeaderFilter.
// TODO: Take care about the order of operators once https://github.com/kubernetes-sigs/gateway-api/issues/480 was solved.
func headersPolicyGatewayAPI(hf *gatewayapi_v1alpha2.HTTPRequestHeaderFilter) (*HeadersPolicy, error) {
	var (
		set         = make(map[string]string, len(hf.Set))
		add         = make(map[string]string, len(hf.Add))
		remove      = sets.NewString()
		hostRewrite = ""
		errlist     = []error{}
	)

	for _, setHeader := range hf.Set {
		key := http.CanonicalHeaderKey(string(setHeader.Name))
		if _, ok := set[key]; ok {
			errlist = append(errlist, fmt.Errorf("duplicate header addition: %q", key))
			continue
		}
		if key == "Host" {
			hostRewrite = setHeader.Value
			continue
		}
		if msgs := validation.IsHTTPHeaderName(key); len(msgs) != 0 {
			errlist = append(errlist, fmt.Errorf("invalid set header %q: %v", key, msgs))
			continue
		}
		set[key] = escapeHeaderValue(setHeader.Value, nil)
	}
	for _, addHeader := range hf.Add {
		key := http.CanonicalHeaderKey(string(addHeader.Name))
		if _, ok := add[key]; ok {
			errlist = append(errlist, fmt.Errorf("duplicate header addition: %q", key))
			continue
		}
		if key == "Host" {
			hostRewrite = addHeader.Value
			continue
		}
		if msgs := validation.IsHTTPHeaderName(key); len(msgs) != 0 {
			errlist = append(errlist, fmt.Errorf("invalid add header %q: %v", key, msgs))
			continue
		}
		add[key] = escapeHeaderValue(addHeader.Value, nil)
	}

	for _, k := range hf.Remove {
		key := http.CanonicalHeaderKey(k)
		if remove.Has(key) {
			errlist = append(errlist, fmt.Errorf("duplicate header removal: %q", key))
			continue
		}
		if msgs := validation.IsHTTPHeaderName(key); len(msgs) != 0 {
			errlist = append(errlist, fmt.Errorf("invalid remove header %q: %v", key, msgs))
			continue
		}
		remove.Insert(key)
	}
	rl := remove.List()

	if len(set) == 0 {
		set = nil
	}
	if len(rl) == 0 {
		rl = nil
	}

	return &HeadersPolicy{
		Add:         add,
		Set:         set,
		HostRewrite: hostRewrite,
		Remove:      rl,
	}, utilerrors.NewAggregate(errlist)
}

func escapeHeaderValue(value string, dynamicHeaders map[string]string) string {
	// Envoy supports %-encoded variables, so literal %'s in the header's value must be escaped.  See:
	// https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#custom-request-response-headers
	// Only allow a specific set of known good Envoy dynamic headers to pass through unescaped
	if !strings.Contains(value, "%") {
		return value
	}
	escapedValue := strings.ReplaceAll(value, "%", "%%")
	for dynamicVar, dynamicVal := range dynamicHeaders {
		escapedValue = strings.ReplaceAll(escapedValue, "%%"+dynamicVar+"%%", dynamicVal)
	}
	for _, envoyVar := range []string{
		"DOWNSTREAM_REMOTE_ADDRESS",
		"DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT",
		"DOWNSTREAM_LOCAL_ADDRESS",
		"DOWNSTREAM_LOCAL_ADDRESS_WITHOUT_PORT",
		"DOWNSTREAM_LOCAL_PORT",
		"DOWNSTREAM_LOCAL_URI_SAN",
		"DOWNSTREAM_PEER_URI_SAN",
		"DOWNSTREAM_LOCAL_SUBJECT",
		"DOWNSTREAM_PEER_SUBJECT",
		"DOWNSTREAM_PEER_ISSUER",
		"DOWNSTREAM_TLS_SESSION_ID",
		"DOWNSTREAM_TLS_CIPHER",
		"DOWNSTREAM_TLS_VERSION",
		"DOWNSTREAM_PEER_FINGERPRINT_256",
		"DOWNSTREAM_PEER_FINGERPRINT_1",
		"DOWNSTREAM_PEER_SERIAL",
		"DOWNSTREAM_PEER_CERT",
		"DOWNSTREAM_PEER_CERT_V_START",
		"DOWNSTREAM_PEER_CERT_V_END",
		"HOSTNAME",
		"PROTOCOL",
		"UPSTREAM_REMOTE_ADDRESS",
		"RESPONSE_FLAGS",
		"RESPONSE_CODE_DETAILS",
	} {
		escapedValue = strings.ReplaceAll(escapedValue, "%%"+envoyVar+"%%", "%"+envoyVar+"%")
	}
	// REQ(header-name)
	var validReqEnvoyVar = regexp.MustCompile(`%(%REQ\([\w-]+\)%)%`)
	escapedValue = validReqEnvoyVar.ReplaceAllString(escapedValue, "$1")
	return escapedValue
}
