package gatewayapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/yaml"
)

func mustUnmarshal(t *testing.T, val string, out interface{}) {
	require.NoError(t, yaml.UnmarshalStrict([]byte(val), out, yaml.DisallowUnknownFields))
}

func TestTranslate(t *testing.T) {
	inputFiles, err := filepath.Glob(filepath.Join("testdata", "*.in.yaml"))
	require.NoError(t, err)

	for _, inputFile := range inputFiles {
		inputFile := inputFile
		t.Run(testName(inputFile), func(t *testing.T) {
			input, err := os.ReadFile(inputFile)
			require.NoError(t, err)

			resources := &Resources{}
			mustUnmarshal(t, string(input), resources)

			output, err := os.ReadFile(strings.ReplaceAll(inputFile, ".in.yaml", ".out.yaml"))
			require.NoError(t, err)

			want := &TranslateResult{}
			mustUnmarshal(t, string(output), want)

			translator := &Translator{
				GatewayClassName: "envoy-gateway-class",
			}

			// Add common test fixtures
			for i := 1; i <= 3; i++ {
				resources.Services = append(resources.Services,
					&v1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "service-" + strconv.Itoa(i),
						},
						Spec: v1.ServiceSpec{
							ClusterIP: "7.7.7.7",
							Ports: []v1.ServicePort{
								{Port: 8080},
								{Port: 8443},
							},
						},
					},
				)
			}

			resources.Namespaces = append(resources.Namespaces, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "envoy-gateway",
				},
			}, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			})

			got := translator.Translate(resources)

			opts := cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")
			require.Empty(t, cmp.Diff(want, got, opts))
		})
	}
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
			err:      "hostname \".example.test.com\" is invalid for a redirect filter: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
		{
			name:     "dot-suffix",
			hostname: "example.test.com.",
			err:      "hostname \"example.test.com.\" is invalid for a redirect filter: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
		{
			name:     "ip-address",
			hostname: "192.168.254.254",
			err:      "hostname: \"192.168.254.254\" cannot be an ip address",
		},
		{
			name:     "dash-prefix",
			hostname: "-example.test.com",
			err:      "hostname \"-example.test.com\" is invalid for a redirect filter: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
		{
			name:     "dash-suffix",
			hostname: "example.test.com-",
			err:      "hostname \"example.test.com-\" is invalid for a redirect filter: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
		{
			name:     "invalid-symbol",
			hostname: "examp!e.test.com",
			err:      "hostname \"examp!e.test.com\" is invalid for a redirect filter: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
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
			err:      fmt.Sprintf("hostname %q is invalid for a redirect filter: [must be no more than 253 characters]", veryLongHostname),
		},
		{
			name:     "empty-hostname",
			hostname: "",
			err:      "hostname \"\" is invalid for a redirect filter: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
		{
			name:     "double-dot",
			hostname: "example..test.com",
			err:      "hostname \"example..test.com\" is invalid for a redirect filter: [a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')]",
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := isValidHostname(tc.hostname)
			if tc.err == "" {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestIsValidCrossNamespaceRef(t *testing.T) {
	type testcase struct {
		name           string
		from           crossNamespaceFrom
		to             crossNamespaceTo
		referenceGrant *v1alpha2.ReferenceGrant
		want           bool
	}

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
			referenceGrant: &v1alpha2.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "referencegrant-1",
					Namespace: "default",
				},
				Spec: v1alpha2.ReferenceGrantSpec{
					From: []v1alpha2.ReferenceGrantFrom{
						{
							Group:     "gateway.networking.k8s.io",
							Kind:      "Gateway",
							Namespace: "envoy-gateway-system",
						},
					},
					To: []v1alpha2.ReferenceGrantTo{
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
			var referenceGrants []*v1alpha2.ReferenceGrant
			if tc.referenceGrant != nil {
				referenceGrants = append(referenceGrants, tc.referenceGrant)
			}

			assert.Equal(t, tc.want, isValidCrossNamespaceRef(tc.from, tc.to, referenceGrants))
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
