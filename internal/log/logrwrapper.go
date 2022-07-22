package log

import (
	"fmt"

	"github.com/go-logr/logr"
)

// LogrWrapper is a nasty hack for turning the logr.Logger we get from NewLogger()
// into something that go-control-plane can accept.
// It seems pretty silly to take a zap logger, which is levelled, turn it into
// a V-style logr Logger, then turn it back again with this, but here we are.
// TODO(youngnick): Reopen the logging library discussion then do something about this.
type LogrWrapper struct {
	logr logr.Logger
}

const LevelDebug int = -2
const LevelInfo int = 0
const LevelWarn int = -1

func (l LogrWrapper) Debugf(template string, args ...interface{}) {

	l.logr.V(LevelDebug).Info(fmt.Sprintf(template, args...))
}

func (l LogrWrapper) Infof(template string, args ...interface{}) {

	l.logr.V(LevelInfo).Info(fmt.Sprintf(template, args...))
}

func (l LogrWrapper) Warnf(template string, args ...interface{}) {

	l.logr.V(LevelWarn).Info(fmt.Sprintf(template, args...))
}

func (l LogrWrapper) Errorf(template string, args ...interface{}) {

	l.logr.Error(fmt.Errorf(template, args...), "")
}

func NewLogrWrapper(log logr.Logger) *LogrWrapper {

	return &LogrWrapper{
		logr: log,
	}
}
