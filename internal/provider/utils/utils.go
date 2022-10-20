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

// Returns a partially hashed name for the string including up to 48 characters of the original name before the hash
func GetHashedName(name string) string {

	h := sha256.New() // Using sha256 instead of sha1 due to Blocklisted import crypto/sha1: weak cryptographic primitive (gosec)
	hsha := h.Sum([]byte(name))
	hashedName := strings.ToLower(fmt.Sprintf("%x", hsha))

	if len(name) > 48 {
		return fmt.Sprintf("%s-%s", name[0:48], hashedName[0:8])
	}
	return fmt.Sprintf("%s-%s", name, hashedName[0:8])
}
