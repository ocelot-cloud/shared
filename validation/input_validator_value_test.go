package validation

import (
	"fmt"
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

func TestValidateUserName(t *testing.T) {
	assert.Nil(t, Validate("validusername", "USER_NAME"))
	assert.Nil(t, Validate("user123", "USER_NAME"))
	assert.NotNil(t, Validate("InvalidUsername", "USER_NAME"))          // Contains uppercase
	assert.NotNil(t, Validate("user!@#", "USER_NAME"))                  // Contains special characters
	assert.NotNil(t, Validate("us", "USER_NAME"))                       // Too short
	assert.NotNil(t, Validate("thisusernameiswaytoolong", "USER_NAME")) // Too long
}

func TestValidateVersion(t *testing.T) {
	assert.Nil(t, Validate("valid.versionname", "VERSION_NAME"))
	assert.Nil(t, Validate("version123", "VERSION_NAME"))
	assert.Nil(t, Validate("version.name123", "VERSION_NAME"))
	assert.NotNil(t, Validate("invalid.versionname!", "VERSION_NAME"))             // Contains special characters other than dot
	assert.NotNil(t, Validate("ta", "VERSION_NAME"))                               // Too short
	assert.NotNil(t, Validate("this.versionname.is.way.too.long", "VERSION_NAME")) // Too long
}

func TestValidatePassword(t *testing.T) {
	assert.Nil(t, Validate("validpassword!", "PASSWORD"))
	assert.Nil(t, Validate("valid_pass123", "PASSWORD"))
	assert.Nil(t, Validate("InvalidPassword", "PASSWORD")) // Contains uppercase
	assert.Nil(t, Validate("valid!@#", "PASSWORD"))        // Contains special characters
	assert.NotNil(t, Validate("1234567", "PASSWORD"))      // Too short
	assert.Nil(t, Validate("12345678", "PASSWORD"))
	assert.NotNil(t, Validate("thispasswordiswaytoolong_xxxxx!", "PASSWORD")) // Too long
}

func TestValidateCookie(t *testing.T) {
	sixtyOneHexDecimalLetters := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde"

	assert.NotNil(t, Validate(sixtyOneHexDecimalLetters, "COOKIE"))
	assert.Nil(t, Validate(sixtyOneHexDecimalLetters+"f", "COOKIE"))
	assert.NotNil(t, Validate(sixtyOneHexDecimalLetters+"ff", "COOKIE"))
	assert.NotNil(t, Validate(sixtyOneHexDecimalLetters+"g", "COOKIE"))
	assert.NotNil(t, Validate("", "COOKIE"))
}

func TestValidateEmail(t *testing.T) {
	assert.Nil(t, Validate("admin@admin.com", "EMAIL"))
	assert.NotNil(t, Validate("@admin.com", "EMAIL"))
	assert.NotNil(t, Validate("admin@.com", "EMAIL"))
	assert.NotNil(t, Validate("admin@admin.", "EMAIL"))
	assert.NotNil(t, Validate("adminadmin.com", "EMAIL"))
	assert.NotNil(t, Validate("admin@admincom", "EMAIL"))

	thirtyCharacters := "abcdefghijklmnopqrstuvwxyz1234"
	validEmail := fmt.Sprintf("%s@%s.de", thirtyCharacters, thirtyCharacters)
	assert.Nil(t, Validate(validEmail, "EMAIL"))
	tooLongEmail := fmt.Sprintf("%s@%s.com", thirtyCharacters, thirtyCharacters)
	assert.NotNil(t, Validate(tooLongEmail, "EMAIL"))
}

func TestValidateNumber(t *testing.T) {
	assert.Nil(t, Validate("0", "NUMBER"))
	assert.Nil(t, Validate("1", "NUMBER"))
	assert.NotNil(t, Validate("-1", "NUMBER"))
	assert.NotNil(t, Validate("a", "NUMBER"))
	assert.NotNil(t, Validate("A", "NUMBER"))
	assert.NotNil(t, Validate("z", "NUMBER"))
	assert.NotNil(t, Validate("Z", "NUMBER"))
	assert.NotNil(t, Validate("-", "NUMBER"))
	assert.NotNil(t, Validate("_", "NUMBER"))
	assert.NotNil(t, Validate(".", "NUMBER"))
	assert.NotNil(t, Validate(",", "NUMBER"))

	twentyDigitNumber := "01234567890123456789"
	assert.Nil(t, Validate(twentyDigitNumber, "NUMBER"))
	assert.NotNil(t, Validate(twentyDigitNumber+"0", "NUMBER"))
}

func TestSearchTerm(t *testing.T) {
	assert.Nil(t, Validate("", "SEARCH_TERM"))
	assert.Nil(t, Validate("a", "SEARCH_TERM"))
	assert.Nil(t, Validate("1", "SEARCH_TERM"))
	assert.Nil(t, Validate("0123456789abcdefghij", "SEARCH_TERM"))
	assert.NotNil(t, Validate("asdf!", "SEARCH_TERM"))
}
