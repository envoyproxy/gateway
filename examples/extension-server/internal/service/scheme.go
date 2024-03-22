package service

import (
	v1 "github.com/exampleorg/envoygateway-extension/api/v1"

	"k8s.io/apimachinery/pkg/runtime"
	runtimeUtil "k8s.io/apimachinery/pkg/util/runtime"
	k8sScheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	// scheme contains all the API types needed by the provider's dynamic clients.
	// Any new non-core types must be added here.
	scheme = runtime.NewScheme()
)

func init() {
	// add standard Kubernetes types to scheme
	runtimeUtil.Must(k8sScheme.AddToScheme(scheme))
	// add our custom types (GlobalLuaScript) to the scheme
	runtimeUtil.Must(v1.AddToScheme(scheme))
}

// GetScheme returns a scheme with types we need to support
func GetScheme() *runtime.Scheme {
	return scheme
}
