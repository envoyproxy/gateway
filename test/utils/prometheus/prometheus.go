// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package prometheus

import (
	"context"
	"fmt"
	"time"

	prom "github.com/prometheus/client_golang/api"
	prompapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	prom.Client

	address   string
	name      string
	namespace string
}

// NewClient returns a prometheus client based on the namespaced name of prometheus-server.
func NewClient(kubeClient client.Client, nn types.NamespacedName) (*Client, error) {
	svc := &corev1.Service{}
	if err := kubeClient.Get(context.Background(), nn, svc); err != nil {
		return nil, fmt.Errorf("failed to get service %s: %w", nn.String(), err)
	}

	var addr string
	for _, ing := range svc.Status.LoadBalancer.Ingress {
		if len(ing.IP) > 0 {
			addr = fmt.Sprintf("http://%s", ing.IP)
		}
	}

	if len(addr) == 0 {
		return nil, fmt.Errorf("no ingress found for %s", nn.String())
	}

	c, err := prom.NewClient(prom.Config{Address: addr})
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:    c,
		address:   addr,
		name:      nn.Name,
		namespace: nn.Namespace,
	}, nil
}

func (c *Client) RawQuery(ctx context.Context, promQL string) (model.Value, error) {
	v, _, err := prompapiv1.NewAPI(c.Client).Query(ctx, promQL, time.Now())
	if err != nil {
		return nil, err
	}

	switch v.Type() {
	case model.ValScalar, model.ValString:
		return v, nil
	case model.ValVector:
		value := v.(model.Vector)
		if len(value) == 0 {
			return nil, fmt.Errorf("value not found (query: %v)", promQL)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported value type: %v", v.Type())
	}
}

func (c *Client) QuerySum(ctx context.Context, promQL string) (float64, error) {
	val, err := c.RawQuery(ctx, promQL)
	if err != nil {
		return 0, err
	}

	got, err := sum(val)
	if err != nil {
		return 0, fmt.Errorf("could not find metric value: %w", err)
	}
	return got, nil
}

func (c *Client) QueryAvg(ctx context.Context, promQL string) (float64, error) {
	val, err := c.RawQuery(ctx, promQL)
	if err != nil {
		return 0, err
	}

	got, err := sum(val)
	if err != nil {
		return 0, fmt.Errorf("could not find metric value: %w", err)
	}

	got = got / float64(val.(model.Vector).Len())
	return got, nil
}

func sum(val model.Value) (float64, error) {
	if val.Type() != model.ValVector {
		return 0, fmt.Errorf("value not a model.Vector; was %s", val.Type().String())
	}

	value := val.(model.Vector)

	valueCount := 0.0
	for _, sample := range value {
		valueCount += float64(sample.Value)
	}

	if valueCount > 0.0 {
		return valueCount, nil
	}
	return 0, fmt.Errorf("value not found")
}
