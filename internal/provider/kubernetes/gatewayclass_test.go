// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/log"
)

func TestGatewayClassHasMatchingController(t *testing.T) {
	testCases := []struct {
		name   string
		obj    client.Object
		expect bool
	}{
		{
			name: "matching controller name",
			obj: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: v1alpha1.GatewayControllerName,
				},
			},
			expect: true,
		},
		{
			name: "non-matching controller name",
			obj: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: "not.configured/controller",
				},
			},
			expect: false,
		},
	}

	// Create the reconciler.
	logger, err := log.NewLogger()
	require.NoError(t, err)
	r := gatewayAPIReconciler{
		classController: v1alpha1.GatewayControllerName,
		log:             logger,
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res := r.hasMatchingController(tc.obj)
			require.Equal(t, tc.expect, res)
		})
	}
}

func TestGatewayOldestClass(t *testing.T) {
	createGatewayClass := func(name string, creationTime time.Time) *gwapiv1b1.GatewayClass {
		return &gwapiv1b1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				CreationTimestamp: metav1.NewTime(creationTime),
			},
			Spec: gwapiv1b1.GatewayClassSpec{
				ControllerName: v1alpha1.GatewayControllerName,
			},
		}
	}

	currentTime := metav1.Now()
	addDuration := time.Duration(10)
	testCases := []struct {
		name    string
		classes map[string]time.Time
		remove  map[string]time.Time
		oldest  string
	}{
		{
			name: "normal",
			classes: map[string]time.Time{
				"class-b": currentTime.Time,
				"class-a": currentTime.Add(1 * addDuration),
			},
			remove: nil,
			oldest: "class-b",
		},
		{
			name: "tie breaker",
			classes: map[string]time.Time{
				"class-aa": currentTime.Time,
				"class-ab": currentTime.Time,
			},
			remove: nil,
			oldest: "class-aa",
		},
		{
			name: "remove from matched",
			classes: map[string]time.Time{
				"class-a": currentTime.Time,
				"class-b": currentTime.Add(1 * addDuration),
				"class-c": currentTime.Add(2 * addDuration),
			},
			remove: map[string]time.Time{
				"class-b": currentTime.Add(1 * addDuration),
			},
			oldest: "class-a",
		},
		{
			name: "remove oldest",
			classes: map[string]time.Time{
				"class-a": currentTime.Time,
				"class-b": currentTime.Add(1 * addDuration),
				"class-c": currentTime.Add(2 * addDuration),
			},
			remove: map[string]time.Time{
				"class-a": currentTime.Time,
			},
			oldest: "class-b",
		},
		{
			name: "remove oldest last",
			classes: map[string]time.Time{
				"class-a": currentTime.Time,
			},
			remove: map[string]time.Time{
				"class-a": currentTime.Time,
			},
			oldest: "",
		},
	}

	for _, tc := range testCases {
		var cc controlledClasses
		for name, timestamp := range tc.classes {
			cc.addMatch(createGatewayClass(name, timestamp))
		}

		for name, timestamp := range tc.remove {
			cc.removeMatch(createGatewayClass(name, timestamp))
		}

		if tc.oldest == "" {
			require.Nil(t, cc.oldestClass)
			return
		}

		require.Equal(t, tc.oldest, cc.oldestClass.Name)
	}
}
