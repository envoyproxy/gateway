// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoy

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-openapi/jsonpointer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

type ProxyTopologyMutator struct {
	client.Client
	Decoder admission.Decoder
}

func (m *ProxyTopologyMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	binding := &corev1.Binding{}
	if err := m.Decoder.Decode(req, binding); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if binding.Target.Name == "" {
		return admission.Allowed("skipped")
	}

	podName := types.NamespacedName{
		Namespace: binding.Namespace,
		Name:      binding.Name,
	}

	pod := &corev1.Pod{}
	if err := m.Get(ctx, podName, pod); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// Skip non-proxy pods
	if v, ok := pod.Labels["app.kubernetes.io/component"]; !ok || v != "proxy" {
		return admission.Allowed("skipped")
	}
	if pod.Labels[gatewayapi.OwningGatewayNameLabel] == "" && pod.Labels[gatewayapi.OwningGatewayClassLabel] == "" {
		return admission.Allowed("skipped")
	}

	nodeName := types.NamespacedName{
		Name: binding.Target.Name,
	}
	node := &corev1.Node{}
	if err := m.Get(ctx, nodeName, node); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	var patch string
	if zone, ok := node.Labels[corev1.LabelTopologyZone]; ok {
		patch = fmt.Sprintf(`[{"op":"replace", "path":"/metadata/labels/%s", "value":"%s"}]`, jsonpointer.Escape(corev1.LabelTopologyZone), zone)
	}

	rawPatch := client.RawPatch(types.JSONPatchType, []byte(patch))
	if err := m.Patch(ctx, pod, rawPatch); err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.Allowed("pod patched")
}

func TopologyWebhook(certDir, certName, keyName, healthProbeBindAddress string, port int) error {
	mgr, _ := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		HealthProbeBindAddress: healthProbeBindAddress,
		WebhookServer: webhook.NewServer(webhook.Options{
			CertDir:  certDir,
			CertName: certName,
			KeyName:  keyName,
			Port:     port,
		}),
	})
	svr := mgr.GetWebhookServer()

	if err := mgr.AddHealthzCheck("healthz", func(req *http.Request) error {
		return svr.StartedChecker()(req)
	}); err != nil {
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", func(req *http.Request) error {
		return svr.StartedChecker()(req)
	}); err != nil {
		return err
	}

	svr.Register("/mutate-pod-topology", &webhook.Admission{
		Handler: &ProxyTopologyMutator{
			Client:  mgr.GetClient(),
			Decoder: admission.NewDecoder(mgr.GetScheme()),
		},
	})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	return mgr.Start(ctx)
}
