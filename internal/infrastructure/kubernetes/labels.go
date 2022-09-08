package kubernetes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// envoyLabels returns the labels used for all Envoy resources.
func envoyLabels() map[string]string {
	return map[string]string{
		"app.gateway.envoyproxy.io/name": "envoy",
	}
}

// envoyPodSelector returns a label selector used to select Envoy pods.
//
// TODO: Update k/v pair to use gatewayclass controller name to distinguish between
//       multiple Envoy Gateways.
func envoyPodSelector(gcName string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: envoyPodLabels(gcName),
	}
}

// envoyPodLabels returns the labels used for Envoy pods.
func envoyPodLabels(gcName string) map[string]string {
	lbls := envoyLabels()
	lbls["gatewayClass"] = gcName
	return lbls
}
