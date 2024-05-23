// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha3"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

const UpstreamTLSChangesMaxTimeout = 30 * time.Second

func init() {
	ConformanceTests = append(ConformanceTests, UpstreamTLSSettingsTest)
}

var UpstreamTLSSettingsTest = suite.ConformanceTest{
	ShortName:   "Upstream tls settings",
	Description: "Use envoy proxy tls settings with upstream",
	Manifests:   []string{"testdata/upstream-tls.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Apply custom TLS settings when making upstream requests.", func(t *testing.T) {
			depNS := "envoy-gateway-system"
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-backend-tls", Namespace: ns}
			gwNN := types.NamespacedName{Name: "backend-namespaces", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{depNS})
			backendTLSPolicy := &v1alpha3.BackendTLSPolicy{}
			btpNN := types.NamespacedName{Name: "policy-btls", Namespace: ns}
			err := suite.Client.Get(context.Background(), btpNN, backendTLSPolicy)
			if err != nil {
				t.Error(err)
			}
			proxyNN := types.NamespacedName{Name: "proxy-config", Namespace: "envoy-gateway-system"}

			config := &v1alpha1.BackendTLSConfig{
				ClientCertificateRef: &gwapiv1.SecretObjectReference{
					Kind:      gatewayapi.KindPtr("Secret"),
					Name:      "client-tls-certificate",
					Namespace: gatewayapi.NamespacePtr(depNS),
				},
				TLSSettings: v1alpha1.TLSSettings{
					MinVersion: ptr.To(v1alpha1.TLSv13),
					MaxVersion: ptr.To(v1alpha1.TLSv13),
				},
			}
			err = UpdateProxyConfig(suite.Client, proxyNN, config)
			if err != nil {
				t.Error(err)
			}
			transport := &nethttp.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec
				},
			}

			expectedRes, err := asExpectedResponse("echo-service-tls-settings-res")
			if err != nil {
				t.Error(err)
			}
			expectOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-tls",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			confirmEchoBackendRes := func(httpRes *http.ExpectedResponse, expectedResBody *Response) error {
				req := http.MakeRequest(t, httpRes, gwAddr, "HTTPS", "https")
				res, err := casePreservingRoundTrip(req, transport, suite)
				if err != nil {
					t.Log(err)
				}
				err = expectNewEchoBackendResponse(res, expectedResBody)
				if err != nil {
					return err
				}
				return nil
			}
			// Reconfigure upstream tls settings
			err = WaitUntil(func(httpRes *http.ExpectedResponse, expectedResBody *Response) error {
				return confirmEchoBackendRes(httpRes, expectedResBody)
			}, UpstreamTLSChangesMaxTimeout, &expectOkResp, expectedRes)
			if err != nil {
				t.Error(err)
			}
			// Ensure that changes to envoy proxy re-configure the upstream tls settings.
			config.TLSSettings = v1alpha1.TLSSettings{
				MinVersion: ptr.To(v1alpha1.TLSv12),
				MaxVersion: ptr.To(v1alpha1.TLSv12),
				Ciphers:    []string{"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"},
			}
			err = UpdateProxyConfig(suite.Client, proxyNN, config)
			if err != nil {
				t.Error(err)
			}
			expectedRes.TLS.Version = "TLSv1.2"
			expectedRes.TLS.CipherSuite = "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
			err = WaitUntil(func(httpRes *http.ExpectedResponse, expectedResBody *Response) error {
				return confirmEchoBackendRes(httpRes, expectedResBody)
			}, UpstreamTLSChangesMaxTimeout, &expectOkResp, expectedRes)
			if err != nil {
				t.Error(err)
			}

			// Cleanup upstream tls settings.
			config.TLSSettings = v1alpha1.TLSSettings{}
			err = UpdateProxyConfig(suite.Client, proxyNN, config)
			if err != nil {
				t.Error(err)
			}
		})
	},
}

