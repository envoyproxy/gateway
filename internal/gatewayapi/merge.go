// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// MergeRouteParentStatus merges RouteParentStatus entries by ParentReference identity.
// It preserves existing entries that are not updated in the new status and appends new entries.
func MergeRouteParentStatus(ns string, old, new []gwapiv1.RouteParentStatus) []gwapiv1.RouteParentStatus {
	merged := make([]gwapiv1.RouteParentStatus, 0, len(old)+len(new))

	for _, oldP := range old {
		found := -1
		for newI, newP := range new {
			if IsParentRefEqual(oldP.ParentRef, newP.ParentRef, ns) {
				found = newI
				break
			}
		}
		if found >= 0 {
			merged = append(merged, new[found])
		} else {
			merged = append(merged, oldP)
		}
	}

	for _, newP := range new {
		found := false
		for _, mergedP := range merged {
			if IsParentRefEqual(newP.ParentRef, mergedP.ParentRef, ns) {
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, newP)
		}
	}

	return merged
}

// MergePolicyAncestorStatus merges PolicyAncestorStatus entries by AncestorRef identity.
// It preserves existing entries that are not updated in the new status and appends new entries.
func MergePolicyAncestorStatus(old, new []gwapiv1.PolicyAncestorStatus) []gwapiv1.PolicyAncestorStatus {
	merged := make([]gwapiv1.PolicyAncestorStatus, 0, len(old)+len(new))

	for _, oldA := range old {
		found := -1
		for newI, newA := range new {
			if isAncestorRefEqual(&oldA.AncestorRef, &newA.AncestorRef) {
				found = newI
				break
			}
		}
		if found >= 0 {
			merged = append(merged, new[found])
		} else {
			merged = append(merged, oldA)
		}
	}

	for _, newA := range new {
		found := false
		for _, mergedA := range merged {
			if isAncestorRefEqual(&newA.AncestorRef, &mergedA.AncestorRef) {
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, newA)
		}
	}

	return merged
}

func isAncestorRefEqual(a, b *gwapiv1.ParentReference) bool {
	if a == nil || b == nil {
		return a == b
	}

	if a.Name != b.Name {
		return false
	}

	if (a.Group == nil) != (b.Group == nil) {
		return false
	}
	if a.Group != nil && *a.Group != *b.Group {
		return false
	}

	if (a.Kind == nil) != (b.Kind == nil) {
		return false
	}
	if a.Kind != nil && *a.Kind != *b.Kind {
		return false
	}

	if (a.Namespace == nil) != (b.Namespace == nil) {
		return false
	}
	if a.Namespace != nil && *a.Namespace != *b.Namespace {
		return false
	}

	if (a.SectionName == nil) != (b.SectionName == nil) {
		return false
	}
	if a.SectionName != nil && *a.SectionName != *b.SectionName {
		return false
	}

	return true
}
