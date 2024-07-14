// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package e2e

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

//go:embed testdata/*.yaml base/*
var Manifests embed.FS

//go:embed testdata/*.yaml upgrade/*
var UpgradeManifests embed.FS

func CheckInstallScheme(t *testing.T, c client.Client) {
	require.NoError(t, gwapiv1a3.Install(c.Scheme()))
	require.NoError(t, gwapiv1a2.Install(c.Scheme()))
	require.NoError(t, gwapiv1b1.Install(c.Scheme()))
	require.NoError(t, gwapiv1.Install(c.Scheme()))
	require.NoError(t, egv1a1.AddToScheme(c.Scheme()))
}
