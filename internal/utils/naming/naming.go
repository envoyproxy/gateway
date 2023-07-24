// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package naming

import "k8s.io/apimachinery/pkg/types"

func ServiceName(nn types.NamespacedName) string {
	return nn.Name + "." + nn.Namespace
}
