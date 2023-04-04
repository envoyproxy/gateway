// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	envoyGatewayChartsDirectory = "charts/gateway-helm"
	templateFileExtension       = ".tmpl"
)

var (
	overlays = struct {
		EGImageTag        string
		EGImageRepository string
	}{
		os.Getenv("RELEASE_TAG"),
		os.Getenv("IMAGE"),
	}

	projectPath string
	chartPath   string
)

func main() {
	var err error
	projectPath, err = os.Getwd()
	if err != nil {
		fmt.Printf("error getting working directory: %v\n", err)
		os.Exit(1)
	}

	if len(overlays.EGImageTag) == 0 || len(overlays.EGImageRepository) == 0 {
		fmt.Printf("missing required env vars, got: %+v\n", overlays)
		os.Exit(1)
	}

	if err := renderChartFiles(envoyGatewayChartsDirectory); err != nil {
		fmt.Printf("unable to render template, error: %+v\n", err)
		os.Exit(1)
	}
}

func renderChartFiles(chart string) (err error) {
	chartPath = path.Join(projectPath, chart)

	templates, err := getTemplateFilesFromChartDir()
	if err != nil {
		fmt.Printf("cannot read dir: %v\n", err)
		os.Exit(1)
	}

	for _, tmpl := range templates {
		filename := strings.ReplaceAll(tmpl, templateFileExtension, "")

		templateBytes, err := os.ReadFile(filepath.Clean(tmpl))
		if err != nil {
			return err
		}

		t, err := template.New(tmpl).Parse(string(templateBytes))
		if err != nil {
			return err
		}

		if err = renderTemplate(t, overlays, filename); err != nil {
			return err
		}
	}
	return nil
}

func getTemplateFilesFromChartDir() ([]string, error) {
	var tmplFiles []string
	err := filepath.WalkDir(chartPath, func(p string, entry fs.DirEntry, err error) error {
		if strings.Contains(entry.Name(), templateFileExtension) {
			tmplFiles = append(tmplFiles, path.Join(chartPath, entry.Name()))
		}
		return nil
	})
	return tmplFiles, err
}

func renderTemplate(t *template.Template, data interface{}, filePath string) (err error) {
	w, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		return err
	}
	defer w.Close()
	return t.Execute(w, data)
}
