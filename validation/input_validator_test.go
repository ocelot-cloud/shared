package validation

import (
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/utils"
	"testing"
)

// TODO I think SingleString will become deprecated then so it can be deleted afterwards.

func TestValidateStruct(t *testing.T) {
	assertStructValidation(t, utils.SingleString{"asdf"}, "no validation tag found for field: Value")
}

func assertStructValidation(t *testing.T, structure interface{}, expectedMessage string) {
	err := ValidateStruct(structure)
	if expectedMessage == "" {
		assert.Nil(t, err)
	} else {
		assert.NotNil(t, err)
		assert.Equal(t, expectedMessage, err.Error())
	}
}
