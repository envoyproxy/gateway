// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package prometheus

import (
	"context"
	"time"

	promapi "github.com/prometheus/client_golang/api"
	promapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/test/utils"
)

func QueryPrometheus(c client.Client, promQL string) (model.Value, error) {
	promNN := types.NamespacedName{Namespace: "monitoring", Name: "prometheus"}
	address, err := utils.RetrieveURL(c, promNN, 80, "/")
	if err != nil {
		return nil, err
	}

	promClient, err := promapi.NewClient(promapi.Config{Address: address})
	if err != nil {
		return nil, err
	}

	promAPIClient := promapiv1.NewAPI(promClient)

	v, _, err := promAPIClient.Query(context.TODO(), promQL, time.Now())
	if err != nil {
		return nil, err
	}

	return v, err
}
