package envoygateway

import (
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
)

var (
	// scheme contains all the API types necessary for the provider's dynamic
	// clients to work. Any new non-core types must be added here.
	//
	// NOTE: The discovery mechanism used by the client doesn't automatically
	// refresh, so only add types here that are guaranteed to exist before the
	// provider starts.
	scheme = runtime.NewScheme()
)

func init() {
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := gwapiv1b1.AddToScheme(scheme); err != nil {
		panic(err)
	}
	if err := gwapiv1a2.AddToScheme(scheme); err != nil {
		panic(err)
	}
}

// GetScheme returns a scheme with types supported by the Kubernetes provider.
func GetScheme() *runtime.Scheme {
	return scheme
}
