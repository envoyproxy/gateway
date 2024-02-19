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
	// ShutdownReadyPort is the port Envoy shutdown manager will listen on.
	ShutdownReadyPort = 19002
	// ShutdownReadyFile is the file used to indicate shutdown readiness.
	ShutdownReadyFile = "/tmp/shutdown-ready"
)

// ShutdownManager serves shutdown manager process for Envoy proxies.
func ShutdownManager(readyTimeout time.Duration) error {
	// Setup HTTP handler
	handler := http.NewServeMux()
	handler.HandleFunc("/shutdown/ready", func(w http.ResponseWriter, r *http.Request) {
		shutdownReadyHandler(w, r, ShutdownReadyFile)
	})

	// Setup HTTP server
	srv := http.Server{
		Handler:           http.TimeoutHandler(handler, readyTimeout, ""),
		Addr:              fmt.Sprintf(":%d", ShutdownReadyPort),
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
// container to block until ready to terminate. After the graceful drain process
// has completed a file will be written to indicate shutdown readiness.
func shutdownReadyHandler(w http.ResponseWriter, req *http.Request, readyFile string) {
	logger.Info("received shutdown ready request")

	// Poll for shutdown readiness
	for {
		_, err := os.Stat(readyFile)
		switch {
		case os.IsNotExist(err):
			time.Sleep(1 * time.Second)
		case err != nil:
			logger.Error(err, "error checking for shutdown readiness")
		case err == nil:
			logger.Info("shutdown readiness detected")
			w.WriteHeader(http.StatusOK)
			return
		}
	}
}

// Shutdown is called from a preStop hook where it will block until envoy can
// gracefully drain open connections prior to pod shutdown.
func Shutdown(drainTimeout time.Duration, minDrainDuration time.Duration, exitAtConnections int) error {
	var startTime = time.Now()
	var allowedToExit = false

	// Reconfigure logger to write to stdout of main process if running in Kubernetes
	if _, k8s := os.LookupEnv("KUBERNETES_SERVICE_HOST"); k8s && os.Getpid() != 1 {
		logger = logging.FileLogger("/proc/1/fd/1", "shutdown-manager", v1alpha1.LogLevelInfo)
	}

	logger.Info(fmt.Sprintf("initiating graceful drain with %.0f second minimum drain period and %.0f second timeout",
		minDrainDuration.Seconds(), drainTimeout.Seconds()))

	// Send request to Envoy admin API to initiate the graceful drain sequence
	if resp, err := http.Post(fmt.Sprintf("http://%s:%d/drain_listeners?graceful&skip_exit",
		bootstrap.EnvoyAdminAddress, bootstrap.EnvoyAdminPort), "application/json", nil); err != nil {
		logger.Error(err, fmt.Sprintf("error %s", "initiating graceful drain"))
	} else {
		if resp.StatusCode != http.StatusOK {
			logger.Error(fmt.Errorf("unexpected response status: %s", resp.Status), fmt.Sprintf("error %s", "initiating graceful drain"))
		}
		resp.Body.Close()
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
			logger.Info(fmt.Sprintf("minimum drain period reached; will exit when total connections is %d", exitAtConnections))
			allowedToExit = true
		}

		if elapsedTime > drainTimeout {
			logger.Info("graceful drain sequence timeout exceeded")
			break
		} else if allowedToExit && conn != nil && *conn == exitAtConnections {
			logger.Info("graceful drain sequence completed")
			break
		}

		time.Sleep(1 * time.Second)
	}

	// Signal to /shutdown endpoint that the shutdown process is complete
	if err := createShutdownReadyFile(); err != nil {
		return err
	}

	return nil
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

// createShutdownReadyFile creates a file to indicate that the shutdown process is complete
func createShutdownReadyFile() error {
	if _, err := os.Create(ShutdownReadyFile); err != nil {
		logger.Error(err, "error creating shutdown ready file")
		return err
	}
	return nil
}
