package utils

import (
	"errors"
	"github.com/ocelot-cloud/shared/assert"
	"strings"
	"testing"
)

func TestLoggingVisually(t *testing.T) {
	logger := ProvideLogger("debug", true)
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")

	logger.Info("This is an info message", "key1", "value1", "key2", "value2")
	logger.Error("This is an info message", ErrorField, "some-error")
	logger.Error("This is an info message", ErrorField, errors.New("some-error"))

	logger.Error("testing normal error", ErrorField, errors.New("some-error"), "key1", "value1")
}

func TestLoggingWithStackTrace(t *testing.T) {
	logger := ProvideLogger("debug", true)
	logger.Error("testing detailed error", ErrorField, subfunction(logger))
}

func subfunction(logger StructuredLogger) error {
	return logger.NewError("an error occurred", "key1", "value1")
}

func TestErrorToString(t *testing.T) {
	logger := ProvideLogger("debug", true)
	testError := logger.NewError("an error occurred", "key1", "value1")

	detailedTestError, ok := testError.(*DetailedError)
	assert.True(t, ok)
	assert.Equal(t, "an error occurred", detailedTestError.ErrorMessage)
	assert.Equal(t, 1, len(detailedTestError.Context))
	assert.Equal(t, "value1", detailedTestError.Context["key1"])
	assert.NotEqual(t, "", detailedTestError.ErrorStack)

	errorString := testError.Error()
	assert.True(t, strings.HasPrefix(errorString, "an error occurred key1=value1\nstack trace:\n"))
}

// TODO maybe I should create a small, reusable logging library in a separate repository
