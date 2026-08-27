package error

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Wrapped error.
type Wrapped interface {
	// Unwrap returns the wrapped error.
	Unwrap() error
}

type WithContext interface {
	// Context returns the error context.
	Context() []any
}

// Traced error with captured stack trace.
type Traced interface {
	// Stack returns a description of the captured stack trace.
	Stack() string
}

// Snapshot is a wrapped error that captures the stack and context.
type Snapshot interface {
	error
	Wrapped
	WithContext
	Traced
}

// Recovered returns an error recovered panic.
func Recovered(p any) (err error) {
	if p == nil {
		return
	}
	err, cast := p.(error)
	if !cast {
		err = Errorf("%+v", p)
	}
	return
}

// New returns a new wrapped error.
func New(m string, kvpair ...any) (newError error) {
	newError = Wrap(
		errors.New(m),
		kvpair...)
	return
}

// Errorf returns a new wrapped error.
func Errorf(m string, v ...any) (newError error) {
	newError = Wrap(fmt.Errorf(m, v...))
	return
}

// Wrap an error.
// Returns `err` when err is `nil` or error.
func Wrap(err error, kvpair ...any) (newError error) {
	if err == nil {
		return err
	}
	if le, cast := err.(*Error); cast {
		le.append(kvpair)
		return le
	}
	bfr := make([]uintptr, 50)
	n := runtime.Callers(2, bfr[:])
	frames := runtime.CallersFrames(bfr[:n])
	stack := []string{""}
	for {
		f, hasNext := frames.Next()
		frame := fmt.Sprintf(
			"%s()\n\t%s:%d",
			f.Function,
			f.File,
			f.Line)
		stack = append(stack, frame)
		if !hasNext {
			break
		}
	}
	newError = &Error{
		stack:   stack,
		wrapped: err,
	}
	newError.(*Error).append(kvpair)
	return newError
}

// Unwrap an error.
// Returns: the original error when not wrapped.
func Unwrap(err error) (out error) {
	if err == nil {
		return
	}
	out = err
	for {
		if wrapped, cast := out.(Wrapped); cast {
			out = wrapped.Unwrap()
		} else {
			break
		}
	}

	return
}

// Error provides an error with context and stack.
type Error struct {
	// Original error.
	wrapped error
	// description of the error.
	description string
	// Context.
	context []any
	// Stack.
	stack []string
}

// Error description.
func (e Error) Error() (s string) {
	if len(e.description) > 0 {
		s = e.causedBy(e.description, e.wrapped.Error())
	} else {
		s = e.wrapped.Error()
	}
	return
}

// Stack returns the stack trace.
// Format:
//
//	package.Function()
//	  file:line
//	package.Function()
//	  file:line
//	...
func (e Error) Stack() (s string) {
	s = strings.Join(e.stack, "\n")
	return
}

// Context returns the`context` key/value pairs.
func (e Error) Context() (context []any) {
	context = e.context
	return
}

// Unwrap the error.
func (e Error) Unwrap() (wrapped error) {
	wrapped = Unwrap(e.wrapped)
	return
}

// append the context.
// And odd number of context is interpreted as:
// a description followed by an even number of key value pairs.
func (e *Error) append(kvpair []any) {
	if len(kvpair) == 0 {
		return
	}
	odd := len(kvpair)%2 != 0
	if description, cast := kvpair[0].(string); odd && cast {
		kvpair = kvpair[1:]
		if len(e.description) > 0 {
			e.description = e.causedBy(description, e.description)
		} else {
			e.description = description
		}
	}

	e.context = append(e.context, kvpair...)
}

// Build caused-by.
func (e *Error) causedBy(error, caused string) (s string) {
	s = fmt.Sprintf(
		"%s | caused by: '%s'",
		error,
		caused)
	return
}
