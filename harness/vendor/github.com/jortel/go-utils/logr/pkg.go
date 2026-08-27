package logr

import (
	"os"
	"strconv"

	"github.com/go-logr/logr"
)

const (
	EnvDevelopment = "LOG_DEVELOPMENT"
	EnvStructured  = "LOG_STRUCTURED"
	EnvLevel       = "LOG_LEVEL"
)

const (
	NORMAL = 0
	EXTRA  = 1
	DEBUG  = 2
)

// New returns a named logger.
func New(name string, level int, kvpair ...any) logr.Logger {
	v := os.Getenv(EnvLevel)
	if v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			level = n
		}
	}
	return logr.New(
		&Sink{
			name:   name,
			fields: fields(kvpair),
			level:  max(NORMAL, level),
		})
}

// WithName returns a named logger.
func WithName(name string, kvpair ...any) logr.Logger {
	return New(name, NORMAL, kvpair...)
}

// Level returns a logger with the adjusted level.
func Level(logger logr.Logger, n int) logr.Logger {
	sink := logger.GetSink()
	if s, cast := sink.(*Sink); cast {
		sink = s.WithLevel(n)
		return logger.WithSink(sink)
	} else {
		return logger
	}
}
