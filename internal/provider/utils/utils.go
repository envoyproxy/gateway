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

// Returns a partially hashed name for the string if it is more than 60 chars. Otherwise returns the original string
func GetHashedName(name string) string {
	if len(name) > 60 {
		h := sha256.New() // Using sha256 instead of sha1 due to Blocklisted import crypto/sha1: weak cryptographic primitive (gosec)
		hsha := h.Sum([]byte(name))
		hashedName := strings.ToLower(fmt.Sprintf("%x", hsha))
		return fmt.Sprintf("%s-%s", name[0:32], hashedName[0:8])
	}
	return name
}
