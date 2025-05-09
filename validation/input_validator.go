package validation

import (
	"fmt"
	"reflect"
	"regexp"
)

// TODO ensure usernames, appnames and version do not contain underscores, maybe add extra test and comment that this is important for formatting
// TODO explicitly dont allow dots in usernames, appnames and version names
// TODO add extra tests for these regexes
var ValidationTypeMap = map[string]*regexp.Regexp{
	"USER_NAME":    regexp.MustCompile("^[a-z0-9]{3,20}$"),
	"APP_NAME":     regexp.MustCompile("^[a-z0-9-]{3,20}$"), // TODO should allow hyphens
	"VERSION_NAME": regexp.MustCompile("^[a-z0-9.]{3,20}$"),
	"SEARCH_TERM":  regexp.MustCompile("^[a-z0-9]{0,20}$"),
	"PASSWORD":     regexp.MustCompile("^[a-zA-Z0-9!@#$%&_,.?]{8,30}$"), // TODO allow more than that?
	// TODO anything else? -> known hosts, ports, host names and ip addresses, (cookies, ValidationCode and secrets? not requests bodies, maybe separate validation function), email,
	"COOKIE": regexp.MustCompile("^[a-f0-9]{64}$"),
	"EMAIL":  regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
	"NUMBER": regexp.MustCompile("^[0-9]{1,20}$"),
}

func ValidateStruct(s interface{}) error {
	reflectionObject := getReflectionObject(s)
	fieldType := reflectionObject.Type()

	if reflectionObject.Kind() != reflect.Struct {
		return fmt.Errorf("input must be a data structure, but was: %s", reflectionObject.Kind())
	}

	for i := 0; i < reflectionObject.NumField(); i++ {
		fieldValue := reflectionObject.Field(i)
		reflectedStructureField := fieldType.Field(i)
		err := validateField(fieldValue, reflectedStructureField)
		if err != nil {
			return err
		}
	}
	return nil
}

func getReflectionObject(s interface{}) reflect.Value {
	object := reflect.ValueOf(s)
	for object.Kind() == reflect.Ptr && !object.IsNil() {
		object = object.Elem()
	}
	return object
}

func validateField(field reflect.Value, structField reflect.StructField) error {
	if !field.CanInterface() {
		return fmt.Errorf("cannot validate non-public fields: %s", structField.Name)
	}

	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return fmt.Errorf("pointer field is nil: %s", structField.Name)
		} else {
			field = field.Elem()
			if field.Kind() == reflect.Ptr {
				return fmt.Errorf("field is double pointer: %s", structField.Name)
			}
		}
	}

	if field.Kind() == reflect.Map {
		return fmt.Errorf("map fields are not allowed: %s", structField.Name)
	}

	if field.Kind() == reflect.String {
		err := validateString(field, structField)
		if err != nil {
			return err
		}
	}

	if field.Kind() == reflect.Array || field.Kind() == reflect.Slice {
		err := validateArrayOrSlice(field, structField)
		if err != nil {
			return err
		}
	}

	if field.Kind() == reflect.Struct {
		if err := ValidateStruct(field.Interface()); err != nil {
			return err
		}
	}
	return nil
}

func validateArrayOrSlice(field reflect.Value, structField reflect.StructField) error {
	if field.Type().Elem().Kind() == reflect.Ptr {
		return fmt.Errorf("field of array or slice of pointers found: %s", structField.Name)
	}

	if field.Type().Elem().Kind() == reflect.String {
		for i := 0; i < field.Len(); i++ {
			if err := validateString(field.Index(i), structField); err != nil {
				return err
			}
		}
	}

	if field.Type().Elem().Kind() == reflect.Struct {
		for i := 0; i < field.Len(); i++ {
			if err := ValidateStruct(field.Index(i).Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateString(field reflect.Value, structField reflect.StructField) error {
	tag := structField.Tag.Get("validate")
	if tag == "" {
		return fmt.Errorf("no validation tag found for field: %s", structField.Name)
	}

	regex, found := ValidationTypeMap[tag]
	if !found {
		return fmt.Errorf("unknown validation type: %s", tag)
	}

	fieldString := field.String() // extra variable to see its content when debugging
	if !regex.MatchString(fieldString) {
		return fmt.Errorf("field does not match regex: %s", structField.Name)
	}

	return nil
}

func Validate(input, validationType string) error {
	regex, found := ValidationTypeMap[validationType]
	if !found {
		return fmt.Errorf("unknown validation type: %s", validationType)
	}

	if validationType == "EMAIL" && len(input) > 64 {
		return fmt.Errorf("invalid input")
	}

	result := regex.MatchString(input)
	if result {
		return nil
	} else {
		return fmt.Errorf("invalid input")
	}
}

// TODO also add a readBody function which both modules can use, which internally does the input validation.
