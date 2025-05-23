package validation

import (
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

type validStruct struct {
	Value string `validate:"user_name"`
}

type noValidationTag struct {
	Value string
}

type nonPublicField struct {
	value string `validate:"user_name"`
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
	Value *string `validate:"user_name"`
}

type invalidPointerString struct {
	Value *string `validate:"unknown-type"`
}

type doublePointerString struct {
	Value **string `validate:"user_name"`
}

type stringArrayStruct struct {
	Value [2]string `validate:"user_name"`
}

type stringSliceStruct struct {
	Value []string `validate:"user_name"`
}

type stringMapStruct struct {
	Value map[string]string `validate:"user_name"`
}

type stringPointerArrayStruct struct {
	Value [2]*string `validate:"user_name"`
}

type stringPointerSliceStruct struct {
	Value []*string `validate:"user_name"`
}

type arrayOfNestedDataStructures struct {
	Value [1]nestedValidStructure
}

type sliceOfNestedDataStructures struct {
	Value []nestedValidStructure
}

type arrayOfNestedDataStructuresPointers struct {
	Value [1]nestedPointerStructure
}

type sliceOfNestedDataStructuresPointers struct {
	Value []nestedPointerStructure
}

type SampleInterface interface {
	SampleFunction()
}

type validSampleInterfaceImplementationStructure struct {
	SampleField string `validate:"user_name"`
}

func (v validSampleInterfaceImplementationStructure) SampleFunction() {}

type invalidSampleInterfaceImplementationStructure struct {
	SampleField string
}

func (v invalidSampleInterfaceImplementationStructure) SampleFunction() {}

func TestValidateStruct(t *testing.T) {
	sampleString := "ocelotcloud"
	sampleStringPointer := &sampleString

	var validSampleInterfaceImplementation SampleInterface = &validSampleInterfaceImplementationStructure{SampleField: "ocelotcloud"}
	var invalidSampleInterfaceImplementation SampleInterface = &invalidSampleInterfaceImplementationStructure{SampleField: "ocelotcloud"}

	testCases := []struct {
		name            string
		input           interface{}
		expectedMessage string
	}{
		{"valid struct", validStruct{"ocelotcloud"}, ""},
		{"invalid value in valid struct", validStruct{"ocelotcloud!!"}, "field does not match regex: Value"},

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

		{"valid array of structs", [2]validStruct{{"ocelotcloud"}, {"another"}}, "input must be a data structure, but was: array"},
		{"valid slice of structs", []validStruct{{"ocelotcloud"}, {"another"}}, "input must be a data structure, but was: slice"},
		{"valid map of structs", map[string]validStruct{"one": {"ocelotcloud"}, "two": {"another"}}, "input must be a data structure, but was: map"},

		{"invalid double pointer string", doublePointerString{&sampleStringPointer}, "field is double pointer: Value"},

		{"valid string array struct", stringArrayStruct{[2]string{"ocelotcloud", "another"}}, ""},
		{"valid string array struct", stringSliceStruct{[]string{"ocelotcloud", "another"}}, ""},
		{"maps are not allowed as fields", stringMapStruct{map[string]string{"one": "ocelotcloud", "two": "another"}}, "map fields are not allowed: Value"},

		{"invalid string array struct", stringArrayStruct{[2]string{"ocelotcloud", "another!!"}}, "field does not match regex: Value"},
		{"invalid string slice struct", stringSliceStruct{[]string{"ocelotcloud", "another!!"}}, "field does not match regex: Value"},

		{"don't allow string point array fields", stringPointerArrayStruct{[2]*string{&sampleString, &sampleString}}, "field of array or slice of pointers found: Value"},
		{"don't allow string point slice fields", stringPointerSliceStruct{[]*string{&sampleString, &sampleString}}, "field of array or slice of pointers found: Value"},

		{"valid array of nested data structures", arrayOfNestedDataStructures{[1]nestedValidStructure{{validStruct{"ocelotcloud"}}}}, ""},
		{"valid slice of nested data structures", sliceOfNestedDataStructures{[]nestedValidStructure{{validStruct{"ocelotcloud"}}}}, ""},

		{"invalid array of nested data structures", arrayOfNestedDataStructures{[1]nestedValidStructure{{validStruct{"!!!"}}}}, "field does not match regex: Value"},
		{"invalid slice of nested data structures", sliceOfNestedDataStructures{[]nestedValidStructure{{validStruct{"!!!"}}}}, "field does not match regex: Value"},

		{"invalid array of nested data structures", arrayOfNestedDataStructuresPointers{[1]nestedPointerStructure{{&validStruct{"!!!"}}}}, "field does not match regex: Value"},
		{"invalid slice of nested data structures", sliceOfNestedDataStructuresPointers{[]nestedPointerStructure{{&validStruct{"!!!"}}}}, "field does not match regex: Value"},

		{"valid interface input", validSampleInterfaceImplementation, ""},
		{"invalid interface input", invalidSampleInterfaceImplementation, "no validation tag found for field: SampleField"},
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
