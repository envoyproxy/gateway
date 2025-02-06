// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/message"
)

const (
	resourcesUpdateTimeout = 1 * time.Minute
	resourcesUpdateTick    = 1 * time.Second
)

var healthzServerPort = 8082

type resourcesParam struct {
	GatewayClassName    string
	GatewayName         string
	GatewayListenerPort string
	HTTPRouteName       string
	BackendName         string
}

func newDefaultResourcesParam() *resourcesParam {
	return &resourcesParam{
		GatewayClassName:    "eg",
		GatewayName:         "eg",
		GatewayListenerPort: "8888",
		HTTPRouteName:       "backend",
		BackendName:         "backend",
	}
}

func newFileProviderConfig(paths []string) (*config.Server, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
		Type: egv1a1.ProviderTypeCustom,
		Custom: &egv1a1.EnvoyGatewayCustomProvider{
			Resource: egv1a1.EnvoyGatewayResourceProvider{
				Type: egv1a1.ResourceProviderTypeFile,
				File: &egv1a1.EnvoyGatewayFileResourceProvider{
					Paths: paths,
				},
			},
			HealthzServerPort: &healthzServerPort,
		},
	}
	return cfg, nil
}

func TestFileProvider(t *testing.T) {
	watchFileBase, _ := os.MkdirTemp(os.TempDir(), "test-files-*")
	watchFilePath := filepath.Join(watchFileBase, "test.yaml")
	watchDirPath, _ := os.MkdirTemp(os.TempDir(), "test-dir-*")
	// Prepare the watched test file.
	writeResourcesFile(t, "testdata/resources.tmpl", watchFilePath, newDefaultResourcesParam())
	require.FileExists(t, watchFilePath)
	require.DirExists(t, watchDirPath)

	cfg, err := newFileProviderConfig([]string{watchFilePath, watchDirPath})
	require.NoError(t, err)
	pResources := new(message.ProviderResources)
	fp, err := New(cfg, pResources)
	require.NoError(t, err)
	// Start file provider.
	go func() {
		if err := fp.Start(context.Background()); err != nil {
			t.Errorf("failed to start file provider: %v", err)
		}
	}()

	// Wait for file provider to be ready.
	waitFileProviderReady(t)

	require.Equal(t, "gateway.envoyproxy.io/gatewayclass-controller", fp.resourcesStore.name)

	t.Run("initial resource load", func(t *testing.T) {
		require.NotZero(t, pResources.GatewayAPIResources.Len())
		resources := pResources.GetResourcesByGatewayClass("eg")
		require.NotNil(t, resources)

		want := &resource.Resources{}
		mustUnmarshal(t, "testdata/resources.all.yaml", want)

		opts := []cmp.Option{
			cmpopts.IgnoreFields(resource.Resources{}, "serviceMap"),
			cmpopts.EquateEmpty(),
		}
		require.Empty(t, cmp.Diff(want, resources, opts...))
	})

	t.Run("rename the watched file then rename it back", func(t *testing.T) {
		// Rename it
		renameFilePath := filepath.Join(watchFileBase, "foobar.yaml")
		err := os.Rename(watchFilePath, renameFilePath)
		require.NoError(t, err)
		require.Eventually(t, func() bool {
			return pResources.GetResourcesByGatewayClass("eg") == nil
		}, resourcesUpdateTimeout, resourcesUpdateTick)

		// Rename it back
		err = os.Rename(renameFilePath, watchFilePath)
		require.NoError(t, err)
		require.Eventually(t, func() bool {
			return pResources.GetResourcesByGatewayClass("eg") != nil
		}, resourcesUpdateTimeout, resourcesUpdateTick)

		resources := pResources.GetResourcesByGatewayClass("eg")
		want := &resource.Resources{}
		mustUnmarshal(t, "testdata/resources.all.yaml", want)

		opts := []cmp.Option{
			cmpopts.IgnoreFields(resource.Resources{}, "serviceMap"),
			cmpopts.EquateEmpty(),
		}
		require.Empty(t, cmp.Diff(want, resources, opts...))
	})

	t.Run("remove the watched file", func(t *testing.T) {
		err := os.Remove(watchFilePath)
		require.NoError(t, err)
		require.Eventually(t, func() bool {
			return pResources.GetResourcesByGatewayClass("eg") == nil
		}, resourcesUpdateTimeout, resourcesUpdateTick)
	})

	t.Run("add a file in watched dir", func(t *testing.T) {
		// Write a new file under watched directory.
		newFilePath := filepath.Join(watchDirPath, "test.yaml")
		writeResourcesFile(t, "testdata/resources.tmpl", newFilePath, newDefaultResourcesParam())

		require.Eventually(t, func() bool {
			return pResources.GetResourcesByGatewayClass("eg") != nil
		}, resourcesUpdateTimeout, resourcesUpdateTick)

		resources := pResources.GetResourcesByGatewayClass("eg")
		want := &resource.Resources{}
		mustUnmarshal(t, "testdata/resources.all.yaml", want)

		opts := []cmp.Option{
			cmpopts.IgnoreFields(resource.Resources{}, "serviceMap"),
			cmpopts.EquateEmpty(),
		}
		require.Empty(t, cmp.Diff(want, resources, opts...))
	})

	t.Run("remove a file in watched dir", func(t *testing.T) {
		newFilePath := filepath.Join(watchDirPath, "test.yaml")
		err := os.Remove(newFilePath)
		require.NoError(t, err)
		require.Eventually(t, func() bool {
			return pResources.GetResourcesByGatewayClass("eg") == nil
		}, resourcesUpdateTimeout, resourcesUpdateTick)
	})

	t.Cleanup(func() {
		_ = os.RemoveAll(watchFileBase)
		_ = os.RemoveAll(watchDirPath)
	})
}

func writeResourcesFile(t *testing.T, tmpl, dst string, params *resourcesParam) {
	dstFile, err := os.Create(dst)
	require.NoError(t, err)

	// Write parameters into target file.
	tmplFile, err := template.ParseFiles(tmpl)
	require.NoError(t, err)

	err = tmplFile.Execute(dstFile, params)
	require.NoError(t, err)
	require.NoError(t, dstFile.Close())
}

func waitFileProviderReady(t *testing.T) {
	require.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/readyz", healthzServerPort))
		if err != nil {
			t.Logf("failed to get from heathlz server")
			return false
		}

		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			t.Logf("failed to get body from response")
			return false
		}

		if string(body) != "ok" {
			t.Logf("the file provider is not ready yet")
			return false
		}
		return true
	}, 3*resourcesUpdateTimeout, resourcesUpdateTick)
}

func mustUnmarshal(t *testing.T, path string, out interface{}) {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, yaml.UnmarshalStrict(content, out, yaml.DisallowUnknownFields))
}
