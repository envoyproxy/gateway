// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build validation
// +build validation

package validation

import (
	"fmt"
	"os"
	"path"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var c client.Client

func TestMain(m *testing.M) {
	kc := os.Getenv("KUBECONFIG")
	if kc == "" {
		kc = path.Join(os.Getenv("HOME"), ".kube/config")
	}

	rest, err := clientcmd.BuildConfigFromFlags("", kc)
	if err != nil {
		panic(fmt.Sprintf("Failed to build config from BuildConfigFromFlags: %v", err))
	}

	c, err = client.New(rest, client.Options{})
	if err != nil {
		panic(fmt.Sprintf("Error initializing client: %v", err))
	}
	_ = egv1a1.AddToScheme(c.Scheme())

	os.Exit(m.Run())
}
