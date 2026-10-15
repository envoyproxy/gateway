// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gateway

import "time"

// Match types supported by route.match_type (Gateway API PathMatchType).
const (
	MatchTypeExact      = "Exact"
	MatchTypePathPrefix = "PathPrefix"
)

// ClusterRequest is the input body for creating or updating a cluster.
type ClusterRequest struct {
	Name       string `json:"name"`
	InternalLB string `json:"internal_lb"`
}

// ClusterResponse is the API representation of a cluster row.
type ClusterResponse struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	InternalLB string    `json:"internal_lb"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ServiceRequest is the input body for creating or updating a service.
type ServiceRequest struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

// ServiceResponse is the API representation of a service row.
type ServiceResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Port      int       `json:"port"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RouteRequest is the input body for creating or updating a route.
type RouteRequest struct {
	ServiceID int64  `json:"service_id"`
	Path      string `json:"path"`
	MatchType string `json:"match_type"`
}

// RouteResponse is the API representation of a route row.
type RouteResponse struct {
	ID        int64     `json:"id"`
	ServiceID int64     `json:"service_id"`
	Path      string    `json:"path"`
	MatchType string    `json:"match_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// clusterEntity mirrors a cluster row.
type clusterEntity struct {
	ID         int64
	Name       string
	InternalLB string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (c clusterEntity) toResponse() ClusterResponse {
	return ClusterResponse{
		ID:         c.ID,
		Name:       c.Name,
		InternalLB: c.InternalLB,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

// serviceEntity mirrors a service row.
type serviceEntity struct {
	ID        int64
	Name      string
	Port      int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s serviceEntity) toResponse() ServiceResponse {
	return ServiceResponse{
		ID:        s.ID,
		Name:      s.Name,
		Port:      s.Port,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

// routeEntity mirrors a route row.
type routeEntity struct {
	ID        int64
	ServiceID int64
	Path      string
	MatchType string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r routeEntity) toResponse() RouteResponse {
	return RouteResponse{
		ID:        r.ID,
		ServiceID: r.ServiceID,
		Path:      r.Path,
		MatchType: r.MatchType,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
