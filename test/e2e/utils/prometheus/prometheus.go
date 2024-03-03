// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

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

func Address(c client.Client, nn types.NamespacedName) (string, error) {
	svc := &corev1.Service{}
	if err := c.Get(context.TODO(), nn, svc); err != nil {
		return "", fmt.Errorf("failed to get service: %w", err)
	}
	for _, ing := range svc.Status.LoadBalancer.Ingress {
		if ing.IP != "" {
			return fmt.Sprintf("http://%s", ing.IP), nil
		}
	}

	return "", fmt.Errorf("no ingress found")
}

func RawQuery(address string, promQL string) (model.Value, error) {
	c, err := prom.NewClient(prom.Config{Address: address})
	if err != nil {
		return nil, err
	}

	v, _, err := prompapiv1.NewAPI(c).Query(context.Background(), promQL, time.Now())
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
		return nil, fmt.Errorf("unhandled value type: %v", v.Type())
	}
}

func QuerySum(address string, promQL string) (float64, error) {
	val, err := RawQuery(address, promQL)
	if err != nil {
		return 0, err
	}
	got, err := sum(val)
	if err != nil {
		return 0, fmt.Errorf("could not find metric value: %w", err)
	}
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
