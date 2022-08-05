package kubernetes

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NamespacedNameStr returns obj's <Namespace>/<Name> string representation.
func NamespacedNameStr(obj client.Object) string {
	return obj.GetNamespace() + string(types.Separator) + obj.GetName()
}
