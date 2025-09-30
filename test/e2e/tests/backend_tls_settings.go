// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTLSSettingsTest)
}

var BackendTLSSettingsTest = suite.ConformanceTest{
	ShortName:   "BackendTLSSettings",
	Description: "Use envoy proxy tls settings with backend",
	Manifests:   []string{"testdata/backend-tls-settings.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		proxyNN := types.NamespacedName{Name: "backend-tls-setting", Namespace: ConformanceInfraNamespace}
		gwNN := types.NamespacedName{Name: "backend-tls-setting", Namespace: ConformanceInfraNamespace}
		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ConformanceInfraNamespace})

		t.Run("Apply custom TLS settings when making backend requests.", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "backend-tls-setting", Namespace: ConformanceInfraNamespace}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			config := &egv1a1.BackendTLSConfig{
				ClientCertificateRef: &gwapiv1.SecretObjectReference{
					Kind:      gatewayapi.KindPtr("Secret"),
					Name:      "client-tls-certificate",
					Namespace: gatewayapi.NamespacePtr(ConformanceInfraNamespace),
				},
				TLSSettings: egv1a1.TLSSettings{
					MinVersion:    ptr.To(egv1a1.TLSv13),
					MaxVersion:    ptr.To(egv1a1.TLSv13),
					ALPNProtocols: []egv1a1.ALPNProtocol{"http/1.1"},
				},
			}
			err := UpdateProxyConfig(suite.Client, proxyNN, config)
			if err != nil {
				t.Error(err)
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
				Namespace: ConformanceInfraNamespace,
			}

			// Reconfigure backend tls settings
			err = WaitUntil(func(httpRes *http.ExpectedResponse, expectedResBody *Response) error {
				return confirmEchoBackendRes(httpRes, expectedResBody, gwAddr, t, suite)
			}, suite.TimeoutConfig.MaxTimeToConsistency, &expectOkResp, expectedRes)
			if err != nil {
				t.Errorf("failed to confirm echo backend for TLS v1.3: %v", err)
			}

			// rotate the client mTLS secret to ensure that a new secret is used.
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/backend-tls-settings-client-cert-rotation.yaml", false)

			err = restartDeploymentAndWaitForRollout(t, &suite.TimeoutConfig, suite.Client, &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-backend",
					Namespace: "gateway-conformance-infra",
				},
			})
			if err != nil {
				t.Errorf("failed to confirm echo backend for client cert rotation: %v", err)
			}

			// confirm new mtls client cert is used when connecting to backend
			expectedResNewMTLSSecret, err := asExpectedResponse("echo-service-tls-settings-new-mtls-secret")
			if err != nil {
				t.Error(err)
			}

			err = WaitUntil(func(httpRes *http.ExpectedResponse, expectedResBody *Response) error {
				return confirmEchoBackendRes(httpRes, expectedResBody, gwAddr, t, suite)
			}, suite.TimeoutConfig.MaxTimeToConsistency, &expectOkResp, expectedResNewMTLSSecret)
			if err != nil {
				t.Error(err)
			}

			t.Logf("updating backend tls settings to use TLSv1.2 and custom cipher suites")
			config.TLSSettings = egv1a1.TLSSettings{
				MinVersion:    ptr.To(egv1a1.TLSv12),
				MaxVersion:    ptr.To(egv1a1.TLSv12),
				ALPNProtocols: nil, // default ALPN protocols, which means h2 preferred over http/1.1.
				Ciphers:       []string{"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"},
			}

			err = UpdateProxyConfig(suite.Client, proxyNN, config)
			if err != nil {
				t.Errorf("failed to confirm echo backend for TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384: %v", err)
			}

			// confirm tls settings can be updated
			expectedUpdatedTLSSettings, err := asExpectedResponse("echo-service-tls-settings-updated-tls-settings")
			if err != nil {
				t.Error(err)
			}

			err = WaitUntil(func(httpRes *http.ExpectedResponse, expectedResBody *Response) error {
				return confirmEchoBackendRes(httpRes, expectedResBody, gwAddr, t, suite)
			}, suite.TimeoutConfig.MaxTimeToConsistency, &expectOkResp, expectedUpdatedTLSSettings)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("UseClientProtocol", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "use-client-protocol", Namespace: ConformanceInfraNamespace}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			// rotate the client mTLS secret to ensure that a new secret is used.
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/backend-tls-settings-client-cert-rotation.yaml", false)
			err := restartDeploymentAndWaitForRollout(t, &suite.TimeoutConfig, suite.Client, &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tls-backend",
					Namespace: "gateway-conformance-infra",
				},
			})
			if err != nil {
				t.Errorf("failed to confirm echo backend for client cert rotation: %v", err)
			}

			config := &egv1a1.BackendTLSConfig{
				ClientCertificateRef: &gwapiv1.SecretObjectReference{
					Kind:      gatewayapi.KindPtr("Secret"),
					Name:      "client-tls-certificate",
					Namespace: gatewayapi.NamespacePtr(ConformanceInfraNamespace),
				},
				TLSSettings: egv1a1.TLSSettings{
					MinVersion: ptr.To(egv1a1.TLSv13),
					MaxVersion: ptr.To(egv1a1.TLSv13),
				},
			}
			err = UpdateProxyConfig(suite.Client, proxyNN, config)
			if err != nil {
				t.Error(err)
			}

			t.Logf("requesting /use-client-protocol with HTTP/1.1")
			expectedRes, err := asExpectedResponse("client-protocol-http1")
			if err != nil {
				t.Error(err)
			}
			expectedRes.TLS.NegotiatedProtocol = "http/1.1"
			expectOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/use-client-protocol",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ConformanceInfraNamespace,
			}
			err = WaitUntil(func(httpRes *http.ExpectedResponse, expectedResBody *Response) error {
				return confirmEchoBackendRes(httpRes, expectedResBody, gwAddr, t, suite)
			}, suite.TimeoutConfig.MaxTimeToConsistency, &expectOkResp, expectedRes)
			if err != nil {
				t.Errorf("failed to confirm echo backend for HTTP/1.1: %v", err)
			}

			// Client use HTTP/2
			t.Logf("requesting /use-client-protocol with HTTP/2")
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path:     "/use-client-protocol",
					Protocol: roundtripper.H2CPriorKnowledgeProtocol,
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ConformanceInfraNamespace,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRoundTrip(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
			if cReq.Protocol != "HTTP/2.0" {
				t.Errorf("expected http/2.0 protocol, got %s", cReq.Protocol)
			}
		})
	},
}

