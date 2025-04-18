// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Functions here are borrowed from the troubleshoot project because they're not exported.
// There's some differences in the implementation, but the core logic is the same.
// 1. Remove `managedFields` from the objects
// 2. Remove json output

package collect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/replicatedhq/troubleshoot/pkg/k8sutil"
	"github.com/replicatedhq/troubleshoot/pkg/k8sutil/discovery"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1b1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextv1clientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	apiextv1b1clientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"
)

// Use for error maps and arrays. These are guaranteed to not result in an error when marshaling.
func marshalErrors(errors interface{}) io.Reader {
	if errors == nil {
		return nil
	}

	val := reflect.ValueOf(errors)
	switch val.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		if val.Len() == 0 {
			return nil
		}
	default:
		// do nothing
	}

	m, _ := json.MarshalIndent(errors, "", "  ")
	return bytes.NewBuffer(m)
}

func getAllNamespaces(ctx context.Context, client *kubernetes.Clientset) ([]byte, *corev1.NamespaceList, []string) {
	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, []string{err.Error()}
	}

	gvk, err := apiutil.GVKForObject(namespaces, scheme.Scheme)
	if err == nil {
		namespaces.GetObjectKind().SetGroupVersionKind(gvk)
	}

	for i, o := range namespaces.Items {
		gvk, err := apiutil.GVKForObject(&o, scheme.Scheme)
		if err == nil {
			namespaces.Items[i].GetObjectKind().SetGroupVersionKind(gvk)
		}
	}

	b, err := yaml.Marshal(namespaces)
	if err != nil {
		return nil, nil, []string{err.Error()}
	}

	return b, namespaces, nil
}

func crs(ctx context.Context, dyn dynamic.Interface, client *kubernetes.Clientset, config *rest.Config, namespaces, includeGroups []string) (map[string][]byte, map[string]string) {
	includeGroupSet := sets.New[string](includeGroups...)
	errorList := make(map[string]string)
	ok, err := discovery.HasResource(client, "apiextensions.k8s.io/v1", "CustomResourceDefinition")
	if err != nil {
		return nil, map[string]string{"discover apiextensions.k8s.io/v1": err.Error()}
	}
	if ok {
		crdClient, err := apiextv1clientset.NewForConfig(config)
		if err != nil {
			errorList["crdClient"] = err.Error()
			return map[string][]byte{}, errorList
		}
		return crsV1(ctx, dyn, crdClient, namespaces, includeGroupSet)
	}

	crdClient, err := apiextv1b1clientset.NewForConfig(config)
	if err != nil {
		errorList["crdClient"] = err.Error()
		return map[string][]byte{}, errorList
	}
	return crsV1beta(ctx, dyn, crdClient, namespaces, includeGroupSet)
}

