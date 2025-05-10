package validation

import (
	"fmt"
	"github.com/ocelot-cloud/shared/assert"
	"testing"
)

const (
	sixtyThreeHexDecimalLetters = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcde"
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
	assert.NotNil(t, ValidateSecret(sixtyThreeHexDecimalLetters))
	assert.Nil(t, ValidateSecret(sixtyThreeHexDecimalLetters+"f"))
	assert.NotNil(t, ValidateSecret(sixtyThreeHexDecimalLetters+"ff"))
	assert.NotNil(t, ValidateSecret(sixtyThreeHexDecimalLetters+"g"))
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

func TestHost(t *testing.T) {
	assert.Nil(t, validate("localhost", "host"))
	assert.Nil(t, validate("localhost123", "host"))
	assert.Nil(t, validate("example.com", "host"))
	assert.Nil(t, validate("my_example-website.com", "host"))
	assert.Nil(t, validate("a.", "host"))
	assert.Nil(t, validate("", "host"))
	assert.Nil(t, validate(sixtyThreeHexDecimalLetters+"a", "host"))
	assert.NotNil(t, validate(sixtyThreeHexDecimalLetters+"ab", "host"))
}

func TestKnownHosts(t *testing.T) {
	sampleKnownHosts := "# 127.0.0.1:2222 SSH-2.0-OpenSSH_9.9\n# 127.0.0.1:2222 SSH-2.0-OpenSSH_9.9\n[127.0.0.1]:2222 ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQClESlJkZf90J0vZZNdAdvl4SpUDt+/VWpiMR8CYbGal8uu09a7UMP9hTeoPacrJtxXRooll7YWv8QRY+/c6UkZHaU4LCOwDJAATHVvKv1ynaGBzGbWK4sGSyTxuzyTYCzcqc1dO+te8qbHh6MI3mC5fF7U+jqU2pJDBfyHb80su4BmyAcSsRc1LgsrHBEYitfsblLWhwzhVRVvD4fRLasfcqpH7ein5peqJPiPOyBsl8+VEpMrH5AzeYsinD5RC84x+0yTOJEQMCdys+EC5i3/Pv3BJ2T/I9VyUoNfF3y9kcxoUIiSj7/kDDhtgAsC87Sv7n5WKrBzkpFpBurLZIaq+ucDUZunE7mbuntc7BI7FIdwxfZl8AgNGAeTAPsbCRORmdYzGNEbgbymMUeNmZYNcrykE8SAsGaaewM+5HnR6x7q7GSHarfIeVSWUDwhMcMCptrsIcSOZlJHEq4hDsb+cILLHQTeOmjuN7O6mLQw5zauIq39YpfzYj9u0PxLBiU=\n# 127.0.0.1:2222 SSH-2.0-OpenSSH_9.9\n[127.0.0.1]:2222 ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBLO699LJQo4+GPThGkZ12YP10xfcf6Zn17nLKi85M1b4wBcb9iaBSLeRAMdszf41pWbW1BHlvXBUkfVbSaiqqh0=\n# 127.0.0.1:2222 SSH-2.0-OpenSSH_9.9\n[127.0.0.1]:2222 ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIERZ7A/6JHp/4VSE3iKJGPWSV6SnYVfzGGamyHwYDsj4\n# 127.0.0.1:2222 SSH-2.0-OpenSSH_9.9"
	assert.Nil(t, validate(sampleKnownHosts, "known_hosts"))
	assert.NotNil(t, validate(sampleKnownHosts+"!", "known_hosts"))
}

func TestResticBackupIdValidations(t *testing.T) {
	sampleResticBackupId := "06b6458017d1e653195d696653c358e4e6a78772aed17582dd6539287332621f"
	assert.Nil(t, validate(sampleResticBackupId, "restic_backup_id"))
	assert.Nil(t, validate(sixtyThreeHexDecimalLetters+"a", "restic_backup_id"))

	assert.NotNil(t, validate(sampleResticBackupId+"a", "restic_backup_id"))
	assert.NotNil(t, validate(sixtyThreeHexDecimalLetters+"g", "restic_backup_id"))
}
