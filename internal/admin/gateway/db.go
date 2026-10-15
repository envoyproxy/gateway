// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package gateway implements the route-management APIs that manage the
// cluster, service and route tables and generate Envoy route files from
// them. See docs/route_generation_schema.sql for the schema.
package gateway

import (
	"database/sql"
	"fmt"
	"time"

	// Register the pgx driver with database/sql under the name "pgx".
	_ "github.com/jackc/pgx/v5/stdlib"
)

// InitDB opens a connection pool to PostgreSQL using the given DSN.
// The DSN is a standard libpq/pgx connection string or URL, e.g.
// "postgres://user:pass@host:5432/dbname".
func InitDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open gateway database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping gateway database: %w", err)
	}

	return db, nil
}
