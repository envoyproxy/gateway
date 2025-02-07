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

type resourcesParam struct {
	GatewayClassName    string
	GatewayName         string
	GatewayListenerPort string
	HTTPRouteName       string
	BackendName         string
}

func newDefaultResourcesParam(name string) *resourcesParam {
	return &resourcesParam{
		GatewayClassName:    name,
		GatewayName:         name,
		GatewayListenerPort: "8888",
		HTTPRouteName:       fmt.Sprintf("backend-%s", name),
		BackendName:         fmt.Sprintf("backend-%s", name),
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
		},
	}
	return cfg, nil
}

func TestFileProvider(t *testing.T) {
	watchFileBase, _ := os.MkdirTemp(os.TempDir(), "test-files-*")
	watchFilePath := filepath.Join(watchFileBase, "test.yaml")
	watchDirPath, _ := os.MkdirTemp(os.TempDir(), "test-dir-*")

	// Prepare the watched test file.
	writeResourcesFile(t, watchFilePath, newDefaultResourcesParam("eg"))
	require.FileExists(t, watchFilePath)
	require.DirExists(t, watchDirPath)

	cfg, err := newFileProviderConfig([]string{watchFilePath, watchDirPath})
	require.NoError(t, err)
	pResources := new(message.ProviderResources)
	fp, err := New(cfg, pResources)
	require.NoError(t, err)

	// Start file provider.
	ctx, cancelFpContext := context.WithCancel(context.Background())
	go func() {
		if err := fp.Start(ctx); err != nil {
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
		mustUnmarshal(t, "eg", want)

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
		expectResourcesToBeAbsent(t, pResources, "eg")

		// Rename it back
		err = os.Rename(renameFilePath, watchFilePath)
		require.NoError(t, err)
		expectResourcesToBePresent(t, pResources, "eg")

		resources := pResources.GetResourcesByGatewayClass("eg")
		want := &resource.Resources{}
		mustUnmarshal(t, "eg", want)

		opts := []cmp.Option{
			cmpopts.IgnoreFields(resource.Resources{}, "serviceMap"),
			cmpopts.EquateEmpty(),
		}
		require.Empty(t, cmp.Diff(want, resources, opts...))
	})

	t.Run("remove the watched file", func(t *testing.T) {
		err := os.Remove(watchFilePath)
		require.NoError(t, err)
		expectResourcesToBeAbsent(t, pResources, "eg")
	})

	t.Run("add a file in watched dir", func(t *testing.T) {
		// Write a new file under watched directory.
		newFilePath := filepath.Join(watchDirPath, "test.yaml")
		writeResourcesFile(t, newFilePath, newDefaultResourcesParam("eg"))
		expectResourcesToBePresent(t, pResources, "eg")

		resources := pResources.GetResourcesByGatewayClass("eg")
		want := &resource.Resources{}
		mustUnmarshal(t, "eg", want)

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
		expectResourcesToBeAbsent(t, pResources, "eg")
	})

	t.Cleanup(func() {
		cancelFpContext()
		err := os.RemoveAll(watchFileBase)
		require.NoError(t, err)
		err = os.RemoveAll(watchDirPath)
		require.NoError(t, err)
	})
}

