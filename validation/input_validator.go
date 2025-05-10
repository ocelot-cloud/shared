package validation

import (
	"encoding/json"
	"fmt"
	"github.com/ocelot-cloud/shared/utils"
	"io"
	"net/http"
	"reflect"
	"regexp"
)

var ValidationTypeMap = map[string]*regexp.Regexp{
	"user_name":        regexp.MustCompile("^[a-z0-9]{3,20}$"),
	"app_name":         regexp.MustCompile("^[a-z0-9-]{3,20}$"),
	"version_name":     regexp.MustCompile("^[a-z0-9.]{3,20}$"),
	"search_term":      regexp.MustCompile("^[a-z0-9]{0,20}$"),
	"password":         regexp.MustCompile("^[a-zA-Z0-9._-]{8,30}$"),
	"email":            regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
	"number":           regexp.MustCompile("^[0-9]{1,20}$"),
	"host":             regexp.MustCompile("^[a-zA-Z0-9:._-]{0,64}$"),
	"known_hosts":      regexp.MustCompile(`^[A-Za-z0-9.:,/_+=#@\[\]| \r\n-]{0,}$`),
	"restic_backup_id": regexp.MustCompile(`^[a-f0-9]{64}$`),
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

func validate(input, validationType string) error {
	regex, found := ValidationTypeMap[validationType]
	if !found {
		return fmt.Errorf("unknown validation type: %s", validationType)
	}

	if validationType == "email" && len(input) > 64 {
		return fmt.Errorf("invalid input")
	}

	result := regex.MatchString(input)
	if result {
		return nil
	} else {
		return fmt.Errorf("invalid input")
	}
}

var secretRegex = regexp.MustCompile("^[a-f0-9]{64}$")

// As we do not receive sensitive data in request bodies such as secrets or cookie values, this is a separate concern in a separate function.
func ValidateSecret(input string) error {
	if !secretRegex.MatchString(input) {
		return fmt.Errorf("invalid input")
	}
	return nil
}

func ReadBody[T any](w http.ResponseWriter, r *http.Request) (*T, error) {
	var result T

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Logger.Warn("Failed to read request body: %v", err)
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return nil, fmt.Errorf("")
	}
	defer utils.Close(r.Body)

	if err = json.Unmarshal(body, &result); err != nil {
		utils.Logger.Warn("Failed to parse request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return nil, fmt.Errorf("")
	}

	if err = ValidateStruct(result); err != nil {
		utils.Logger.Info("invalid input: %v", err)
		http.Error(w, "invalid input", http.StatusBadRequest)
		return nil, fmt.Errorf("")
	}

	return &result, nil
}
