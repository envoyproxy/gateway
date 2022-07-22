package infrastructure

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	"github.com/envoyproxy/gateway/internal/ir"
)

// Manager provides the scaffolding for managing infrastructure.
type Manager struct {
	Kubernetes *kubernetes.Infra
}

// NewManager returns a new infrastructure Manager.
func NewManager(infra *ir.Infra) (*Manager, error) {
	mgr := new(Manager)

	if *infra.GetProvider() == v1alpha1.ProviderTypeKubernetes {
		cli, err := client.New(clicfg.GetConfigOrDie(), client.Options{})
		if err != nil {
			return nil, err
		}
		mgr.Kubernetes = kubernetes.NewInfra(cli)
	} else {
		// Kube is the only supported provider type.
		return nil, fmt.Errorf("unsupported provider type %v", infra.Provider)
	}

	return mgr, nil
}
