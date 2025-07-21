// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package console

import (
	"bytes"
	"fmt"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateLoading(t *testing.T) {
	// Test each template using the same method as the handler
	testCases := []struct {
		templateName string
		contentBlock string
		expectedText string
	}{
		{"index.html", "index-content", "Welcome to the Envoy Gateway Admin Console"},
		{"server_info.html", "server-info-content", "displays detailed information about the Envoy Gateway server"},
		{"config_dump.html", "config-dump-content", "current configuration state"},
		{"stats.html", "stats-content", "Envoy Gateway statistics"},
		{"pprof.html", "pprof-content", "Performance profiling"},
	}

	for _, tc := range testCases {
		t.Run(tc.templateName, func(t *testing.T) {
			// Create a new template instance for each page (same as handler)
			tmpl := template.New(tc.templateName)

			// Parse base template and specific page template
			tmpl, err := tmpl.ParseFS(templateFiles, "templates/base.html", "templates/"+tc.templateName)
			require.NoError(t, err)

			// Create a wrapper template that calls the correct content block
			wrapperTemplate := fmt.Sprintf(`{{define "content"}}{{template "%s" .}}{{end}}`, tc.contentBlock)
			tmpl, err = tmpl.Parse(wrapperTemplate)
			require.NoError(t, err)

			var buf bytes.Buffer
			data := struct {
				Title          string
				MetricsAddress string
				EnablePprof    bool
			}{
				Title:          "Test",
				MetricsAddress: "localhost:19001",
				EnablePprof:    true,
			}

			// Execute the base template (same as handler)
			err = tmpl.ExecuteTemplate(&buf, "base.html", data)
			require.NoError(t, err)

			content := buf.String()
			assert.Contains(t, content, tc.expectedText, "Template %s should contain expected text", tc.templateName)

			// Verify the template produces valid HTML structure
			assert.Contains(t, content, "<html", "Template should contain HTML structure")
			assert.Contains(t, content, "</html>", "Template should be properly closed")
			assert.Contains(t, content, "Envoy Gateway", "Template should contain Envoy Gateway branding")
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
