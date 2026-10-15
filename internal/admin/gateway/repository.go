// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gateway

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// Sentinel errors returned by the repository and mapped to HTTP statuses by the
// service layer.
var (
	// ErrNotFound is returned when an active row does not exist for the given id.
	ErrNotFound = errors.New("resource not found")
	// ErrConflict is returned when a unique constraint is violated (duplicate
	// active cluster name, service name, or route service_id+path+match_type).
	ErrConflict = errors.New("resource already exists")
)

// isUniqueViolation reports whether err is a PostgreSQL unique-constraint
// violation (SQLSTATE 23505). The partial unique indexes uq_service_active and
// uq_route_active surface as this error.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// pgx surfaces the SQLSTATE in the error string; checking the code keeps the
	// repository decoupled from the concrete pgconn.PgError type.
	return strings.Contains(err.Error(), "23505") ||
		strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}

// Repository provides CRUD access to the cluster, service and route tables.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a Repository backed by the given database handle.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// --- cluster ---

// CreateCluster inserts a new active cluster row and returns the stored entity.
func (r *Repository) CreateCluster(ctx context.Context, e clusterEntity) (clusterEntity, error) {
	const q = `
		INSERT INTO cluster (name, internal_lb)
		VALUES ($1, $2)
		RETURNING id, name, internal_lb, created_at, updated_at`
	out, err := scanCluster(r.db.QueryRowContext(ctx, q, e.Name, e.InternalLB))
	if err != nil {
		if isUniqueViolation(err) {
			return clusterEntity{}, ErrConflict
		}
		return clusterEntity{}, err
	}
	return out, nil
}

// GetCluster returns the active cluster with the given id.
func (r *Repository) GetCluster(ctx context.Context, id int64) (clusterEntity, error) {
	const q = `
		SELECT id, name, internal_lb, created_at, updated_at
		FROM cluster
		WHERE id = $1 AND deleted = 0`
	out, err := scanCluster(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return clusterEntity{}, ErrNotFound
	}
	return out, err
}

// GetClusterByName returns the active cluster with the given name.
func (r *Repository) GetClusterByName(ctx context.Context, name string) (clusterEntity, error) {
	const q = `
		SELECT id, name, internal_lb, created_at, updated_at
		FROM cluster
		WHERE name = $1 AND deleted = 0`
	out, err := scanCluster(r.db.QueryRowContext(ctx, q, name))
	if errors.Is(err, sql.ErrNoRows) {
		return clusterEntity{}, ErrNotFound
	}
	return out, err
}

// ListClusters returns all active clusters ordered by name.
func (r *Repository) ListClusters(ctx context.Context) ([]clusterEntity, error) {
	const q = `
		SELECT id, name, internal_lb, created_at, updated_at
		FROM cluster
		WHERE deleted = 0
		ORDER BY name`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []clusterEntity
	for rows.Next() {
		out, err := scanCluster(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, out)
	}
	return result, rows.Err()
}

// UpdateCluster updates an active cluster in place and returns the stored entity.
func (r *Repository) UpdateCluster(ctx context.Context, e clusterEntity) (clusterEntity, error) {
	const q = `
		UPDATE cluster
		SET name = $2, internal_lb = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted = 0
		RETURNING id, name, internal_lb, created_at, updated_at`
	out, err := scanCluster(r.db.QueryRowContext(ctx, q, e.ID, e.Name, e.InternalLB))
	if errors.Is(err, sql.ErrNoRows) {
		return clusterEntity{}, ErrNotFound
	}
	if err != nil {
		if isUniqueViolation(err) {
			return clusterEntity{}, ErrConflict
		}
		return clusterEntity{}, err
	}
	return out, nil
}

// DeleteCluster soft-deletes a cluster.
func (r *Repository) DeleteCluster(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE cluster SET deleted = 1, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $1 AND deleted = 0`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// --- service ---

// CreateService inserts a new active service row and returns the stored entity.
func (r *Repository) CreateService(ctx context.Context, e serviceEntity) (serviceEntity, error) {
	const q = `
		INSERT INTO service (name, port)
		VALUES ($1, $2)
		RETURNING id, name, port, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, q, e.Name, e.Port)
	out, err := scanService(row)
	if err != nil {
		if isUniqueViolation(err) {
			return serviceEntity{}, ErrConflict
		}
		return serviceEntity{}, err
	}
	return out, nil
}

// GetService returns the active service with the given id.
func (r *Repository) GetService(ctx context.Context, id int64) (serviceEntity, error) {
	const q = `
		SELECT id, name, port, created_at, updated_at
		FROM service
		WHERE id = $1 AND deleted = 0`
	out, err := scanService(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return serviceEntity{}, ErrNotFound
	}
	return out, err
}

