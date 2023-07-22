// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package naming

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestServiceName(t *testing.T) {
	cases := []struct {
		nn       types.NamespacedName
		expected string
	}{
		{
			nn: types.NamespacedName{
				Name:      "foo",
				Namespace: "bar",
			},
			expected: "foo.bar",
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			got := ServiceName(c.nn)
			assert.Equal(t, c.expected, got)
		})
	}
}
