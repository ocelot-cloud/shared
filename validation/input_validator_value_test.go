package validation

import (
	"fmt"
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

func TestValidateUserName(t *testing.T) {
	assert.Nil(t, validate("validusername", "USER_NAME"))
	assert.Nil(t, validate("user123", "USER_NAME"))
	assert.NotNil(t, validate("user.123", "USER_NAME"))
	assert.NotNil(t, validate("user-123", "USER_NAME"))
	assert.NotNil(t, validate("user_123", "USER_NAME"))
	assert.NotNil(t, validate("InvalidUsername", "USER_NAME"))          // Contains uppercase
	assert.NotNil(t, validate("user!@#", "USER_NAME"))                  // Contains special characters
	assert.NotNil(t, validate("us", "USER_NAME"))                       // Too short
	assert.NotNil(t, validate("thisusernameiswaytoolong", "USER_NAME")) // Too long
}

func TestValidateAppName(t *testing.T) {
	assert.Nil(t, validate("validappname", "APP_NAME"))
	assert.Nil(t, validate("app123", "APP_NAME"))
	assert.Nil(t, validate("app-123", "APP_NAME"))
	assert.NotNil(t, validate("app_123", "APP_NAME"))
	assert.NotNil(t, validate("app.123", "APP_NAME"))
	assert.NotNil(t, validate("InvalidAppName", "APP_NAME"))          // Contains uppercase
	assert.NotNil(t, validate("app!@#", "APP_NAME"))                  // Contains special characters
	assert.NotNil(t, validate("ap", "APP_NAME"))                      // Too short
	assert.NotNil(t, validate("thisappnameiswaytoolong", "APP_NAME")) // Too long
}

func TestValidateVersion(t *testing.T) {
	assert.Nil(t, validate("valid.versionname", "VERSION_NAME"))
	assert.Nil(t, validate("version123", "VERSION_NAME"))
	assert.Nil(t, validate("version.name123", "VERSION_NAME"))
	assert.NotNil(t, validate("version_name123", "VERSION_NAME"))
	assert.NotNil(t, validate("invalid.versionname!", "VERSION_NAME"))             // Contains special characters other than dot
	assert.NotNil(t, validate("ta", "VERSION_NAME"))                               // Too short
	assert.NotNil(t, validate("this.versionname.is.way.too.long", "VERSION_NAME")) // Too long
}

func TestValidatePassword(t *testing.T) {
	assert.Nil(t, validate("validpassword!", "PASSWORD"))
	assert.Nil(t, validate("valid_pass123", "PASSWORD"))
	assert.Nil(t, validate("InvalidPassword", "PASSWORD")) // Contains uppercase
	assert.Nil(t, validate("valid!@#", "PASSWORD"))        // Contains special characters
	assert.NotNil(t, validate("1234567", "PASSWORD"))      // Too short
	assert.Nil(t, validate("12345678", "PASSWORD"))
	assert.NotNil(t, validate("thispasswordiswaytoolong_xxxxx!", "PASSWORD")) // Too long
}

func TestValidateCookie(t *testing.T) {
	sixtyOneHexDecimalLetters := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcde"

	assert.NotNil(t, ValidateSecret(sixtyOneHexDecimalLetters))
	assert.Nil(t, ValidateSecret(sixtyOneHexDecimalLetters+"f"))
	assert.NotNil(t, ValidateSecret(sixtyOneHexDecimalLetters+"ff"))
	assert.NotNil(t, ValidateSecret(sixtyOneHexDecimalLetters+"g"))
	assert.NotNil(t, ValidateSecret(""))
}

func TestValidateEmail(t *testing.T) {
	assert.Nil(t, validate("admin@admin.com", "EMAIL"))
	assert.NotNil(t, validate("@admin.com", "EMAIL"))
	assert.NotNil(t, validate("admin@.com", "EMAIL"))
	assert.NotNil(t, validate("admin@admin.", "EMAIL"))
	assert.NotNil(t, validate("adminadmin.com", "EMAIL"))
	assert.NotNil(t, validate("admin@admincom", "EMAIL"))

	thirtyCharacters := "abcdefghijklmnopqrstuvwxyz1234"
	validEmail := fmt.Sprintf("%s@%s.de", thirtyCharacters, thirtyCharacters)
	assert.Nil(t, validate(validEmail, "EMAIL"))
	tooLongEmail := fmt.Sprintf("%s@%s.com", thirtyCharacters, thirtyCharacters)
	assert.NotNil(t, validate(tooLongEmail, "EMAIL"))
}

func TestValidateNumber(t *testing.T) {
	assert.Nil(t, validate("0", "NUMBER"))
	assert.Nil(t, validate("1", "NUMBER"))
	assert.NotNil(t, validate("-1", "NUMBER"))
	assert.NotNil(t, validate("a", "NUMBER"))
	assert.NotNil(t, validate("A", "NUMBER"))
	assert.NotNil(t, validate("z", "NUMBER"))
	assert.NotNil(t, validate("Z", "NUMBER"))
	assert.NotNil(t, validate("-", "NUMBER"))
	assert.NotNil(t, validate("_", "NUMBER"))
	assert.NotNil(t, validate(".", "NUMBER"))
	assert.NotNil(t, validate(",", "NUMBER"))

	twentyDigitNumber := "01234567890123456789"
	assert.Nil(t, validate(twentyDigitNumber, "NUMBER"))
	assert.NotNil(t, validate(twentyDigitNumber+"0", "NUMBER"))
}

func TestSearchTerm(t *testing.T) {
	assert.Nil(t, validate("", "SEARCH_TERM"))
	assert.Nil(t, validate("a", "SEARCH_TERM"))
	assert.Nil(t, validate("1", "SEARCH_TERM"))
	assert.Nil(t, validate("0123456789abcdefghij", "SEARCH_TERM"))
	assert.NotNil(t, validate("asdf!", "SEARCH_TERM"))
}
