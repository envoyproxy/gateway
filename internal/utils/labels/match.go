// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package labels

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
)

func SelectorMatch(selector *metav1.LabelSelector, labels map[string]string) (bool, error) {
	s, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return false, fmt.Errorf("invalid label selector is generated: %w", err)
	}

	return s.Matches(klabels.Set(labels)), nil
}

func Matches(left, right map[string]string) (bool, error) {
	s := metav1.SetAsLabelSelector(left)
	return SelectorMatch(s, right)
}
