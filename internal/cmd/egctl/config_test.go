// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	"github.com/envoyproxy/gateway/internal/utils/file"
	netutil "github.com/envoyproxy/gateway/internal/utils/net"
)

const (
	defaultNamespace           = "default"
	defaultEnvoyGatewayPodName = "eg"
)

var _ kube.PortForwarder = &fakePortForwarder{}

type fakePortForwarder struct {
	responseBody []byte
	localPort    int
	l            net.Listener
	mux          *http.ServeMux
}

func newFakePortForwarder(b []byte) (kube.PortForwarder, error) {
	p, err := netutil.LocalAvailablePort()
	if err != nil {
		return nil, err
	}

	fw := &fakePortForwarder{
		responseBody: b,
		localPort:    p,
		mux:          http.NewServeMux(),
	}
	fw.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(fw.responseBody)
	})

	return fw, nil
}

func (fw *fakePortForwarder) Start() error {
	l, err := net.Listen("tcp", fw.Address())
	if err != nil {
		return err
	}
	fw.l = l

	go func() {
		// nolint: gosec
		if err := http.Serve(l, fw.mux); err != nil {
			log.Fatal(err)
		}
	}()

	return nil
}

func (fw *fakePortForwarder) Stop() {}

func (fw *fakePortForwarder) Address() string {
	return fmt.Sprintf("localhost:%d", fw.localPort)
}

func (fw *fakePortForwarder) WaitForStop() {}

func TestExtractAllConfigDump(t *testing.T) {
	input, err := readInputConfig("in.all.json")
	require.NoError(t, err)
	fw, err := newFakePortForwarder(input)
	require.NoError(t, err)
	err = fw.Start()
	require.NoError(t, err)

	cases := []struct {
		output       string
		expected     string
		resourceType string
	}{
		{
			output:   "json",
			expected: "out.all.json",
		},
		{
			output:   "yaml",
			expected: "out.all.yaml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			configDump, err := extractConfigDump(fw, true, AllEnvoyConfigType)
			require.NoError(t, err)
			aggregated := sampleAggregatedConfigDump(configDump)
			got, err := marshalEnvoyProxyConfig(aggregated, tc.output)
			require.NoError(t, err)
			if *overrideTestData {
				require.NoError(t, file.Write(string(got), filepath.Join("testdata", "config", "out", tc.expected)))
			}
			out, err := readOutputConfig(tc.expected)
			require.NoError(t, err)
			if tc.output == "yaml" {
				assert.YAMLEq(t, string(out), string(got))
			} else {
				assert.JSONEq(t, string(out), string(got))
			}
		})
	}

	fw.Stop()
}

func TestExtractSubResourcesConfigDump(t *testing.T) {
	input, err := readInputConfig("in.all.json")
	require.NoError(t, err)
	fw, err := newFakePortForwarder(input)
	require.NoError(t, err)
	err = fw.Start()
	require.NoError(t, err)

	cases := []struct {
		output       string
		expected     string
		resourceType envoyConfigType
	}{
		{
			output:       "json",
			resourceType: BootstrapEnvoyConfigType,
			expected:     "out.bootstrap.json",
		},
		{
			output:       "yaml",
			resourceType: BootstrapEnvoyConfigType,
			expected:     "out.bootstrap.yaml",
		}, {
			output:       "json",
			resourceType: ClusterEnvoyConfigType,
			expected:     "out.cluster.json",
		},
		{
			output:       "yaml",
			resourceType: ClusterEnvoyConfigType,
			expected:     "out.cluster.yaml",
		}, {
			output:       "json",
			resourceType: ListenerEnvoyConfigType,
			expected:     "out.listener.json",
		},
		{
			output:       "yaml",
			resourceType: ListenerEnvoyConfigType,
			expected:     "out.listener.yaml",
		}, {
			output:       "json",
			resourceType: RouteEnvoyConfigType,
			expected:     "out.route.json",
		},
		{
			output:       "yaml",
			resourceType: RouteEnvoyConfigType,
			expected:     "out.route.yaml",
		},
		{
			output:       "json",
			resourceType: EndpointEnvoyConfigType,
			expected:     "out.endpoints.json",
		},
		{
			output:       "yaml",
			resourceType: EndpointEnvoyConfigType,
			expected:     "out.endpoints.yaml",
		},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			configDump, err := extractConfigDump(fw, false, tc.resourceType)
			require.NoError(t, err)
			aggregated := sampleAggregatedConfigDump(configDump)
			got, err := marshalEnvoyProxyConfig(aggregated, tc.output)
			require.NoError(t, err)
			if *overrideTestData {
				require.NoError(t, file.Write(string(got), filepath.Join("testdata", "config", "out", tc.expected)))
			}
			out, err := readOutputConfig(tc.expected)
			require.NoError(t, err)
			if tc.output == "yaml" {
				assert.YAMLEq(t, string(out), string(got))
			} else {
				assert.JSONEq(t, string(out), string(got))
			}
		})
	}

	fw.Stop()
}

