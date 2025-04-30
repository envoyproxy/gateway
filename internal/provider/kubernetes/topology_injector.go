// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-openapi/jsonpointer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/metrics"
)

type ProxyTopologyInjector struct {
	client.Client
	Decoder admission.Decoder
}

func (m *ProxyTopologyInjector) Handle(ctx context.Context, req admission.Request) admission.Response {
	binding := &corev1.Binding{}
	if err := m.Decoder.Decode(req, binding); err != nil {
		klog.Error(err, "decoding binding failed", "request.ObjectKind", req.Object.Object.GetObjectKind())
		topologyInjectorEventsTotal.WithFailure(metrics.ReasonError).Increment()
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if binding.Target.Name == "" {
		topologyInjectorEventsTotal.WithStatus(statusNoAction).Increment()
		return admission.Allowed("skipped")
	}

	podName := types.NamespacedName{
		Namespace: binding.Namespace,
		Name:      binding.Name,
	}

	pod := &corev1.Pod{}
	if err := m.Get(ctx, podName, pod); err != nil {
		klog.Error(err, "get pod failed", "pod", podName.String())
		topologyInjectorEventsTotal.WithFailure(metrics.ReasonError).Increment()
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// Skip non-proxy pods
	if !hasEnvoyProxyLabels(pod.Labels) {
		klog.V(1).Info("skipping pod due to missing labels", "pod", podName)
		topologyInjectorEventsTotal.WithStatus(statusNoAction).Increment()
		return admission.Allowed("skipped")
	}

	nodeName := types.NamespacedName{
		Name: binding.Target.Name,
	}
	node := &corev1.Node{}
	if err := m.Get(ctx, nodeName, node); err != nil {
		klog.Error(err, "get node failed", "node", node.Name)
		topologyInjectorEventsTotal.WithFailure(metrics.ReasonError).Increment()
		return admission.Errored(http.StatusInternalServerError, err)
	}

	var patch string
	if zone, ok := node.Labels[corev1.LabelTopologyZone]; ok {
		patch = fmt.Sprintf(`[{"op":"replace", "path":"/metadata/labels/%s", "value":"%s"}]`, jsonpointer.Escape(corev1.LabelTopologyZone), zone)
	}

	rawPatch := client.RawPatch(types.JSONPatchType, []byte(patch))
	if err := m.Patch(ctx, pod, rawPatch); err != nil {
		klog.Error(err, "patch pod failed", "pod", podName.String())
		topologyInjectorEventsTotal.WithFailure(metrics.ReasonError).Increment()
		return admission.Errored(http.StatusInternalServerError, err)
	}
	klog.V(1).Info("patch pod succeeded", "pod", podName.String())
	topologyInjectorEventsTotal.WithSuccess().Increment()
	return admission.Allowed("pod patched")
}

func hasEnvoyProxyLabels(labels map[string]string) bool {
	if labels["app.kubernetes.io/component"] != "proxy" {
		return false
	}

	if labels[gatewayapi.OwningGatewayNameLabel] == "" && labels[gatewayapi.OwningGatewayClassLabel] == "" {
		return false
	}

	if labels["app.kubernetes.io/managed-by"] != "envoy-gateway" {
		return false
	}

	if labels["app.kubernetes.io/name"] != "envoy" {
		return false
	}

	return true
}
