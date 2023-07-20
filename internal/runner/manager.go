// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

// RunnerManager is an interface that defines methods for managing runners.
type RunnerManager interface {
	// Initializes the runner manager with a server configuration.
	Init(conf config.Server)
	// Registers one or more runners.
	Register(runner Runner, parent string)
	// Starts all runners and waits for them to finish.
	Run(ctx context.Context) error
	// Starts all registered runners.
	StartAll(ctx context.Context) error
	// Starts a specific runner by name.
	Start(ctx context.Context, name string) error
	// ShutDown Shuts down a specific runner by name.
	ShutDown(ctx context.Context, name string)
	// ShutDownAll Shuts down all runners.
	ShutDownAll(ctx context.Context)
	// Remove removes a specific runner by name from manager.
	Remove(ctx context.Context, name string)
	// RemoveAll removes all runners from manager.
	RemoveAll(ctx context.Context)
	// Retrieves a runner by name.
	Get(name string) Runner
	// Lists all registered runners.
	List() []Runner
	// Lists the names of all registered runners.
	ListNames() []string
	// Status shows the statuses of all registered runners.
	Status() []any
}

// genericManager is a concrete implementation of the RunnerManager interface.
type genericManager struct {
	runners       []Runner // A slice to hold all registered runners.
	config.Server          // Embedded server configuration.
}

// runnerManager is a global instance of the genericManager.
var runnerManager RunnerManager = &genericManager{runners: []Runner{}}

// Manager returns the global instance of the runner genericManager.
func Manager() RunnerManager {
	return runnerManager
}

// Init initializes the runner genericManager with a server configuration.
func (m *genericManager) Init(conf config.Server) {
	m.Server = conf
	m.Logger = m.Logger.WithName("eg-manager")

	m.Logger.Info("initialized eg-manager")
}

// Register registers one or more runners.
func (m *genericManager) Register(r Runner, parent string) {
	if parent != RootParentRunner {
		parentRunner := m.Get(parent)
		parentRunner.AddChildren(r)
	}
	r.SetParent(parent)
	m.runners = append(m.runners, r)
	m.Logger.Info("registered runner", "status", r.Status())
}

// Run starts all runners and waits for them to finish.
func (m *genericManager) Run(ctx context.Context) error {
	if err := m.StartAll(ctx); err != nil {
		return err
	}

	// Wait until done
	<-ctx.Done()
	m.ShutDownAll(ctx)

	return nil
}

// StartAll starts all registered runners.
func (m *genericManager) StartAll(ctx context.Context) error {
	m.Logger.Info("starting all runners", "status", m.Status())
	for _, r := range m.runners {
		if err := m.Start(ctx, r.Name()); err != nil {
			return err
		}
	}

	m.Logger.Info("started all runners", "status", m.Status())

	return nil
}

// Start starts a specific runner by name.
func (m *genericManager) Start(ctx context.Context, name string) error {
	runner := m.Get(name)

	m.Logger.Info("starting runner", "status", runner.Status())
	if runner == nil {
		return fmt.Errorf("runner is not existed: %s", name)
	}
	m.Logger.Info("started runner", "status", runner.Status())
	return runner.Start(ctx)
}

// ShutDownAll shuts down all runners.
func (m *genericManager) ShutDownAll(ctx context.Context) {
	m.Logger.Info("shutdown all runners", "status", m.Status())
	for _, r := range m.runners {
		m.ShutDown(ctx, r.Name())
	}
}

// ShutDown shuts down a specific runner by name.
func (m *genericManager) ShutDown(ctx context.Context, name string) {
	runner := m.Get(name)

	if runner != nil {
		m.Logger.Info("shutdown runner", "status", runner.Status())
		runner.ShutDown(ctx)
		// Iterated shutdown all children runners
		if len(runner.GetChildrenNames()) != 0 {
			for _, c := range runner.GetChildrenNames() {
				m.ShutDown(ctx, c)
			}
		}
	}
}

// RemoveAll removes all runners from genericManager.
func (m *genericManager) RemoveAll(ctx context.Context) {
	m.ShutDownAll(ctx)
	m.Logger.Info("remove all runners", "status", m.Status())
	m.runners = []Runner{}
}

// Remove removes a specific runner by name from genericManager.
func (m *genericManager) Remove(ctx context.Context, name string) {
	for index, runner := range m.runners {
		if runner.Name() == name {
			m.ShutDown(ctx, name)
			m.Logger.Info("remove runner", "status", runner.Status())
			m.runners = append(m.runners[:index], m.runners[index+1:]...)
			// Iterated removes all children runners
			if len(runner.GetChildrenNames()) != 0 {
				for _, c := range runner.GetChildrenNames() {
					m.Remove(ctx, c)
				}
			}
		}
	}
}

// Status shows the statuses of all registered runners.
func (m *genericManager) Status() []any {
	statuses := []any{"current_live_runnners", m.ListNames(), "total_count", fmt.Sprint(len(m.List()))}
	for _, r := range m.List() {
		statuses = append(statuses, r.Name(), r.Status())
	}
	return statuses
}

// Get retrieves a runner by name.
func (m *genericManager) Get(name string) Runner {
	for _, r := range m.runners {
		if r.Name() == name {
			return r
		}
	}

	return nil
}

// List lists all registered runners.
func (m *genericManager) List() []Runner {
	return m.runners
}

// ListNames lists the names of all registered runners.
func (m *genericManager) ListNames() []string {
	names := []string{}
	for _, r := range m.runners {
		names = append(names, r.Name())
	}

	return names
}
