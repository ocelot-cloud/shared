package validation

import (
	"fmt"
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

func TestValidateUserName(t *testing.T) {
	assert.Nil(t, validate("validusername", "user_name"))
	assert.Nil(t, validate("user123", "user_name"))
	assert.NotNil(t, validate("user.123", "user_name"))
	assert.NotNil(t, validate("user-123", "user_name"))
	assert.NotNil(t, validate("user_123", "user_name"))
	assert.NotNil(t, validate("InvalidUsername", "user_name"))          // Contains uppercase
	assert.NotNil(t, validate("user!@#", "user_name"))                  // Contains special characters
	assert.NotNil(t, validate("us", "user_name"))                       // Too short
	assert.NotNil(t, validate("thisusernameiswaytoolong", "user_name")) // Too long
}

func TestValidateAppName(t *testing.T) {
	assert.Nil(t, validate("validappname", "app_name"))
	assert.Nil(t, validate("app123", "app_name"))
	assert.Nil(t, validate("app-123", "app_name"))
	assert.NotNil(t, validate("app_123", "app_name"))
	assert.NotNil(t, validate("app.123", "app_name"))
	assert.NotNil(t, validate("InvalidAppName", "app_name"))          // Contains uppercase
	assert.NotNil(t, validate("app!@#", "app_name"))                  // Contains special characters
	assert.NotNil(t, validate("ap", "app_name"))                      // Too short
	assert.NotNil(t, validate("thisappnameiswaytoolong", "app_name")) // Too long
}

func TestValidateVersion(t *testing.T) {
	assert.Nil(t, validate("valid.versionname", "version_name"))
	assert.Nil(t, validate("version123", "version_name"))
	assert.Nil(t, validate("version.name123", "version_name"))
	assert.NotNil(t, validate("version_name123", "version_name"))
	assert.NotNil(t, validate("invalid.versionname!", "version_name"))             // Contains special characters other than dot
	assert.NotNil(t, validate("ta", "version_name"))                               // Too short
	assert.NotNil(t, validate("this.versionname.is.way.too.long", "version_name")) // Too long
}

func TestValidatePassword(t *testing.T) {
	assert.Nil(t, validate("validpassword._-", "password"))
	assert.NotNil(t, validate("validpassword!", "password"))
	assert.Nil(t, validate("valid_pass123", "password"))
	assert.Nil(t, validate("InvalidPassword", "password")) // Contains uppercase
	assert.NotNil(t, validate("valid!@#", "password"))     // Contains special characters
	assert.NotNil(t, validate("1234567", "password"))      // Too short
	assert.Nil(t, validate("12345678", "password"))
	assert.NotNil(t, validate("thispasswordiswaytoolong_xxxxx!", "password")) // Too long
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
	assert.Nil(t, validate("admin@admin.com", "email"))
	assert.NotNil(t, validate("@admin.com", "email"))
	assert.NotNil(t, validate("admin@.com", "email"))
	assert.NotNil(t, validate("admin@admin.", "email"))
	assert.NotNil(t, validate("adminadmin.com", "email"))
	assert.NotNil(t, validate("admin@admincom", "email"))

	thirtyCharacters := "abcdefghijklmnopqrstuvwxyz1234"
	validEmail := fmt.Sprintf("%s@%s.de", thirtyCharacters, thirtyCharacters)
	assert.Nil(t, validate(validEmail, "email"))
	tooLongEmail := fmt.Sprintf("%s@%s.com", thirtyCharacters, thirtyCharacters)
	assert.NotNil(t, validate(tooLongEmail, "email"))
}

func TestValidateNumber(t *testing.T) {
	assert.Nil(t, validate("0", "number"))
	assert.Nil(t, validate("1", "number"))
	assert.NotNil(t, validate("-1", "number"))
	assert.NotNil(t, validate("a", "number"))
	assert.NotNil(t, validate("A", "number"))
	assert.NotNil(t, validate("z", "number"))
	assert.NotNil(t, validate("Z", "number"))
	assert.NotNil(t, validate("-", "number"))
	assert.NotNil(t, validate("_", "number"))
	assert.NotNil(t, validate(".", "number"))
	assert.NotNil(t, validate(",", "number"))

	twentyDigitNumber := "01234567890123456789"
	assert.Nil(t, validate(twentyDigitNumber, "number"))
	assert.NotNil(t, validate(twentyDigitNumber+"0", "number"))
}

func TestSearchTerm(t *testing.T) {
	assert.Nil(t, validate("", "search_term"))
	assert.Nil(t, validate("a", "search_term"))
	assert.Nil(t, validate("1", "search_term"))
	assert.Nil(t, validate("0123456789abcdefghij", "search_term"))
	assert.NotNil(t, validate("asdf!", "search_term"))
}
