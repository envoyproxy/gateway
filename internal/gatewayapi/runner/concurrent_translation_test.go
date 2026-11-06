// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/telepresenceio/watchable"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
)

type gatewayTranslationGate struct {
	started chan string
	gates   map[string]<-chan struct{}
}

func (g *gatewayTranslationGate) OnStart(_ context.Context, span sdktrace.ReadWriteSpan) {
	if span.Name() != "GatewayApiRunner.ResoureTranslationCycle.TranslateGatewayClass" {
		return
	}
	for _, attr := range span.Attributes() {
		if attr.Key == "gatewayclass.name" {
			name := attr.Value.AsString()
			g.started <- name
			if gate := g.gates[name]; gate != nil {
				<-gate
			}
			return
		}
	}
}

func (*gatewayTranslationGate) OnEnd(sdktrace.ReadOnlySpan)      {}
func (*gatewayTranslationGate) Shutdown(context.Context) error   { return nil }
func (*gatewayTranslationGate) ForceFlush(context.Context) error { return nil }

func translationTestRunner(t *testing.T) *Runner {
	t.Helper()
	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)
	return New(&Config{
		Server: *cfg, ProviderResources: new(message.ProviderResources),
		XdsIR: new(message.XdsIR), InfraIR: new(message.InfraIR), RunnerErrors: new(message.RunnerErrors),
	})
}

func translationTestResources(name string) *resource.Resources {
	resources := resource.NewResources()
	resources.GatewayClass = &gwapiv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       gwapiv1.GatewayClassSpec{ControllerName: egv1a1.GatewayControllerName},
	}
	resources.Namespaces = []*corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "default"}}}
	resources.Gateways = []*gwapiv1.Gateway{{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: name, UID: types.UID(name)},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: gwapiv1.ObjectName(name),
			Listeners:        []gwapiv1.Listener{{Name: "http", Protocol: gwapiv1.HTTPProtocolType, Port: 8080}},
		},
	}}
	return resources
}

func receiveGatewayTranslation[T any](t *testing.T, values <-chan T) T {
	t.Helper()
	select {
	case value := <-values:
		return value
	case <-time.After(5 * time.Second):
		t.Fatal("GatewayClass translation did not complete")
		var zero T
		return zero
	}
}

func TestConcurrentGatewayClassesPublishBeforeBatchCompletes(t *testing.T) {
	slowGate, otherGates := make(chan struct{}), make(chan struct{})
	releaseSlow := sync.OnceFunc(func() { close(slowGate) })
	releaseOthers := sync.OnceFunc(func() { close(otherGates) })
	gate := &gatewayTranslationGate{
		started: make(chan string, gatewayClassConcurrency+1),
		gates:   make(map[string]<-chan struct{}),
	}
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(gate))
	previousTracer := tracer
	tracer = provider.Tracer("gateway-concurrency-test")
	t.Cleanup(func() {
		tracer = previousTracer
		require.NoError(t, provider.Shutdown(context.Background()))
	})
	runner := translationTestRunner(t)
	resources := make(resource.ControllerResources, gatewayClassConcurrency+1)
	sharedRoute := &gwapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "shared", UID: "shared"},
		Spec: gwapiv1.HTTPRouteSpec{
			Rules: []gwapiv1.HTTPRouteRule{{Filters: []gwapiv1.HTTPRouteFilter{{
				Type:            gwapiv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gwapiv1.HTTPRequestRedirectFilter{Hostname: new(gwapiv1.PreciseHostname("example.com"))},
			}}}},
		},
	}
	names := make([]string, len(resources))
	for i := range resources {
		name := fmt.Sprintf("class-%d", i)
		names[i] = name
		resources[i] = translationTestResources(name)
		resources[i].HTTPRoutes = []*gwapiv1.HTTPRoute{sharedRoute}
		sharedRoute.Spec.ParentRefs = append(sharedRoute.Spec.ParentRefs, gwapiv1.ParentReference{Name: gwapiv1.ObjectName(name)})
		if i == 0 {
			gate.gates[name] = slowGate
		} else if i < gatewayClassConcurrency {
			gate.gates[name] = otherGates
		}
	}
	input := make(chan watchable.Snapshot[string, *resource.ControllerResourcesContext], 1)
	input <- watchable.Snapshot[string, *resource.ControllerResourcesContext]{State: map[string]*resource.ControllerResourcesContext{
		"controller": {Resources: &resources},
	}}
	close(input)
	done := make(chan struct{})
	go func() {
		runner.subscribeAndTranslate(input)
		close(done)
	}()
	t.Cleanup(func() {
		releaseSlow()
		releaseOthers()
		receiveGatewayTranslation(t, done)
	})
	started := make([]string, gatewayClassConcurrency)
	for i := range started {
		started[i] = receiveGatewayTranslation(t, gate.started)
	}
	require.ElementsMatch(t, names[:gatewayClassConcurrency], started)
	select {
	case name := <-gate.started:
		t.Fatalf("another class started while all workers were blocked: %s", name)
	case <-time.After(50 * time.Millisecond):
	}
	releaseOthers()
	require.Equal(t, names[gatewayClassConcurrency], receiveGatewayTranslation(t, gate.started))
	require.Eventually(t, func() bool {
		for _, name := range names[1:] {
			if _, found := runner.XdsIR.Load("default/" + name); !found {
				return false
			}
		}
		return true
	}, 5*time.Second, time.Millisecond)
	_, found := runner.XdsIR.Load("default/" + names[0])
	require.False(t, found)
	select {
	case <-done:
		t.Fatal("runner returned while the first class was still translating")
	default:
	}
	releaseSlow()
	receiveGatewayTranslation(t, done)
	require.Len(t, runner.XdsIR.LoadAll(), len(resources))
	status, found := runner.ProviderResources.HTTPRouteStatuses.Load(types.NamespacedName{Namespace: "default", Name: "shared"})
	require.True(t, found)
	parents := make([]string, 0, len(status.Parents))
	for _, parent := range status.Parents {
		parents = append(parents, string(parent.ParentRef.Name))
	}
	require.Equal(t, names, parents)
	require.Empty(t, sharedRoute.Status.Parents, "workers must not mutate shared input statuses")
}

func TestGatewayClassPanicDoesNotSweepExistingIR(t *testing.T) {
	for _, phase := range []string{"translation", "publication"} {
		t.Run(phase, func(t *testing.T) {
			runner := translationTestRunner(t)
			runner.XdsIR.Store("default/private", &message.XdsIRWithContext{XdsIR: &ir.Xds{}})
			runner.keyCache.IR["default/private"] = true
			private := translationTestResources("private")
			if phase == "translation" {
				private.Gateways = []*gwapiv1.Gateway{nil}
			} else {
				runner.InfraIR.Close()
			}
			resources := resource.ControllerResources{private, translationTestResources("edge")}
			input := make(chan watchable.Snapshot[string, *resource.ControllerResourcesContext], 1)
			input <- watchable.Snapshot[string, *resource.ControllerResourcesContext]{State: map[string]*resource.ControllerResourcesContext{
				"controller": {Resources: &resources},
			}}
			close(input)
			done := make(chan struct{})
			go func() {
				runner.subscribeAndTranslate(input)
				close(done)
			}()
			receiveGatewayTranslation(t, done)
			_, found := runner.XdsIR.Load("default/private")
			require.True(t, found, "a failed class must retain its existing configuration")
			_, found = runner.XdsIR.Load("default/edge")
			require.Equal(t, phase == "translation", found, "an independent class must publish unless publication itself fails")
		})
	}
}
