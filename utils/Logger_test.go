package utils

import (
	"testing"
)

func TestLoggingVisually(t *testing.T) {
	t.Helper()

	logger := ProvideLogger("debug", true)
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")
}
