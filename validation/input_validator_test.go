package validation

import (
	"github.com/ocelot-cloud/shared/assert"
	"github.com/ocelot-cloud/shared/utils"
	"testing"
)

func TestValidateStruct(t *testing.T) {
	assertStructValidation(t, utils.SingleString{"asdf"}, "")
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
