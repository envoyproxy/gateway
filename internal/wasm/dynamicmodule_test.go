// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package wasm

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/sets"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
)

// fakeELF returns content that passes the ELF shared object check.
func fakeELF(name string) []byte {
	return append([]byte{0x7f, 'E', 'L', 'F'}, []byte(name)...)
}

type tarEntry struct {
	name    string
	content []byte
	dir     bool
	// symlink and hardlink are the targets of a link entry.
	symlink  string
	hardlink string
}

// newTarLayer returns a gzipped tar layer, as found in container images.
func newTarLayer(t *testing.T, mediaType types.MediaType, entries ...tarEntry) v1.Layer {
	t.Helper()
	var b bytes.Buffer
	tw := tar.NewWriter(&b)
	for _, e := range entries {
		h := &tar.Header{Name: e.name, Mode: 0o644, Size: int64(len(e.content)), Typeflag: tar.TypeReg}
		switch {
		case e.dir:
			h = &tar.Header{Name: e.name, Mode: 0o755, Typeflag: tar.TypeDir}
		case e.symlink != "":
			h = &tar.Header{Name: e.name, Mode: 0o777, Typeflag: tar.TypeSymlink, Linkname: e.symlink}
		case e.hardlink != "":
			h = &tar.Header{Name: e.name, Mode: 0o644, Typeflag: tar.TypeLink, Linkname: e.hardlink}
		}
		require.NoError(t, tw.WriteHeader(h))
		_, err := tw.Write(e.content)
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	raw := b.Bytes()
	l, err := tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(raw)), nil
	}, tarball.WithMediaType(mediaType))
	require.NoError(t, err)
	return l
}

// newImage returns an OCI image with the given layers. A layer with a title is
// added as a raw file layer, the way `oras push` does.
func newImage(t *testing.T, layers ...mutate.Addendum) v1.Image {
	t.Helper()
	img := mutate.MediaType(empty.Image, types.OCIManifestSchema1)
	img, err := mutate.Append(img, layers...)
	require.NoError(t, err)
	return img
}

func tarLayer(t *testing.T, entries ...tarEntry) mutate.Addendum {
	return mutate.Addendum{Layer: newTarLayer(t, types.OCILayer, entries...)}
}

func orasLayer(title string, content []byte) mutate.Addendum {
	return mutate.Addendum{
		Layer:       static.NewLayer(content, "application/vnd.oci.image.layer.v1.tar"),
		Annotations: map[string]string{ociImageTitleAnnotation: title},
	}
}

