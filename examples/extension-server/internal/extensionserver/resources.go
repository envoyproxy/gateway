package extensionserver

import (
	"encoding/json"
	"fmt"
	extensionpb "github.com/envoyproxy/gateway/proto/extension"
	extensionv1alpha1 "github.com/exampleorg/envoygateway-extension/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func (s *Server) getCustomBackendMtlsPolicy(untyped *extensionpb.ExtensionResource) (*extensionv1alpha1.CustomBackendMtlsPolicy, error) {
	resourceInfo := &unstructured.Unstructured{}
	if err := json.Unmarshal(untyped.GetUnstructuredBytes(), &resourceInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal extension resource: %w", err)
	}

	resourceGVK := resourceInfo.GroupVersionKind().String()
	obj := &extensionv1alpha1.CustomBackendMtlsPolicy{}
	gvk, err := apiutil.GVKForObject(obj, scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to get GVK for resource: %w", err)
	}

	if gvk.String() != resourceGVK {
		return nil, nil
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(resourceInfo.Object, &obj); err != nil {
		return nil, fmt.Errorf("error converting resourceInfo from unstructured: %w", err)
	}

	return obj, nil
}
