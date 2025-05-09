package validation

import (
	"fmt"
	"github.com/ocelot-cloud/shared/utils"
	"reflect"
	"regexp"
)

type ValidationType string

// TODO ensure usernames, appnames and version do not contain underscores, maybe add extra test and comment that this is important for formatting

// TODO explicitly dont allow dots in usernames, appnames and version names
// TODO add extra tests for these regexes
var validationTypeMap = map[string]string{
	"USER_NAME":    "^[a-zA-Z0-9]{3,20}$",
	"APP_NAME":     "^[a-zA-Z0-9-]{3,20}$", // TODO should allow hyphens
	"VERSION_NAME": "^[a-zA-Z0-9.]{3,20}$",
	"SEARCH_TERM":  "^[a-zA-Z0-9-]{0,20}$",
	"PASSWORD":     "^[a-zA-Z0-9-]{8,30}$", // TODO allow more than that?
	// TODO anything else? -> known hosts, ports, host names and ip addresses, (cookies and secrets? not requests bodies, maybe separate validation function)
	"INTEGER": "^[0-9]{1,30}$", // relevant for ID's
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

	regex, found := validationTypeMap[tag]
	if !found {
		return fmt.Errorf("unknown validation type: %s", tag)
	}

	fieldString := field.String() // extra variable to see its content when debugging
	matched, err := regexp.MatchString(regex, fieldString)
	if err != nil {
		utils.Logger.Error("error for field validation '%s' when matching regex: %v ", structField.Name, err)
		return fmt.Errorf("validation failed")
	}
	if !matched {
		return fmt.Errorf("field does not match regex: %s", structField.Name)
	}

	return nil
}

// TODO also add a readBody function which both modules can use, which internally does the input validation.