// UpdateProxyConfig updates the proxy configuration with BackendTLS settings.
func UpdateProxyConfig(client client.Client, proxyNN types.NamespacedName, config *v1alpha1.BackendTLSConfig) error {
	proxyConfig := &v1alpha1.EnvoyProxy{}
	err := client.Get(context.Background(), proxyNN, proxyConfig)
	if err != nil {
		return err
	}

	proxyConfig.Spec.BackendTLS = config
	err = client.Update(context.Background(), proxyConfig)
	if err != nil {
		return err
	}
	return nil
}

// WaitUntil repeatedly calls the provided check function until it returns true or the timeout is reached.
// It returns true if the check function returns true within the timeout period, otherwise false.
func WaitUntil(check func(httpRes *http.ExpectedResponse, expectedResBody *Response) error, timeout time.Duration, expectedResponse *http.ExpectedResponse, expectResBody *Response) error {
	end := time.Now().Add(timeout)
	var err error
	for time.Now().Before(end) {
		if err = check(expectedResponse, expectResBody); err == nil {
			return nil
		}
		time.Sleep(1 * time.Second) // Wait a bit before trying again
	}
	return err
}

func expectNewEchoBackendResponse(respBody interface{}, expect *Response) error {
	res := &Response{}
	bytes, err := json.Marshal(respBody)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, res)
	if err != nil {
		return err
	}

	if ok, err := hasAllFieldsAndValues(res.TLS, expect.TLS); !ok {
		return err
	}
	return nil
}

func asExpectedResponse(fileName string) (*Response, error) {
	var res Response
	filename := path.Join("testdata", "expect", fmt.Sprintf("%s.yaml", fileName))

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// Function to check if obj1 has all the fields of obj2, including nested fields, and matching values
func hasAllFieldsAndValues(obj1, obj2 interface{}) (bool, error) {
	return hasAllFieldsAndValuesRecursive(reflect.ValueOf(obj1), reflect.ValueOf(obj2))
}

func hasAllFieldsAndValuesRecursive(v1, v2 reflect.Value) (bool, error) {
	if v1.Kind() == reflect.Ptr {
		v1 = v1.Elem()
	}
	if v2.Kind() == reflect.Ptr {
		v2 = v2.Elem()
	}

	if v1.Kind() != reflect.Struct || v2.Kind() != reflect.Struct {
		return false, fmt.Errorf("both parameters must be structs")
	}

	t1 := v1.Type()
	t2 := v2.Type()

	for i := 0; i < t2.NumField(); i++ {
		field2 := t2.Field(i)
		field1, found := t1.FieldByName(field2.Name)

		if !found {
			fmt.Printf("Field %s is missing in obj1\n", field2.Name)
			return false, nil
		}

		value1 := v1.FieldByName(field2.Name)
		value2 := v2.Field(i)

		// Recursively check nested fields
		if field2.Type.Kind() == reflect.Struct {
			hasFields, err := hasAllFieldsAndValuesRecursive(value1, value2)
			if err != nil || !hasFields {
				return hasFields, err
			}
		} else {
			// Check if the field types and values are the same
			if field1.Type != field2.Type {
				return false, fmt.Errorf("field %s has different type in obj1: %v (expected %v)", field2.Name, field1.Type, field2.Type)
			}
			if !reflect.DeepEqual(value1.Interface(), value2.Interface()) {
				return false, fmt.Errorf("field %s has different value in obj1: %v (expected %v)", field2.Name, value1.Interface(), value2.Interface())
			}
		}
	}

	return true, nil
}

// Response defines echo server response
type Response struct {
	Path      string              `json:"path"`
	Host      string              `json:"host"`
	Method    string              `json:"method"`
	Proto     string              `json:"proto"`
	Headers   map[string][]string `json:"headers"`
	Namespace string              `json:"namespace"`
	Ingress   string              `json:"ingress"`
	Service   string              `json:"service"`
	Pod       string              `json:"pod"`
	TLS       TLSInfo             `json:"tls"`
}

type TLSInfo struct {
	Version            string   `json:"version"`
	PeerCertificates   []string `json:"peerCertificates"`
	ServerName         string   `json:"serverName"`
	NegotiatedProtocol string   `json:"negotiatedProtocol"`
	CipherSuite        string   `json:"cipherSuite"`
}
