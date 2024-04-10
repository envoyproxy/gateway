// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/field"
	"github.com/envoyproxy/gateway/internal/utils/file"
)

var (
	overrideTestData = flag.Bool("override-testdata", false, "if override the test output data.")
)

func mustUnmarshal(t *testing.T, val []byte, out interface{}) {
	require.NoError(t, yaml.UnmarshalStrict(val, out, yaml.DisallowUnknownFields))
}

func TestTranslate(t *testing.T) {
	testCasesConfig := []struct {
		name                    string
		EnvoyPatchPolicyEnabled bool
	}{
		{
			name:                    "envoypatchpolicy-invalid-feature-disabled",
			EnvoyPatchPolicyEnabled: false,
		},
	}

	inputFiles, err := filepath.Glob(filepath.Join("testdata", "*.in.yaml"))
	require.NoError(t, err)

	for _, inputFile := range inputFiles {
		inputFile := inputFile
		t.Run(testName(inputFile), func(t *testing.T) {
			input, err := os.ReadFile(inputFile)
			require.NoError(t, err)

			resources := &Resources{}
			mustUnmarshal(t, input, resources)
			envoyPatchPolicyEnabled := true

			for _, config := range testCasesConfig {
				if config.name == strings.Split(filepath.Base(inputFile), ".")[0] {
					envoyPatchPolicyEnabled = config.EnvoyPatchPolicyEnabled
				}
			}

			translator := &Translator{
				GatewayControllerName:   egv1a1.GatewayControllerName,
				GatewayClassName:        "envoy-gateway-class",
				GlobalRateLimitEnabled:  true,
				EnvoyPatchPolicyEnabled: envoyPatchPolicyEnabled,
				Namespace:               "envoy-gateway-system",
				MergeGateways:           IsMergeGatewaysEnabled(resources),
			}

			// Add common test fixtures
			for i := 1; i <= 4; i++ {
				svcName := "service-" + strconv.Itoa(i)
				epSliceName := "endpointslice-" + strconv.Itoa(i)
				resources.Services = append(resources.Services,
					&corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      svcName,
						},
						Spec: corev1.ServiceSpec{
							ClusterIP: "1.1.1.1",
							Ports: []corev1.ServicePort{
								{
									Name:       "http",
									Port:       8080,
									TargetPort: intstr.IntOrString{IntVal: 8080},
									Protocol:   corev1.ProtocolTCP,
								},
								{
									Name:       "https",
									Port:       8443,
									TargetPort: intstr.IntOrString{IntVal: 8443},
									Protocol:   corev1.ProtocolTCP,
								},
								{
									Name:       "tcp",
									Port:       8163,
									TargetPort: intstr.IntOrString{IntVal: 8163},
									Protocol:   corev1.ProtocolTCP,
								},
								{
									Name:       "udp",
									Port:       8162,
									TargetPort: intstr.IntOrString{IntVal: 8162},
									Protocol:   corev1.ProtocolUDP,
								},
							},
						},
					},
				)
				resources.EndpointSlices = append(resources.EndpointSlices,
					&discoveryv1.EndpointSlice{
						ObjectMeta: metav1.ObjectMeta{
							Name:      epSliceName,
							Namespace: "default",
							Labels: map[string]string{
								discoveryv1.LabelServiceName: svcName,
							},
						},
						AddressType: discoveryv1.AddressTypeIPv4,
						Ports: []discoveryv1.EndpointPort{
							{
								Name:     ptr.To("http"),
								Port:     ptr.To[int32](8080),
								Protocol: ptr.To(corev1.ProtocolTCP),
							},
							{
								Name:     ptr.To("https"),
								Port:     ptr.To[int32](8443),
								Protocol: ptr.To(corev1.ProtocolTCP),
							},
							{
								Name:     ptr.To("tcp"),
								Port:     ptr.To[int32](8163),
								Protocol: ptr.To(corev1.ProtocolTCP),
							},
							{
								Name:     ptr.To("udp"),
								Port:     ptr.To[int32](8162),
								Protocol: ptr.To(corev1.ProtocolUDP),
							},
						},
						Endpoints: []discoveryv1.Endpoint{
							{
								Addresses: []string{
									"7.7.7.7",
								},
								Conditions: discoveryv1.EndpointConditions{
									Ready: ptr.To(true),
								},
							},
						},
					},
				)
			}

			resources.Services = append(resources.Services,
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "mirror-service",
					},
					Spec: corev1.ServiceSpec{
						ClusterIP: "2.2.2.2",
						Ports: []corev1.ServicePort{
							{
								Name:       "http",
								Port:       8080,
								TargetPort: intstr.IntOrString{IntVal: 8080},
								Protocol:   corev1.ProtocolTCP,
							},
						},
					},
				},
			)
			resources.EndpointSlices = append(resources.EndpointSlices,
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mirror-service-endpointslice",
						Namespace: "default",
						Labels: map[string]string{
							discoveryv1.LabelServiceName: "mirror-service",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Ports: []discoveryv1.EndpointPort{
						{
							Name:     ptr.To("http"),
							Port:     ptr.To[int32](8080),
							Protocol: ptr.To(corev1.ProtocolTCP),
						},
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{
								"7.6.5.4",
							},
							Conditions: discoveryv1.EndpointConditions{
								Ready: ptr.To(true),
							},
						},
					},
				},
			)

			resources.Namespaces = append(resources.Namespaces, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "envoy-gateway",
				},
			}, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			})

			got := translator.Translate(resources)
			require.NoError(t, field.SetValue(got, "LastTransitionTime", metav1.NewTime(time.Time{})))
			outputFilePath := strings.ReplaceAll(inputFile, ".in.yaml", ".out.yaml")
			out, err := yaml.Marshal(got)
			require.NoError(t, err)

			if *overrideTestData {
				overrideOutputConfig(t, string(out), outputFilePath)
			}

			output, err := os.ReadFile(outputFilePath)
			require.NoError(t, err)

			want := &TranslateResult{}
			mustUnmarshal(t, output, want)

			opts := []cmp.Option{
				cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime"),
				cmpopts.EquateEmpty(),
			}

			require.Empty(t, cmp.Diff(want, got, opts...))
		})
	}
}

