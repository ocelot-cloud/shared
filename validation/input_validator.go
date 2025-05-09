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
	"PASSWORD":     "^[a-zA-Z0-9-]{8,30}$",
	// TODO anything else? -> known hosts, ports, host names and ip addresses, (cookies and secrets? not requests bodies, maybe separate validation function)
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

		/* TODO is that needed?
		if structField.Tag == "" || !field.CanInterface() {
			continue
		}
		*/

		if field.Kind() == reflect.String {
			err := validateString(field, structField)
			if err != nil {
				return err
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
		return fmt.Errorf("field %s does not match regex", structField.Name)
	}

	return nil
}

/* TODO
* simplify input validation by using my "reflection" approach. Does this in store first.
  * also check nested structures/interfaces, slices, arrays, nil, maps, pointers, simple string input (should cause error since its no data structure?) etc.
  * other types than string needed to be checked?
  * fail if a string field was found which does not have "validate" tag, or when its value is empty, it should be a regex
  * can I use constants as tags? if not, maybe do sth like "validate:user", and the validate function checks -> if x == "user" then validate it for user regex; unknown validation type should throw error
  * also add a readBody function which both modules can use, which internally does the input validation.
*/
