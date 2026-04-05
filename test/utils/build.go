// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// BuildGoBinaryOnDemand builds the binary unless the env variable is set.
// If the environment variable is set, it will use that path instead.
func BuildGoBinaryOnDemand(env, binaryName, packagePath string) (string, error) {
	if envPath := os.Getenv(env); envPath != "" {
		if !filepath.IsAbs(envPath) {
			envPath = filepath.Join(FindProjectRoot(), envPath)
		}
		if _, err := os.Stat(envPath); err != nil {
			return "", fmt.Errorf("%s path does not exist: %s", env, envPath)
		}
		fmt.Fprintf(os.Stderr, "Using %s : %s\n", env, envPath)
		return envPath, nil
	}

	return BuildGoBinary(binaryName, packagePath)
}

// BuildGoBinary builds a Go binary with the given name and package path.
func BuildGoBinary(binaryName, packagePath string) (string, error) {
	projectRoot := FindProjectRoot()
	outputDir := filepath.Join(projectRoot, "out")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	platformBinaryName := fmt.Sprintf("%s-%s-%s", binaryName, runtime.GOOS, runtime.GOARCH)
	binaryPath := filepath.Join(outputDir, platformBinaryName)

	cmd := exec.Command("go", "build", "-o", binaryPath, packagePath)
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	var stderr strings.Builder
	cmd.Stdout = io.Discard
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to build %s: %w\nstderr: %s", binaryName, err, stderr.String())
	}

	if err := os.Chmod(binaryPath, 0o755); err != nil {
		return "", fmt.Errorf("failed to make binary executable: %w", err)
	}
	return binaryPath, nil
}

// FindProjectRoot finds the root of the project by looking for go.mod.
func FindProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			panic("could not find project root (go.mod)")
		}
		dir = parent
	}
}
