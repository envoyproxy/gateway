package gatewayapi

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

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
				gatewayClassName: "envoy-gateway-class",
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

			sort.Slice(got.IR.HTTP, func(i, j int) bool { return got.IR.HTTP[i].Name < got.IR.HTTP[j].Name })

			assert.EqualValues(t, want, got)
		})
	}
}

func testName(inputFile string) string {
	_, fileName := filepath.Split(inputFile)
	return strings.TrimSuffix(fileName, ".in.yaml")
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
		t.Run(tc.name, func(t *testing.T) {
			var referenceGrants []*v1alpha2.ReferenceGrant
			if tc.referenceGrant != nil {
				referenceGrants = append(referenceGrants, tc.referenceGrant)
			}

			assert.Equal(t, tc.want, isValidCrossNamespaceRef(tc.from, tc.to, referenceGrants))
		})
	}
}