func TestTranslateWithExtensionKinds(t *testing.T) {
	inputFiles, err := filepath.Glob(filepath.Join("testdata/extensions", "*.in.yaml"))
	require.NoError(t, err)

	for _, inputFile := range inputFiles {
		inputFile := inputFile
		t.Run(testName(inputFile), func(t *testing.T) {
			input, err := os.ReadFile(inputFile)
			require.NoError(t, err)

			resources := &Resources{}
			mustUnmarshal(t, input, resources)

			translator := &Translator{
				GatewayControllerName:  egv1a1.GatewayControllerName,
				GatewayClassName:       "envoy-gateway-class",
				GlobalRateLimitEnabled: true,
				ExtensionGroupKinds:    []schema.GroupKind{{Group: "foo.example.io", Kind: "Foo"}},
				MergeGateways:          IsMergeGatewaysEnabled(resources),
			}

			// Add common test fixtures
			for i := 1; i <= 3; i++ {
				svcName := "service-" + strconv.Itoa(i)
				epSliceName := "endpointslice-" + strconv.Itoa(i)
				resources.Services = append(resources.Services,
					&corev1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      svcName,
						},
						Spec: corev1.ServiceSpec{
							ClusterIP: "1.1.1.1",
							Ports: []corev1.ServicePort{
								{
									Name:       "http",
									Port:       8080,
									TargetPort: intstr.IntOrString{IntVal: 8080},
									Protocol:   corev1.ProtocolTCP,
								},
								{
									Name:       "https",
									Port:       8443,
									TargetPort: intstr.IntOrString{IntVal: 8443},
									Protocol:   corev1.ProtocolTCP,
								},
								{
									Name:       "tcp",
									Port:       8163,
									TargetPort: intstr.IntOrString{IntVal: 8163},
									Protocol:   corev1.ProtocolTCP,
								},
								{
									Name:       "udp",
									Port:       8162,
									TargetPort: intstr.IntOrString{IntVal: 8162},
									Protocol:   corev1.ProtocolUDP,
								},
							},
						},
					},
				)
				resources.EndpointSlices = append(resources.EndpointSlices,
					&discoveryv1.EndpointSlice{
						ObjectMeta: metav1.ObjectMeta{
							Name:      epSliceName,
							Namespace: "default",
							Labels: map[string]string{
								discoveryv1.LabelServiceName: svcName,
							},
						},
						AddressType: discoveryv1.AddressTypeIPv4,
						Ports: []discoveryv1.EndpointPort{
							{
								Name:     ptr.To("http"),
								Port:     ptr.To[int32](8080),
								Protocol: ptr.To(corev1.ProtocolTCP),
							},
							{
								Name:     ptr.To("https"),
								Port:     ptr.To[int32](8443),
								Protocol: ptr.To(corev1.ProtocolTCP),
							},
							{
								Name:     ptr.To("tcp"),
								Port:     ptr.To[int32](8163),
								Protocol: ptr.To(corev1.ProtocolTCP),
							},
							{
								Name:     ptr.To("udp"),
								Port:     ptr.To[int32](8162),
								Protocol: ptr.To(corev1.ProtocolUDP),
							},
						},
						Endpoints: []discoveryv1.Endpoint{
							{
								Addresses: []string{
									"7.7.7.7",
								},
								Conditions: discoveryv1.EndpointConditions{
									Ready: ptr.To(true),
								},
							},
						},
					},
				)
			}

			resources.Services = append(resources.Services,
				&corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "mirror-service",
					},
					Spec: corev1.ServiceSpec{
						ClusterIP: "2.2.2.2",
						Ports: []corev1.ServicePort{
							{
								Port:       8080,
								TargetPort: intstr.IntOrString{IntVal: 8080},
								Protocol:   corev1.ProtocolTCP,
							},
						},
					},
				},
			)
			resources.EndpointSlices = append(resources.EndpointSlices,
				&discoveryv1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mirror-service-endpointslice",
						Namespace: "default",
						Labels: map[string]string{
							discoveryv1.LabelServiceName: "mirror-service",
						},
					},
					AddressType: discoveryv1.AddressTypeIPv4,
					Ports: []discoveryv1.EndpointPort{
						{
							Name:     ptr.To("http"),
							Port:     ptr.To[int32](8080),
							Protocol: ptr.To(corev1.ProtocolTCP),
						},
					},
					Endpoints: []discoveryv1.Endpoint{
						{
							Addresses: []string{
								"7.6.5.4",
							},
							Conditions: discoveryv1.EndpointConditions{
								Ready: ptr.To(true),
							},
						},
					},
				},
			)

			resources.Namespaces = append(resources.Namespaces, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "envoy-gateway",
				},
			}, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			})

			got := translator.Translate(resources)
			require.NoError(t, field.SetValue(got, "LastTransitionTime", metav1.NewTime(time.Time{})))

			outputFilePath := strings.ReplaceAll(inputFile, ".in.yaml", ".out.yaml")
			out, err := yaml.Marshal(got)
			require.NoError(t, err)

			if *overrideTestData {
				require.NoError(t, file.Write(string(out), outputFilePath))
			}

			output, err := os.ReadFile(outputFilePath)
			require.NoError(t, err)

			want := &TranslateResult{}
			mustUnmarshal(t, output, want)

			opts := cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")
			require.Empty(t, cmp.Diff(want, got, opts))
		})
	}
}

