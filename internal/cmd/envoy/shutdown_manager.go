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
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

// TODO: Remove the global logger and localize the scope of the logger.
var logger = logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo).WithName("shutdown-manager")

const (
	// ShutdownManagerPort is the port Envoy shutdown manager will listen on.
	ShutdownManagerPort = 19002
	// ShutdownManagerHealthCheckPath is the path used for health checks.
	ShutdownManagerHealthCheckPath = "/healthz"
	// ShutdownManagerReadyPath is the path used to indicate shutdown readiness.
	ShutdownManagerReadyPath = "/shutdown/ready"
	// ShutdownReadyFile is the file used to indicate shutdown readiness.
	ShutdownReadyFile = "/tmp/shutdown-ready"
)

// ShutdownManager serves shutdown manager process for Envoy proxies.
func ShutdownManager(readyTimeout time.Duration) error {
	// Setup HTTP handler
	handler := http.NewServeMux()
	handler.HandleFunc(ShutdownManagerHealthCheckPath, func(_ http.ResponseWriter, _ *http.Request) {})
	handler.HandleFunc(ShutdownManagerReadyPath, func(w http.ResponseWriter, _ *http.Request) {
		shutdownReadyHandler(w, readyTimeout, ShutdownReadyFile)
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
		logger.Info(fmt.Sprintf("received %s", (r.(syscall.Signal)).String()))

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
func shutdownReadyHandler(w http.ResponseWriter, readyTimeout time.Duration, readyFile string) {
	startTime := time.Now()

	logger.Info("received shutdown ready request")

	// Poll for shutdown readiness
	for {
		// Check if ready timeout is exceeded
		elapsedTime := time.Since(startTime)
		if elapsedTime > readyTimeout {
			logger.Info("shutdown readiness timeout exceeded")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err := os.Stat(readyFile)
		switch {
		case os.IsNotExist(err):
			time.Sleep(1 * time.Second)
		case err != nil:
			logger.Error(err, "error checking for shutdown readiness")
		default:
			logger.Info("shutdown readiness detected")
			return
		}
	}
}

// Shutdown is called from a preStop hook on the shutdown-manager container where
// it will initiate a drain sequence on the Envoy proxy and block until
// connections are drained or a timeout is exceeded.
func Shutdown(drainTimeout, minDrainDuration time.Duration, exitAtConnections int) error {
	startTime := time.Now()
	allowedToExit := false

	// Reconfigure logger to write to stdout of main process if running in Kubernetes
	if _, k8s := os.LookupEnv("KUBERNETES_SERVICE_HOST"); k8s && os.Getpid() != 1 {
		logger = logging.FileLogger("/proc/1/fd/1", "shutdown-manager", egv1a1.LogLevelInfo)
	}

	logger.Info(fmt.Sprintf("initiating drain with %.0f second minimum drain period and %.0f second timeout",
		minDrainDuration.Seconds(), drainTimeout.Seconds()))

	// Start failing active health checks
	if err := postEnvoyAdminAPI("healthcheck/fail"); err != nil {
		logger.Error(err, "error failing active health checks")
	}

	// Poll total connections from Envoy admin API until minimum drain period has
	// been reached and total connections reaches threshold or timeout is exceeded
	for {
		elapsedTime := time.Since(startTime)

		conn, err := getTotalConnections(bootstrap.EnvoyAdminPort)
		if err != nil {
			logger.Error(err, "error getting total connections")
		}

		if elapsedTime > minDrainDuration && !allowedToExit {
			logger.Info(fmt.Sprintf("minimum drain period reached; will exit when total connections reaches %d", exitAtConnections))
			allowedToExit = true
		}

		if elapsedTime > drainTimeout {
			logger.Info("drain sequence timeout exceeded")
			break
		} else if allowedToExit && conn != nil && *conn <= exitAtConnections {
			logger.Info("drain sequence completed")
			break
		}

		time.Sleep(1 * time.Second)
	}

	// Signal to shutdownReadyHandler that drain process is complete
	if _, err := os.Create(ShutdownReadyFile); err != nil {
		logger.Error(err, "error creating shutdown ready file")
		return err
	}

	return nil
}

// postEnvoyAdminAPI sends a POST request to the Envoy admin API
func postEnvoyAdminAPI(path string) error {
	resp, err := http.Post(fmt.Sprintf("http://%s:%d/%s",
		"localhost", bootstrap.EnvoyAdminPort, path), "application/json", nil)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("unexcepted nil response from Envoy admin API")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}
	return nil
}

func getTotalConnections(port int) (*int, error) {
	return getDownstreamCXActive(port)
}

// Define struct to decode JSON response into; expecting a single stat in the response in the format:
// {"stats":[{"name":"server.total_connections","value":123}]}
type envoyStatsResponse struct {
	Stats []struct {
		Name  string
		Value int
	}
}

func getStatsFromEnvoyStatsEndpoint(port int, statFilter string) (*envoyStatsResponse, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s//stats?filter=%s&format=json",
		net.JoinHostPort("localhost", strconv.Itoa(port)), statFilter))
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	r := &envoyStatsResponse{}
	// Decode JSON response into struct
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	// Defensive check for empty stats
	if len(r.Stats) == 0 {
		return nil, fmt.Errorf("no stats found")
	}

	return r, nil
}

// getDownstreamCXActive retrieves the total number of open connections from Envoy's listener downstream_cx_active stat
func getDownstreamCXActive(port int) (*int, error) {
	// Send request to Envoy admin API to retrieve listener.\.$.downstream_cx_active stat
	statFilter := "^listener\\..*\\.downstream_cx_active$"
	r, err := getStatsFromEnvoyStatsEndpoint(port, statFilter)
	if err != nil {
		return nil, fmt.Errorf("error getting listener downstream_cx_active stat: %w", err)
	}

	totalConnection := filterDownstreamCXActive(r)
	logger.Info(fmt.Sprintf("total downstream connections: %d", *totalConnection))
	return totalConnection, nil
}

// skipConnectionRE is a regex to match connection stats to be excluded from total connections count
// e.g. admin, ready and stat listener and stats from worker thread
var skipConnectionRE = regexp.MustCompile(`admin|19001|19003|worker`)

func filterDownstreamCXActive(r *envoyStatsResponse) *int {
	totalConnection := 0
	for _, stat := range r.Stats {
		if excluded := skipConnectionRE.MatchString(stat.Name); !excluded {
			totalConnection += stat.Value
		}
	}

	return &totalConnection
}
