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
