package zerologadapter

// This package adapts zerolog to logr.Logger so that we can pass a zerologger to controller runtime

import (
	"github.com/go-logr/logr"
	"github.com/rs/zerolog"
)

type ZerologSink struct {
	logger zerolog.Logger
}

func NewLogr(logger zerolog.Logger) logr.Logger {
	sink := &ZerologSink{logger: logger}
	return logr.New(sink)
}

// Implement logr.LogSink interface methods
func (z *ZerologSink) Init(info logr.RuntimeInfo) {
	// No init needed, but we need to add this func to meet the logr interface requirements
}

func (z *ZerologSink) Info(level int, msg string, keysAndValues ...interface{}) {
	// Directly use the logger to log at Info level
	if z.Enabled(level) {
		z.logger.Info().Fields(makeFields(keysAndValues...)).Msg(msg)
	}
}

func (z *ZerologSink) Error(err error, msg string, keysAndValues ...interface{}) {
	// Directly use the logger to log at Error level, attaching the error if present
	if err != nil {
		z.logger.Error().Fields(makeFields(keysAndValues...)).Err(err).Msg(msg)
	} else {
		z.logger.Error().Fields(makeFields(keysAndValues...)).Msg(msg)
	}
}

func (z *ZerologSink) Enabled(level int) bool {
	return zerolog.GlobalLevel() <= zerolog.Level(level)
}

func (z *ZerologSink) WithValues(keysAndValues ...interface{}) logr.LogSink {
	newLogger := z.logger.With().Fields(makeFields(keysAndValues...)).Logger()
	return &ZerologSink{logger: newLogger}
}

func (z *ZerologSink) WithName(name string) logr.LogSink {
	newLogger := z.logger.With().Str("name", name).Logger()
	return &ZerologSink{logger: newLogger}
}

// Helper function to convert keys and values to a format zerolog can use
func makeFields(keysAndValues ...interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key, ok := keysAndValues[i].(string)
			if !ok {
				continue // or handle the error as appropriate
			}
			fields[key] = keysAndValues[i+1]
		}
	}
	return fields
}
