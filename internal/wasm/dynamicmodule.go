// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package wasm

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// This file implements the extraction of an Envoy dynamic module shared library from an OCI image.
// There is no image specification for dynamic modules, so two common layouts are supported:
//
//   - Container images, where the shared library is a file in one of the image's tar layers.
//     Layers are searched from the top of the image down, so a file in an upper layer shadows
//     the same file in a lower one.
//   - OCI artifacts pushed with ORAS (e.g. `oras push registry/module:v1 libmodule.so`), where a
//     layer is the raw shared library and its file name is carried by the
//     "org.opencontainers.image.title" annotation.

const (
	// ociImageTitleAnnotation is the annotation ORAS uses to record the file name of a layer.
	ociImageTitleAnnotation = "org.opencontainers.image.title"
	// orasUnpackAnnotation marks an ORAS layer that is a (gzipped) tarball of a directory.
	orasUnpackAnnotation = "io.deis.oras.content.unpack"
	// tarWhiteoutPrefix marks a file deleted by an upper layer of a container image.
	tarWhiteoutPrefix = ".wh."
	// tarOpaqueWhiteout marks a directory whose content in the lower layers is hidden.
	tarOpaqueWhiteout = ".wh..wh..opq"
	// maxLinkHops is the maximum number of links followed to find the shared library.
	maxLinkHops = 8
	// sharedLibraryExt is the file extension of a shared library.
	sharedLibraryExt = ".so"
)

// extractDynamicModuleBinary extracts the dynamic module shared library from the image.
// If modulePath is set, the file at that path is returned. Otherwise, the image must contain
// exactly one shared library (a file whose name ends with ".so").
func extractDynamicModuleBinary(img v1.Image, modulePath string, maxSize int64) ([]byte, error) {
	manifest, err := img.Manifest()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve manifest: %w", err)
	}
	layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("could not fetch layers: %w", err)
	}
	if len(layers) == 0 {
		return nil, errors.New("number of layers must be greater than zero")
	}
	if len(layers) != len(manifest.Layers) {
		return nil, fmt.Errorf("image has %d layers but its manifest lists %d", len(layers), len(manifest.Layers))
	}

	// Shared libraries are often installed as a symlink to a versioned file
	// (e.g. libmodule.so -> libmodule.so.1), so follow links to the actual file.
	p := normalizeImagePath(modulePath)
	for range maxLinkHops {
		s, err := searchImage(manifest, layers, p, maxSize)
		if err != nil {
			return nil, err
		}
		if s.found == "" {
			if s.path != "" {
				return nil, fmt.Errorf("%s not found in the image", s.path)
			}
			return nil, fmt.Errorf("no shared library (*%s) found in the image", sharedLibraryExt)
		}
		if s.linkTarget == "" {
			return s.data, nil
		}
		p = s.linkTarget
	}
	return nil, fmt.Errorf("too many links to find the shared library in the image (last target %s)", p)
}

// searchImage searches the layers of the image for the shared library at the normalized path p,
// or for the only shared library if p is empty.
func searchImage(manifest *v1.Manifest, layers []v1.Layer, p string, maxSize int64) (*moduleSearch, error) {
	s := &moduleSearch{
		path:    p,
		maxSize: maxSize,
		seen:    make(map[string]struct{}),
	}

	// Walk the layers from the top of the image down.
	for i := len(layers) - 1; i >= 0; i-- {
		var err error
		annotations := manifest.Layers[i].Annotations
		if title, ok := annotations[ociImageTitleAnnotation]; ok && annotations[orasUnpackAnnotation] != "true" {
			err = s.searchRawLayer(layers[i], title)
		} else {
			err = s.searchTarLayer(layers[i])
		}
		if err != nil {
			return nil, err
		}
		if s.done {
			break
		}
	}
	return s, nil
}

// moduleSearch holds the state of a search for the shared library across the image layers.
type moduleSearch struct {
	// path is the normalized path of the shared library, empty to search for the only shared library.
	path    string
	maxSize int64
	// seen records the files present in the layers searched so far.
	// A file in a lower layer with the same path is shadowed.
	seen map[string]struct{}
	// deleted records the paths deleted by whiteouts in the layers searched so far.
	// A file in a lower layer at, or under, one of these paths is shadowed.
	deleted []string
	// found is the path of the shared library found, and data its content.
	found string
	data  []byte
	// linkTarget is the normalized path the file found links to, if it is a link.
	linkTarget string
	// done is true when the search has a definitive result.
	done bool
}

