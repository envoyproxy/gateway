// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package logging

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestZapLogLevel(t *testing.T) {
	level, err := zapcore.ParseLevel("warn")
	if err != nil {
		t.Errorf("ParseLevel error %v", err)
	}
	zc := zap.NewDevelopmentConfig()
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(zc.EncoderConfig), zapcore.AddSync(os.Stdout), zap.NewAtomicLevelAt(level))
	zapLogger := zap.New(core, zap.AddCaller())
	log := zapLogger.Sugar()
	log.Info("ok", "k1", "v1")
	log.Error(errors.New("new error"), "error")
}

func TestLogger(t *testing.T) {
	logger := NewLogger(v1alpha1.DefaultEnvoyGatewayLogging())
	logger.Info("kv msg", "key", "value")
	logger.Sugar().Infof("template %s %d", "string", 123)

	logger.WithName(string(v1alpha1.LogComponentGlobalRateLimitRunner)).WithValues("runner", v1alpha1.LogComponentGlobalRateLimitRunner).Info("msg", "k", "v")

	defaultLogger := DefaultLogger(v1alpha1.LogLevelInfo)
	assert.True(t, defaultLogger.logging != nil)
	assert.True(t, defaultLogger.sugaredLogger != nil)
}

func TestLoggerWithName(t *testing.T) {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		// Restore the original stdout and close the pipe
		os.Stdout = originalStdout
		err := w.Close()
		assert.NoError(t, err)
	}()

	config := v1alpha1.DefaultEnvoyGatewayLogging()
	config.Level[v1alpha1.LogComponentInfrastructureRunner] = v1alpha1.LogLevelDebug

	logger := NewLogger(config).WithName(string(v1alpha1.LogComponentInfrastructureRunner))
	logger.Info("info message")
	logger.Sugar().Debugf("debug message")

	// Read from the pipe (captured stdout)
	outputBytes := make([]byte, 200)
	_, err := r.Read(outputBytes)
	assert.NoError(t, err)
	capturedOutput := string(outputBytes)
	assert.Contains(t, capturedOutput, string(v1alpha1.LogComponentInfrastructureRunner))
	assert.Contains(t, capturedOutput, "info message")
	assert.Contains(t, capturedOutput, "debug message")
}