func TestRecursiveFileProvider(t *testing.T) {
	// Create nested temporary directories with the following structure:
	// |- baseDir/
	//    | config_base.yaml
	//    |- subDir1/
	//       |- config_1.yaml
	// 	  |- subDir2/
	// 	     |- config_2.yaml
	baseDir, _ := os.MkdirTemp(os.TempDir(), "test-base-*")
	subDir1, _ := os.MkdirTemp(baseDir, "subdir-1")
	subDir2, _ := os.MkdirTemp(baseDir, "subdir-2")
	require.DirExists(t, baseDir)
	require.DirExists(t, subDir1)
	require.DirExists(t, subDir2)

	// Create the watched files.
	configBase := filepath.Join(baseDir, "config_base.yaml")
	writeResourcesFile(t, configBase, newDefaultResourcesParam("eg"))
	configSubDir1 := filepath.Join(subDir1, "config_1.yaml")
	writeResourcesFile(t, configSubDir1, newDefaultResourcesParam("eg1"))
	configSubDir2 := filepath.Join(subDir2, "config_2.yaml")
	writeResourcesFile(t, configSubDir2, newDefaultResourcesParam("eg2"))

	// Define locations for move tests
	moveLocation1 := filepath.Join(baseDir, "new_config.yaml")
	unwatchedDir, err := os.MkdirTemp(os.TempDir(), "unwatched-dir-*")
	require.NoError(t, err)
	moveLocation2 := filepath.Join(unwatchedDir, "new_config.yaml")

	require.FileExists(t, configBase)
	require.FileExists(t, configSubDir1)
	require.FileExists(t, configSubDir2)

	// Create the file provider configuration.
	cfg, err := newFileProviderConfig([]string{baseDir})
	require.NoError(t, err)
	pResources := new(message.ProviderResources)
	fp, err := New(cfg, pResources)
	require.NoError(t, err)

	// Start file provider.
	ctx, cancelFpContext := context.WithCancel(context.Background())
	go func() {
		if err := fp.Start(ctx); err != nil {
			t.Errorf("failed to start file provider: %v", err)
		}
	}()

	// Wait for file provider to be ready.
	waitFileProviderReady(t)
	require.Equal(t, "gateway.envoyproxy.io/gatewayclass-controller", fp.resourcesStore.name)

	t.Run("initial resource load", func(t *testing.T) {
		require.NotZero(t, pResources.GatewayAPIResources.Len())

		testClasses := []string{"eg", "eg1", "eg2"}
		for _, className := range testClasses {
			resources := pResources.GetResourcesByGatewayClass(className)
			require.NotNil(t, resources)

			want := &resource.Resources{}
			mustUnmarshal(t, className, want)

			opts := []cmp.Option{
				cmpopts.IgnoreFields(resource.Resources{}, "serviceMap"),
				cmpopts.EquateEmpty(),
			}
			require.Empty(t, cmp.Diff(want, resources, opts...))
		}
	})

	t.Run("rename a file in watched sub dir and then rename it back", func(t *testing.T) {
		// Rename the file
		renameFilePath := filepath.Join(subDir1, "config_1_renamed.yaml")
		err := os.Rename(configSubDir1, renameFilePath)
		require.NoError(t, err)
		expectResourcesToBePresent(t, pResources, "eg", "eg1", "eg2")

		// Rename the file back
		err = os.Rename(renameFilePath, configSubDir1)
		require.NoError(t, err)
		expectResourcesToBePresent(t, pResources, "eg", "eg1", "eg2")
	})

	t.Run("remove a file from watched sub dir", func(t *testing.T) {
		err := os.Remove(configSubDir2)
		require.NoError(t, err)

		expectResourcesToBeAbsent(t, pResources, "eg2")
		expectResourcesToBePresent(t, pResources, "eg", "eg1")
	})

	t.Run("add a new file to a watched sub dir", func(t *testing.T) {
		writeResourcesFile(t, configSubDir2, newDefaultResourcesParam("eg2"))
		expectResourcesToBePresent(t, pResources, "eg", "eg1", "eg2")
	})

	t.Run("move a file from one subdirectory to other subdirectory", func(t *testing.T) {
		moveFile(t, configSubDir2, moveLocation1)
		expectResourcesToBePresent(t, pResources, "eg", "eg1", "eg2")
	})

	t.Run("move a file from one subdirectory to an unwatched subdirectory", func(t *testing.T) {
		moveFile(t, moveLocation1, moveLocation2)
		expectResourcesToBePresent(t, pResources, "eg", "eg1")
		expectResourcesToBeAbsent(t, pResources, "eg2")
	})

	t.Cleanup(func() {
		cancelFpContext()
		err := os.RemoveAll(baseDir)
		require.NoError(t, err)
		err = os.RemoveAll(unwatchedDir)
		require.NoError(t, err)
	})
}

func writeResourcesFile(t *testing.T, dst string, params *resourcesParam) {
	t.Helper()

	dstFile, err := os.Create(dst)
	require.NoError(t, err)

	// Write parameters into target file.
	tmplFile, err := template.ParseFiles("testdata/resources.tmpl")
	require.NoError(t, err)

	err = tmplFile.Execute(dstFile, params)
	require.NoError(t, err)
	require.NoError(t, dstFile.Close())
}

func waitFileProviderReady(t *testing.T) {
	t.Helper()

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

// Generates the expected YAML for resources based on `testdata/resources.all.tmpl`
// and ensures that `out` unmarshalls to the same configuration.
func mustUnmarshal(t *testing.T, className string, out any) {
	t.Helper()

	// Create a temporary file to write the expected configuration.
	dstFile, err := os.CreateTemp("", "expected-*")
	require.NoError(t, err)
	defer os.Remove(dstFile.Name())

	// Write the template file.
	tmplFile, err := template.ParseFiles("testdata/resources.all.tmpl")
	require.NoError(t, err)
	err = tmplFile.Execute(dstFile, newDefaultResourcesParam(className))
	require.NoError(t, err)
	require.NoError(t, dstFile.Close())

	content, err := os.ReadFile(dstFile.Name())
	require.NoError(t, err)
	require.NoError(t, yaml.UnmarshalStrict(content, out, yaml.DisallowUnknownFields))
}

// expectResourcesToBePresent waits for the resources to be present in the provider.
func expectResourcesToBePresent(t *testing.T, resources *message.ProviderResources, classNames ...string) {
	t.Helper()

	require.Eventually(t, func() bool {
		for _, name := range classNames {
			if resources.GetResourcesByGatewayClass(name) == nil {
				return false
			}
		}
		return true
	}, resourcesUpdateTimeout, resourcesUpdateTick)
}

// expectResourcesToBeAbsent waits for the resources to be absent from the provider.
func expectResourcesToBeAbsent(t *testing.T, resources *message.ProviderResources, classNames ...string) {
	t.Helper()

	require.Eventually(t, func() bool {
		for _, name := range classNames {
			if resources.GetResourcesByGatewayClass(name) != nil {
				return false
			}
		}
		return true
	}, resourcesUpdateTimeout, resourcesUpdateTick)
}

// moveFile moves a file from sourcePath to destPath.
func moveFile(t *testing.T, sourcePath, destPath string) {
	t.Helper()

	inputFile, err := os.Open(sourcePath)
	require.NoError(t, err)

	outputFile, err := os.Create(destPath)
	require.NoError(t, err)
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	require.NoError(t, err)
	inputFile.Close()

	err = os.Remove(sourcePath)
	require.NoError(t, err)
}
