// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gateway

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalid is returned when a request fails validation.
var ErrInvalid = errors.New("invalid request")

// Service holds the business logic for managing gateway services and routes.
type Service struct {
	repo *Repository
}

// NewService creates a Service backed by the given repository.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// --- cluster operations ---

// CreateCluster validates the request and inserts a new cluster.
func (s *Service) CreateCluster(ctx context.Context, req ClusterRequest) (ClusterResponse, error) {
	e, err := validateClusterRequest(req)
	if err != nil {
		return ClusterResponse{}, err
	}
	out, err := s.repo.CreateCluster(ctx, e)
	if err != nil {
		return ClusterResponse{}, err
	}
	return out.toResponse(), nil
}

// GetCluster returns a single cluster by id.
func (s *Service) GetCluster(ctx context.Context, id int64) (ClusterResponse, error) {
	out, err := s.repo.GetCluster(ctx, id)
	if err != nil {
		return ClusterResponse{}, err
	}
	return out.toResponse(), nil
}

// ListClusters returns all clusters.
func (s *Service) ListClusters(ctx context.Context) ([]ClusterResponse, error) {
	entities, err := s.repo.ListClusters(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ClusterResponse, 0, len(entities))
	for _, e := range entities {
		result = append(result, e.toResponse())
	}
	return result, nil
}

// UpdateCluster validates the request and updates an existing cluster.
func (s *Service) UpdateCluster(ctx context.Context, id int64, req ClusterRequest) (ClusterResponse, error) {
	e, err := validateClusterRequest(req)
	if err != nil {
		return ClusterResponse{}, err
	}
	e.ID = id
	out, err := s.repo.UpdateCluster(ctx, e)
	if err != nil {
		return ClusterResponse{}, err
	}
	return out.toResponse(), nil
}

// DeleteCluster soft-deletes a cluster.
func (s *Service) DeleteCluster(ctx context.Context, id int64) error {
	return s.repo.DeleteCluster(ctx, id)
}

// --- service operations ---

// CreateService validates the request and inserts a new service.
func (s *Service) CreateService(ctx context.Context, req ServiceRequest) (ServiceResponse, error) {
	e, err := validateServiceRequest(req)
	if err != nil {
		return ServiceResponse{}, err
	}
	out, err := s.repo.CreateService(ctx, e)
	if err != nil {
		return ServiceResponse{}, err
	}
	return out.toResponse(), nil
}

// GetService returns a single service by id.
func (s *Service) GetService(ctx context.Context, id int64) (ServiceResponse, error) {
	out, err := s.repo.GetService(ctx, id)
	if err != nil {
		return ServiceResponse{}, err
	}
	return out.toResponse(), nil
}

// ListServices returns services filtered by name.
func (s *Service) ListServices(ctx context.Context, name string) ([]ServiceResponse, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	entities, err := s.repo.ListServices(ctx, name)
	if err != nil {
		return nil, err
	}
	result := make([]ServiceResponse, 0, len(entities))
	for _, e := range entities {
		result = append(result, e.toResponse())
	}
	return result, nil
}

// UpdateService validates the request and updates an existing service.
func (s *Service) UpdateService(ctx context.Context, id int64, req ServiceRequest) (ServiceResponse, error) {
	e, err := validateServiceRequest(req)
	if err != nil {
		return ServiceResponse{}, err
	}
	e.ID = id
	out, err := s.repo.UpdateService(ctx, e)
	if err != nil {
		return ServiceResponse{}, err
	}
	return out.toResponse(), nil
}

// DeleteService soft-deletes a service and cascade soft-deletes its routes.
func (s *Service) DeleteService(ctx context.Context, id int64) error {
	return s.repo.DeleteService(ctx, id)
}

// --- route operations ---

// CreateRoute validates the request and inserts a new route. The referenced
// service must exist and be active.
func (s *Service) CreateRoute(ctx context.Context, req RouteRequest) (RouteResponse, error) {
	e, err := validateRouteRequest(req)
	if err != nil {
		return RouteResponse{}, err
	}
	if err := s.ensureServiceActive(ctx, e.ServiceID); err != nil {
		return RouteResponse{}, err
	}
	out, err := s.repo.CreateRoute(ctx, e)
	if err != nil {
		return RouteResponse{}, err
	}
	return out.toResponse(), nil
}

// GetRoute returns a single route by id.
func (s *Service) GetRoute(ctx context.Context, id int64) (RouteResponse, error) {
	out, err := s.repo.GetRoute(ctx, id)
	if err != nil {
		return RouteResponse{}, err
	}
	return out.toResponse(), nil
}

