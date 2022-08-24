package kubernetes

// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=gatewayclasses;gateways;httproutes;referencepolicies;referencegrants,verbs=get;list;watch
// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=gatewayclasses/status;gateways/status;httproutes/status,verbs=update

// +kubebuilder:rbac:groups="",resources=secrets;services;namespaces,verbs=get;list;watch
