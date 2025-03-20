package utils

import (
	"github.com/ocelot-cloud/shared/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestZipUnzip(t *testing.T) {
	dir := t.TempDir()
	defer os.RemoveAll(dir)
	sampleContent := "hello"
	sampleFileName := "file.txt"
	sampleFile := filepath.Join(dir, sampleFileName)
	err := os.WriteFile(sampleFile, []byte(sampleContent), 0644)
	assert.Nil(t, err)
	zipBytes, err := ZipDirectoryToBytes(dir)
	assert.Nil(t, err)

	unzippedDir, err := UnzipToTempDir(zipBytes)
	assert.Nil(t, err)

	unzippedFile := filepath.Join(unzippedDir, sampleFileName)
	unzippedBytes, err := os.ReadFile(unzippedFile)
	assert.Nil(t, err)

	assert.Equal(t, string(unzippedBytes), sampleContent)
}

func TestHash(t *testing.T) {
	hashedString, err := Hash("hello")
	assert.Nil(t, err)
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", hashedString)
}

func TestSaltAndHash(t *testing.T) {
	sampleString := "hello"
	saltedAndHashedString, err := SaltAndHash(sampleString)
	assert.Nil(t, err)

	saltedAndHashedString2, err := SaltAndHash(sampleString)
	assert.Nil(t, err)

	assert.NotEqual(t, saltedAndHashedString, saltedAndHashedString2)

	assert.True(t, DoesMatchSaltedHash(sampleString, saltedAndHashedString))
	assert.True(t, DoesMatchSaltedHash(sampleString, saltedAndHashedString2))
	assert.False(t, DoesMatchSaltedHash(sampleString+"x", saltedAndHashedString))
	assert.False(t, DoesMatchSaltedHash(sampleString+"x", saltedAndHashedString2))
}
