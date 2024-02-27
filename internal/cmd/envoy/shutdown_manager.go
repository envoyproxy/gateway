// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

var (
	logger = logging.DefaultLogger(v1alpha1.LogLevelInfo).WithName("shutdown-manager")
)

const (
	// ShutdownManagerPort is the port Envoy shutdown manager will listen on.
	ShutdownManagerPort = 19002
	// ShutdownManagerHealthCheckPath is the path used for health checks.
	ShutdownManagerHealthCheckPath = "/healthz"
	// ShutdownManagerReadyPath is the path used to indicate shutdown readiness.
	ShutdownManagerReadyPath = "/shutdown/ready"
	// ShutdownActiveFile is the file used to indicate shutdown readiness.
	ShutdownActiveFile = "/tmp/shutdown-active"
)

// ShutdownManager serves shutdown manager process for Envoy proxies.
func ShutdownManager(readyTimeout time.Duration) error {
	// Setup HTTP handler
	handler := http.NewServeMux()
	handler.HandleFunc(ShutdownManagerHealthCheckPath, func(_ http.ResponseWriter, _ *http.Request) {})
	handler.HandleFunc(ShutdownManagerReadyPath, func(w http.ResponseWriter, _ *http.Request) {
		shutdownReadyHandler(w, readyTimeout, ShutdownActiveFile)
	})

	// Setup HTTP server
	srv := http.Server{
		Handler:           handler,
		Addr:              fmt.Sprintf(":%d", ShutdownManagerPort),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}

	// Setup signal handling
	c := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, os.Interrupt, syscall.SIGTERM)

		r := <-s
		logger.Info(fmt.Sprintf("received %s", unix.SignalName(r.(syscall.Signal))))

		// Shutdown HTTP server without interrupting active connections
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Error(err, "server shutdown error")
		}
		close(c)
	}()

	// Start HTTP server
	logger.Info("starting shutdown manager")
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Error(err, "starting shutdown manager failed")
	}

	// Wait until done
	<-c
	return nil
}

// shutdownReadyHandler handles the endpoint used by a preStop hook on the Envoy
// container to block until ready to terminate. When the graceful drain process
// has begun a file will be written to indicate shutdown is in progress. This
// handler will poll for the file and block until it is removed.
func shutdownReadyHandler(w http.ResponseWriter, readyTimeout time.Duration, activeFile string) {
	var startTime = time.Now()

	logger.Info("received shutdown ready request")

	// Since we cannot establish order between preStop hooks, wait a bit giving the exec
	// preStop hook a chance to touch the signaling file before starting to poll for it.
	time.Sleep(2 * time.Second)

	// Exit early if active file failed to show up
	if _, err := os.Stat(activeFile); os.IsNotExist(err) {
		logger.Info("shutdown active file not found")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Poll for shutdown readiness
	for {
		// Check if ready timeout is exceeded
		elapsedTime := time.Since(startTime)
		if elapsedTime > readyTimeout {
			logger.Info("shutdown readiness timeout exceeded")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err := os.Stat(activeFile)
		switch {
		case os.IsNotExist(err):
			logger.Info("shutdown readiness detected")
			return
		case err != nil:
			logger.Error(err, "error checking for shutdown readiness")
		case err == nil:
			time.Sleep(1 * time.Second)
		}
	}
}

// Shutdown is called from a preStop hook on the shutdown-manager container where
// it will initiate a graceful drain sequence on the Envoy proxy and block until
// connections are drained or a timeout is exceeded.
func Shutdown(drainTimeout time.Duration, minDrainDuration time.Duration, exitAtConnections int) error {
	var startTime = time.Now()
	var allowedToExit = false

	// Reconfigure logger to write to stdout of main process if running in Kubernetes
	if _, k8s := os.LookupEnv("KUBERNETES_SERVICE_HOST"); k8s && os.Getpid() != 1 {
		logger = logging.FileLogger("/proc/1/fd/1", "shutdown-manager", v1alpha1.LogLevelInfo)
	}

	// Signal to shutdownReadyHandler that drain process is started
	if _, err := os.Create(ShutdownActiveFile); err != nil {
		logger.Error(err, "error creating shutdown active file")
		return err
	}

	logger.Info(fmt.Sprintf("initiating graceful drain with %.0f second minimum drain period and %.0f second timeout",
		minDrainDuration.Seconds(), drainTimeout.Seconds()))

	// Start failing active health checks
	if err := postEnvoyAdminAPI("healthcheck/fail"); err != nil {
		logger.Error(err, "error failing active health checks")
	}

	// Initiate graceful drain sequence
	if err := postEnvoyAdminAPI("drain_listeners?graceful&skip_exit"); err != nil {
		logger.Error(err, "error initiating graceful drain")
	}

	// Poll total connections from Envoy admin API until minimum drain period has
	// been reached and total connections reaches threshold or timeout is exceeded
	for {
		elapsedTime := time.Since(startTime)

		conn, err := getTotalConnections()
		if err != nil {
			logger.Error(err, "error getting total connections")
		}

		if elapsedTime > minDrainDuration && !allowedToExit {
			logger.Info(fmt.Sprintf("minimum drain period reached; will exit when total connections reaches %d", exitAtConnections))
			allowedToExit = true
		}

		if elapsedTime > drainTimeout {
			logger.Info("graceful drain sequence timeout exceeded")
			break
		} else if allowedToExit && conn != nil && *conn <= exitAtConnections {
			logger.Info("graceful drain sequence completed")
			break
		}

		time.Sleep(1 * time.Second)
	}

	// Signal to shutdownReadyHandler that drain process is complete
	if err := os.Remove(ShutdownActiveFile); err != nil {
		logger.Error(err, "error removing shutdown active file")
		return err
	}

	return nil
}

// postEnvoyAdminAPI sends a POST request to the Envoy admin API
func postEnvoyAdminAPI(path string) error {
	if resp, err := http.Post(fmt.Sprintf("http://%s:%d/%s",
		bootstrap.EnvoyAdminAddress, bootstrap.EnvoyAdminPort, path), "application/json", nil); err != nil {
		return err
	} else {
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected response status: %s", resp.Status)
		}
		return nil
	}
}

// getTotalConnections retrieves the total number of open connections from Envoy's server.total_connections stat
func getTotalConnections() (*int, error) {
	// Send request to Envoy admin API to retrieve server.total_connections stat
	if resp, err := http.Get(fmt.Sprintf("http://%s:%d//stats?filter=^server\\.total_connections$&format=json",
		bootstrap.EnvoyAdminAddress, bootstrap.EnvoyAdminPort)); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
		} else {
			// Define struct to decode JSON response into; expecting a single stat in the response in the format:
			// {"stats":[{"name":"server.total_connections","value":123}]}
			var r *struct {
				Stats []struct {
					Name  string
					Value int
				}
			}

			// Decode JSON response into struct
			if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
				return nil, err
			}

			// Defensive check for empty stats
			if len(r.Stats) == 0 {
				return nil, fmt.Errorf("no stats found")
			}

			// Log and return total connections
			c := r.Stats[0].Value
			logger.Info(fmt.Sprintf("total connections: %d", c))
			return &c, nil
		}
	}
}
