// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

// loadFromFilesAndDirs loads resources from specific files and directories.
func loadFromFilesAndDirs(files, dirs []string) ([]*gatewayapi.Resources, error) {
	var rs []*gatewayapi.Resources

	for _, file := range files {
		r, err := loadFromFile(file)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}

	for _, dir := range dirs {
		r, err := loadFromDir(dir)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r...)
	}

	return rs, nil
}

// loadFromFile loads resources from a specific file.
func loadFromFile(path string) (*gatewayapi.Resources, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %s is not exist", path)
		}
		return nil, err
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return convertKubernetesYAMLToResources(string(bytes))
}

// loadFromDir loads resources from all the files under a specific directory excluding subdirectories.
func loadFromDir(path string) ([]*gatewayapi.Resources, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var rs []*gatewayapi.Resources
	for _, entry := range entries {
		// Ignoring subdirectories and all hidden files and directories.
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		r, err := loadFromFile(filepath.Join(path, entry.Name()))
		if err != nil {
			return nil, err
		}

		rs = append(rs, r)
	}

	return rs, nil
}

// TODO(sh2): This function is copied and updated from internal/cmd/egctl/translate.go.
// This function should be able to process arbitrary number of resources, so we
// need to come up with a way to extend the GatewayClass and EnvoyProxy field to array
// instead of single variable in gatewayapi.Resources structure.
//
// - This issue is tracked by https://github.com/envoyproxy/gateway/issues/3207
//
// convertKubernetesYAMLToResources converts a Kubernetes YAML string into GatewayAPI Resources.
func convertKubernetesYAMLToResources(str string) (*gatewayapi.Resources, error) {
	resources := gatewayapi.NewResources()
	var useDefaultNamespace bool
	providedNamespaceMap := sets.New[string]()
	requiredNamespaceMap := sets.New[string]()
	yamls := strings.Split(str, "\n---")
	combinedScheme := envoygateway.GetScheme()
	for _, y := range yamls {
		if strings.TrimSpace(y) == "" {
			continue
		}
		var obj map[string]interface{}
		err := yaml.Unmarshal([]byte(y), &obj)
		if err != nil {
			return nil, err
		}
		un := unstructured.Unstructured{Object: obj}
		gvk := un.GroupVersionKind()
		name, namespace := un.GetName(), un.GetNamespace()
		if namespace == "" {
			// When kubectl applies a resource in yaml which doesn't have a namespace,
			// the current namespace is applied. Here we do the same thing before translating
			// the GatewayAPI resource. Otherwise, the resource can't pass the namespace validation
			useDefaultNamespace = true
			namespace = config.DefaultNamespace
		}
		requiredNamespaceMap.Insert(namespace)
		kobj, err := combinedScheme.New(gvk)
		if err != nil {
			return nil, err
		}
		err = combinedScheme.Convert(&un, kobj, nil)
		if err != nil {
			return nil, err
		}

		objType := reflect.TypeOf(kobj)
		if objType.Kind() != reflect.Ptr {
			return nil, fmt.Errorf("expected pointer type, but got %s", objType.Kind().String())
		}
		kobjVal := reflect.ValueOf(kobj).Elem()
		spec := kobjVal.FieldByName("Spec")

		switch gvk.Kind {
		case gatewayapi.KindEnvoyProxy:
			typedSpec := spec.Interface()
			envoyProxy := &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(egv1a1.EnvoyProxySpec),
			}
			resources.EnvoyProxyForGatewayClass = envoyProxy
		case gatewayapi.KindGatewayClass:
			typedSpec := spec.Interface()
			gatewayClass := &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1.GatewayClassSpec),
			}
			// fill controller name by default controller name when gatewayclass controller name empty.
			if gatewayClass.Spec.ControllerName == "" {
				gatewayClass.Spec.ControllerName = egv1a1.GatewayControllerName
			}
			resources.GatewayClass = gatewayClass
		case gatewayapi.KindGateway:
			typedSpec := spec.Interface()
			gateway := &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1.GatewaySpec),
			}
			resources.Gateways = append(resources.Gateways, gateway)
		case gatewayapi.KindTCPRoute:
			typedSpec := spec.Interface()
			tcpRoute := &gwapiv1a2.TCPRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: gatewayapi.KindTCPRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1a2.TCPRouteSpec),
			}
			resources.TCPRoutes = append(resources.TCPRoutes, tcpRoute)
		case gatewayapi.KindUDPRoute:
			typedSpec := spec.Interface()
			udpRoute := &gwapiv1a2.UDPRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: gatewayapi.KindUDPRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1a2.UDPRouteSpec),
			}
			resources.UDPRoutes = append(resources.UDPRoutes, udpRoute)
		case gatewayapi.KindTLSRoute:
			typedSpec := spec.Interface()
			tlsRoute := &gwapiv1a2.TLSRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: gatewayapi.KindTLSRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1a2.TLSRouteSpec),
			}
			resources.TLSRoutes = append(resources.TLSRoutes, tlsRoute)
		case gatewayapi.KindHTTPRoute:
			typedSpec := spec.Interface()
			httpRoute := &gwapiv1.HTTPRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: gatewayapi.KindHTTPRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1.HTTPRouteSpec),
			}
			resources.HTTPRoutes = append(resources.HTTPRoutes, httpRoute)
		case gatewayapi.KindGRPCRoute:
			typedSpec := spec.Interface()
			grpcRoute := &gwapiv1.GRPCRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: gatewayapi.KindGRPCRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1.GRPCRouteSpec),
			}
			resources.GRPCRoutes = append(resources.GRPCRoutes, grpcRoute)
		case gatewayapi.KindNamespace:
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}
			resources.Namespaces = append(resources.Namespaces, namespace)
			providedNamespaceMap.Insert(name)
		case gatewayapi.KindService:
			typedSpec := spec.Interface()
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(corev1.ServiceSpec),
			}
			resources.Services = append(resources.Services, service)
		case egv1a1.KindEnvoyPatchPolicy:
			typedSpec := spec.Interface()
			envoyPatchPolicy := &egv1a1.EnvoyPatchPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindEnvoyPatchPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      name,
				},
				Spec: typedSpec.(egv1a1.EnvoyPatchPolicySpec),
			}
			resources.EnvoyPatchPolicies = append(resources.EnvoyPatchPolicies, envoyPatchPolicy)
		case egv1a1.KindClientTrafficPolicy:
			typedSpec := spec.Interface()
			clientTrafficPolicy := &egv1a1.ClientTrafficPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindClientTrafficPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      name,
				},
				Spec: typedSpec.(egv1a1.ClientTrafficPolicySpec),
			}
			resources.ClientTrafficPolicies = append(resources.ClientTrafficPolicies, clientTrafficPolicy)
		case egv1a1.KindBackendTrafficPolicy:
			typedSpec := spec.Interface()
			backendTrafficPolicy := &egv1a1.BackendTrafficPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindBackendTrafficPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      name,
				},
				Spec: typedSpec.(egv1a1.BackendTrafficPolicySpec),
			}
			resources.BackendTrafficPolicies = append(resources.BackendTrafficPolicies, backendTrafficPolicy)
		case egv1a1.KindSecurityPolicy:
			typedSpec := spec.Interface()
			securityPolicy := &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind:       egv1a1.KindSecurityPolicy,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      name,
				},
				Spec: typedSpec.(egv1a1.SecurityPolicySpec),
			}
			resources.SecurityPolicies = append(resources.SecurityPolicies, securityPolicy)
		}
	}

	if useDefaultNamespace {
		if !providedNamespaceMap.Has(config.DefaultNamespace) {
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: config.DefaultNamespace,
				},
			}
			resources.Namespaces = append(resources.Namespaces, namespace)
			providedNamespaceMap.Insert(config.DefaultNamespace)
		}
	}

	return resources, nil
}
