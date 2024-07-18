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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var _ tbcollect.Collector = &EnvoyGatewayResource{}

// EnvoyGatewayResource defines a collector for the Envoy Gateway resource from the given namespace.
// This is most like the CusterResource collector, but remove unnecessary types like StatefulSet, CornJob etc.
type EnvoyGatewayResource struct {
	BundlePath   string
	Namespace    string
	ClientConfig *rest.Config
}

func (eg EnvoyGatewayResource) Title() string {
	return "envoy-gateway-resource"
}

func (eg EnvoyGatewayResource) IsExcluded() (bool, error) {
	return false, nil
}

func (eg EnvoyGatewayResource) GetRBACErrors() []error {
	return nil
}

func (eg EnvoyGatewayResource) HasRBACErrors() bool {
	return false
}

func (eg EnvoyGatewayResource) CheckRBAC(_ context.Context, _ tbcollect.Collector, _ *troubleshootv1b2.Collect, _ *rest.Config, _ string) error {
	return nil
}

func (eg EnvoyGatewayResource) Collect(_ chan<- interface{}) (tbcollect.CollectorResult, error) {
	ctx := context.Background()
	output := tbcollect.NewResult()
	client, err := kubernetes.NewForConfig(eg.ClientConfig)
	if err != nil {
		return nil, err
	}
	namespaceNames := []string{eg.Namespace}

	// pods
	pods, podErrors, _ := pods(ctx, client, namespaceNames)
	for k, v := range pods {
		_ = output.SaveResult(eg.BundlePath, path.Join(constants.CLUSTER_RESOURCES_DIR, constants.CLUSTER_RESOURCES_PODS, k), bytes.NewBuffer(v))
	}
	_ = output.SaveResult(eg.BundlePath, path.Join(constants.CLUSTER_RESOURCES_DIR, fmt.Sprintf("%s-errors.json", constants.CLUSTER_RESOURCES_PODS)), marshalErrors(podErrors))

	// services
	services, servicesErrors := services(ctx, client, namespaceNames)
	for k, v := range services {
		_ = output.SaveResult(eg.BundlePath, path.Join(constants.CLUSTER_RESOURCES_DIR, constants.CLUSTER_RESOURCES_SERVICES, k), bytes.NewBuffer(v))
	}
	_ = output.SaveResult(eg.BundlePath, path.Join(constants.CLUSTER_RESOURCES_DIR, fmt.Sprintf("%s-errors.json", constants.CLUSTER_RESOURCES_SERVICES)), marshalErrors(servicesErrors))

	// deployments
	deployments, deploymentsErrors := deployments(ctx, client, namespaceNames)
	for k, v := range deployments {
		_ = output.SaveResult(eg.BundlePath, path.Join(constants.CLUSTER_RESOURCES_DIR, constants.CLUSTER_RESOURCES_DEPLOYMENTS, k), bytes.NewBuffer(v))
	}
	_ = output.SaveResult(eg.BundlePath, path.Join(constants.CLUSTER_RESOURCES_DIR, fmt.Sprintf("%s-errors.json", constants.CLUSTER_RESOURCES_DEPLOYMENTS)), marshalErrors(deploymentsErrors))

	// daemonsets
	daemonsets, daemonsetsErrors := daemonsets(ctx, client, namespaceNames)
	for k, v := range daemonsets {
		_ = output.SaveResult(eg.BundlePath, path.Join(constants.CLUSTER_RESOURCES_DIR, constants.CLUSTER_RESOURCES_DAEMONSETS, k), bytes.NewBuffer(v))
	}
	_ = output.SaveResult(eg.BundlePath, path.Join(constants.CLUSTER_RESOURCES_DIR, fmt.Sprintf("%s-errors.json", constants.CLUSTER_RESOURCES_DAEMONSETS)), marshalErrors(daemonsetsErrors))

	// jobs
	jobs, jobsErrors := jobs(ctx, client, namespaceNames)
	for k, v := range jobs {
		_ = output.SaveResult(eg.BundlePath, path.Join(constants.CLUSTER_RESOURCES_DIR, constants.CLUSTER_RESOURCES_JOBS, k), bytes.NewBuffer(v))
	}
	_ = output.SaveResult(eg.BundlePath, path.Join(constants.CLUSTER_RESOURCES_DIR, fmt.Sprintf("%s-errors.json", constants.CLUSTER_RESOURCES_JOBS)), marshalErrors(jobsErrors))

	// ConfigMaps
	configMaps, configMapsErrors := configMaps(ctx, client, namespaceNames)
	for k, v := range configMaps {
		_ = output.SaveResult(eg.BundlePath, path.Join(constants.CLUSTER_RESOURCES_DIR, constants.CLUSTER_RESOURCES_CONFIGMAPS, k), bytes.NewBuffer(v))
	}
	_ = output.SaveResult(eg.BundlePath, path.Join(constants.CLUSTER_RESOURCES_DIR, fmt.Sprintf("%s-errors.json", constants.CLUSTER_RESOURCES_CONFIGMAPS)), marshalErrors(configMapsErrors))

	return output, nil
}
