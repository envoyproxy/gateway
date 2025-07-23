// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package console

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
)

// Handler manages the console web interface
type Handler struct {
	cfg               *config.Server
	templates         map[string]*template.Template
	providerResources *message.ProviderResources
}

// NewHandler creates a new console handler
func NewHandler(cfg *config.Server, providerResources *message.ProviderResources) *Handler {
	return &Handler{
		cfg:               cfg,
		templates:         make(map[string]*template.Template),
		providerResources: providerResources,
	}
}

// RegisterRoutes registers all console routes with the provided mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Web UI endpoints
	mux.HandleFunc("/", h.handleIndex)
	mux.HandleFunc("/pprof", h.handlePprof)
	mux.HandleFunc("/server_info", h.handleServerInfo)
	mux.HandleFunc("/config_dump", h.handleConfigDump)
	mux.HandleFunc("/stats", h.handleStats)

	// API endpoints
	mux.HandleFunc("/api/info", h.handleAPIInfo)
	mux.HandleFunc("/api/server_info", h.handleAPIServerInfo)
	mux.HandleFunc("/api/config_dump", h.handleAPIConfigDump)
	mux.HandleFunc("/api/metrics", h.handleAPIMetrics)

	// Static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err == nil {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}
}

// handleIndex serves the main console page
func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := struct {
		Title     string
		Timestamp time.Time
	}{
		Title:     "Envoy Gateway Admin Console",
		Timestamp: time.Now(),
	}

	if err := h.renderTemplate(w, "index.html", data); err != nil {
		// Content-Type is already set by renderTemplate
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal server error"))
		return
	}
}

// handlePprof serves the pprof page
func (h *Handler) handlePprof(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	enablePprof := h.cfg.EnvoyGateway.GetEnvoyGatewayAdmin().EnablePprof
	data := struct {
		Title       string
		EnablePprof bool
	}{
		Title:       "Performance Profiling",
		EnablePprof: enablePprof,
	}

	if err := h.renderTemplate(w, "pprof.html", data); err != nil {
		// Content-Type is already set by renderTemplate
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal server error"))
		return
	}
}

// handleServerInfo serves the server info page
func (h *Handler) handleServerInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := struct {
		Title string
	}{
		Title: "Server Information",
	}

	if err := h.renderTemplate(w, "server_info.html", data); err != nil {
		// Content-Type is already set by renderTemplate
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal server error"))
		return
	}
}

// handleConfigDump serves the config dump page
func (h *Handler) handleConfigDump(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := struct {
		Title string
	}{
		Title: "Configuration Dump",
	}

	if err := h.renderTemplate(w, "config_dump.html", data); err != nil {
		// Content-Type is already set by renderTemplate
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal server error"))
		return
	}
}

// handleStats serves the stats page
func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get metrics server address
	metricsAddress := fmt.Sprintf("%s:%d", egv1a1.GatewayMetricsHost, egv1a1.GatewayMetricsPort)

	data := struct {
		Title          string
		MetricsAddress string
	}{
		Title:          "Statistics",
		MetricsAddress: metricsAddress,
	}

	if err := h.renderTemplate(w, "stats.html", data); err != nil {
		// Content-Type is already set by renderTemplate
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal server error"))
		return
	}
}

// renderTemplate renders a template with the given data
func (h *Handler) renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) error {
	if len(h.templates) == 0 {
		if err := h.loadTemplates(); err != nil {
			// Set Content-Type before returning error
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			return err
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get the specific template
	t, exists := h.templates[tmpl]
	if !exists {
		return fmt.Errorf("template %s not found", tmpl)
	}

	return t.ExecuteTemplate(w, "base.html", data)
}

// loadTemplates loads all HTML templates
func (h *Handler) loadTemplates() error {
	// Template mappings: template name -> content block name
	templateMappings := map[string]string{
		"index.html":       "index-content",
		"server_info.html": "server-info-content",
		"config_dump.html": "config-dump-content",
		"stats.html":       "stats-content",
		"pprof.html":       "pprof-content",
	}

	// Create individual templates for each page
	for templateName, contentBlock := range templateMappings {
		// Create a new template instance for each page
		tmpl := template.New(templateName)

		// Parse all templates together
		tmpl, err := tmpl.ParseFS(templateFiles, "templates/base.html", "templates/"+templateName)
		if err != nil {
			return fmt.Errorf("failed to parse templates for %s: %w", templateName, err)
		}

		// Create a wrapper template that calls the correct content block
		wrapperTemplate := fmt.Sprintf(`{{define "content"}}{{template "%s" .}}{{end}}`, contentBlock)
		tmpl, err = tmpl.Parse(wrapperTemplate)
		if err != nil {
			return fmt.Errorf("failed to parse wrapper for %s: %w", templateName, err)
		}

		h.templates[templateName] = tmpl
	}

	return nil
}