func confirmEchoBackendRes(httpRes *http.ExpectedResponse, expectedResBody *Response, gwAddr string, t *testing.T, suite *suite.ConformanceTestSuite) error {
	transport := &nethttp.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		},
	}
	req := http.MakeRequest(t, httpRes, gwAddr, "HTTP", "http")
	res, err := casePreservingRoundTrip(&req, transport, suite)
	if err != nil {
		return err
	}
	err = expectNewEchoBackendResponse(res, expectedResBody)
	if err != nil {
		return err
	}
	return nil
}

// UpdateProxyConfig updates the proxy configuration with BackendTLS settings.
func UpdateProxyConfig(client client.Client, proxyNN types.NamespacedName, config *egv1a1.BackendTLSConfig) error {
	proxyConfig := &egv1a1.EnvoyProxy{}
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

	if cmp.Equal(res.TLS, expect.TLS) {
		return nil
	}
	return fmt.Errorf("mismatch found between returned and expected response. Difference: %s", cmp.Diff(res.TLS, expect.TLS))
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

// Response defines echo server response
type Response struct {
	TLS TLSInfo `json:"tls"`
}

type TLSInfo struct {
	Version            string   `json:"version"`
	PeerCertificates   []string `json:"peerCertificates"`
	ServerName         string   `json:"serverName"`
	NegotiatedProtocol string   `json:"negotiatedProtocol"`
	CipherSuite        string   `json:"cipherSuite"`
}

func restartDeploymentAndWaitForRollout(t *testing.T, timeoutConfig *config.TimeoutConfig, c client.Client, dp *appsv1.Deployment) error {
	t.Helper()
	const restartAnnotation = "kubectl.kubernetes.io/restartedAt"
	restartTime := time.Now().Format(time.RFC3339)
	ctx := context.Background()

	if timeoutConfig == nil {
		t.Fatalf("timeoutConfig cannot be nil")
	}

	if err := c.Get(context.Background(), types.NamespacedName{Name: dp.Name, Namespace: dp.Namespace}, dp); err != nil {
		return err
	}

	// Update an annotation to trigger a rolling update
	if dp.Spec.Template.Annotations == nil {
		dp.Spec.Template.Annotations = make(map[string]string)
	}
	dp.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartTime

	if err := c.Update(ctx, dp); err != nil {
		return err
	}

	return wait.PollUntilContextTimeout(ctx, 1*time.Second, timeoutConfig.CreateTimeout, true, func(ctx context.Context) (bool, error) {
		// wait for replicaset with the same annotation to reach ready status
		podList := &corev1.PodList{}
		listOpts := []client.ListOption{
			client.InNamespace(dp.Namespace),
			client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(dp.Spec.Selector.MatchLabels)},
		}

		err := c.List(ctx, podList, listOpts...)
		if err != nil {
			return false, err
		}

		rolled := int32(0)
		for i := range podList.Items {
			rs := &podList.Items[i]
			if rs.Annotations[restartAnnotation] == restartTime {
				rolled++
			}
		}

		// all pods are rolled
		if rolled == int32(len(podList.Items)) && rolled >= *dp.Spec.Replicas {
			return true, nil
		}

		return false, nil
	})
}