func overrideOutputConfig(t *testing.T, data string, filepath string) {
	t.Helper()
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	require.NoError(t, err)
	defer file.Close()
	write := bufio.NewWriter(file)
	_, err = write.WriteString(data)
	require.NoError(t, err)
	write.Flush()
}

func testName(inputFile string) string {
	_, fileName := filepath.Split(inputFile)
	return strings.TrimSuffix(fileName, ".in.yaml")
}

func TestIsValidHostname(t *testing.T) {
	type testcase struct {
		name     string
		hostname string
		err      string
	}

	translator := &Translator{}

	// Setting up a hostname that is 256+ characters for a test case that does not also trip the max label size
	veryLongHostname := "a"
	label := 0
	for i := 0; i < 256; i++ {
		if label > 10 {
			veryLongHostname += "."
			label = 0
		} else {
			veryLongHostname += string(veryLongHostname[0])
		}
		label++
	}
	veryLongHostname += ".com"

	testcases := []*testcase{
		{
			name:     "good-hostname",
			hostname: "example.test.com",
			err:      "",
		},
		{
			name:     "dot-prefix",
			hostname: ".example.test.com",
			err:      "hostname \".example.test.com\" is invalid: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
		{
			name:     "dot-suffix",
			hostname: "example.test.com.",
			err:      "hostname \"example.test.com.\" is invalid: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
		{
			name:     "ip-address",
			hostname: "192.168.254.254",
			err:      "hostname: \"192.168.254.254\" cannot be an ip address",
		},
		{
			name:     "dash-prefix",
			hostname: "-example.test.com",
			err:      "hostname \"-example.test.com\" is invalid: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
		{
			name:     "dash-suffix",
			hostname: "example.test.com-",
			err:      "hostname \"example.test.com-\" is invalid: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
		{
			name:     "invalid-symbol",
			hostname: "examp!e.test.com",
			err:      "hostname \"examp!e.test.com\" is invalid: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
		{
			name:     "long-label",
			hostname: "example.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.com",
			err:      "label: \"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\" in hostname \"example.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.com\" cannot exceed 63 characters",
		},
		{
			name:     "long-label-last-index",
			hostname: "example.abc.commmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm",
			err:      "label: \"commmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm\" in hostname \"example.abc.commmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmmm\" cannot exceed 63 characters",
		},
		{
			name:     "way-too-long-hostname",
			hostname: veryLongHostname,
			err:      fmt.Sprintf("hostname %q is invalid: [must be no more than 253 characters]", veryLongHostname),
		},
		{
			name:     "empty-hostname",
			hostname: "",
			err:      "hostname \"\" is invalid: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
		{
			name:     "double-dot",
			hostname: "example..test.com",
			err:      "hostname \"example..test.com\" is invalid: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := translator.validateHostname(tc.hostname)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestIsValidCrossNamespaceRef(t *testing.T) {
	type testcase struct {
		name           string
		from           crossNamespaceFrom
		to             crossNamespaceTo
		referenceGrant *v1beta1.ReferenceGrant
		want           bool
	}

	translator := &Translator{}

	var testcases []*testcase

	baseCase := func() *testcase {
		return &testcase{
			name: "reference covered by reference grant (all resources of kind)",
			from: crossNamespaceFrom{
				group:     "gateway.networking.k8s.io",
				kind:      "Gateway",
				namespace: "envoy-gateway-system",
			},
			to: crossNamespaceTo{
				group:     "",
				kind:      "Secret",
				namespace: "default",
				name:      "tls-secret-1",
			},
			referenceGrant: &v1beta1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "referencegrant-1",
					Namespace: "default",
				},
				Spec: v1beta1.ReferenceGrantSpec{
					From: []v1beta1.ReferenceGrantFrom{
						{
							Group:     "gateway.networking.k8s.io",
							Kind:      "Gateway",
							Namespace: "envoy-gateway-system",
						},
					},
					To: []v1beta1.ReferenceGrantTo{
						{
							Group: "",
							Kind:  "Secret",
						},
					},
				},
			},
			want: true,
		}
	}

	testcases = append(testcases, baseCase())

	modified := baseCase()
	modified.name = "reference covered by reference grant (named resource)"
	modified.referenceGrant.Spec.To[0].Name = ObjectNamePtr("tls-secret-1")
	testcases = append(testcases, modified)

	modified = baseCase()
	modified.name = "no reference grants"
	modified.referenceGrant = nil
	modified.want = false
	testcases = append(testcases, modified)

	modified = baseCase()
	modified.name = "reference not covered by reference grant (wrong from namespace)"
	modified.referenceGrant.Spec.From[0].Namespace = "wrong-namespace"
	modified.want = false
	testcases = append(testcases, modified)

	modified = baseCase()
	modified.name = "reference not covered by reference grant (wrong from kind)"
	modified.referenceGrant.Spec.From[0].Kind = "WrongKind"
	modified.want = false
	testcases = append(testcases, modified)

	modified = baseCase()
	modified.name = "reference not covered by reference grant (wrong from group)"
	modified.referenceGrant.Spec.From[0].Group = "wrong.group.k8s.io"
	modified.want = false
	testcases = append(testcases, modified)

	modified = baseCase()
	modified.name = "reference not covered by reference grant (wrong to name)"
	modified.referenceGrant.Spec.To[0].Name = ObjectNamePtr("wrong-name")
	modified.want = false
	testcases = append(testcases, modified)

	modified = baseCase()
	modified.name = "reference not covered by reference grant (wrong to namespace)"
	modified.referenceGrant.Namespace = "wrong-namespace"
	modified.want = false
	testcases = append(testcases, modified)

	modified = baseCase()
	modified.name = "reference not covered by reference grant (wrong to kind)"
	modified.referenceGrant.Spec.To[0].Kind = "WrongKind"
	modified.want = false
	testcases = append(testcases, modified)

	modified = baseCase()
	modified.name = "reference not covered by reference grant (wrong to group)"
	modified.referenceGrant.Spec.To[0].Group = "wrong.group.k8s.io"
	modified.want = false
	testcases = append(testcases, modified)

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var referenceGrants []*v1beta1.ReferenceGrant
			if tc.referenceGrant != nil {
				referenceGrants = append(referenceGrants, tc.referenceGrant)
			}

			assert.Equal(t, tc.want, translator.validateCrossNamespaceRef(tc.from, tc.to, referenceGrants))
		})
	}
}

func TestServicePortToContainerPort(t *testing.T) {
	testCases := []struct {
		servicePort   int32
		containerPort int32
	}{
		{
			servicePort:   99,
			containerPort: 10099,
		},
		{
			servicePort:   1023,
			containerPort: 11023,
		},
		{
			servicePort:   1024,
			containerPort: 1024,
		},
		{
			servicePort:   8080,
			containerPort: 8080,
		},
	}

	for _, tc := range testCases {
		got := servicePortToContainerPort(tc.servicePort)
		assert.Equal(t, tc.containerPort, got)
	}
}
