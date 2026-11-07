// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/logging"
)

type fakeBootstrapValidator struct {
	calls []*egv1a1.ProxyBootstrap
	err   error
}

func (f *fakeBootstrapValidator) Validate(cfg *egv1a1.ProxyBootstrap) error {
	f.calls = append(f.calls, cfg)
	return f.err
}

func newBootstrapValidatorResources() (*resource.Resources, *egv1a1.ProxyBootstrap, *egv1a1.ProxyBootstrap) {
	gcBootstrap := &egv1a1.ProxyBootstrap{Value: new("gatewayclass")}
	gwBootstrap := &egv1a1.ProxyBootstrap{Value: new("gateway")}

	gc := &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "test-gc"}}
	gcProxy := &egv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{Namespace: "envoy-gateway-system", Name: "gc-proxy"},
		Spec:       egv1a1.EnvoyProxySpec{Bootstrap: gcBootstrap},
	}
	gwProxy := &egv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "gw-proxy"},
		Spec:       egv1a1.EnvoyProxySpec{Bootstrap: gwBootstrap},
	}
	gw := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "gw"},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: "test-gc",
			Infrastructure: &gwapiv1.GatewayInfrastructure{
				ParametersRef: &gwapiv1.LocalParametersReference{
					Group: gwapiv1.Group(egv1a1.GroupVersion.Group),
					Kind:  egv1a1.KindEnvoyProxy,
					Name:  "gw-proxy",
				},
			},
		},
	}

	return &resource.Resources{
		GatewayClass:              gc,
		EnvoyProxyForGatewayClass: gcProxy,
		EnvoyProxiesForGateways:   []*egv1a1.EnvoyProxy{gwProxy},
		Gateways:                  []*gwapiv1.Gateway{gw},
	}, gcBootstrap, gwBootstrap
}

func findCondition(conds []metav1.Condition, condType string) *metav1.Condition {
	for i := range conds {
		if conds[i].Type == condType {
			return &conds[i]
		}
	}
	return nil
}

func TestGetRelevantGatewaysBootstrapValidator(t *testing.T) {
	t.Run("validates bootstrap of GatewayClass and Gateway EnvoyProxy", func(t *testing.T) {
		resources, gcBootstrap, gwBootstrap := newBootstrapValidatorResources()
		fake := &fakeBootstrapValidator{}
		tr := &Translator{
			GatewayClassName:   "test-gc",
			BootstrapValidator: fake,
			Logger:             logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
		}

		accepted, failed := tr.GetRelevantGateways(resources)

		require.Len(t, accepted, 1)
		require.Empty(t, failed)
		require.ElementsMatch(t, []*egv1a1.ProxyBootstrap{gcBootstrap, gwBootstrap}, fake.calls)
	})

	t.Run("validator error rejects GatewayClass and Gateway", func(t *testing.T) {
		resources, _, _ := newBootstrapValidatorResources()
		fake := &fakeBootstrapValidator{err: errors.New("bad bootstrap")}
		tr := &Translator{
			GatewayClassName:   "test-gc",
			BootstrapValidator: fake,
			Logger:             logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
		}

		accepted, failed := tr.GetRelevantGateways(resources)

		require.Empty(t, accepted)
		require.Len(t, failed, 1)
		require.Len(t, fake.calls, 2)

		gcCond := findCondition(resources.GatewayClass.Status.Conditions, string(gwapiv1.GatewayClassConditionStatusAccepted))
		require.NotNil(t, gcCond)
		require.Equal(t, metav1.ConditionFalse, gcCond.Status)
		require.Contains(t, gcCond.Message, "bad bootstrap")

		gwCond := findCondition(failed[0].Status.Conditions, string(gwapiv1.GatewayConditionAccepted))
		require.NotNil(t, gwCond)
		require.Equal(t, metav1.ConditionFalse, gwCond.Status)
		require.Equal(t, string(gwapiv1.GatewayReasonInvalidParameters), gwCond.Reason)
		require.Contains(t, gwCond.Message, "bad bootstrap")
	})
}
