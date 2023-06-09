// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package logging

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
)

type Logger struct {
	logr.Logger
	logging       *v1alpha1.EnvoyGatewayLogging
	sugaredLogger *zap.SugaredLogger
}

func NewLogger(logging *v1alpha1.EnvoyGatewayLogging) Logger {
	logger := initZapLogger(logging.Level[v1alpha1.LogComponentGatewayDefault])

	return Logger{
		Logger:        zapr.NewLogger(logger),
		logging:       logging,
		sugaredLogger: logger.Sugar(),
	}
}

func DefaultLogger(level v1alpha1.LogLevel) Logger {
	logging := v1alpha1.DefaultEnvoyGatewayLogging()
	logger := initZapLogger(level)

	return Logger{
		Logger:        zapr.NewLogger(logger),
		logging:       logging,
		sugaredLogger: logger.Sugar(),
	}
}

// WithName returns a new Logger instance with the specified name element added
// to the Logger's name.  Successive calls with WithName append additional
// suffixes to the Logger's name.  It's strongly recommended that name segments
// contain only letters, digits, and hyphens (see the package documentation for
// more information).
func (l Logger) WithName(name string) Logger {
	logLevel := l.logging.Level[v1alpha1.EnvoyGatewayLogComponent(name)]
	logger := initZapLogger(v1alpha1.DefaultEnvoyGatewayLoggingLevel(name, logLevel))

	return Logger{
		Logger:        zapr.NewLogger(logger).WithName(name),
		logging:       l.logging,
		sugaredLogger: l.sugaredLogger,
	}
}

// WithValues returns a new Logger instance with additional key/value pairs.
// See Info for documentation on how key/value pairs work.
func (l Logger) WithValues(keysAndValues ...interface{}) Logger {
	l.Logger = l.Logger.WithValues(keysAndValues...)
	return l
}

// A Sugar wraps the base Logger functionality in a slower, but less
// verbose, API. Any Logger can be converted to a SugaredLogger with its Sugar
// method.
//
// Unlike the Logger, the SugaredLogger doesn't insist on structured logging.
// For each log level, it exposes four methods:
//
//   - methods named after the log level for log.Print-style logging
//   - methods ending in "w" for loosely-typed structured logging
//   - methods ending in "f" for log.Printf-style logging
//   - methods ending in "ln" for log.Println-style logging
//
// For example, the methods for InfoLevel are:
//
//	Info(...any)           Print-style logging
//	Infow(...any)          Structured logging (read as "info with")
//	Infof(string, ...any)  Printf-style logging
//	Infoln(...any)         Println-style logging
func (l Logger) Sugar() *zap.SugaredLogger {
	return l.sugaredLogger
}

func initZapLogger(level v1alpha1.LogLevel) *zap.Logger {
	parseLevel, _ := zapcore.ParseLevel(string(level))
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), zapcore.AddSync(os.Stdout), zap.NewAtomicLevelAt(parseLevel))

	return zap.New(core, zap.AddCaller())
}
