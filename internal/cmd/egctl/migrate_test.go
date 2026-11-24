// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/cmd/egctl/migrate"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/utils/test"
)

const migrateTestData = "testdata/migrate/deprecated"

// runMigrate runs `egctl x migrate deprecated` with args and returns stdout, stderr and the error.
func runMigrate(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd := newMigrateCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.SetArgs(append([]string{"deprecated"}, args...))
	err := cmd.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := errors.AsType[*migrate.ExitError](err); ok {
		return exitErr.Code
	}
	return -1
}

func TestMigrateDeprecatedGolden(t *testing.T) {
	inputs, err := filepath.Glob(filepath.Join(migrateTestData, "*.in.yaml"))
	require.NoError(t, err)
	require.NotEmpty(t, inputs)

	// Run from the testdata directory, so the report names the input file without a directory.
	t.Chdir(migrateTestData)
	for _, in := range inputs {
		in = filepath.Base(in)
		name := strings.TrimSuffix(in, ".in.yaml")
		t.Run(name, func(t *testing.T) {
			stdout, report, err := runMigrate(t, "-f", in)
			code := exitCode(err)
			require.Contains(t, []int{0, migrate.ExitCodeError}, code, "unexpected error: %v", err)

			outFile := name + ".out.yaml"
			reportFile := name + ".report"
			if test.OverrideTestData() {
				require.NoError(t, os.WriteFile(outFile, []byte(stdout), 0o600))
				require.NoError(t, os.WriteFile(reportFile, []byte(report), 0o600))
				return
			}
			wantOut, err := os.ReadFile(outFile)
			require.NoError(t, err)
			wantReport, err := os.ReadFile(reportFile)
			require.NoError(t, err)
			require.Equal(t, string(wantOut), stdout)
			require.Equal(t, string(wantReport), report)

			// Migrating the output again finds nothing new to migrate.
			second := filepath.Join(t.TempDir(), "out.yaml")
			require.NoError(t, os.WriteFile(second, []byte(stdout), 0o600))
			stdout2, _, err := runMigrate(t, "-f", second)
			require.Equal(t, code, exitCode(err))
			require.Equal(t, stdout, stdout2)
		})
	}
}

func TestMigrateDeprecatedUntouchedIsByteForByte(t *testing.T) {
	in := filepath.Join(migrateTestData, "untouched.in.yaml")
	want, err := os.ReadFile(in)
	require.NoError(t, err)
	stdout, stderr, err := runMigrate(t, "-f", in)
	require.NoError(t, err)
	require.Equal(t, string(want), stdout)
	require.Equal(t, "Nothing to migrate.\n", stderr)
}

// TestMigrateDeprecatedTranslateEquivalence checks that the migrated resources translate to the
// same xDS as the original ones. The telemetry host rule isn't covered: a Service backendRef is
// resolved through EDS rather than DNS, which is intended.
func TestMigrateDeprecatedTranslateEquivalence(t *testing.T) {
	in := filepath.Join(migrateTestData, "translate-equivalence.in.yaml")
	migrated, _, err := runMigrate(t, "-f", in)
	require.NoError(t, err)
	out := filepath.Join(t.TempDir(), "migrated.yaml")
	require.NoError(t, os.WriteFile(out, []byte(migrated), 0o600))

	translateXds := func(file string) string {
		var b bytes.Buffer
		require.NoError(t, translate(&b, file, gatewayAPIType, []string{xdsType}, yamlOutput,
			string(AllEnvoyConfigType), false, "envoy-gateway-system", "cluster.local"))
		return b.String()
	}
	before := translateXds(in)
	require.Contains(t, before, "proxy_protocol")
	require.Contains(t, before, "brotli")
	require.Equal(t, before, translateXds(out))
}

