// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

const (
	dockerUsername       = "testuser"
	dockerPassword       = "testpassword"
	dockerEmail          = "your-email@example.com"
	testNS               = "gateway-conformance-infra"
	testGW               = "same-namespace"
	testEEP              = "oci-wasm-source-test"
	pullSecret           = "registry-secret"
	httpRouteWithWasm    = "http-with-oci-wasm-source"
	httpRouteWithoutWasm = "http-without-wasm"
)

func init() {
	ConformanceTests = append(ConformanceTests, OCIWasmTest)
}

// OCIWasmTest tests Wasm extension for an http route with OCI Wasm configured.
var OCIWasmTest = suite.ConformanceTest{
	ShortName:   "WasmOCIImageCodeSource",
	Description: "Test OCI Wasm extension",
	Manifests:   []string{"testdata/wasm-oci.yaml", "testdata/wasm-oci-registry-test-server.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		// Get the LoadBalancer IP of the registry
		registryNN := types.NamespacedName{Name: "oci-registry", Namespace: testNS}
		registryIP, err := WaitForLoadBalancerAddress(t, suite.Client, 10*time.Second, registryNN)
		if err != nil {
			t.Fatalf("failed to get registry IP: %v", err)
		}
		registryAddr := net.JoinHostPort(registryIP, "5000")

		// Push the wasm image to the registry
		digest := pushWasmImageForTest(t, suite, registryAddr)

		// Create the pull secret for the wasm image
		secret := createPullSecretForWasmTest(t, suite, registryAddr, dockerPassword)

		// Create the EnvoyExtensionPolicy referencing the wasm image
		eep := createEEPForWasmTest(t, suite, registryAddr, digest, true)

		// Wait for the EnvoyExtensionPolicy to be accepted
		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(testNS),
			Name:      gwapiv1.ObjectName(testGW),
		}

		EnvoyExtensionPolicyMustBeAccepted(
			t, suite.Client,
			types.NamespacedName{Name: testEEP, Namespace: testNS},
			suite.ControllerName,
			ancestorRef)

		// HTTPRoute configured with the correct wasm extension should modify the response
		t.Run("http route with oci wasm source", func(t *testing.T) {
			// Wait for the HTTPRoute to be accepted
			routeNN := types.NamespacedName{Name: httpRouteWithWasm, Namespace: testNS}
			gwNN := types.NamespacedName{Name: testGW, Namespace: testNS}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			// Make a request to the gateway and expect the wasm filter to add a response header
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/wasm-oci",
				},

				// Set the expected request properties to empty strings.
				// This is a workaround to avoid the test failure.
				// These values can't be extracted from the json format response
				// body because the test wasm code appends a "Hello, world" text
				// to the response body, invalidating the json format.
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host:    "",
						Method:  "",
						Path:    "",
						Headers: nil,
					},
				},
				Namespace: "",

				Response: http.Response{
					StatusCode: 200,
					Headers: map[string]string{
						"x-wasm-custom": "FOO", // response header added by wasm
					},
				},
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		// HTTPRoute without wasm should not modify the response
		t.Run("http route without wasm", func(t *testing.T) {
			EnvoyExtensionPolicyMustBeAccepted(
				t, suite.Client,
				types.NamespacedName{Name: testEEP, Namespace: testNS},
				suite.ControllerName,
				ancestorRef)

			ns := testNS
			routeNN := types.NamespacedName{Name: httpRouteWithoutWasm, Namespace: ns}
			gwNN := types.NamespacedName{Name: testGW, Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/no-wasm",
				},
				Response: http.Response{
					StatusCode:    200,
					AbsentHeaders: []string{"x-wasm-custom"},
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})

		// Verify that the wasm module can't be loaded if the pull secret is missing
		// even if the wasm image is already cached.
		t.Run("without pull secret", func(t *testing.T) {
			// Delete the EnvoyExtensionPolicy with pull secret
			_ = suite.Client.Delete(context.Background(), eep)

			// Create the EnvoyExtensionPolicy without pull secret
			createEEPForWasmTest(t, suite, registryAddr, digest, false)

			defer func() {
				_ = suite.Client.Delete(context.Background(), eep)
			}()

			// Wait for the EnvoyExtensionPolicy to be failed due to missing pull secret
			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(testNS),
				Name:      gwapiv1.ObjectName(testGW),
			}

			EnvoyExtensionPolicyMustFail(
				t, suite.Client,
				types.NamespacedName{Name: testEEP, Namespace: testNS},
				suite.ControllerName,
				ancestorRef, "UNAUTHORIZED: authentication required")
		})

		// Verify that the wasm module can't be loaded if the password is incorrect
		// even if the wasm image is already cached.
		t.Run("with wrong password", func(t *testing.T) {
			// Delete the EnvoyExtensionPolicy with pull secret
			_ = suite.Client.Delete(context.Background(), eep)

			// Delete the pull secret
			_ = suite.Client.Delete(context.Background(), secret)

			// Create the pull secret with a wrong password
			secret = createPullSecretForWasmTest(t, suite, registryAddr, "wrongpassword")

			// Create the EnvoyExtensionPolicy without pull secret
			eep = createEEPForWasmTest(t, suite, registryAddr, digest, true)

			defer func() {
				_ = suite.Client.Delete(context.Background(), eep)
				_ = suite.Client.Delete(context.Background(), secret)
			}()

			// Wait for the EnvoyExtensionPolicy to be failed due to missing pull secret
			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(testNS),
				Name:      gwapiv1.ObjectName(testGW),
			}

			EnvoyExtensionPolicyMustFail(
				t, suite.Client,
				types.NamespacedName{Name: testEEP, Namespace: testNS},
				suite.ControllerName,
				ancestorRef, "UNAUTHORIZED: authentication required")
		})
	},
}

