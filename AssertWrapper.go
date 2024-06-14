package shared

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func AssertTrue(t *testing.T, condition bool, msg ...string) {
	assert.True(t, condition, msg)
}

func AssertFalse(t *testing.T, condition bool, msg ...string) {
	assert.False(t, condition, msg)
}

func AssertEqual(t *testing.T, expected interface{}, actual interface{}, msg ...string) {
	assert.Equal(t, expected, actual, msg)
}

func AssertPanics(t *testing.T, f PanicFunc, msg ...string) {
	assert.Panics(t, assert.PanicTestFunc(f), msg)
}

func AssertFail(t *testing.T, msg string) {
	assert.Fail(t, msg)
}

func AssertNil(t *testing.T, object interface{}, msg ...string) {
	assert.Nil(t, object, msg)
}

func AssertNotNil(t *testing.T, object interface{}, msg ...string) {
	assert.NotNil(t, object, msg)
}

type PanicFunc func()