func TestMigrateDeprecatedModes(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(migrateTestData, "backendtrafficpolicy.in.yaml"))
	require.NoError(t, err)
	untouched, err := os.ReadFile(filepath.Join(migrateTestData, "untouched.in.yaml"))
	require.NoError(t, err)
	wantOut, err := os.ReadFile(filepath.Join(migrateTestData, "backendtrafficpolicy.out.yaml"))
	require.NoError(t, err)

	setup := func(t *testing.T) (string, string, string) {
		dir := t.TempDir()
		changed := filepath.Join(dir, "nested", "btp.yaml")
		same := filepath.Join(dir, "same.yml")
		require.NoError(t, os.MkdirAll(filepath.Dir(changed), 0o750))
		require.NoError(t, os.WriteFile(changed, src, 0o600))    //nolint:gosec // test temp dir
		require.NoError(t, os.WriteFile(same, untouched, 0o600)) //nolint:gosec // test temp dir
		return dir, changed, same
	}

	t.Run("check-only", func(t *testing.T) {
		dir, changed, _ := setup(t)
		stdout, stderr, err := runMigrate(t, "-f", dir, "--mode", "check-only")
		require.Equal(t, migrate.ExitCodeWouldChange, exitCode(err))
		require.Empty(t, stdout)
		require.Contains(t, stderr, "spec.compression")
		got, err := os.ReadFile(changed)
		require.NoError(t, err)
		require.Equal(t, string(src), string(got))
	})

	t.Run("in-place", func(t *testing.T) {
		dir, changed, same := setup(t)
		before, err := os.Stat(same)
		require.NoError(t, err)
		stdout, _, err := runMigrate(t, "-f", dir, "--mode", "in-place")
		require.NoError(t, err)
		require.Empty(t, stdout)
		got, err := os.ReadFile(changed)
		require.NoError(t, err)
		require.Equal(t, string(wantOut), string(got))
		after, err := os.Stat(same)
		require.NoError(t, err)
		require.Equal(t, before.ModTime(), after.ModTime(), "a file with nothing to migrate must not be written")
	})
}

func TestMigrateOptionsValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{name: "no input", args: nil, want: "exactly one of"},
		{name: "both inputs", args: []string{"-f", "x.yaml", "--from-cluster"}, want: "exactly one of"},
		{name: "bad mode", args: []string{"-f", "x.yaml", "--mode", "apply"}, want: "invalid --mode"},
		{name: "in-place stdin", args: []string{"-f", "-", "--mode", "in-place"}, want: "no file to write back to"},
		{name: "in-place cluster", args: []string{"--from-cluster", "--mode", "in-place"}, want: "not supported with --from-cluster yet"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := runMigrate(t, tc.args...)
			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestMigrateDeprecatedFromCluster(t *testing.T) {
	gatewayTarget := &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
		LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
			Group: gwapiv1.GroupName, Kind: "Gateway", Name: "eg",
		},
	}
	plain := &egv1a1.ClientTrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "plain", Namespace: "default", ResourceVersion: "42", Generation: 3,
			Annotations: map[string]string{
				"kubectl.kubernetes.io/last-applied-configuration": `{"spec":{"targetRef":{}}}`,
			},
		},
		Spec: egv1a1.ClientTrafficPolicySpec{
			PolicyTargetReferences: egv1a1.PolicyTargetReferences{TargetRef: gatewayTarget},
		},
		Status: gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{{ControllerName: "x"}}},
	}
	helm := &egv1a1.ClientTrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "helm", Namespace: "default",
			Labels: map[string]string{"app.kubernetes.io/managed-by": "Helm"},
		},
		Spec: egv1a1.ClientTrafficPolicySpec{
			PolicyTargetReferences: egv1a1.PolicyTargetReferences{TargetRef: gatewayTarget},
		},
	}
	argo := &egv1a1.BackendTrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "argo", Namespace: "apps",
			Annotations: map[string]string{"argocd.argoproj.io/tracking-id": "app:gateway.envoyproxy.io/BackendTrafficPolicy:apps/argo"},
		},
		Spec: egv1a1.BackendTrafficPolicySpec{
			PolicyTargetReferences: egv1a1.PolicyTargetReferences{TargetRef: gatewayTarget},
		},
	}
	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "envoy-gateway-config", Namespace: "envoy-gateway-system"},
		Data: map[string]string{"envoy-gateway.yaml": "apiVersion: gateway.envoyproxy.io/v1alpha1\n" +
			"kind: EnvoyGateway\nextensionApis:\n  disableLua: true\n"},
	}

	cli := fake.NewClientBuilder().WithScheme(envoygateway.GetScheme()).
		WithObjects(plain, helm, argo, config).Build()
	orig := newK8sClientFunc
	newK8sClientFunc = func() (client.Client, error) { return cli, nil }
	t.Cleanup(func() { newK8sClientFunc = orig })

	t.Run("all namespaces", func(t *testing.T) {
		stdout, stderr, err := runMigrate(t, "--from-cluster", "-A")
		require.Equal(t, migrate.ExitCodeError, exitCode(err))

		// Server-set fields are stripped.
		for _, s := range []string{"resourceVersion", "generation", "status:", "last-applied-configuration", "managedFields"} {
			require.NotContains(t, stdout, s)
		}
		// The unmanaged policy is migrated, the managed ones are printed unchanged.
		docs := strings.Split(stdout, "\n---\n")
		require.Len(t, docs, 4)
		require.Contains(t, docs[0], "name: helm")
		require.Contains(t, docs[0], "targetRef:\n")
		require.Contains(t, docs[1], "name: plain")
		require.Contains(t, docs[1], "targetRefs:\n")
		require.NotContains(t, docs[1], "targetRef:\n")
		require.Contains(t, docs[2], "name: argo")
		require.Contains(t, docs[2], "targetRef:\n")
		require.Contains(t, docs[3], "disableLua: true")

		require.Regexp(t, `default/plain\s+spec.targetRef\s+migrated`, stderr)
		require.Regexp(t, `default/helm\s+spec.targetRef\s+error\s+managed by Helm`, stderr)
		require.Regexp(t, `apps/argo\s+spec.targetRef\s+error\s+managed by Argo CD`, stderr)
		require.Regexp(t, `envoy-gateway-system/envoy-gateway-config\s+data.envoy-gateway.yaml:extensionApis.disableLua\s+error`, stderr)
	})

	t.Run("one namespace", func(t *testing.T) {
		stdout, _, err := runMigrate(t, "--from-cluster", "-n", "default", "--mode", "check-only")
		require.Equal(t, migrate.ExitCodeError, exitCode(err))
		require.Empty(t, stdout)

		stdout, stderr, err := runMigrate(t, "--from-cluster", "-n", "apps")
		require.Equal(t, migrate.ExitCodeError, exitCode(err))
		require.NotContains(t, stdout, "envoy-gateway-config")
		require.NotContains(t, stderr, "default/")
	})
}

