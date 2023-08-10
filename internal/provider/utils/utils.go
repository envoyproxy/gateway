// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NamespacedName creates and returns object's NamespacedName.
func NamespacedName(obj client.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
}

// GetHashedName returns a partially hashed name for the string including up to 48 characters of the original name before the hash.
// Input `nsName` should be formatted as `{Namespace}/{ResourceName}`.
func GetHashedName(nsName string) string {

	h := sha256.New() // Using sha256 instead of sha1 due to Blocklisted import crypto/sha1: weak cryptographic primitive (gosec)
	h.Write([]byte(nsName))
	hashedName := strings.ToLower(fmt.Sprintf("%x", h.Sum(nil)))

	// replace `/` with `-` to create a valid K8s resource name
	resourceName := strings.ReplaceAll(nsName, "/", "-")

	if len(resourceName) > 48 {
		return fmt.Sprintf("%s-%s", resourceName[0:48], hashedName[0:8])
	}
	return fmt.Sprintf("%s-%s", resourceName, hashedName[0:8])
}
