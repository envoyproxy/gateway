// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

const dummyClusterIP = "1.2.3.4"

// LoadResourcesFromYAMLBytes will load Resources from given Kubernetes YAML string.
// TODO: This function should be able to process arbitrary number of resources, tracked by https://github.com/envoyproxy/gateway/issues/3207.
func LoadResourcesFromYAMLBytes(yamlBytes []byte, addMissingResources bool) (*Resources, error) {
	r, err := loadKubernetesYAMLToResources(yamlBytes, addMissingResources)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// loadKubernetesYAMLToResources converts a Kubernetes YAML string into GatewayAPI Resources.
// TODO: add support for kind:
//   - BackendLPPolicy (gateway.networking.k8s.io/v1alpha2)
//   - ReferenceGrant (gateway.networking.k8s.io/v1alpha2)
func loadKubernetesYAMLToResources(input []byte, addMissingResources bool) (*Resources, error) {
	resources := NewResources()
	var useDefaultNamespace bool
	providedNamespaceMap := sets.New[string]()
	requiredNamespaceMap := sets.New[string]()
	combinedScheme := envoygateway.GetScheme()

	if err := IterYAMLBytes(input, func(yamlByte []byte) error {
		var obj map[string]interface{}
		err := yaml.Unmarshal(yamlByte, &obj)
		if err != nil {
			return err
		}

		un := &unstructured.Unstructured{Object: obj}
		gvk := un.GroupVersionKind()
		name, namespace := un.GetName(), un.GetNamespace()
		if len(namespace) == 0 {
			useDefaultNamespace = true
			namespace = config.DefaultNamespace
		}

		// Perform local validation and apply default values for gateway-api related resources only.
		if gvk.Group == egv1a1.GroupName || gvk.Group == gwapiv1.GroupName {
			if err = defaultValidator.Validate(yamlByte); err != nil {
				return fmt.Errorf("local validation error: %w", err)
			}

			un, err = gatewaySchemaDefaulter.ApplyDefault(un)
			if err != nil {
				return fmt.Errorf("failed to apply default values for %s/%s: %w", un.GetKind(), un.GetName(), err)
			}
		}

		requiredNamespaceMap.Insert(namespace)
		kobj, err := combinedScheme.New(gvk)
		if err != nil {
			return err
		}
		err = combinedScheme.Convert(un, kobj, nil)
		if err != nil {
			return err
		}

		objType := reflect.TypeOf(kobj)
		if objType.Kind() != reflect.Ptr {
			return fmt.Errorf("expected pointer type, but got %s", objType.Kind().String())
		}
		kobjVal := reflect.ValueOf(kobj).Elem()
		spec := kobjVal.FieldByName("Spec")
		data := kobjVal.FieldByName("Data")
		stringData := kobjVal.FieldByName("StringData")

		switch gvk.Kind {
		case KindEnvoyProxy:
			typedSpec := spec.Interface()
			envoyProxy := &egv1a1.EnvoyProxy{
				TypeMeta: metav1.TypeMeta{
					Kind: KindEnvoyProxy,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(egv1a1.EnvoyProxySpec),
			}
			// TODO: only support loading one envoyproxy for now.
			resources.EnvoyProxyForGatewayClass = envoyProxy
		case KindGatewayClass:
			typedSpec := spec.Interface()
			gatewayClass := &gwapiv1.GatewayClass{
				TypeMeta: metav1.TypeMeta{
					Kind: KindGatewayClass,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1.GatewayClassSpec),
			}
			// fill controller name by default controller name when gatewayclass controller name empty.
			if addMissingResources && len(gatewayClass.Spec.ControllerName) == 0 {
				gatewayClass.Spec.ControllerName = egv1a1.GatewayControllerName
			}
			resources.GatewayClass = gatewayClass
		case KindGateway:
			typedSpec := spec.Interface()
			gateway := &gwapiv1.Gateway{
				TypeMeta: metav1.TypeMeta{
					Kind: KindGateway,
				},
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
			grpcRoute := &gwapiv1.GRPCRoute{
				TypeMeta: metav1.TypeMeta{
					Kind: KindGRPCRoute,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1.GRPCRouteSpec),
			}
			resources.GRPCRoutes = append(resources.GRPCRoutes, grpcRoute)
		case KindNamespace:
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}
			resources.Namespaces = append(resources.Namespaces, namespace)
			providedNamespaceMap.Insert(name)
		case KindService:
			typedSpec := spec.Interface()
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(corev1.ServiceSpec),
			}
			if addMissingResources && len(service.Spec.ClusterIP) == 0 {
				// fill with dummy IP when service clusterIP is empty
				service.Spec.ClusterIP = dummyClusterIP
			}
			resources.Services = append(resources.Services, service)
		case KindEnvoyPatchPolicy:
			typedSpec := spec.Interface()
			envoyPatchPolicy := &egv1a1.EnvoyPatchPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind: egv1a1.KindEnvoyPatchPolicy,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(egv1a1.EnvoyPatchPolicySpec),
			}
			resources.EnvoyPatchPolicies = append(resources.EnvoyPatchPolicies, envoyPatchPolicy)
		case KindClientTrafficPolicy:
			typedSpec := spec.Interface()
			clientTrafficPolicy := &egv1a1.ClientTrafficPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind: KindClientTrafficPolicy,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(egv1a1.ClientTrafficPolicySpec),
			}
			resources.ClientTrafficPolicies = append(resources.ClientTrafficPolicies, clientTrafficPolicy)
		case KindBackendTrafficPolicy:
			typedSpec := spec.Interface()
			backendTrafficPolicy := &egv1a1.BackendTrafficPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind: KindBackendTrafficPolicy,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(egv1a1.BackendTrafficPolicySpec),
			}
			resources.BackendTrafficPolicies = append(resources.BackendTrafficPolicies, backendTrafficPolicy)
		case KindSecurityPolicy:
			typedSpec := spec.Interface()
			securityPolicy := &egv1a1.SecurityPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind: KindSecurityPolicy,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(egv1a1.SecurityPolicySpec),
			}
			resources.SecurityPolicies = append(resources.SecurityPolicies, securityPolicy)
		case KindHTTPRouteFilter:
			typedSpec := spec.Interface()
			httpRouteFilter := &egv1a1.HTTPRouteFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       KindHTTPRouteFilter,
					APIVersion: egv1a1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      name,
				},
				Spec: typedSpec.(egv1a1.HTTPRouteFilterSpec),
			}
			resources.HTTPRouteFilters = append(resources.HTTPRouteFilters, httpRouteFilter)
		case KindBackend:
			typedSpec := spec.Interface()
			backend := &egv1a1.Backend{
				TypeMeta: metav1.TypeMeta{
					Kind: KindBackend,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(egv1a1.BackendSpec),
			}
			resources.Backends = append(resources.Backends, backend)
		case KindSecret:
			typedData := data.Interface()
			typedStringData := stringData.Interface()
			typedSecretType := kobjVal.FieldByName("Type").Interface()

			secret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind: KindSecret,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Type:       typedSecretType.(corev1.SecretType),
				Data:       typedData.(map[string][]byte),
				StringData: typedStringData.(map[string]string),
			}
			resources.Secrets = append(resources.Secrets, secret)
		case KindConfigMap:
			typedData := data.Interface()
			configMap := &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind: KindConfigMap,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Data: typedData.(map[string]string),
			}
			resources.ConfigMaps = append(resources.ConfigMaps, configMap)
		case KindBackendTLSPolicy:
			typedSpec := spec.Interface()
			backendTLSPolicy := &gwapiv1a3.BackendTLSPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind: KindBackendTLSPolicy,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(gwapiv1a3.BackendTLSPolicySpec),
			}
			resources.BackendTLSPolicies = append(resources.BackendTLSPolicies, backendTLSPolicy)
		case KindEnvoyExtensionPolicy:
			typedSpec := spec.Interface()
			envoyExtensionPolicy := &egv1a1.EnvoyExtensionPolicy{
				TypeMeta: metav1.TypeMeta{
					Kind: KindEnvoyExtensionPolicy,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: typedSpec.(egv1a1.EnvoyExtensionPolicySpec),
			}
			resources.EnvoyExtensionPolicies = append(resources.EnvoyExtensionPolicies, envoyExtensionPolicy)
		}

		return nil
	}); err != nil {
		return nil, err
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

	if addMissingResources {
		// Sort the required namespace in order to keep the consistency for test.
		sortedRequiredNamespace := requiredNamespaceMap.UnsortedList()
		sort.Strings(sortedRequiredNamespace)
		for _, ns := range sortedRequiredNamespace {
			if !providedNamespaceMap.Has(ns) {
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
		// Sort the required service in order to keep the consistency for test.
		sortedRequiredService := make([]string, 0, len(requiredServiceMap))
		for key := range requiredServiceMap {
			sortedRequiredService = append(sortedRequiredService, key)
		}
		sort.Strings(sortedRequiredService)

		providedServiceMap := map[string]*corev1.Service{}
		for _, service := range resources.Services {
			providedServiceMap[service.Namespace+"/"+service.Name] = service
		}

		for _, key := range sortedRequiredService {
			service := requiredServiceMap[key]
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

		// Add EnvoyProxy if it does not exist.
		if resources.EnvoyProxyForGatewayClass == nil {
			if err := addDefaultEnvoyProxy(resources); err != nil {
				return nil, err
			}
		}
	}

	return resources, nil
}

func addMissingServices(requiredServices map[string]*corev1.Service, obj interface{}) {
	var objNamespace string
	protocol := ir.TCPProtocolType

	var refs []gwapiv1.BackendRef
	switch route := obj.(type) {
	case *gwapiv1.HTTPRoute:
		objNamespace = route.Namespace
		for _, rule := range route.Spec.Rules {
			for _, httpBakcendRef := range rule.BackendRefs {
				refs = append(refs, httpBakcendRef.BackendRef)
			}
		}
	case *gwapiv1.GRPCRoute:
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
		protocol = ir.UDPProtocolType
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
			Protocol: corev1.Protocol(protocol),
			Port:     port,
		}
		if service, found := requiredServices[key]; !found {
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: ns,
				},
				Spec: corev1.ServiceSpec{
					// Just a dummy IP
					ClusterIP: dummyClusterIP,
					Ports:     []corev1.ServicePort{servicePort},
				},
			}
			requiredServices[key] = service
		} else {
			inserted := false
			for _, port := range service.Spec.Ports {
				if port.Protocol == servicePort.Protocol && port.Port == servicePort.Port {
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
				Value: &defaultBootstrapStr,
			},
		},
	}
	resources.EnvoyProxyForGatewayClass = ep
	ns := gwapiv1.Namespace(namespace)
	resources.GatewayClass.Spec.ParametersRef = &gwapiv1.ParametersReference{
		Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
		Kind:      KindEnvoyProxy,
		Name:      defaultEnvoyProxyName,
		Namespace: &ns,
	}
	return nil
}

// IterYAMLBytes iters every valid YAML resource from YAML bytes
// and process each of them by calling `handle` callback.
func IterYAMLBytes(input []byte, handle func([]byte) error) error {
	reader := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewBuffer(input)))
	for {
		yamlBytes, err := reader.Read()
		if errors.Is(err, io.EOF) || len(yamlBytes) == 0 {
			break
		} else if err != nil {
			return err
		}
		if err = handle(yamlBytes); err != nil {
			return err
		}
	}
	return nil
}