func TestExtractDynamicModuleBinary(t *testing.T) {
	libA := fakeELF("liba")
	libB := fakeELF("libb")

	cases := []struct {
		name       string
		img        func(t *testing.T) v1.Image
		modulePath string
		maxSize    int64
		want       []byte
		wantErr    string
	}{
		{
			name: "container image with path",
			img: func(t *testing.T) v1.Image {
				return newImage(t, tarLayer(t,
					tarEntry{name: "usr/", dir: true},
					tarEntry{name: "usr/lib/liba.so", content: libA},
					tarEntry{name: "usr/lib/libb.so", content: libB},
				))
			},
			modulePath: "/usr/lib/libb.so",
			want:       libB,
		},
		{
			name: "container image with path in tar prefixed with ./",
			img: func(t *testing.T) v1.Image {
				return newImage(t, tarLayer(t, tarEntry{name: "./usr/lib/liba.so", content: libA}))
			},
			modulePath: "usr/lib/liba.so",
			want:       libA,
		},
		{
			name: "container image with path to a relative symlink",
			img: func(t *testing.T) v1.Image {
				return newImage(t, tarLayer(t,
					tarEntry{name: "usr/lib/liba.so.1", content: libA},
					tarEntry{name: "usr/lib/liba.so", symlink: "liba.so.1"},
				))
			},
			modulePath: "/usr/lib/liba.so",
			want:       libA,
		},
		{
			name: "container image with a symlink chain across layers",
			img: func(t *testing.T) v1.Image {
				return newImage(t,
					tarLayer(t, tarEntry{name: "opt/liba.so.1.0", content: libA}),
					tarLayer(t,
						tarEntry{name: "usr/lib/liba.so.1", symlink: "/opt/liba.so.1.0"},
						tarEntry{name: "usr/lib/liba.so", symlink: "liba.so.1"},
					),
				)
			},
			want: libA,
		},
		{
			name: "container image with path to a hard link",
			img: func(t *testing.T) v1.Image {
				return newImage(t, tarLayer(t,
					tarEntry{name: "lib/liba.so.1", content: libA},
					tarEntry{name: "lib/liba.so", hardlink: "lib/liba.so.1"},
				))
			},
			modulePath: "/lib/liba.so",
			want:       libA,
		},
		{
			name: "container image with a dangling symlink",
			img: func(t *testing.T) v1.Image {
				return newImage(t, tarLayer(t, tarEntry{name: "lib/liba.so", symlink: "liba.so.1"}))
			},
			modulePath: "/lib/liba.so",
			wantErr:    "lib/liba.so.1 not found in the image",
		},
		{
			name: "container image with a symlink loop",
			img: func(t *testing.T) v1.Image {
				return newImage(t, tarLayer(t,
					tarEntry{name: "lib/liba.so", symlink: "libb.so"},
					tarEntry{name: "lib/libb.so", symlink: "liba.so"},
				))
			},
			modulePath: "/lib/liba.so",
			wantErr:    "too many links",
		},
		{
			name: "container image with an opaque whiteout",
			img: func(t *testing.T) v1.Image {
				return newImage(t,
					tarLayer(t,
						tarEntry{name: "lib/liba.so", content: libA},
						tarEntry{name: "opt/libb.so", content: libB},
					),
					tarLayer(t, tarEntry{name: "lib/.wh..wh..opq"}),
				)
			},
			want: libB,
		},
		{
			name: "container image with the only shared library",
			img: func(t *testing.T) v1.Image {
				return newImage(t,
					tarLayer(t, tarEntry{name: "etc/config.json", content: []byte("{}")}),
					tarLayer(t, tarEntry{name: "lib/liba.so", content: libA}),
				)
			},
			want: libA,
		},
		{
			name: "docker layer",
			img: func(t *testing.T) v1.Image {
				return newImage(t, mutate.Addendum{Layer: newTarLayer(t, types.DockerLayer, tarEntry{name: "liba.so", content: libA})})
			},
			want: libA,
		},
		{
			name: "upper layer shadows lower layer",
			img: func(t *testing.T) v1.Image {
				return newImage(t,
					tarLayer(t, tarEntry{name: "lib/liba.so", content: libB}),
					tarLayer(t, tarEntry{name: "lib/liba.so", content: libA}),
				)
			},
			want: libA,
		},
		{
			name: "file deleted by an upper layer",
			img: func(t *testing.T) v1.Image {
				return newImage(t,
					tarLayer(t, tarEntry{name: "lib/liba.so", content: libA}),
					tarLayer(t, tarEntry{name: "lib/.wh.liba.so"}),
				)
			},
			modulePath: "/lib/liba.so",
			wantErr:    "lib/liba.so not found in the image",
		},
		{
			name: "directory deleted by an upper layer",
			img: func(t *testing.T) v1.Image {
				return newImage(t,
					tarLayer(t, tarEntry{name: "lib/old/liba.so", content: libA}),
					tarLayer(t, tarEntry{name: "lib/.wh.old"}, tarEntry{name: "lib/new/libb.so", content: libB}),
				)
			},
			want: libB,
		},
		{
			name: "multiple shared libraries without path",
			img: func(t *testing.T) v1.Image {
				return newImage(t,
					tarLayer(t, tarEntry{name: "lib/liba.so", content: libA}),
					tarLayer(t, tarEntry{name: "lib/libb.so", content: libB}),
				)
			},
			wantErr: "found multiple shared libraries (lib/libb.so, lib/liba.so) in the image",
		},
		{
			name: "path not found",
			img: func(t *testing.T) v1.Image {
				return newImage(t, tarLayer(t, tarEntry{name: "lib/liba.so", content: libA}))
			},
			modulePath: "/lib/libc.so",
			wantErr:    "lib/libc.so not found in the image",
		},
		{
			name: "no shared library",
			img: func(t *testing.T) v1.Image {
				return newImage(t, tarLayer(t, tarEntry{name: "README.md", content: []byte("hi")}))
			},
			wantErr: "no shared library (*.so) found in the image",
		},
		{
			name: "file exceeds the maximum size",
			img: func(t *testing.T) v1.Image {
				return newImage(t, tarLayer(t, tarEntry{name: "liba.so", content: libA}))
			},
			maxSize: 4,
			wantErr: "exceeds maximum size 4",
		},
		{
			name: "oras artifact",
			img: func(t *testing.T) v1.Image {
				return newImage(t, orasLayer("liba.so", libA))
			},
			want: libA,
		},
		{
			name: "oras artifact with multiple files and path",
			img: func(t *testing.T) v1.Image {
				return newImage(t, orasLayer("liba.so", libA), orasLayer("libb.so", libB))
			},
			modulePath: "libb.so",
			want:       libB,
		},
		{
			name: "oras artifact exceeds the maximum size",
			img: func(t *testing.T) v1.Image {
				return newImage(t, orasLayer("liba.so", libA))
			},
			maxSize: 4,
			wantErr: "liba.so size exceeds maximum size 4",
		},
		{
			name: "empty image",
			img: func(_ *testing.T) v1.Image {
				return mutate.MediaType(empty.Image, types.OCIManifestSchema1)
			},
			wantErr: "number of layers must be greater than zero",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			maxSize := tc.maxSize
			if maxSize == 0 {
				maxSize = maxWasmSize
			}
			got, err := extractDynamicModuleBinary(tc.img(t), tc.modulePath, maxSize)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestIsValidDynamicModuleBinary(t *testing.T) {
	require.True(t, isValidDynamicModuleBinary(fakeELF("lib")))
	require.False(t, isValidDynamicModuleBinary([]byte{0x7f, 'E', 'L'}))
	require.False(t, isValidDynamicModuleBinary(append(wasmMagicNumber, 0x1, 0x0, 0x0, 0x0)))
}

func TestCacheDynamicModuleFromOCIImage(t *testing.T) {
	s := httptest.NewServer(registry.New())
	t.Cleanup(s.Close)
	u, err := url.Parse(s.URL)
	require.NoError(t, err)

	libA := fakeELF("liba")
	libB := fakeELF("libb")

	img := newImage(t, tarLayer(t,
		tarEntry{name: "usr/lib/liba.so", content: libA},
		tarEntry{name: "usr/lib/libb.so", content: libB},
	))
	ref := fmt.Sprintf("%s/dynamic-modules/test:v1", u.Host)
	require.NoError(t, crane.Push(img, ref))
	digest, err := img.Digest()
	require.NoError(t, err)

	wasmImg := newImage(t, tarLayer(t, tarEntry{name: "plugin.wasm", content: append(wasmMagicNumber, 0x1, 0x0, 0x0, 0x0)}))
	wasmRef := fmt.Sprintf("%s/wasm/test:v1", u.Host)
	require.NoError(t, crane.Push(wasmImg, wasmRef))

	tmpDir := t.TempDir()
	options := defaultCacheOptions()
	options.CacheDir = tmpDir
	options.InsecureRegistries = sets.New("*")
	cache := newLocalFileCache(options, logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo))
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)
	cache.Start(ctx)

	sum := func(b []byte) string {
		s := sha256.Sum256(b)
		return hex.EncodeToString(s[:])
	}

	t.Run("modules from the same image are cached separately", func(t *testing.T) {
		for _, tc := range []struct {
			path string
			want []byte
		}{
			{path: "/usr/lib/liba.so", want: libA},
			{path: "/usr/lib/libb.so", want: libB},
		} {
			file, checksum, err := cache.Get("oci://"+ref, &GetOptions{
				ModuleType:   ModuleTypeDynamicModule,
				ModulePath:   tc.path,
				Checksum:     digest.Hex,
				ResourceName: "envoyproxy/default/proxy/dynamic-module/" + tc.path,
			})
			require.NoError(t, err)
			require.Equal(t, sum(tc.want), checksum)
			require.True(t, strings.HasSuffix(file, digest.Hex+".so"), file)
			got, err := os.ReadFile(file)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		}
		require.Len(t, cache.modules, 2)
	})

	t.Run("image digest mismatch", func(t *testing.T) {
		_, _, err := cache.Get("oci://"+ref, &GetOptions{
			ModuleType: ModuleTypeDynamicModule,
			ModulePath: "/usr/lib/liba.so",
			Checksum:   strings.Repeat("0", 64),
		})
		require.ErrorContains(t, err, "which does not match")
	})

	t.Run("ambiguous shared library", func(t *testing.T) {
		_, _, err := cache.Get("oci://"+ref, &GetOptions{ModuleType: ModuleTypeDynamicModule})
		require.ErrorContains(t, err, "found multiple shared libraries")
	})

	t.Run("wasm image is not a dynamic module", func(t *testing.T) {
		_, _, err := cache.Get("oci://"+wasmRef, &GetOptions{
			ModuleType: ModuleTypeDynamicModule,
			ModulePath: "plugin.wasm",
		})
		require.ErrorContains(t, err, "fetched dynamic module binary from")
	})

	t.Run("multi-platform image", func(t *testing.T) {
		otherArch := "arm64"
		if runtime.GOARCH == otherArch {
			otherArch = "amd64"
		}
		libHost, libOther := fakeELF("host"), fakeELF("other")
		idx := mutate.AppendManifests(mutate.IndexMediaType(empty.Index, types.OCIImageIndex),
			mutate.IndexAddendum{
				Add:        newImage(t, orasLayer("libmodule.so", libOther)),
				Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: "linux", Architecture: otherArch}},
			},
			mutate.IndexAddendum{
				Add:        newImage(t, orasLayer("libmodule.so", libHost)),
				Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: "linux", Architecture: runtime.GOARCH}},
			},
		)
		idxRef, err := name.ParseReference(fmt.Sprintf("%s/dynamic-modules/multi:v1", u.Host))
		require.NoError(t, err)
		require.NoError(t, remote.WriteIndex(idxRef, idx))
		idxDigest, err := idx.Digest()
		require.NoError(t, err)

		// The image digest is the one of the index, both in the URL and in the checksum.
		for _, ref := range []string{
			idxRef.String(),
			fmt.Sprintf("%s/dynamic-modules/multi@%s", u.Host, idxDigest),
		} {
			file, checksum, err := cache.Get("oci://"+ref, &GetOptions{
				ModuleType: ModuleTypeDynamicModule,
				Checksum:   idxDigest.Hex,
			})
			require.NoError(t, err)
			require.Equal(t, sum(libHost), checksum)
			got, err := os.ReadFile(file)
			require.NoError(t, err)
			require.Equal(t, libHost, got)
		}
	})

	t.Run("http source is not supported", func(t *testing.T) {
		_, _, err := cache.Get("https://example.com/liba.so", &GetOptions{ModuleType: ModuleTypeDynamicModule})
		require.ErrorContains(t, err, "unsupported dynamic module downloading URL scheme")
	})
}

