// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gateway

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
)

// Handler exposes the gateway management REST API. All routes are protected by
// the auth middleware, which validates the caller's bearer token via
// go-authentication.
type Handler struct {
	service   *Service
	generator *Generator
	auth      *authMiddleware
	logger    logr.Logger
}

// Config configures a gateway management Handler.
type Config struct {
	Service   *Service
	Generator *Generator
	// AuthURL is the go-authentication base URL used to validate bearer tokens.
	AuthURL string
	Logger  logr.Logger
}

// NewHandler creates a gateway management Handler.
func NewHandler(cfg *Config) *Handler {
	return &Handler{
		service:   cfg.Service,
		generator: cfg.Generator,
		auth:      newAuthMiddleware(cfg.AuthURL, cfg.Logger),
		logger:    cfg.Logger,
	}
}

// RegisterRoutes registers all /api/gateway/* routes on the given mux, each
// wrapped with the auth middleware. It uses Go 1.22+ method+path patterns.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	a := h.auth.wrap

	// Cluster CRUD. The resource id is passed as an ?id= query parameter so the
	// path stays constant (simplifies the go-auth api_permission rows).
	mux.HandleFunc("POST /api/gateway/clusters", a(h.createCluster))
	mux.HandleFunc("GET /api/gateway/clusters", a(h.listClusters))
	mux.HandleFunc("PUT /api/gateway/clusters", a(h.updateCluster))
	mux.HandleFunc("DELETE /api/gateway/clusters", a(h.deleteCluster))

	// Service CRUD. The resource id is passed as an ?id= query parameter so the
	// path stays constant (simplifies the go-auth api_permission rows).
	mux.HandleFunc("POST /api/gateway/services", a(h.createService))
	mux.HandleFunc("GET /api/gateway/services", a(h.listServices))
	mux.HandleFunc("PUT /api/gateway/services", a(h.updateService))
	mux.HandleFunc("DELETE /api/gateway/services", a(h.deleteService))

	// Route CRUD.
	mux.HandleFunc("POST /api/gateway/routes", a(h.createRoute))
	mux.HandleFunc("GET /api/gateway/routes", a(h.listRoutes))
	mux.HandleFunc("PUT /api/gateway/routes", a(h.updateRoute))
	mux.HandleFunc("DELETE /api/gateway/routes", a(h.deleteRoute))

	// Generation.
	mux.HandleFunc("POST /api/gateway/generate", a(h.generate))
}

// --- cluster handlers ---

func (h *Handler) createCluster(w http.ResponseWriter, r *http.Request) {
	var req ClusterRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := h.service.CreateCluster(r.Context(), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) listClusters(w http.ResponseWriter, r *http.Request) {
	// GET /api/gateway/clusters?id=N returns a single cluster.
	if strings.TrimSpace(r.URL.Query().Get("id")) != "" {
		h.getCluster(w, r)
		return
	}
	resp, err := h.service.ListClusters(r.Context())
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getCluster(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	resp, err := h.service.GetCluster(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) updateCluster(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req ClusterRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := h.service.UpdateCluster(r.Context(), id, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) deleteCluster(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.service.DeleteCluster(r.Context(), id); err != nil {
		h.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- service handlers ---

func (h *Handler) createService(w http.ResponseWriter, r *http.Request) {
	var req ServiceRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := h.service.CreateService(r.Context(), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) listServices(w http.ResponseWriter, r *http.Request) {
	// GET /api/gateway/services?id=N returns a single service.
	if strings.TrimSpace(r.URL.Query().Get("id")) != "" {
		h.getService(w, r)
		return
	}
	resp, err := h.service.ListServices(r.Context(), r.URL.Query().Get("name"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getService(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	resp, err := h.service.GetService(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) updateService(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req ServiceRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := h.service.UpdateService(r.Context(), id, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) deleteService(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.service.DeleteService(r.Context(), id); err != nil {
		h.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- route handlers ---

func (h *Handler) createRoute(w http.ResponseWriter, r *http.Request) {
	var req RouteRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := h.service.CreateRoute(r.Context(), req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) listRoutes(w http.ResponseWriter, r *http.Request) {
	// GET /api/gateway/routes?id=N returns a single route.
	if strings.TrimSpace(r.URL.Query().Get("id")) != "" {
		h.getRoute(w, r)
		return
	}
	var serviceID int64
	if raw := strings.TrimSpace(r.URL.Query().Get("service_id")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "service_id must be a positive integer")
			return
		}
		serviceID = parsed
	}
	resp, err := h.service.ListRoutes(r.Context(), serviceID, r.URL.Query().Get("match_type"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	resp, err := h.service.GetRoute(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) updateRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req RouteRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := h.service.UpdateRoute(r.Context(), id, req)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) deleteRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.service.DeleteRoute(r.Context(), id); err != nil {
		h.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- generation handler ---

func (h *Handler) generate(w http.ResponseWriter, r *http.Request) {
	cluster := strings.TrimSpace(r.URL.Query().Get("cluster"))
	if cluster == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "cluster query parameter is required")
		return
	}

	yaml, err := h.generator.Generate(r.Context(), cluster)
	if err != nil {
		if errors.Is(err, ErrInvalid) {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		h.logger.Error(err, "failed to generate route files", "cluster", cluster)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate route files")
		return
	}

	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(yaml))
}

// --- helpers ---

// parseID extracts and validates the ?id= query parameter.
func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "id query parameter must be a positive integer")
		return 0, false
	}
	return id, true
}

// decodeJSON decodes the request body into dst, writing a 400 on failure.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return false
	}
	return true
}

// writeJSON writes v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a structured JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{
		"errorCode": code,
		"message":   message,
	})
}

// writeServiceError maps service/repository errors to HTTP responses.
func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Resource was not found")
	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusConflict, "CONFLICT", "An active resource already exists with the same unique fields")
	default:
		h.logger.Error(err, "gateway management operation failed")
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Operation failed")
	}
}
