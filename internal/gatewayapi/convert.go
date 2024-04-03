// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"bufio"
	"fmt"
	"os"
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
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

func ReadKubernetesYAMLBytes(filePath string) ([]byte, error) {
	// Get input from stdin
	if filePath == "-" {
		scanner := bufio.NewScanner(os.Stdin)
		var input string
		for {
			if !scanner.Scan() {
				break
			}
			input += scanner.Text() + "\n"
		}
		return []byte(input), nil
	}
	// Get input from file
	return os.ReadFile(filePath)
}

// ConvertKubernetesYAMLToResources converts a Kubernetes YAML string into GatewayAPI Resources
func ConvertKubernetesYAMLToResources(str string, addMissingResources bool) (*Resources, error) {
	resources := NewResources()
	var useDefaultNamespace bool
	providedNamespaceMap := map[string]struct{}{}
	requiredNamespaceMap := map[string]struct{}{}
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
		requiredNamespaceMap[namespace] = struct{}{}
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
		case KindEnvoyProxy:
			typedSpec := spec.Interface()
			envoyProxy := &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(egv1a1.EnvoyProxySpec),
			}
			resources.EnvoyProxy = envoyProxy
		case KindGatewayClass:
			typedSpec := spec.Interface()
			gatewayClass := &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1.GatewayClassSpec),
			}
			resources.GatewayClass = gatewayClass
		case KindGateway:
			typedSpec := spec.Interface()
			gateway := &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1.GatewaySpec),
			}
			resources.Gateways = append(resources.Gateways, gateway)
		case KindTCPRoute:
			typedSpec := spec.Interface()
			tcpRoute := &gwapiv1a2.TCPRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: KindTCPRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1a2.TCPRouteSpec),
			}
			resources.TCPRoutes = append(resources.TCPRoutes, tcpRoute)
		case KindUDPRoute:
			typedSpec := spec.Interface()
			udpRoute := &gwapiv1a2.UDPRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: KindUDPRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1a2.UDPRouteSpec),
			}
			resources.UDPRoutes = append(resources.UDPRoutes, udpRoute)
		case KindTLSRoute:
			typedSpec := spec.Interface()
			tlsRoute := &gwapiv1a2.TLSRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: KindTLSRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1a2.TLSRouteSpec),
			}
			resources.TLSRoutes = append(resources.TLSRoutes, tlsRoute)
		case KindHTTPRoute:
			typedSpec := spec.Interface()
			httpRoute := &gwapiv1.HTTPRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: KindHTTPRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1.HTTPRouteSpec),
			}
			resources.HTTPRoutes = append(resources.HTTPRoutes, httpRoute)
		case KindGRPCRoute:
			typedSpec := spec.Interface()
			grpcRoute := &gwapiv1a2.GRPCRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: KindGRPCRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1a2.GRPCRouteSpec),
			}
			resources.GRPCRoutes = append(resources.GRPCRoutes, grpcRoute)
		case KindNamespace:
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}
			resources.Namespaces = append(resources.Namespaces, ns)
			providedNamespaceMap[name] = struct{}{}
		case KindService:
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
		if _, found := providedNamespaceMap[config.DefaultNamespace]; !found {
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: config.DefaultNamespace,
				},
			}
			resources.Namespaces = append(resources.Namespaces, namespace)
			providedNamespaceMap[config.DefaultNamespace] = struct{}{}
		}
	}

	if addMissingResources {
		for ns := range requiredNamespaceMap {
			if _, found := providedNamespaceMap[ns]; !found {
				namespace := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: ns,
					},
				}
				resources.Namespaces = append(resources.Namespaces, namespace)
			}
		}

		requiredServiceMap := map[string]*corev1.Service{}
		for _, route := range resources.TCPRoutes {
			addMissingServices(requiredServiceMap, route)
		}
		for _, route := range resources.UDPRoutes {
			addMissingServices(requiredServiceMap, route)
		}
		for _, route := range resources.TLSRoutes {
			addMissingServices(requiredServiceMap, route)
		}
		for _, route := range resources.HTTPRoutes {
			addMissingServices(requiredServiceMap, route)
		}
		for _, route := range resources.GRPCRoutes {
			addMissingServices(requiredServiceMap, route)
		}

		providedServiceMap := map[string]*corev1.Service{}
		for _, service := range resources.Services {
			providedServiceMap[service.Namespace+"/"+service.Name] = service
		}

		for key, service := range requiredServiceMap {
			if provided, found := providedServiceMap[key]; !found {
				resources.Services = append(resources.Services, service)
			} else {
				providedPorts := sets.NewString()
				for _, port := range provided.Spec.Ports {
					portKey := fmt.Sprintf("%s-%d", port.Protocol, port.Port)
					providedPorts.Insert(portKey)
				}

				for _, port := range service.Spec.Ports {
					name := fmt.Sprintf("%s-%d", port.Protocol, port.Port)
					if !providedPorts.Has(name) {
						servicePort := corev1.ServicePort{
							Name:     name,
							Protocol: port.Protocol,
							Port:     port.Port,
						}
						provided.Spec.Ports = append(provided.Spec.Ports, servicePort)
					}
				}
			}
		}

		// Add EnvoyProxy if it does not exist
		if resources.EnvoyProxy == nil {
			if err := addDefaultEnvoyProxy(resources); err != nil {
				return nil, err
			}
		}
	}

	return resources, nil
}