func crsV1(ctx context.Context, client dynamic.Interface, crdClient apiextv1clientset.ApiextensionsV1Interface,
	namespaces []string, includeGroups sets.Set[string],
) (map[string][]byte, map[string]string) {
	customResources := make(map[string][]byte)
	errorList := make(map[string]string)

	crds, err := crdClient.CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		errorList["crdList"] = err.Error()
		return customResources, errorList
	}

	metaAccessor := meta.NewAccessor()

	// Loop through CRDs to fetch the CRs
	for _, crd := range crds.Items {
		// A resource that contains '/' is a subresource type, and it has no
		// object instances
		if strings.ContainsAny(crd.Name, "/") {
			continue
		}

		var version string
		if len(crd.Spec.Versions) > 0 {
			versions := []string{}
			for _, v := range crd.Spec.Versions {
				versions = append(versions, v.Name)
			}

			version = versions[0]
			if len(versions) > 1 {
				version = selectCRDVersionByPriority(versions)
			}
		}

		if len(includeGroups) > 0 &&
			!includeGroups.Has(crd.Spec.Group) {
			continue
		}

		gvr := schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  version,
			Resource: crd.Spec.Names.Plural,
		}
		isNamespacedResource := crd.Spec.Scope == apiextensionsv1.NamespaceScoped

		// Fetch all resources of given type
		customResourceList, err := client.Resource(gvr).List(ctx, metav1.ListOptions{})
		if err != nil {
			errorList[crd.Name] = err.Error()
			continue
		}

		if len(customResourceList.Items) == 0 {
			continue
		}

		if !isNamespacedResource {
			objects := []map[string]interface{}{}
			for _, item := range customResourceList.Items {
				objects = append(objects, item.Object)
			}
			err := storeCustomResource(crd.Name, objects, customResources)
			if err != nil {
				errorList[crd.Name] = err.Error()
				continue
			}
		} else {
			// Group fetched resources by the namespace
			perNamespace := map[string][]map[string]interface{}{}
			errors := []string{}

			for _, item := range customResourceList.Items {
				omitManagedFields(&item)
				ns, err := metaAccessor.Namespace(&item)
				if err != nil {
					errors = append(errors, err.Error())
					continue
				}
				if perNamespace[ns] == nil {
					perNamespace[ns] = []map[string]interface{}{}
				}
				perNamespace[ns] = append(perNamespace[ns], item.Object)
			}

			if len(errors) > 0 {
				errorList[crd.Name] = strings.Join(errors, "\n")
			}

			// Only include resources from requested namespaces
			for _, ns := range namespaces {
				if len(perNamespace[ns]) == 0 {
					continue
				}

				namespacedName := fmt.Sprintf("%s/%s", crd.Name, ns)
				err := storeCustomResource(namespacedName, perNamespace[ns], customResources)
				if err != nil {
					errorList[namespacedName] = err.Error()
					continue
				}
			}
		}
	}

	return customResources, errorList
}

func omitManagedFields(o runtime.Object) runtime.Object {
	a, err := meta.Accessor(o)
	if err != nil {
		// The object is not a `metav1.Object`, ignore it.
		return o
	}
	a.SetManagedFields(nil)
	return o
}

func crsV1beta(ctx context.Context, client dynamic.Interface, crdClient apiextv1b1clientset.ApiextensionsV1beta1Interface,
	namespaces []string, includeGroups sets.Set[string],
) (map[string][]byte, map[string]string) {
	customResources := make(map[string][]byte)
	errorList := make(map[string]string)

	crds, err := crdClient.CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		errorList["crdList"] = err.Error()
		return customResources, errorList
	}

	metaAccessor := meta.NewAccessor()

	// Loop through CRDs to fetch the CRs
	for _, crd := range crds.Items {
		// A resource that contains '/' is a subresource type, and it has no
		// object instances
		if strings.ContainsAny(crd.Name, "/") {
			continue
		}

		if len(includeGroups) > 0 &&
			includeGroups.Has(crd.Spec.Group) {
			continue
		}

		gvr := schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  crd.Spec.Version,
			Resource: crd.Spec.Names.Plural,
		}

		if len(crd.Spec.Versions) > 0 {
			versions := []string{}
			for _, v := range crd.Spec.Versions {
				versions = append(versions, v.Name)
			}

			version := versions[0]
			if len(versions) > 1 {
				version = selectCRDVersionByPriority(versions)
			}
			gvr.Version = version
		}

		isNamespacedResource := crd.Spec.Scope == apiextensionsv1b1.NamespaceScoped

		// Fetch all resources of given type
		customResourceList, err := client.Resource(gvr).List(ctx, metav1.ListOptions{})
		if err != nil {
			errorList[crd.Name] = err.Error()
			continue
		}

		if len(customResourceList.Items) == 0 {
			continue
		}

		if !isNamespacedResource {
			objects := []map[string]interface{}{}
			for _, item := range customResourceList.Items {
				objects = append(objects, item.Object)
			}

			err = storeCustomResource(crd.Name, objects, customResources)
			if err != nil {
				errorList[crd.Name] = err.Error()
				continue
			}

		} else {
			// Group fetched resources by the namespace
			perNamespace := map[string][]map[string]interface{}{}
			errors := []string{}

			for _, item := range customResourceList.Items {
				ns, err := metaAccessor.Namespace(&item)
				if err != nil {
					errors = append(errors, err.Error())
					continue
				}
				if perNamespace[ns] == nil {
					perNamespace[ns] = []map[string]interface{}{}
				}
				perNamespace[ns] = append(perNamespace[ns], item.Object)
			}

			if len(errors) > 0 {
				errorList[crd.Name] = strings.Join(errors, "\n")
			}

			// Only include resources from requested namespaces
			for _, ns := range namespaces {
				if len(perNamespace[ns]) == 0 {
					continue
				}

				namespacedName := fmt.Sprintf("%s/%s", crd.Name, ns)
				err := storeCustomResource(namespacedName, perNamespace[ns], customResources)
				if err != nil {
					errorList[namespacedName] = err.Error()
					continue
				}
			}
		}
	}

	return customResources, errorList
}

