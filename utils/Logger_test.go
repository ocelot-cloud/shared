package utils

import (
	"errors"
	"testing"
)

func TestLoggingVisually(t *testing.T) {
	logger := ProvideLogger("debug", true)
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")

	logger.Info("This is an info message", "key1", "value1", "key2", "value2")
	logger.Error("This is an info message", "error", "some-error")             // TODO
	logger.Error("This is an info message", "error", errors.New("some-error")) // TODO
}

/* TODO I want smart error logging with structured logging and stack traces, plan:
type OcError struct {
	errorMessage string
	errorStack   string
	Context 	 map[string]any
}

* logger.CreateError()
* high level function logs the error; prints the error message and context in a single line, and below that with pretty formatting is the stack trace
* func AddContext(string...) -> make sure its two args, the odd index arg must be string, the even index arg can be of any type
* when some methods are not used correctly, e.g. type errors etc, then an error should be logged as the developer used it wrongly
* actually very lightweight as the error object is not recreated all the time, but you rather add changes directly to the existing error object
* condition: all errors must be created through my logging library, so that the stack trace is always available, and operations work; using operations on "normal" errors should work but will cause an error log for the developer
* maybe I should create a small, reusable library for that
*/
