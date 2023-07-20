// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package runner provides an interface and a generic implementation for running tasks.
package runner

import (
	"context"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

const (
	// RootParentRunner is the root parent runner name.
	RootParentRunner = "root-runner"
)

// Runner is an interface that defines methods for starting, shutting down, and subscribing to a runner.
type Runner interface {
	// Name returns the name of the runner.
	Name() string
	// GetParent returns the name of the parent runner.
	GetParent() string
	// SetParent sets the name of the parent runner.
	SetParent(parent string)
	// GetChildren returns the list of the child runner.
	GetChildren() map[string]Runner
	// GetChildrenNames returns the names of the child runner.
	GetChildrenNames() []string
	// AddChildren add the name of the child runner.
	AddChildren(runner Runner)
	// RemoveChildren removes the name of the child runner.
	RemoveChildren(name string)
	// Status shows the statuses of the runner.
	Status() []any
	// Start starts the runner with the provided context.
	Start(ctx context.Context) error
	// ShutDown shuts down the runner with the provided context.
	ShutDown(ctx context.Context)
	// SubscribeAndTranslate handles subscription and translation for the runner.
	SubscribeAndTranslate(ctx context.Context)
}

// GenericRunner is a generic implementation of the Runner interface.
// It embeds a GenericConfig and a string representing the name of the runner.
type GenericRunner[T any] struct {
	GenericConfig[T]                   // GenericConfig is a generic configuration for the runner.
	Runner           string            // Runner is the name of the runner.
	Parent           string            // ParentRunner is the name of the parent-runner.
	Children         map[string]Runner // ChildrenRunners is the names of the child-runners.
}

// GenericConfig is a generic configuration structure for a runner.
// It embeds a Server configuration and a generic Resources configuration.
type GenericConfig[T any] struct {
	config.Server   // Server is the server configuration.
	Resources     T // Resources is the resources configuration.
}

// New creates a new GenericRunner with the provided name, resources, and global configuration.
func New[T any](name string, resources T, globalConfig config.Server) *GenericRunner[T] {
	return &GenericRunner[T]{
		GenericConfig: GenericConfig[T]{
			globalConfig,
			resources,
		},
		Runner:   name,
		Children: map[string]Runner{},
	}
}

// Name returns the name of the GenericRunner.
func (r *GenericRunner[T]) Name() string {
	return r.Runner
}

// Parent returns the name of the parent GenericRunner.
func (r *GenericRunner[T]) GetParent() string {
	return r.Parent
}

// SetParent sets the name of the parent runner.
func (r *GenericRunner[T]) SetParent(parent string) {
	r.Parent = parent
}

// Children returns the name of the children GenericRunner.
func (r *GenericRunner[T]) GetChildren() map[string]Runner {
	return r.Children
}

// GetChildrenNames returns the names of the child runner.
func (r *GenericRunner[T]) GetChildrenNames() []string {
	names := []string{}
	for k := range r.GetChildren() {
		names = append(names, k)
	}
	return names
}

// AddChildren add the name of the child runner.
func (r *GenericRunner[T]) AddChildren(runner Runner) {
	r.Children[runner.Name()] = runner
}

// RemoveChildren removes the name of the child runner.
func (r *GenericRunner[T]) RemoveChildren(name string) {
	delete(r.Children, name)
}

// Status shows the statuses of the runner.
func (r *GenericRunner[T]) Status() []any {
	return []any{"name", r.Name(), "parent_runner", r.GetParent(), "children_runners", r.GetChildrenNames()}
}

// Config returns the GenericConfig of the GenericRunner.
func (r *GenericRunner[T]) Config() GenericConfig[T] {
	return r.GenericConfig
}

// Init initializes the GenericRunner with the provided context.
// It also sets the logger name and values for the runner.
func (r *GenericRunner[T]) Init(ctx context.Context) {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name(), "parent", r.GetParent())
}
