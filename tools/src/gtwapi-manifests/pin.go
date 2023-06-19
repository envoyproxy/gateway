// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build pin
// +build pin

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"helm.sh/helm/v3/pkg/kube"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/cli-runtime/pkg/resource"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/yaml"
)

var (
	addToScheme sync.Once
	nopLogger   = func(_ string, _ ...interface{}) {}
)

func newClient() *kube.Client {
	addToScheme.Do(func() {
		if err := apiextv1.AddToScheme(scheme.Scheme); err != nil {
			// This should never happen.
			panic(err)
		}
		if err := apiextv1beta1.AddToScheme(scheme.Scheme); err != nil {
			panic(err)
		}
	})

	testFactory := cmdtesting.NewTestFactory()

	c := &kube.Client{
		Factory: testFactory.WithNamespace("default"),
		Log:     nopLogger,
	}

	return c
}

func main() {
	outputDir := env("OUTPUT_DIR", "charts/gateway-helm")
	gtwAPIVer := env("GATEWAY_API_VERSION", "v0.7.1")

	genGatewayAPI(outputDir, gtwAPIVer)
}

func genGatewayAPI(outputDir string, version string) {
	fmt.Printf("start to gen Gateway API resources, version: %s output: %s \n", version, outputDir)
	gtwAPIUrl := fmt.Sprintf("https://github.com/kubernetes-sigs/gateway-api/releases/download/%s/experimental-install.yaml", version)
	resp, err := http.Get(gtwAPIUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		resp.Body.Close()
	}()

	c := newClient()
	resources, err := c.Build(resp.Body, false)
	if err != nil {
		fmt.Printf("build helm fail: %v", err)
		os.Exit(-1)
	}

	writeCRDs(path.Join(outputDir, "crds"), resources)
	writeOthers(path.Join(outputDir, "templates", "gateway-api-admission"), resources)
}

func writeCRDs(outputDir string, resources kube.ResourceList) error {
	fileName := path.Join(outputDir, "gatewayapi-crds.yaml")
	if err := os.Remove(fileName); err != nil {
		return err
	}

	crds := resources.Filter(func(r *resource.Info) bool {
		// only need CRD
		return r.Mapping.GroupVersionKind.Kind == "CustomResourceDefinition"
	})

	crdYamls := make([]string, 0, len(crds))
	for _, r := range crds {
		out, _ := yaml.Marshal(r.Object)

		crdYamls = append(crdYamls, string(out))
	}

	out := strings.Join(crdYamls, "\n---\n")
	if err := os.WriteFile(fileName, []byte(out), 0o755); err != nil {
		return err
	}

	return nil
}

func writeOthers(outputDir string, resources kube.ResourceList) error {
	if err := os.RemoveAll(outputDir); err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	crds := resources.Filter(func(r *resource.Info) bool {
		return r.Mapping.GroupVersionKind.Kind != "CustomResourceDefinition"
	})

	for _, r := range crds {
		out, _ := yaml.Marshal(r.Object)
		n := path.Join(outputDir, fmt.Sprintf("%s_%s.yaml", r.Name, r.Mapping.GroupVersionKind.Kind))
		if err := os.WriteFile(n, out, 0o755); err != nil {
			return err
		}
	}

	return nil
}

func env(key, defaultVal string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	return v
}
