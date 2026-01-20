// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package logging

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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
	logger := NewLogger(os.Stdout, egv1a1.DefaultEnvoyGatewayLogging())
	logger.Info("kv msg", "key", "value")
	logger.Sugar().Infof("template %s %d", "string", 123)

	logger.WithName(string(egv1a1.LogComponentGlobalRateLimitRunner)).WithValues("runner", egv1a1.LogComponentGlobalRateLimitRunner).Info("msg", "k", "v")

	defaultLogger := DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)
	assert.NotNil(t, defaultLogger.logging)
	assert.NotNil(t, defaultLogger.sugaredLogger)

	fileLogger := FileLogger("/dev/stderr", "fl-test", egv1a1.LogLevelInfo)
	assert.NotNil(t, fileLogger.logging)
	assert.NotNil(t, fileLogger.sugaredLogger)
}

func TestLoggerWithName(t *testing.T) {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		// Restore the original stdout and close the pipe
		os.Stdout = originalStdout
		err := w.Close()
		require.NoError(t, err)
	}()

	config := egv1a1.DefaultEnvoyGatewayLogging()
	config.Level[egv1a1.LogComponentInfrastructureRunner] = egv1a1.LogLevelDebug

	logger := NewLogger(os.Stdout, config).WithName(string(egv1a1.LogComponentInfrastructureRunner))
	logger.Info("info message")
	logger.Sugar().Debugf("debug message")

	// Read from the pipe (captured stdout)
	outputBytes := make([]byte, 200)
	_, err := r.Read(outputBytes)
	require.NoError(t, err)
	capturedOutput := string(outputBytes)
	assert.Contains(t, capturedOutput, string(egv1a1.LogComponentInfrastructureRunner))
	assert.Contains(t, capturedOutput, "info message")
	assert.Contains(t, capturedOutput, "debug message")
}

func TestLoggerSugarName(t *testing.T) {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		// Restore the original stdout and close the pipe
		os.Stdout = originalStdout
		err := w.Close()
		require.NoError(t, err)
	}()

	const logName = "loggerName"

	config := egv1a1.DefaultEnvoyGatewayLogging()
	config.Level[logName] = egv1a1.LogLevelDebug

	logger := NewLogger(os.Stdout, config).WithName(logName)

	logger.Sugar().Debugf("debugging message")

	// Read from the pipe (captured stdout)
	outputBytes := make([]byte, 200)
	_, err := r.Read(outputBytes)
	require.NoError(t, err)
	capturedOutput := string(outputBytes)
	assert.Contains(t, capturedOutput, "debugging message", logName)
}

func TestLoggerWithTrace(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := NewLogger(buffer, egv1a1.DefaultEnvoyGatewayLogging())

	traceID := trace.TraceID{0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe, 0xba, 0xbe, 0xfa, 0xce, 0xb0, 0x0c, 0x12, 0x34, 0x56, 0x78}
	spanID := trace.SpanID{0xba, 0xad, 0xf0, 0x0d, 0xfe, 0xed, 0xfa, 0xce}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	logger.WithTrace(ctx).Info("hello tracing")

	output := buffer.String()
	assert.Contains(t, output, traceID.String())
	assert.Contains(t, output, spanID.String())
}
