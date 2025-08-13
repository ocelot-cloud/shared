package assert

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func True(t *testing.T, condition bool, msg ...string) {
	assert.True(t, condition, msg)
}

func False(t *testing.T, condition bool, msg ...string) {
	assert.False(t, condition, msg)
}

func Equal(t *testing.T, expected any, actual any, msg ...string) {
	assert.Equal(t, expected, actual, msg)
}

func Panics(t *testing.T, f PanicFunc, msg ...string) {
	assert.Panics(t, assert.PanicTestFunc(f), msg)
}

func Fail(t *testing.T, msg string) {
	assert.Fail(t, msg)
}

func Nil(t *testing.T, object any, msg ...string) {
	assert.Nil(t, object, msg)
}

func NotNil(t *testing.T, object any, msg ...string) {
	assert.NotNil(t, object, msg)
}

func NotEqual(t *testing.T, expected any, actual any, msg ...string) {
	assert.NotEqual(t, expected, actual, msg)
}

type PanicFunc func()
