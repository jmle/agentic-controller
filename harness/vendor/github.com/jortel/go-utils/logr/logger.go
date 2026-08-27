package logr

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-logr/logr"
	liberr "github.com/jortel/go-utils/error"
	log "github.com/sirupsen/logrus"
)

// Sink -.
type Sink struct {
	delegate    *log.Logger
	fields      log.Fields
	name        string
	development bool
	structured  bool
	level       int
}

// Init builds the delegate logger.
func (s *Sink) Init(_ logr.RuntimeInfo) {
	s.delegate = log.New()
	v := os.Getenv(EnvDevelopment)
	b, _ := strconv.ParseBool(v)
	s.development = b
	if s.development {
		fmt := new(log.TextFormatter)
		fmt.TimestampFormat = "2006-01-02 15:04:05"
		fmt.FullTimestamp = true
		s.delegate.SetFormatter(fmt)
	} else {
		fmt := new(log.TextFormatter)
		fmt.FullTimestamp = true
		fmt.DisableColors = true
		fmt.DisableQuote = true
		s.delegate.SetFormatter(fmt)
	}
	v = os.Getenv(EnvStructured)
	s.structured, _ = strconv.ParseBool(v)
	if s.structured {
		fmt := new(log.JSONFormatter)
		fmt.PrettyPrint = true
		s.delegate.Formatter = fmt
	}
}

// Info logs at info.
func (s *Sink) Info(_ int, message string, kvpair ...any) {
	fields := fields(kvpair)
	entry := s.delegate.WithFields(s.fields)
	entry = entry.WithFields(fields)
	entry.Info(s.named(message))
}

// Error logs an error.
func (s *Sink) Error(err error, message string, kvpair ...any) {
	if err == nil {
		return
	}
	snapshot, cast := err.(liberr.Snapshot)
	if cast {
		err = snapshot.Unwrap()
		if context := snapshot.Context(); context != nil {
			context = append(
				context,
				kvpair...)
			kvpair = context
		}
		if s.structured {
			fields := fields(kvpair)
			fields["error"] = snapshot.Error()
			fields["stack"] = snapshot.Stack()
			fields["logger"] = s.name
			entry := s.delegate.WithFields(s.fields)
			entry = entry.WithFields(fields)
			entry.Error(s.named(message))
		} else {
			fields := fields(kvpair)
			entry := s.delegate.WithFields(s.fields)
			entry = entry.WithFields(fields)
			if message != "" {
				entry.Error(s.named(message), "\n", snapshot.Error(), snapshot.Stack())
			} else {
				entry.Error(s.named(snapshot.Error()), snapshot.Stack())
			}
		}
	} else {
		if wrapped, cast := err.(liberr.Wrapped); cast {
			err = wrapped.Unwrap()
		}
		err = liberr.Wrap(err)
		s.Error(err, message, kvpair...)
	}
}

// Enabled returns whether logger is enabled.
func (s *Sink) Enabled(level int) bool {
	return s.level >= level
}

// WithName returns a logger with name.
func (s *Sink) WithName(name string) logr.LogSink {
	return &Sink{
		delegate:    s.delegate,
		fields:      s.fields,
		name:        name,
		development: s.development,
		structured:  s.structured,
		level:       s.level,
	}
}

// WithValues returns a logger with values.
func (s *Sink) WithValues(kvpair ...any) logr.LogSink {
	return &Sink{
		delegate:    s.delegate,
		fields:      fields(kvpair),
		name:        s.name,
		development: s.development,
		structured:  s.structured,
		level:       s.level,
	}
}

func (s *Sink) WithLevel(n int) logr.LogSink {
	return &Sink{
		delegate:    s.delegate,
		fields:      s.fields,
		name:        s.name,
		development: s.development,
		structured:  s.structured,
		level:       max(0, n),
	}
}

func (s *Sink) named(message string) (m string) {
	if s.name != "" {
		m = "[" + s.name + "] "
	}
	m = m + message
	return
}

// fields returns fields for kvpair.
func fields(kvpair []any) log.Fields {
	fields := log.Fields{}
	for i := range kvpair {
		if i%2 != 0 {
			key := fmt.Sprintf("%v", kvpair[i-1])
			v := fmt.Sprintf("%+v", kvpair[i])
			fields[key] = v
		}
	}
	return fields
}
