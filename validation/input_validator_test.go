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

type nonPublicField struct {
	value string `validate:"USER_NAME"`
}

type unknownTag struct {
	Value string `validate:"unknown-type"`
}

type nestedValidStructure struct {
	SomeTag validStruct
}

type nestedInvalidStructure struct {
	SomeTag unknownTag
}

type nestedPointerStructure struct {
	SomeTag *validStruct
}

type ignoreNumberFields struct {
	SomeNumber int
}

type pointerString struct {
	Value *string `validate:"USER_NAME"`
}

type invalidPointerString struct {
	Value *string `validate:"unknown-type"`
}

// TODO I think SingleString will become deprecated then so it can be deleted afterwards.

func TestValidateStruct(t *testing.T) {
	sampleString := "ocelotcloud"

	testCases := []struct {
		name            string
		input           interface{}
		expectedMessage string
	}{
		{"valid struct", validStruct{"ocelotcloud"}, ""},

		{"no validation tag", noValidationTag{"asdf"}, "no validation tag found for field: Value"},
		{"unknown validation tag", unknownTag{"asdf"}, "unknown validation type: unknown-type"},
		{"non-public field", nonPublicField{"asdf"}, "cannot validate non-public fields: value"},

		{"nested valid structure", nestedValidStructure{validStruct{"asdf"}}, ""},
		{"nested invalid structure", nestedInvalidStructure{unknownTag{"asdf"}}, "unknown validation type: unknown-type"},

		{"string input fails", "some-string", "input must be a data structure, but was: string"},
		{"float input fails", 1.23, "input must be a data structure, but was: float64"},
		{"integer input fails", 123, "input must be a data structure, but was: int"},

		{"ignore fields which are neither strings nor structs", ignoreNumberFields{123}, ""},

		{"valid struct as pointer", &validStruct{"ocelotcloud"}, ""},
		{"valid struct with nested pointer structure", nestedPointerStructure{&validStruct{"ocelotcloud"}}, ""},
		{"valid struct with pointer string", pointerString{&sampleString}, ""},
		{"invalid struct with pointer string", invalidPointerString{&sampleString}, "unknown validation type: unknown-type"},

		{"invalid nil pointer field", pointerString{nil}, "pointer field is nil: Value"},

		{"valid array of structs", [2]validStruct{{"ocelotcloud"}, {"another"}}, ""},
		{"valid slice of structs", []validStruct{{"ocelotcloud"}, {"another"}}, ""},
		{"valid array of pointer structs", [2]pointerString{{&sampleString}, {&sampleString}}, ""},
		{"valid slice of pointer structs", []pointerString{{&sampleString}, {&sampleString}}, ""},
		{"invalid array of pointer structs", [2]pointerString{{&sampleString}, {nil}}, "pointer field is nil: Value"},
		{"invalid slice of pointer structs", []pointerString{{&sampleString}, {nil}}, "pointer field is nil: Value"},
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
