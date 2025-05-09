package validation

import (
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

type noValidationTag struct {
	Value string
}

type noRegexTag struct {
	Value string `validate:"a(b"`
}

// TODO I think SingleString will become deprecated then so it can be deleted afterwards.

func TestValidateStruct(t *testing.T) {
	testCases := []struct {
		name            string
		input           interface{}
		expectedMessage string
	}{
		{"no validation tag", noValidationTag{"asdf"}, "no validation tag found for field: Value"},
		{"no regex in validation tag", noRegexTag{"asdf"}, "validation failed"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateStruct(tc.input)
			if tc.expectedMessage == "" {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedMessage, err.Error())
			}
		})
	}
}