// Selects the newest version by kube-aware priority.
func selectCRDVersionByPriority(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	sort.Slice(versions, func(i, j int) bool {
		return version.CompareKubeAwareVersionStrings(versions[i], versions[j]) < 0
	})
	return versions[len(versions)-1]
}

func storeCustomResource(name string, objects any, m map[string][]byte) error {
	y, err := yaml.Marshal(objects)
	if err != nil {
		return err
	}

	m[fmt.Sprintf("%s.yaml", name)] = y
	return nil
}

func pods(ctx context.Context, client *kubernetes.Clientset, namespaces []string) (map[string][]byte, map[string]string, []corev1.Pod) {
	podsByNamespace := make(map[string][]byte)
	errorsByNamespace := make(map[string]string)
	unhealthyPods := []corev1.Pod{}

	for _, namespace := range namespaces {
		pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			errorsByNamespace[namespace] = err.Error()
			continue
		}

		if len(pods.Items) == 0 {
			continue
		}

		gvk, err := apiutil.GVKForObject(pods, scheme.Scheme)
		if err == nil {
			pods.GetObjectKind().SetGroupVersionKind(gvk)
		}

		for i, o := range pods.Items {
			gvk, err := apiutil.GVKForObject(&o, scheme.Scheme)
			if err == nil {
				pods.Items[i].GetObjectKind().SetGroupVersionKind(gvk)
				pods.Items[i].SetManagedFields(nil)
			}
		}

		b, err := yaml.Marshal(pods)
		if err != nil {
			errorsByNamespace[namespace] = err.Error()
			continue
		}

		for _, pod := range pods.Items {
			if k8sutil.IsPodUnhealthy(&pod) {
				unhealthyPods = append(unhealthyPods, pod)
			}
		}

		podsByNamespace[namespace+".yaml"] = b
	}

	return podsByNamespace, errorsByNamespace, unhealthyPods
}

func services(ctx context.Context, client *kubernetes.Clientset, namespaces []string) (map[string][]byte, map[string]string) {
	servicesByNamespace := make(map[string][]byte)
	errorsByNamespace := make(map[string]string)

	for _, namespace := range namespaces {
		services, err := client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			errorsByNamespace[namespace] = err.Error()
			continue
		}

		if len(services.Items) == 0 {
			continue
		}

		gvk, err := apiutil.GVKForObject(services, scheme.Scheme)
		if err == nil {
			services.GetObjectKind().SetGroupVersionKind(gvk)
		}

		for i, o := range services.Items {
			gvk, err := apiutil.GVKForObject(&o, scheme.Scheme)
			if err == nil {
				services.Items[i].GetObjectKind().SetGroupVersionKind(gvk)
				services.Items[i].SetManagedFields(nil)
			}
		}

		b, err := yaml.Marshal(services)
		if err != nil {
			errorsByNamespace[namespace] = err.Error()
			continue
		}

		servicesByNamespace[namespace+".yaml"] = b
	}

	return servicesByNamespace, errorsByNamespace
}

func deployments(ctx context.Context, client *kubernetes.Clientset, namespaces []string) (map[string][]byte, map[string]string) {
	deploymentsByNamespace := make(map[string][]byte)
	errorsByNamespace := make(map[string]string)

	for _, namespace := range namespaces {
		deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			errorsByNamespace[namespace] = err.Error()
			continue
		}

		if len(deployments.Items) == 0 {
			continue
		}

		gvk, err := apiutil.GVKForObject(deployments, scheme.Scheme)
		if err == nil {
			deployments.GetObjectKind().SetGroupVersionKind(gvk)
		}

		for i, o := range deployments.Items {
			gvk, err := apiutil.GVKForObject(&o, scheme.Scheme)
			if err == nil {
				deployments.Items[i].GetObjectKind().SetGroupVersionKind(gvk)
				deployments.Items[i].SetManagedFields(nil)
			}
		}

		b, err := yaml.Marshal(deployments)
		if err != nil {
			errorsByNamespace[namespace] = err.Error()
			continue
		}

		deploymentsByNamespace[namespace+".yaml"] = b
	}

	return deploymentsByNamespace, errorsByNamespace
}

