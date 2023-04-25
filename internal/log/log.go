// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package log

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

func NewLogger() (logr.Logger, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return logr.Logger{}, err
	}
	return zapr.NewLogger(zapLogger), nil

	// var logger *zap.Logger

	// encoder := getJsonEncoder()
	// core := zapcore.NewTee(
	// 	zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel), // log console
	// )

	// logger = zap.New(core, zap.AddCallerSkip(0), zap.AddCaller())

	// // singleton
	// zap.ReplaceGlobals(logger)
	// return zapr.NewLogger(logger), nil
}

// func getJsonEncoder() zapcore.Encoder {
// 	encoderConfig := zap.NewProductionEncoderConfig()
// 	encoderConfig.TimeKey = "time"
// 	encoderConfig.LevelKey = "level"
// 	encoderConfig.NameKey = "logger"
// 	encoderConfig.CallerKey = "caller"
// 	encoderConfig.FunctionKey = zapcore.OmitKey
// 	encoderConfig.MessageKey = "message"
// 	encoderConfig.StacktraceKey = "stacktrace"
// 	encoderConfig.LineEnding = zapcore.DefaultLineEnding
// 	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
// 	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
// 	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
// 	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

// 	return zapcore.NewJSONEncoder(encoderConfig)
// }
