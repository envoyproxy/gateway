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
