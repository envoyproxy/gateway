// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/test"
)

func TestMergePolicy(t *testing.T) {
	baseDir := "testdata"
	caseFiles, err := filepath.Glob(filepath.Join(baseDir, "*policy_*.in.yaml"))
	require.NoError(t, err)

	for _, caseFile := range caseFiles {
		caseName := strings.TrimPrefix(strings.TrimSuffix(caseFile, ".in.yaml"), baseDir+"/")
		policyType := strings.SplitN(caseName, "_", 2)[0]

		t.Run(caseName, func(t *testing.T) {
			switch policyType {
			case "backendtrafficpolicy":
				runMergePolicyTest[*egv1a1.BackendTrafficPolicy](t, caseFile)
			case "securitypolicy":
				runMergePolicyTest[*egv1a1.SecurityPolicy](t, caseFile)
			default:
				t.Fatalf("unsupported policy type %q in %s", policyType, caseFile)
			}
		})
	}
}

func runMergePolicyTest[T client.Object](t *testing.T, caseFile string) {
	t.Helper()

	for _, mergeType := range []egv1a1.MergeType{egv1a1.StrategicMerge, egv1a1.JSONMerge} {
		patchedInput := strings.Replace(caseFile, ".in.yaml", ".patch.yaml", 1)
		var output string
		if mergeType == egv1a1.StrategicMerge {
			output = strings.Replace(caseFile, ".in.yaml", ".strategicmerge.out.yaml", 1)
		} else {
			output = strings.Replace(caseFile, ".in.yaml", ".jsonmerge.out.yaml", 1)
		}

		original := readObject[T](t, caseFile)
		patch := readObject[T](t, patchedInput)

		got, err := Merge(original, patch, mergeType)
		require.NoError(t, err)

		if test.OverrideTestData() {
			b, err := yaml.Marshal(got)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(output, b, 0o600))
			continue
		}

		expected := readObject[T](t, output)
		require.Equal(t, expected, got)
	}
}

func readObject[T client.Object](t *testing.T, path string) T {
	t.Helper()
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	btp := new(T)
	err = yaml.Unmarshal(b, btp)
	require.NoError(t, err)
	return *btp
}
