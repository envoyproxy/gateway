// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package http

import "k8s.io/apimachinery/pkg/util/sets"

var SupportedRedirectCodes = sets.New[int32](301, 302, 303, 307, 308)
