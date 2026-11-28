// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/moby/moby/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

const (
	dynamicModuleOCIProxy   = "dynamic-module-oci-proxy"
	dynamicModuleOCIGateway = "dynamic-module-oci-gateway"
	dynamicModuleOCIRoute   = "http-with-oci-dynamic-module"
	dynamicModuleOCIEEP     = "oci-dynamic-module-test"
	dynamicModuleOCIPath    = "/lib/libheader_mutation.so"

	// dynamicModuleOCIDockerfile packs the shared library of the dynamic module test image,
	// which is built for the Envoy version under test, into an image with only that file.
	dynamicModuleOCIDockerfile = `FROM envoyproxy/gateway-dynamic-module-test:latest AS module
FROM scratch
COPY --from=module /usr/local/lib/libheader_mutation.so ` + dynamicModuleOCIPath + `
`
)

func init() {
	ConformanceTests = append(ConformanceTests, OCIDynamicModuleTest)
}

// OCIDynamicModuleTest tests a dynamic module loaded from an OCI image in a private registry.
var OCIDynamicModuleTest = suite.ConformanceTest{
	ShortName:   "DynamicModuleOCIImageSource",
	Description: "Test dynamic module extension loaded from an OCI image",
	Manifests:   []string{"testdata/dynamic-module-oci.yaml", "testdata/wasm-oci-registry-test-server.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		registryNN := types.NamespacedName{Name: "oci-registry", Namespace: testNS}
		registryIP, err := WaitForLoadBalancerAddress(t, suite.Client, 10*time.Second, registryNN)
		if err != nil {
			t.Fatalf("failed to get registry IP: %v", err)
		}
		registryAddr := net.JoinHostPort(registryIP, "5000")

		registryCAPEM := createRegistryTLSForWasmTest(t, suite, registryIP)
		digest := pushDynamicModuleImageForTest(t, suite, registryAddr, registryCAPEM)
		secret := createPullSecretForWasmTest(t, suite, registryAddr, dockerPassword)
		t.Cleanup(func() {
			_ = suite.Client.Delete(context.Background(), secret)
		})
		addDynamicModuleToEnvoyProxyForOCITest(t, suite, registryAddr, digest)

		t.Run("http route with oci dynamic module source", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: dynamicModuleOCIRoute, Namespace: testNS}
			gwNN := types.NamespacedName{Name: dynamicModuleOCIGateway, Namespace: testNS}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: dynamicModuleOCIEEP, Namespace: testNS}, suite.ControllerName, ancestorRef)

			// Wait for the Envoy proxy pods to be running and ready.
			WaitForPods(t, suite.Client, GetGatewayResourceNamespace(), map[string]string{
				"gateway.envoyproxy.io/owning-gateway-name":      gwNN.Name,
				"gateway.envoyproxy.io/owning-gateway-namespace": gwNN.Namespace,
			}, corev1.PodRunning, &PodReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/dynamic-module-oci",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"x-dynamic-module": "true",
					},
				},
				Namespace: testNS,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}

// pushDynamicModuleImageForTest builds an image with the dynamic module shared library,
// pushes it to the test registry and returns its digest.
func pushDynamicModuleImageForTest(t *testing.T, suite *suite.ConformanceTestSuite, registryAddr string, registryCAPEM []byte) string {
	podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}
	WaitForPods(
		t, suite.Client, testNS,
		map[string]string{"app": "oci-registry"}, corev1.PodRunning, &podReady)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	tag := fmt.Sprintf("%s/testdynamicmodule:v1.0.0", registryAddr)

	cli, err := client.New(client.FromEnv)
	if err != nil {
		t.Fatalf("failed to create docker client: %v", err)
	}

	var buildContext bytes.Buffer
	tw := tar.NewWriter(&buildContext)
	if err = tw.WriteHeader(&tar.Header{Name: "Dockerfile", Mode: 0o644, Size: int64(len(dynamicModuleOCIDockerfile))}); err != nil {
		t.Fatalf("failed to write build context: %v", err)
	}
	if _, err = tw.Write([]byte(dynamicModuleOCIDockerfile)); err != nil {
		t.Fatalf("failed to write build context: %v", err)
	}
	if err = tw.Close(); err != nil {
		t.Fatalf("failed to write build context: %v", err)
	}

	res, err := cli.ImageBuild(ctx, &buildContext, client.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{tag},
		Remove:     true,
	})
	if err != nil {
		t.Fatalf("failed to build image: %v", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()
	if err = printDockerCLIResponse(res.Body); err != nil {
		t.Fatalf("failed to build image: %v", err)
	}

	ref, err := name.ParseReference(tag)
	if err != nil {
		t.Fatalf("failed to parse reference: %v", err)
	}
	img, err := daemon.Image(ref)
	if err != nil {
		t.Fatalf("failed to retrieve image: %v", err)
	}

	authOption := remote.WithAuth(&authn.Basic{
		Username: dockerUsername,
		Password: dockerPassword,
	})
	transportOption := remote.WithTransport(registryTLSTransport(t, registryCAPEM))

	const retries = 5
	for range retries {
		if err = remote.Write(ref, img, authOption, transportOption); err == nil {
			break
		}
		tlog.Logf(t, "failed to push image: %v", err)
	}
	if err != nil {
		t.Fatalf("failed to push image: %v", err)
	}

	// The digest of the image in the registry, which the EnvoyProxy uses to verify the image.
	desc, err := remote.Get(ref, authOption, transportOption)
	if err != nil {
		t.Fatalf("failed to retrieve image: %v", err)
	}

	tlog.Logf(t, "pushed image %s with digest: %s", tag, desc.Digest.Hex)
	return desc.Digest.Hex
}

// addDynamicModuleToEnvoyProxyForOCITest registers the dynamic module from the image in the test registry
// in the EnvoyProxy of the test Gateway.
func addDynamicModuleToEnvoyProxyForOCITest(t *testing.T, suite *suite.ConformanceTestSuite, registryAddr, digest string) {
	dynamicModules := []egv1a1.DynamicModuleEntry{
		{
			Name: "header-mutation",
			Source: egv1a1.DynamicModuleSource{
				Type: new(egv1a1.ImageDynamicModuleSourceType),
				Image: &egv1a1.ImageDynamicModuleSource{
					URL:    fmt.Sprintf("%s/testdynamicmodule:v1.0.0", registryAddr),
					SHA256: &digest,
					Path:   new(dynamicModuleOCIPath),
					PullSecretRef: &gwapiv1.SecretObjectReference{
						Name: gwapiv1.ObjectName(pullSecret),
					},
					TLS: &egv1a1.WasmCodeSourceTLSConfig{
						CACertificateRef: gwapiv1.SecretObjectReference{
							Name:  gwapiv1.ObjectName(registryCAConfigMap),
							Group: gatewayapi.GroupPtr(""),
							Kind:  gatewayapi.KindPtr(resource.KindConfigMap),
						},
					},
				},
			},
		},
	}

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ep := &egv1a1.EnvoyProxy{}
		if err := suite.Client.Get(context.Background(), types.NamespacedName{Name: dynamicModuleOCIProxy, Namespace: testNS}, ep); err != nil {
			return err
		}
		ep.Spec.DynamicModules = dynamicModules
		return suite.Client.Update(context.Background(), ep)
	})
	if err != nil {
		t.Fatalf("failed to update envoy proxy: %v", err)
	}
}