func daemonsets(ctx context.Context, client *kubernetes.Clientset, namespaces []string) (map[string][]byte, map[string]string) {
	daemonsetsByNamespace := make(map[string][]byte)
	errorsByNamespace := make(map[string]string)

	for _, namespace := range namespaces {
		daemonsets, err := client.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			errorsByNamespace[namespace] = err.Error()
			continue
		}

		if len(daemonsets.Items) == 0 {
			continue
		}

		gvk, err := apiutil.GVKForObject(daemonsets, scheme.Scheme)
		if err == nil {
			daemonsets.GetObjectKind().SetGroupVersionKind(gvk)
		}

		for i, o := range daemonsets.Items {
			gvk, err := apiutil.GVKForObject(&o, scheme.Scheme)
			if err == nil {
				daemonsets.Items[i].GetObjectKind().SetGroupVersionKind(gvk)
				daemonsets.Items[i].SetManagedFields(nil)
			}
		}

		b, err := yaml.Marshal(daemonsets)
		if err != nil {
			errorsByNamespace[namespace] = err.Error()
			continue
		}

		daemonsetsByNamespace[namespace+".yaml"] = b
	}

	return daemonsetsByNamespace, errorsByNamespace
}

func jobs(ctx context.Context, client *kubernetes.Clientset, namespaces []string) (map[string][]byte, map[string]string) {
	jobsByNamespace := make(map[string][]byte)
	errorsByNamespace := make(map[string]string)

	for _, namespace := range namespaces {
		nsJobs, err := client.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			errorsByNamespace[namespace] = err.Error()
			continue
		}

		if len(nsJobs.Items) == 0 {
			continue
		}

		gvk, err := apiutil.GVKForObject(nsJobs, scheme.Scheme)
		if err == nil {
			nsJobs.GetObjectKind().SetGroupVersionKind(gvk)
		}

		for i, o := range nsJobs.Items {
			gvk, err := apiutil.GVKForObject(&o, scheme.Scheme)
			if err == nil {
				nsJobs.Items[i].GetObjectKind().SetGroupVersionKind(gvk)
				nsJobs.Items[i].SetManagedFields(nil)
			}
		}

		b, err := yaml.Marshal(nsJobs)
		if err != nil {
			errorsByNamespace[namespace] = err.Error()
			continue
		}

		jobsByNamespace[namespace+".yaml"] = b
	}

	return jobsByNamespace, errorsByNamespace
}

func configMaps(ctx context.Context, client kubernetes.Interface, namespaces []string) (map[string][]byte, map[string]string) {
	configmapByNamespace := make(map[string][]byte)
	errorsByNamespace := make(map[string]string)

	for _, namespace := range namespaces {
		configmaps, err := client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			errorsByNamespace[namespace] = err.Error()
			continue
		}

		if len(configmaps.Items) == 0 {
			continue
		}

		gvk, err := apiutil.GVKForObject(configmaps, scheme.Scheme)
		if err == nil {
			configmaps.GetObjectKind().SetGroupVersionKind(gvk)
		}

		for i, o := range configmaps.Items {
			gvk, err := apiutil.GVKForObject(&o, scheme.Scheme)
			if err == nil {
				configmaps.Items[i].GetObjectKind().SetGroupVersionKind(gvk)
				configmaps.Items[i].SetManagedFields(nil)
			}
		}

		b, err := yaml.Marshal(configmaps)
		if err != nil {
			errorsByNamespace[namespace] = err.Error()
			continue
		}

		configmapByNamespace[namespace+".yaml"] = b
	}

	return configmapByNamespace, errorsByNamespace
}
