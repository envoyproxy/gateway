// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build celvalidation
// +build celvalidation

package celvalidation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var c client.Client

func TestMain(m *testing.M) {
	// Setup the test environment.
	testEnv, restCfg, err := startEnv()
	if err != nil {
		panic(fmt.Sprintf("Failed to start testenv: %v", err))
	}

	_, cancel := context.WithCancel(ctrl.SetupSignalHandler())
	defer func() {
		cancel()
		if err := testEnv.Stop(); err != nil {
			panic(fmt.Sprintf("Failed to stop testenv: %v", err))
		}
	}()

	c, err = client.New(restCfg, client.Options{})
	if err != nil {
		panic(fmt.Sprintf("Error initializing client: %v", err))
	}
	_ = egv1a1.AddToScheme(c.Scheme())

	os.Exit(m.Run())
}

func startEnv() (*envtest.Environment, *rest.Config, error) {
	log.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))
	egAPIs := filepath.Join("..", "..", "charts", "gateway-helm", "crds", "generated")

	env := &envtest.Environment{
		CRDDirectoryPaths: []string{egAPIs},
	}
	cfg, err := env.Start()
	if err != nil {
		return env, nil, err
	}
	return env, cfg, nil
}