func pushWasmImageForTest(t *testing.T, suite *suite.ConformanceTestSuite, registryAddr string) string {
	// Wait for the registry pod to be ready
	podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}
	WaitForPods(
		t, suite.Client, testNS,
		map[string]string{"app": "oci-registry"}, corev1.PodRunning, podReady)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	var (
		cli    *client.Client
		tar    io.Reader
		res    dockertypes.ImageBuildResponse
		digest v1.Hash
		err    error
	)

	tag := fmt.Sprintf("%s/testwasm:v1.0.0", registryAddr)

	if cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); err != nil {
		t.Fatalf("failed to create docker client: %v", err)
	}

	if tar, err = archive.TarWithOptions("testdata/wasm", &archive.TarOptions{}); err != nil {
		t.Fatalf("failed to create tar: %v", err)
	}

	opts := dockertypes.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{tag},
		Remove:     true,
	}
	if res, err = cli.ImageBuild(ctx, tar, opts); err != nil {
		t.Fatalf("failed to build image: %v", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if err = printDockerCLIResponse(res.Body); err != nil {
		t.Fatalf("failed to print docker cli response: %v", err)
	}

	ref, err := name.ParseReference(tag, name.Insecure)
	if err != nil {
		t.Fatalf("failed to parse reference: %v", err)
	}

	// Retrieve the image from the local Docker daemon
	img, err := daemon.Image(ref)
	if err != nil {
		t.Fatalf("failed to retrieve image: %v", err)
	}

	authOption := remote.WithAuth(&authn.Basic{
		Username: dockerUsername,
		Password: dockerPassword,
	})

	const retries = 5
	for i := 0; i < retries; i++ {
		// Push the image to the remote registry
		// err = crane.Push(img, tag)
		err = remote.Write(ref, img, authOption)
		if err == nil {
			break
		}
		tlog.Logf(t, "failed to push image: %v", err)
	}
	if err != nil {
		t.Fatalf("failed to push image: %v", err)
	}

	if img, err = remote.Image(ref, authOption); err != nil {
		t.Fatalf("failed to retrieve image: %v", err)
	}
	if digest, err = img.Digest(); err != nil {
		t.Fatalf("failed to get image digest: %v", err)
	}

	tlog.Logf(t, "pushed image %s with digest: %s", tag, digest.Hex)
	return digest.Hex
}

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func printDockerCLIResponse(rd io.Reader) error {
	var lastLine string

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lastLine = scanner.Text()
		fmt.Println(scanner.Text())
	}

	errLine := &ErrorLine{}
	_ = json.Unmarshal([]byte(lastLine), errLine)
	if errLine.Error != "" {
		return errors.New(errLine.Error)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func createPullSecretForWasmTest(t *testing.T, suite *suite.ConformanceTestSuite, registryAddr string, password string) *corev1.Secret {
	// Create Docker config JSON
	dockerConfigJSON := fmt.Sprintf(`{"auths":{"%s":{"username":"%s","password":"%s","email":"%s","auth":"%s"}}}`,
		registryAddr, dockerUsername, password, dockerEmail,
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", dockerUsername, password))))

	// Create a Secret object
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pullSecret,
			Namespace: testNS,
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: []byte(dockerConfigJSON),
		},
	}

	// Create the secret in the specified namespace
	_ = suite.Client.Delete(context.Background(), secret)
	if err := suite.Client.Create(context.Background(), secret); err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}
	return secret
}

func createEEPForWasmTest(
	t *testing.T, suite *suite.ConformanceTestSuite,
	registryAddr string, digest string, withPullSecret bool,
) *egv1a1.EnvoyExtensionPolicy {
	eep := &egv1a1.EnvoyExtensionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testEEP,
			Namespace: testNS,
		},
		Spec: egv1a1.EnvoyExtensionPolicySpec{
			PolicyTargetReferences: egv1a1.PolicyTargetReferences{
				TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
							Group: "gateway.networking.k8s.io",
							Kind:  "HTTPRoute",
							Name:  httpRouteWithWasm,
						},
					},
				},
			},

			Wasm: []egv1a1.Wasm{
				{
					Name:   ptr.To("wasm-filter"),
					RootID: ptr.To("my_root_id"),
					Code: egv1a1.WasmCodeSource{
						Type: egv1a1.ImageWasmCodeSourceType,
						Image: &egv1a1.ImageWasmCodeSource{
							URL:    fmt.Sprintf("%s/testwasm:v1.0.0", registryAddr),
							SHA256: &digest,
						},
					},
				},
			},
		},
	}
	if withPullSecret {
		eep.Spec.Wasm[0].Code.Image.PullSecretRef = &gwapiv1.SecretObjectReference{
			Name: gwapiv1.ObjectName(pullSecret),
		}
	}
	// Create the EnvoyExtensionPolicy in the specified namespace
	if err := suite.Client.Create(context.Background(), eep); err != nil {
		t.Fatalf("failed to create envoy extension policy: %v", err)
	}
	return eep
}