func addMissingServices(requiredServices map[string]*corev1.Service, obj interface{}) {
	var objNamespace string
	protocol := corev1.Protocol(TCPProtocol)

	var refs []gwapiv1.BackendRef
	switch route := obj.(type) {
	case *gwapiv1.HTTPRoute:
		objNamespace = route.Namespace
		for _, rule := range route.Spec.Rules {
			for _, httpBakcendRef := range rule.BackendRefs {
				refs = append(refs, httpBakcendRef.BackendRef)
			}
		}
	case *gwapiv1a2.GRPCRoute:
		objNamespace = route.Namespace
		for _, rule := range route.Spec.Rules {
			for _, gRPCBakcendRef := range rule.BackendRefs {
				refs = append(refs, gRPCBakcendRef.BackendRef)
			}
		}
	case *gwapiv1a2.TLSRoute:
		objNamespace = route.Namespace
		for _, rule := range route.Spec.Rules {
			refs = append(refs, rule.BackendRefs...)
		}
	case *gwapiv1a2.TCPRoute:
		objNamespace = route.Namespace
		for _, rule := range route.Spec.Rules {
			refs = append(refs, rule.BackendRefs...)
		}
	case *gwapiv1a2.UDPRoute:
		protocol = UDPProtocol
		objNamespace = route.Namespace
		for _, rule := range route.Spec.Rules {
			refs = append(refs, rule.BackendRefs...)
		}
	}

	for _, ref := range refs {
		if ref.Kind == nil || *ref.Kind != KindService {
			continue
		}

		ns := objNamespace
		if ref.Namespace != nil {
			ns = string(*ref.Namespace)
		}
		name := string(ref.Name)
		key := ns + "/" + name

		port := int32(*ref.Port)
		servicePort := corev1.ServicePort{
			Name:     fmt.Sprintf("%s-%d", protocol, port),
			Protocol: protocol,
			Port:     port,
		}
		if service, found := requiredServices[key]; !found {
			svc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: ns,
				},
				Spec: corev1.ServiceSpec{
					// Just a dummy IP
					ClusterIP: "127.0.0.1",
					Ports:     []corev1.ServicePort{servicePort},
				},
			}
			requiredServices[key] = svc
		} else {
			inserted := false
			for _, p := range service.Spec.Ports {
				if p.Protocol == servicePort.Protocol && p.Port == servicePort.Port {
					inserted = true
					break
				}
			}

			if !inserted {
				service.Spec.Ports = append(service.Spec.Ports, servicePort)
			}
		}
	}
}

func addDefaultEnvoyProxy(resources *Resources) error {
	if resources.GatewayClass == nil {
		return fmt.Errorf("the GatewayClass resource is required")
	}

	defaultEnvoyProxyName := "default-envoy-proxy"
	namespace := resources.GatewayClass.Namespace
	defaultBootstrapStr, err := bootstrap.GetRenderedBootstrapConfig(nil)
	if err != nil {
		return err
	}
	ep := &egv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      defaultEnvoyProxyName,
		},
		Spec: egv1a1.EnvoyProxySpec{
			Bootstrap: &egv1a1.ProxyBootstrap{
				Value: defaultBootstrapStr,
			},
		},
	}
	resources.EnvoyProxy = ep
	ns := gwapiv1.Namespace(namespace)
	resources.GatewayClass.Spec.ParametersRef = &gwapiv1.ParametersReference{
		Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
		Kind:      KindEnvoyProxy,
		Name:      defaultEnvoyProxyName,
		Namespace: &ns,
	}
	return nil
}