func TestMigrateDeprecatedEmptyCompressor(t *testing.T) {
	for _, replacement := range []string{"[]", "null"} {
		t.Run(replacement, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "policy.yaml")
			data := "apiVersion: gateway.envoyproxy.io/v1alpha1\nkind: BackendTrafficPolicy\nmetadata:\n  name: example\nspec:\n  compressor: " + replacement + "\n  compression:\n  - type: Gzip\n"
			require.NoError(t, os.WriteFile(file, []byte(data), 0o600))
			stdout, _, err := runMigrate(t, "-f", file)
			require.NoError(t, err)
			var policy map[string]any
			require.NoError(t, yaml.Unmarshal([]byte(stdout), &policy))
			spec := policy["spec"].(map[string]any)
			require.NotContains(t, spec, "compression")
			require.Equal(t, []any{map[string]any{"type": "Gzip", "gzip": map[string]any{}}}, spec["compressor"])
		})
	}
}

func TestMigrateDeprecatedFileBoundary(t *testing.T) {
	for _, prefix := range []string{"", "---\n"} {
		t.Run(prefix, func(t *testing.T) {
			dir := t.TempDir()
			first, second := filepath.Join(dir, "first.yaml"), filepath.Join(dir, "second.yaml")
			original := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: first"
			require.NoError(t, os.WriteFile(first, []byte(original), 0o600))
			require.NoError(t, os.WriteFile(second, []byte(prefix+strings.ReplaceAll(original, "first", "second")), 0o600))
			stdout, _, err := runMigrate(t, "-f", first, "-f", second)
			require.NoError(t, err)
			decoder := yaml.NewDecoder(strings.NewReader(stdout))
			var names []string
			for {
				var resource struct{ Metadata struct{ Name string } }
				err := decoder.Decode(&resource)
				if errors.Is(err, io.EOF) {
					break
				}
				require.NoError(t, err)
				if resource.Metadata.Name != "" {
					names = append(names, resource.Metadata.Name)
				}
			}
			require.Equal(t, []string{"first", "second"}, names)
			_, _, err = runMigrate(t, "-f", first, "--mode", "in-place")
			require.NoError(t, err)
			data, err := os.ReadFile(first)
			require.NoError(t, err)
			require.Equal(t, original, string(data))
		})
	}
}