func TestHTTPServerServesDynamicModule(t *testing.T) {
	s := httptest.NewServer(registry.New())
	t.Cleanup(s.Close)
	u, err := url.Parse(s.URL)
	require.NoError(t, err)

	lib := fakeELF("liba")
	ref := fmt.Sprintf("%s/dynamic-modules/test:v1", u.Host)
	require.NoError(t, crane.Push(newImage(t, orasLayer("liba.so", lib)), ref))

	tmpDir := t.TempDir()
	server := NewHTTPServerWithFileCache(
		ServerOptions{Salt: []byte("salt")},
		CacheOptions{CacheDir: tmpDir, InsecureRegistries: sets.New("*")},
		"envoy-gateway-system",
		logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
	)

	servingURL, checksum, err := server.Get("oci://"+ref, &GetOptions{ModuleType: ModuleTypeDynamicModule})
	require.NoError(t, err)
	want := sha256.Sum256(lib)
	require.Equal(t, hex.EncodeToString(want[:]), checksum)
	require.True(t, strings.HasPrefix(servingURL, "http://envoy-gateway.envoy-gateway-system.svc.cluster.local:18002/"), servingURL)
	require.Equal(t, ".so", filepath.Ext(servingURL))

	// The Wasm module from the same URL gets a different serving path.
	wasmPath := generateUnguessablePath("oci://"+ref, []byte("salt"), "wasm")
	require.NotEqual(t, wasmPath, filepath.Base(servingURL))

	servingPath, err := url.Parse(servingURL)
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, servingPath.Path, nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, lib, rec.Body.Bytes())
}