// ListRoutes returns routes filtered by service_id and match_type.
func (s *Service) ListRoutes(ctx context.Context, serviceID int64, matchType string) ([]RouteResponse, error) {
	matchType = strings.TrimSpace(matchType)
	if matchType != "" {
		normalized, err := normalizeMatchType(matchType)
		if err != nil {
			return nil, err
		}
		matchType = normalized
	}
	entities, err := s.repo.ListRoutes(ctx, serviceID, matchType)
	if err != nil {
		return nil, err
	}
	result := make([]RouteResponse, 0, len(entities))
	for _, e := range entities {
		result = append(result, e.toResponse())
	}
	return result, nil
}

// UpdateRoute validates the request and updates an existing route.
func (s *Service) UpdateRoute(ctx context.Context, id int64, req RouteRequest) (RouteResponse, error) {
	e, err := validateRouteRequest(req)
	if err != nil {
		return RouteResponse{}, err
	}
	if err := s.ensureServiceActive(ctx, e.ServiceID); err != nil {
		return RouteResponse{}, err
	}
	e.ID = id
	out, err := s.repo.UpdateRoute(ctx, e)
	if err != nil {
		return RouteResponse{}, err
	}
	return out.toResponse(), nil
}

// DeleteRoute soft-deletes a single route.
func (s *Service) DeleteRoute(ctx context.Context, id int64) error {
	return s.repo.DeleteRoute(ctx, id)
}

func (s *Service) ensureServiceActive(ctx context.Context, serviceID int64) error {
	exists, err := s.repo.ServiceExists(ctx, serviceID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: service_id %d does not reference an active service", ErrInvalid, serviceID)
	}
	return nil
}

// --- validation ---

func validateClusterRequest(req ClusterRequest) (clusterEntity, error) {
	name := strings.ToUpper(strings.TrimSpace(req.Name))
	if name == "" || len(name) > 32 {
		return clusterEntity{}, fmt.Errorf("%w: name is required and must be 1-32 chars", ErrInvalid)
	}

	internalLB := strings.TrimSpace(req.InternalLB)
	if internalLB == "" || len(internalLB) > 255 {
		return clusterEntity{}, fmt.Errorf("%w: internal_lb is required and must be 1-255 chars", ErrInvalid)
	}
	if strings.Contains(internalLB, "://") {
		return clusterEntity{}, fmt.Errorf("%w: internal_lb must be a hostname only, without a scheme", ErrInvalid)
	}

	return clusterEntity{
		Name:       name,
		InternalLB: internalLB,
	}, nil
}

func validateServiceRequest(req ServiceRequest) (serviceEntity, error) {
	name := strings.ToLower(strings.TrimSpace(req.Name))
	if name == "" || len(name) > 128 {
		return serviceEntity{}, fmt.Errorf("%w: name is required and must be 1-128 chars", ErrInvalid)
	}

	if req.Port < 1 || req.Port > 65535 {
		return serviceEntity{}, fmt.Errorf("%w: port must be between 1 and 65535", ErrInvalid)
	}

	return serviceEntity{
		Name: name,
		Port: req.Port,
	}, nil
}

func validateRouteRequest(req RouteRequest) (routeEntity, error) {
	if req.ServiceID <= 0 {
		return routeEntity{}, fmt.Errorf("%w: service_id is required", ErrInvalid)
	}

	path := strings.TrimSpace(req.Path)
	if path == "" || !strings.HasPrefix(path, "/") {
		return routeEntity{}, fmt.Errorf("%w: path is required and must start with /", ErrInvalid)
	}
	if len(path) > 512 {
		return routeEntity{}, fmt.Errorf("%w: path must be at most 512 chars", ErrInvalid)
	}

	matchType, err := normalizeMatchType(req.MatchType)
	if err != nil {
		return routeEntity{}, err
	}

	return routeEntity{
		ServiceID: req.ServiceID,
		Path:      path,
		MatchType: matchType,
	}, nil
}

// normalizeMatchType accepts case-insensitive input and returns the canonical
// Gateway API PathMatchType value.
func normalizeMatchType(matchType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(matchType)) {
	case "exact":
		return MatchTypeExact, nil
	case "pathprefix":
		return MatchTypePathPrefix, nil
	default:
		return "", fmt.Errorf("%w: match_type must be Exact or PathPrefix", ErrInvalid)
	}
}
