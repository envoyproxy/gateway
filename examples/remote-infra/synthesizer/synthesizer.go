// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package synthesizer

import (
	"context"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// InfraSynthesizer reconciles Kubernetes resources (Service, Deployment) for
// a proxy fleet described by the Envoy Gateway remote infrastructure IR.
type InfraSynthesizer struct {
	KubernetesClient client.Client
	Namespace        string
}

// templateParams defines the template variables to inject into the bootstrap yaml.
type templateParams struct {
	// ServiceClusterName is the generated name of the Envoy ServiceCluster.
	ServiceClusterName string
}

//go:embed basic_bootstrap.yaml.tpl
var bootstrapTmplStr string

var bootstrapTmpl = template.Must(template.New("bootstrap.yaml").Funcs(template.FuncMap{
	"base64": func(data []byte) string {
		return base64.StdEncoding.EncodeToString(data)
	},
}).Parse(bootstrapTmplStr))

// CreateOrUpdate brings the Service and Deployment for the given IR into the
// desired state, creating them if absent and patching them otherwise.
func (is *InfraSynthesizer) CreateOrUpdate(ctx context.Context, ir *Infra) error {
	if err := validateIR(ir); err != nil {
		return err
	}

	desiredSvc, err := is.GetService(ir)
	if err != nil {
		return fmt.Errorf("build service: %w", err)
	}
	existingSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      desiredSvc.Name,
			Namespace: desiredSvc.Namespace,
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, is.KubernetesClient, existingSvc, func() error {
		mergeServiceSpec(existingSvc, desiredSvc)
		return nil
	}); err != nil {
		return fmt.Errorf("apply service %s/%s: %w", desiredSvc.Namespace, desiredSvc.Name, err)
	}

	desiredDeploy, err := is.GetDeployment(ctx, ir)
	if err != nil {
		return fmt.Errorf("build deployment: %w", err)
	}
	existingDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      desiredDeploy.Name,
			Namespace: desiredDeploy.Namespace,
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, is.KubernetesClient, existingDeploy, func() error {
		mergeDeploymentSpec(existingDeploy, desiredDeploy)
		return nil
	}); err != nil {
		return fmt.Errorf("apply deployment %s/%s: %w", desiredDeploy.Namespace, desiredDeploy.Name, err)
	}

	return nil
}

// Delete removes the Deployment and Service for the given IR. NotFound errors
// are ignored so callers can retry safely.
func (is *InfraSynthesizer) Delete(ctx context.Context, ir *Infra) error {
	if err := validateIR(ir); err != nil {
		return err
	}

	name := resourceName(ir)

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: is.Namespace,
		},
	}
	if err := is.KubernetesClient.Delete(ctx, deploy); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete deployment %s/%s: %w", is.Namespace, name, err)
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: is.Namespace,
		},
	}
	if err := is.KubernetesClient.Delete(ctx, svc); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete service %s/%s: %w", is.Namespace, name, err)
	}

	return nil
}

// validateIR returns an error if the IR is missing fields the synthesizer
// needs in order to derive a resource name.
func validateIR(ir *Infra) error {
	if ir == nil || ir.Proxy == nil {
		return errors.New("infra IR is missing a Proxy")
	}
	if ir.Proxy.Name == "" {
		return errors.New("infra IR Proxy is missing a Name")
	}
	return nil
}

// resourceName derives the Service / Deployment object name from the IR. The
// upstream IR uses "/" to separate the gateway namespace and name, which is
// not a legal character in a Kubernetes object name.
func resourceName(ir *Infra) string {
	return strings.ReplaceAll(ir.Proxy.Name, "/", "-")
}
