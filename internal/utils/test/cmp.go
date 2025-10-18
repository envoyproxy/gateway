// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CmpResources(t *testing.T, x, y any, cmpOpts ...cmp.Option) {
	opts := []cmp.Option{
		cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion"),
		cmpopts.IgnoreFields(metav1.TypeMeta{}, "Kind", "APIVersion"),
		cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime"),
		cmpopts.EquateEmpty(),
	}
	opts = append(opts, cmpOpts...)

	require.Empty(t, cmp.Diff(x, y, opts...))
}
