// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

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
		return admission.Allowed("internal error, skipped")
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
		return admission.Allowed("internal error, skipped")
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
		return admission.Allowed("internal error, skipped")
	}

	if zone, ok := node.Labels[corev1.LabelTopologyZone]; ok {
		if binding.Annotations == nil {
			binding.Annotations = map[string]string{}
		}
		binding.Annotations[corev1.LabelTopologyZone] = zone
	} else {
		return admission.Allowed("Skipping injection due to missing topology label on node")
	}

	marshaledBinding, err := json.Marshal(binding)
	if err != nil {
		klog.Errorf("failed to marshal Pod Binding: %v", err)
		return admission.Allowed(fmt.Sprintf("failed to marshal binding, skipped: %v", err))
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledBinding)
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