// ListServices returns active services, optionally filtered by name.
func (r *Repository) ListServices(ctx context.Context, name string) ([]serviceEntity, error) {
	const q = `
		SELECT id, name, port, created_at, updated_at
		FROM service
		WHERE deleted = 0
		  AND ($1 = '' OR name = $1)
		ORDER BY name`
	rows, err := r.db.QueryContext(ctx, q, name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []serviceEntity
	for rows.Next() {
		out, err := scanService(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, out)
	}
	return result, rows.Err()
}

// UpdateService updates an active service in place and returns the stored entity.
func (r *Repository) UpdateService(ctx context.Context, e serviceEntity) (serviceEntity, error) {
	const q = `
		UPDATE service
		SET name = $2, port = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted = 0
		RETURNING id, name, port, created_at, updated_at`
	out, err := scanService(r.db.QueryRowContext(ctx, q, e.ID, e.Name, e.Port))
	if errors.Is(err, sql.ErrNoRows) {
		return serviceEntity{}, ErrNotFound
	}
	if err != nil {
		if isUniqueViolation(err) {
			return serviceEntity{}, ErrConflict
		}
		return serviceEntity{}, err
	}
	return out, nil
}

// DeleteService soft-deletes a service and cascade soft-deletes all of its
// active routes in a single transaction.
func (r *Repository) DeleteService(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Cascade: soft-delete the service's routes first.
	if _, err := tx.ExecContext(ctx,
		`UPDATE route SET deleted = 1, updated_at = CURRENT_TIMESTAMP
		 WHERE service_id = $1 AND deleted = 0`, id); err != nil {
		return err
	}

	// Soft-delete the service itself.
	res, err := tx.ExecContext(ctx,
		`UPDATE service SET deleted = 1, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $1 AND deleted = 0`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}

	return tx.Commit()
}

// ServiceExists reports whether an active service with the given id exists.
func (r *Repository) ServiceExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM service WHERE id = $1 AND deleted = 0)`, id).
		Scan(&exists)
	return exists, err
}

// --- route ---

// CreateRoute inserts a new active route row and returns the stored entity.
func (r *Repository) CreateRoute(ctx context.Context, e routeEntity) (routeEntity, error) {
	const q = `
		INSERT INTO route (service_id, path, match_type)
		VALUES ($1, $2, $3)
		RETURNING id, service_id, path, match_type, created_at, updated_at`
	out, err := scanRoute(r.db.QueryRowContext(ctx, q, e.ServiceID, e.Path, e.MatchType))
	if err != nil {
		if isUniqueViolation(err) {
			return routeEntity{}, ErrConflict
		}
		return routeEntity{}, err
	}
	return out, nil
}

// GetRoute returns the active route with the given id.
func (r *Repository) GetRoute(ctx context.Context, id int64) (routeEntity, error) {
	const q = `
		SELECT id, service_id, path, match_type, created_at, updated_at
		FROM route
		WHERE id = $1 AND deleted = 0`
	out, err := scanRoute(r.db.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return routeEntity{}, ErrNotFound
	}
	return out, err
}

// ListRoutes returns active routes, optionally filtered by service_id and/or match_type.
// A serviceID of 0 means "any service".
func (r *Repository) ListRoutes(ctx context.Context, serviceID int64, matchType string) ([]routeEntity, error) {
	const q = `
		SELECT id, service_id, path, match_type, created_at, updated_at
		FROM route
		WHERE deleted = 0
		  AND ($1 = 0 OR service_id = $1)
		  AND ($2 = '' OR match_type = $2)
		ORDER BY service_id, path`
	rows, err := r.db.QueryContext(ctx, q, serviceID, matchType)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []routeEntity
	for rows.Next() {
		out, err := scanRoute(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, out)
	}
	return result, rows.Err()
}

// UpdateRoute updates an active route in place and returns the stored entity.
func (r *Repository) UpdateRoute(ctx context.Context, e routeEntity) (routeEntity, error) {
	const q = `
		UPDATE route
		SET service_id = $2, path = $3, match_type = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted = 0
		RETURNING id, service_id, path, match_type, created_at, updated_at`
	out, err := scanRoute(r.db.QueryRowContext(ctx, q, e.ID, e.ServiceID, e.Path, e.MatchType))
	if errors.Is(err, sql.ErrNoRows) {
		return routeEntity{}, ErrNotFound
	}
	if err != nil {
		if isUniqueViolation(err) {
			return routeEntity{}, ErrConflict
		}
		return routeEntity{}, err
	}
	return out, nil
}

// DeleteRoute soft-deletes a single route.
func (r *Repository) DeleteRoute(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE route SET deleted = 1, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $1 AND deleted = 0`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// --- generation support ---

// ListRoutesForService returns all active routes for a service, ordered so that
// Exact matches precede PathPrefix matches (and then by path).
func (r *Repository) ListRoutesForService(ctx context.Context, serviceID int64) ([]routeEntity, error) {
	const q = `
		SELECT id, service_id, path, match_type, created_at, updated_at
		FROM route
		WHERE service_id = $1 AND deleted = 0
		ORDER BY CASE match_type WHEN 'Exact' THEN 0 ELSE 1 END, path`
	rows, err := r.db.QueryContext(ctx, q, serviceID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []routeEntity
	for rows.Next() {
		out, err := scanRoute(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, out)
	}
	return result, rows.Err()
}

// rowScanner is implemented by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanCluster(s rowScanner) (clusterEntity, error) {
	var e clusterEntity
	err := s.Scan(&e.ID, &e.Name, &e.InternalLB, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func scanService(s rowScanner) (serviceEntity, error) {
	var e serviceEntity
	err := s.Scan(&e.ID, &e.Name, &e.Port, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func scanRoute(s rowScanner) (routeEntity, error) {
	var e routeEntity
	err := s.Scan(&e.ID, &e.ServiceID, &e.Path, &e.MatchType, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}
