package validation

import (
	"fmt"
	"reflect"
	"regexp"
)

type ValidationType int

const (
	USER_NAME ValidationType = iota
	APP_NAME
	VERSION_NAME
	SEARCH_TERM
	// TODO anything else?
)

// TODO just dummy stuff, re-check, and test
var validationTypeMap = map[ValidationType]string{
	USER_NAME:    "^[a-zA-Z0-9_]{3,16}$",
	APP_NAME:     "^[a-zA-Z0-9_]{3,16}$",
	VERSION_NAME: "^[0-9]+\\.[0-9]+\\.[0-9]+$",
	SEARCH_TERM:  "^[a-zA-Z0-9_ ]{3,16}$",
}

func ValidateStruct(s interface{}) error {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		if structField.Tag == "" || !field.CanInterface() {
			continue
		}

		tag := structField.Tag.Get("validate")
		if tag != "" && field.Kind() == reflect.String {
			matched, _ := regexp.MatchString(tag, field.String())
			if !matched {
				return fmt.Errorf("field %s does not match regex", structField.Name)
			}
		}

		if field.Kind() == reflect.Struct {
			if err := ValidateStruct(field.Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

/* TODO
* simplify input validation by using my "reflection" approach. Does this in store first.
  * also check nested structures, slices, arrays, nil, maps, pointers, simple string input (should cause error since its no data structure?) etc.
  * other types than string needed to be checked?
  * fail if a string field was found which does not have "validate" tag, or when its value is empty, should be a regex
  * can I use constants as tags? if not, maybe do sth like "validate:user", and the validate function checks -> if x == "user" then validate it for user regex; unknown validation type should throw error
  * also add a readBody function which both modules can use, which internally does the input validation.
*/
