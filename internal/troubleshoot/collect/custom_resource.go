// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package collect

import (
	"bytes"
	"context"
	"fmt"
	"path"

	troubleshootv1b2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	tbcollect "github.com/replicatedhq/troubleshoot/pkg/collect"
	"github.com/replicatedhq/troubleshoot/pkg/constants"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var _ tbcollect.Collector = &CustomResource{}

// CustomResource defines a custom resource collector, which collect custom resources from the cluster,
// with the given IncludeGroups and Namespaces.
type CustomResource struct {
	BundlePath    string
	Namespace     string
	ClientConfig  *rest.Config
	Namespaces    []string
	IncludeGroups []string
}

func (cr *CustomResource) Title() string {
	return "custom-resource"
}

func (cr *CustomResource) IsExcluded() (bool, error) {
	return false, nil
}

func (cr *CustomResource) GetRBACErrors() []error {
	return nil
}

func (cr *CustomResource) HasRBACErrors() bool {
	return false
}

func (cr *CustomResource) CheckRBAC(_ context.Context, _ tbcollect.Collector, _ *troubleshootv1b2.Collect, _ *rest.Config, _ string) error {
	return nil
}

func (cr *CustomResource) Collect(_ chan<- interface{}) (tbcollect.CollectorResult, error) {
	ctx := context.Background()
	output := tbcollect.NewResult()
	client, err := kubernetes.NewForConfig(cr.ClientConfig)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(cr.ClientConfig)
	if err != nil {
		return nil, err
	}

	// namespaces
	var namespaceNames []string
	switch {
	case len(cr.Namespaces) > 0:
		namespaceNames = cr.Namespaces
	case cr.Namespace != "":
		namespaceNames = append(namespaceNames, cr.Namespace)
	default:
		_, namespaceList, _ := getAllNamespaces(ctx, client)
		if namespaceList != nil {
			for i := range namespaceList.Items {
				namespace := &namespaceList.Items[i]
				namespaceNames = append(namespaceNames, namespace.Name)
			}
		}
	}

	// crs
	customResources, crErrors := crs(ctx, dynamicClient, client, cr.ClientConfig, namespaceNames, cr.IncludeGroups)
	for k, v := range customResources {
		p := path.Join(constants.CLUSTER_RESOURCES_DIR, k)
		_ = output.SaveResult(cr.BundlePath, p, bytes.NewBuffer(v))
	}
	errPath := path.Join(constants.CLUSTER_RESOURCES_DIR, fmt.Sprintf("%s-errors.json", constants.CLUSTER_RESOURCES_CUSTOM_RESOURCES))
	_ = output.SaveResult(cr.BundlePath, errPath, marshalErrors(crErrors))

	return output, nil
}