// These inputs previously passed check-only without migrating deprecated fields.
func TestMigrateDeprecatedReviewRegressions(t *testing.T) {
	tests := []struct {
		name, input, field string
		want               any
	}{
		{
			name: "nested typed list",
			input: `apiVersion: v1
kind: List
items:
- apiVersion: gateway.envoyproxy.io/v1alpha1
  kind: ClientTrafficPolicyList
  items:
  - metadata:
      name: policy
    spec:
      targetRef: {name: eg}
`,
			field: "items[0].items[0].spec.targetRefs[0].name",
			want:  "eg",
		},
		{
			name: "aliased identity",
			input: `metadata:
  annotations:
    version: &version gateway.envoyproxy.io/v1alpha1
    kind: &kind ClientTrafficPolicy
    name: &name policy
    namespace: &namespace default
  name: *name
  namespace: *namespace
apiVersion: *version
kind: *kind
spec:
  targetRef: {name: eg}
`,
			field: "spec.targetRefs[0].name",
			want:  "eg",
		},
		{
			name: "null policy fields",
			input: `apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
spec:
  targetRef: null
  targetRefs: [{name: eg}]
  enableProxyProtocol: null
  headers:
    preserveXRequestID: null
    requestID: PreserveOrGenerate
  tls:
    clientValidation:
      optional: null
      mode: RequireAndVerify
`,
			field: "spec",
			want: map[string]any{
				"targetRefs": []any{map[string]any{"name": "eg"}},
				"headers":    map[string]any{"requestID": "PreserveOrGenerate"},
				"tls":        map[string]any{"clientValidation": map[string]any{"mode": "RequireAndVerify"}},
			},
		},
		{
			name: "null proxy fields",
			input: `apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
spec:
  luaValidation: null
  telemetry:
    tracing:
      provider:
        host: null
        port: null
        backendRef: null
`,
			field: "spec",
			want: map[string]any{"telemetry": map[string]any{
				"tracing": map[string]any{"provider": map[string]any{}},
			}},
		},
		{
			name: "null compression and header",
			input: `apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
spec:
  compression: null
  loadBalancer:
    consistentHash:
      type: SourceIP
      header: null
`,
			field: "spec",
			want: map[string]any{"loadBalancer": map[string]any{
				"consistentHash": map[string]any{"type": "SourceIP"},
			}},
		},
		{
			name: "custom cluster domain",
			input: `apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
spec:
  telemetry:
    tracing:
      provider:
        host: otel.monitoring.svc.corp.internal.
        port: 4318
`,
			field: "spec.telemetry.tracing.provider.backendRefs",
			want:  []any{map[string]any{"name": "otel", "namespace": "monitoring", "port": 4318}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "input.yaml")
			require.NoError(t, os.WriteFile(file, []byte(tt.input), 0o600))
			out, _, err := runMigrate(t, "-f", file, "--mode", "check-only")
			require.Equal(t, migrate.ExitCodeWouldChange, exitCode(err), "%v", err)
			require.Empty(t, out)
			original, err := os.ReadFile(file)
			require.NoError(t, err)
			require.Equal(t, tt.input, string(original))

			out, report, err := runMigrate(t, "-f", file)
			require.NoError(t, err)
			if tt.name == "aliased identity" {
				require.Contains(t, report, "policy")
				require.NotContains(t, report, "*kind")
			}
			var doc yaml.Node
			require.NoError(t, yaml.Unmarshal([]byte(out), &doc))
			got := migrate.Lookup(doc.Content[0], tt.field)
			require.NotNil(t, got)
			var value any
			require.NoError(t, got.Decode(&value))
			require.Equal(t, tt.want, value)

			_, _, err = runMigrate(t, "-f", file, "--mode", "in-place")
			require.NoError(t, err)
			data, err := os.ReadFile(file)
			require.NoError(t, err)
			require.Equal(t, out, string(data))
			_, _, err = runMigrate(t, "-f", file, "--mode", "check-only")
			require.NoError(t, err)
		})
	}
}
