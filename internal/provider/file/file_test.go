// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"bytes"
	"context"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type resourcesParam struct {
	GatewayClassName    string
	GatewayName         string
	GatewayListenerPort string
	HTTPRouteName       string
	HTTPRouteHostname   string
	BackendName         string
	EndpointPort        string
}

func newResourcesParam1() *resourcesParam {
	return &resourcesParam{
		GatewayClassName:    "eg-1",
		GatewayName:         "eg-1",
		GatewayListenerPort: "8801",
		HTTPRouteName:       "backend-1",
		HTTPRouteHostname:   "www.test1.com",
		BackendName:         "backend-1",
		EndpointPort:        "3001",
	}
}

func newResourcesParam2() *resourcesParam {
	return &resourcesParam{
		GatewayClassName:    "eg-2",
		GatewayName:         "eg-2",
		GatewayListenerPort: "8802",
		HTTPRouteName:       "backend-2",
		HTTPRouteHostname:   "www.test2.com",
		BackendName:         "backend-2",
		EndpointPort:        "3002",
	}
}

func newFileProviderConfig(paths []string) (*config.Server, error) {
	cfg, err := config.New(os.Stdout)
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
		},
	}
	return cfg, nil
}

func TestFileProvider(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	watchFileBase, _ := os.MkdirTemp(os.TempDir(), "test-files-*")
	watchFilePath := filepath.Join(watchFileBase, "test.yaml")
	watchDirPath, _ := os.MkdirTemp(os.TempDir(), "test-dir-*")
	// Prepare the watched test file.
	writeResourcesFile(t, watchFilePath, newResourcesParam1())
	require.FileExists(t, watchFilePath)
	require.DirExists(t, watchDirPath)

	cfg, err := newFileProviderConfig([]string{watchFilePath, watchDirPath})
	require.NoError(t, err)
	pResources := new(message.ProviderResources)
	fp, err := New(ctx, cfg, pResources)
	require.NoError(t, err)
	// Start file provider.
	go func() {
		if err := fp.Start(ctx); err != nil {
			t.Errorf("failed to start file provider: %v", err)
		}
	}()

	// Wait for file provider to be ready.
	waitFileProviderReady(t)

	require.Equal(t, "gateway.envoyproxy.io/gatewayclass-controller", fp.store.name)

	t.Run("initial resource load", func(t *testing.T) {
		// Wait for the first reconcile to kick in.
		require.Eventually(t, func() bool {
			return pResources.GatewayAPIResources.Len() > 0
		}, resourcesUpdateTimeout, resourcesUpdateTick)
		resources := pResources.GetResourcesByGatewayClass("eg-1")
		require.NotNil(t, resources)

		want := &resource.Resources{}
		mustUnmarshal(t, "testdata/resources.1.yaml", want)
		cmpResources(t, want, resources)
	})

	t.Run("rename the watched file then rename it back", func(t *testing.T) {
		// Rename it first, the watched file is losed.
		renameFilePath := filepath.Join(watchFileBase, "foobar.yaml")
		err := os.Rename(watchFilePath, renameFilePath)
		require.NoError(t, err)
		require.Eventually(t, func() bool {
			return pResources.GetResourcesByGatewayClass("eg-1") == nil
		}, resourcesUpdateTimeout, resourcesUpdateTick)

		// Rename it back, the watched file is resumed.
		err = os.Rename(renameFilePath, watchFilePath)
		require.NoError(t, err)
		require.Eventually(t, func() bool {
			return pResources.GetResourcesByGatewayClass("eg-1") != nil
		}, resourcesUpdateTimeout, resourcesUpdateTick)

		resources := pResources.GetResourcesByGatewayClass("eg-1")
		want := &resource.Resources{}
		mustUnmarshal(t, "testdata/resources.1.yaml", want)
		cmpResources(t, want, resources)
	})

	t.Run("remove the watched file", func(t *testing.T) {
		require.NoError(t, os.Remove(watchFilePath))
		require.Eventually(t, func() bool {
			return len(pResources.GetResources()) == 0
		}, resourcesUpdateTimeout, resourcesUpdateTick)
	})

	t.Run("add a new file in watched dir", func(t *testing.T) {
		// Write a new file under empty watched directory.
		newFilePath := filepath.Join(watchDirPath, "test.yaml")
		writeResourcesFile(t, newFilePath, newResourcesParam1())

		require.Eventually(t, func() bool {
			return pResources.GetResourcesByGatewayClass("eg-1") != nil
		}, resourcesUpdateTimeout, resourcesUpdateTick)

		resources := pResources.GetResourcesByGatewayClass("eg-1")
		want := &resource.Resources{}
		mustUnmarshal(t, "testdata/resources.1.yaml", want)
		cmpResources(t, want, resources)
	})

	t.Run("rename the file then rename it back in watched dir", func(t *testing.T) {
		// Rename it first.
		srcFilePath := filepath.Join(watchDirPath, "test.yaml")
		dstFilePath := filepath.Join(watchDirPath, "foobar.yaml")
		err := os.Rename(srcFilePath, dstFilePath)
		require.NoError(t, err)
		require.Eventually(t, func() bool {
			return pResources.GetResourcesByGatewayClass("eg-1") != nil
		}, resourcesUpdateTimeout, resourcesUpdateTick)

		// Rename it back.
		err = os.Rename(dstFilePath, srcFilePath)
		require.NoError(t, err)
		require.Eventually(t, func() bool {
			return pResources.GetResourcesByGatewayClass("eg-1") != nil
		}, resourcesUpdateTimeout, resourcesUpdateTick)

		resources := pResources.GetResourcesByGatewayClass("eg-1")
		want := &resource.Resources{}
		mustUnmarshal(t, "testdata/resources.1.yaml", want)
		cmpResources(t, want, resources)
	})

	t.Run("update file content in watched dir", func(t *testing.T) {
		// Rewrite the file under watched directory.
		newFilePath := filepath.Join(watchDirPath, "test.yaml")
		writeResourcesFile(t, newFilePath, newResourcesParam2())

		require.Eventually(t, func() bool {
			return pResources.GetResourcesByGatewayClass("eg-1") == nil &&
				pResources.GetResourcesByGatewayClass("eg-2") != nil
		}, resourcesUpdateTimeout, resourcesUpdateTick)
	})

	t.Run("add another file with new gatewayclass in watched dir", func(t *testing.T) {
		// The test.yaml was changed by previous case, safe to use resources param 1 here.
		newFilePath := filepath.Join(watchDirPath, "another.yaml")
		writeResourcesFile(t, newFilePath, newResourcesParam1())

		require.Eventually(t, func() bool {
			return pResources.GetResourcesByGatewayClass("eg-1") != nil &&
				pResources.GetResourcesByGatewayClass("eg-2") != nil
		}, resourcesUpdateTimeout, resourcesUpdateTick)

		resources1 := pResources.GetResourcesByGatewayClass("eg-1")
		want1 := &resource.Resources{}
		mustUnmarshal(t, "testdata/resources.1.yaml", want1)
		cmpResources(t, want1, resources1)

		resources2 := pResources.GetResourcesByGatewayClass("eg-2")
		want2 := &resource.Resources{}
		mustUnmarshal(t, "testdata/resources.2.yaml", want2)
		cmpResources(t, want2, resources2)
	})

	t.Run("remove all files in watched dir", func(t *testing.T) {
		fp1 := filepath.Join(watchDirPath, "test.yaml")
		fp2 := filepath.Join(watchDirPath, "another.yaml")
		require.NoError(t, os.Remove(fp1))
		require.NoError(t, os.Remove(fp2))
		require.Eventually(t, func() bool {
			return len(pResources.GetResources()) == 0
		}, resourcesUpdateTimeout, resourcesUpdateTick)
	})

	t.Cleanup(func() {
		cancel()
		_ = os.RemoveAll(watchFileBase)
		_ = os.RemoveAll(watchDirPath)
	})
}

func writeResourcesFile(t *testing.T, dst string, params *resourcesParam) {
	var buf bytes.Buffer

	// Write parameters into target file.
	tmplFile, err := template.ParseFiles("testdata/resources.tmpl")
	require.NoError(t, err)

	err = tmplFile.Execute(&buf, params)
	require.NoError(t, err)
	// Write file in an atomic way, prevent unnecessary reconcile.
	require.NoError(t, os.WriteFile(dst, buf.Bytes(), 0o600))
}

func waitFileProviderReady(t *testing.T) {
	require.Eventually(t, func() bool {
		resp, err := http.Get("http://localhost:8081/readyz")
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

func cmpResources(t *testing.T, x, y interface{}) {
	opts := []cmp.Option{
		cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion"),
		cmpopts.EquateEmpty(),
	}
	require.Empty(t, cmp.Diff(x, y, opts...))
}
