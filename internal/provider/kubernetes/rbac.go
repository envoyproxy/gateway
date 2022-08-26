package kubernetes

// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=gatewayclasses;gateways;httproutes;referencepolicies;referencegrants,verbs=get;list;watch
// +kubebuilder:rbac:groups="gateway.networking.k8s.io",resources=gatewayclasses/status;gateways/status;httproutes/status,verbs=update

// RBAC for Infra Manager to manage Envoy.
// +kubebuilder:rbac:groups="",resources=secrets;namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services;serviceaccounts,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
