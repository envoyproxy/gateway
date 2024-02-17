// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoy

import (
	"context"
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
)

var (
	logger = logging.DefaultLogger(v1alpha1.LogLevelInfo).WithName("shutdown-manager")
)

const (
	// ShutdownReadyPort is the port Envoy shutdown manager will listen on.
	ShutdownReadyPort = 19002
	// ShutdownReadyPath is the path Envoy shutdown manager will listen on.
	ShutdownReadyPath = "/shutdown"
	// ShutdownReadyFile is the file used to indicate shutdown readiness.
	ShutdownReadyFile = "/tmp/shutdown-ready"
)

// Shutdown is called from a preStop hook where it will block until envoy can
// gracefully drain open connections prior to pod shutdown.
func Shutdown(timeout time.Duration, minDrainDuration time.Duration, exitAtConnections int) error {
	logger.Info(fmt.Sprintf("initiating graceful drain with %.0f second minimum drain period and %.0f second timeout",
		minDrainDuration.Seconds(), timeout.Seconds()))

	time.Sleep(minDrainDuration) // TODO: Implement graceful shutdown

	if _, err := os.Create(ShutdownReadyFile); err != nil {
		logger.Error(err, "error creating shutdown ready file")
		return err
	}

	return nil
}

// ShutdownManager serves shutdown manager process for Envoy proxies.
func ShutdownManager() error {
	// Setup HTTP handler
	handler := http.NewServeMux()
	handler.HandleFunc(ShutdownReadyPath, func(w http.ResponseWriter, r *http.Request) {
		shutdownReadyHandler(w, r, ShutdownReadyFile)
	})

	// Setup HTTP server
	srv := http.Server{
		Handler:           handler,
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
	for {
		_, err := os.Stat(readyFile)
		switch {
		case os.IsNotExist(err):
			time.Sleep(1 * time.Second)
		case err != nil:
			logger.Error(err, "error checking for shutdown readiness")
		case err == nil:
			logger.Info("shutdown readiness detected")
			return
		}
	}
}