func TestLabelSelectorBadInput(t *testing.T) {
	podNamespace = "default"

	cases := []struct {
		name   string
		args   []string
		labels []string
	}{
		{
			name:   "no label, no pod name",
			args:   []string{},
			labels: []string{},
		},
		{
			name:   "wrong label, no pod name",
			args:   []string{},
			labels: []string{"foo=bar"},
		},
		{
			name:   "no label, wrong pod name",
			args:   []string{"eg"},
			labels: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			labelSelectors = tc.labels
			_, err := retrieveConfigDump(tc.args, false, AllEnvoyConfigType)
			require.Error(t, err, "error not found")
		})
	}
}

func readInputConfig(filename string) ([]byte, error) {
	b, err := os.ReadFile(path.Join("testdata", "config", "in", filename))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func readOutputConfig(filename string) ([]byte, error) {
	b, err := os.ReadFile(path.Join("testdata", "config", "out", filename))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func sampleAggregatedConfigDump(configDump protoreflect.ProtoMessage) aggregatedConfigDump {
	return aggregatedConfigDump{
		defaultNamespace: {
			defaultEnvoyGatewayPodName: configDump,
		},
	}
}

type fakeCLIClient struct {
	pods []corev1.Pod
	cm   *corev1.ConfigMap
}

func (f *fakeCLIClient) RESTConfig() *rest.Config {
	return nil
}

func (f *fakeCLIClient) Pod(types.NamespacedName) (*corev1.Pod, error) {
	return nil, nil
}

func (f *fakeCLIClient) PodsForSelector(string, ...string) (*corev1.PodList, error) {
	return &corev1.PodList{Items: f.pods}, nil
}

func (f *fakeCLIClient) PodExec(types.NamespacedName, string, string) (stdout string, stderr string, err error) {
	return "", "", nil
}

func (f *fakeCLIClient) Kube() kubernetes.Interface {
	return fake.NewSimpleClientset(f.cm)
}

func TestFetchRunningRateLimitPods(t *testing.T) {

	cases := []struct {
		caseName      string
		rlPods        []corev1.Pod
		namespace     string
		labelSelector []string
		expectErr     error
	}{
		{
			caseName: "normally obtain the rate limit pod of Running phase",
			rlPods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "envoy-ratelimit-666457bc4c-c2td5",
						Namespace: "envoy-gateway-system",
						Labels: map[string]string{
							"app.kubernetes.io/name":       "envoy-ratelimit",
							"app.kubernetes.io/component":  "ratelimit",
							"app.kubernetes.io/managed-by": "envoy-gateway",
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
							{
								Type:   corev1.ContainersReady,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			namespace:     "envoy-gateway-system",
			labelSelector: ratelimit.LabelSelector(),
			expectErr:     nil,
		},
		{
			caseName: "unable to obtain rate limit pod",
			rlPods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "envoy-ratelimit-666457bc4c-c2td5",
						Namespace: "envoy-gateway-system",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
					},
				},
			},
			namespace:     "envoy-gateway-system",
			labelSelector: ratelimit.LabelSelector(),
			expectErr:     fmt.Errorf("please check that the rate limit instance starts properly"),
		},
	}

	for _, tc := range cases {

		t.Run(tc.caseName, func(t *testing.T) {

			fakeCli := &fakeCLIClient{
				pods: tc.rlPods,
			}

			_, err := fetchRunningRateLimitPods(fakeCli, tc.namespace, tc.labelSelector)
			require.Equal(t, tc.expectErr, err)

		})

	}
}

func TestCheckEnableGlobalRateLimit(t *testing.T) {

	cases := []struct {
		caseName    string
		egConfigMap *corev1.ConfigMap
		expect      bool
	}{
		{
			caseName: "global rate limit feature is enabled",
			expect:   true,
			egConfigMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "envoy-gateway-config",
					Namespace: "envoy-gateway-system",
				},
				Data: map[string]string{
					"envoy-gateway.yaml": `
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
provider:
  type: Kubernetes
gateway:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
rateLimit:
  backend:
    type: Redis
    redis:
      url: redis.redis-system.svc.cluster.local:6379
`,
				},
			},
		},
		{
			caseName: "global rate limit feature is not enabled",
			expect:   false,
			egConfigMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "envoy-gateway-config",
					Namespace: "envoy-gateway-system",
				},
				Data: map[string]string{
					"envoy-gateway.yaml": `
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
provider:
  type: Kubernetes
gateway:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
`,
				},
			},
		},
	}

	for _, tc := range cases {

		t.Run(tc.caseName, func(t *testing.T) {

			fakeCli := &fakeCLIClient{
				cm: tc.egConfigMap,
			}

			actual, err := checkEnableGlobalRateLimit(fakeCli)
			require.Equal(t, tc.expect, actual)
			require.NoError(t, err)

		})

	}
}

