package validation

import (
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

type validStruct struct {
	Value string `validate:"USER_NAME"`
}

type noValidationTag struct {
	Value string
}

type unknownTag struct {
	Value string `validate:"unknown-type"`
}

// TODO I think SingleString will become deprecated then so it can be deleted afterwards.

func TestValidateStruct(t *testing.T) {
	testCases := []struct {
		name            string
		input           interface{}
		expectedMessage string
	}{
		{"valid struct", validStruct{"ocelotcloud"}, ""},
		{"no validation tag", noValidationTag{"asdf"}, "no validation tag found for field: Value"},
		{"unknown validation tag", unknownTag{"asdf"}, "unknown validation type: unknown-type"},
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