// match reports whether the file at the normalized path p is the shared library being searched for.
func (s *moduleSearch) match(p string) bool {
	if s.path != "" {
		return p == s.path
	}
	return strings.HasSuffix(p, sharedLibraryExt)
}

// visit records a file found in a layer and reports whether its content needs to be read.
func (s *moduleSearch) visit(p string) (bool, error) {
	if _, shadowed := s.seen[p]; shadowed {
		return false, nil
	}
	for _, d := range s.deleted {
		if p == d || strings.HasPrefix(p, d+"/") {
			return false, nil
		}
	}
	s.seen[p] = struct{}{}
	if !s.match(p) {
		return false, nil
	}
	if s.found != "" {
		return false, fmt.Errorf("found multiple shared libraries (%s, %s) in the image, the path of the module must be specified", s.found, p)
	}
	return true, nil
}

func (s *moduleSearch) setFound(p string, data []byte, linkTarget string) {
	s.found = p
	s.data = data
	s.linkTarget = linkTarget
	// When the path is given, the first match is the result.
	// Otherwise, keep searching to make sure the shared library is the only one.
	s.done = s.path != ""
}

// searchRawLayer searches a layer whose content is the file named title.
func (s *moduleSearch) searchRawLayer(layer v1.Layer, title string) error {
	p := normalizeImagePath(title)
	read, err := s.visit(p)
	if err != nil || !read {
		return err
	}

	// The layer is the raw file, so read it without decompression.
	r, err := layer.Compressed()
	if err != nil {
		return fmt.Errorf("could not get layer content: %w", err)
	}
	defer r.Close()

	data, err := io.ReadAll(io.LimitReader(r, s.maxSize+1))
	if err != nil {
		return fmt.Errorf("could not read %s: %w", p, err)
	}
	if int64(len(data)) > s.maxSize {
		return fmt.Errorf("%s size exceeds maximum size %d", p, s.maxSize)
	}
	s.setFound(p, data, "")
	return nil
}

// searchTarLayer searches a container image layer, which is a (possibly compressed) tarball.
func (s *moduleSearch) searchTarLayer(layer v1.Layer) error {
	r, err := layer.Uncompressed()
	if err != nil {
		return fmt.Errorf("could not get layer content: %w", err)
	}
	defer r.Close()

	// Files deleted by this layer only shadow the lower layers,
	// so record them once the whole layer has been searched.
	var whiteouts []string
	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("could not read layer: %w", err)
		}

		p := normalizeImagePath(h.Name)
		if base := path.Base(p); base == tarOpaqueWhiteout {
			whiteouts = append(whiteouts, path.Dir(p))
			continue
		} else if strings.HasPrefix(base, tarWhiteoutPrefix) {
			whiteouts = append(whiteouts, path.Join(path.Dir(p), strings.TrimPrefix(base, tarWhiteoutPrefix)))
			continue
		}
		var linkTarget string
		switch h.Typeflag {
		case tar.TypeReg:
		case tar.TypeSymlink:
			// A relative symlink target is relative to the directory of the link.
			// The target is only used to look up another file in the image, and
			// normalizeImagePath keeps it within the image root.
			linkTarget = h.Linkname
			if !path.IsAbs(linkTarget) {
				linkTarget = path.Dir(p) + "/" + linkTarget
			}
			linkTarget = normalizeImagePath(linkTarget)
		case tar.TypeLink:
			// A hard link target is relative to the root of the layer.
			linkTarget = normalizeImagePath(h.Linkname)
		default:
			continue
		}

		read, err := s.visit(p)
		if err != nil {
			return err
		}
		if !read {
			continue
		}
		if linkTarget != "" {
			s.setFound(p, nil, linkTarget)
			if s.done {
				return nil
			}
			continue
		}
		if h.Size > s.maxSize {
			return fmt.Errorf("%s size %d exceeds maximum size %d", p, h.Size, s.maxSize)
		}
		data := make([]byte, h.Size)
		if _, err := io.ReadFull(tr, data); err != nil {
			return fmt.Errorf("could not read %s: %w", p, err)
		}
		s.setFound(p, data, "")
		if s.done {
			return nil
		}
	}

	s.deleted = append(s.deleted, whiteouts...)
	return nil
}

// normalizeImagePath returns the path of a file in an image relative to the image root,
// e.g. "/usr/lib/libmodule.so" and "./usr/lib/libmodule.so" are both "usr/lib/libmodule.so".
func normalizeImagePath(p string) string {
	if p == "" {
		return ""
	}
	return strings.TrimPrefix(path.Clean("/"+p), "/")
}