func TestExtractRateLimitConfig(t *testing.T) {

	cases := []struct {
		caseName     string
		responseBody []byte
		rlPod        types.NamespacedName
	}{
		{
			caseName:     "rate limit configuration is extract normally",
			responseBody: []byte("default/eg/http.httproute/default/backend/rule/0/match/0/*-key-rule-0-match-0_httproute/default/backend/rule/0/match/0/*-value-rule-0-match-0: unit=HOUR requests_per_unit=3, shadow_mode: false"),
			rlPod: types.NamespacedName{
				Name:      "envoy-ratelimit-666457bc4c-c2td5",
				Namespace: "envoy-gateway-system",
			},
		},
	}

	for _, tc := range cases {

		t.Run(tc.caseName, func(t *testing.T) {

			fw, err := newFakePortForwarder(tc.responseBody)
			require.NoError(t, err)

			out, err := extractRateLimitConfig(fw, tc.rlPod)
			require.NoError(t, err)
			require.NotEmpty(t, out)

		})

	}
}

func TestCheckRateLimitPodStatusReady(t *testing.T) {

	cases := []struct {
		caseName string
		status   corev1.PodStatus
		expect   bool
	}{
		{
			caseName: "rate limit pod is ready",
			expect:   true,
			status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
					{
						Type:   corev1.ContainersReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
		{
			caseName: "rate limit pod is not ready",
			expect:   false,
			status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionFalse,
					},
					{
						Type:   corev1.ContainersReady,
						Status: corev1.ConditionFalse,
					},
				},
			},
		},
		{
			caseName: "rate limit pod is running failed",
			expect:   false,
			status: corev1.PodStatus{
				Phase: corev1.PodFailed,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			actual := checkRateLimitPodStatusReady(tc.status)
			require.Equal(t, tc.expect, actual)
		})
	}

}
