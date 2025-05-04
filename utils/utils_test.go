package utils

import (
	"archive/zip"
	"bytes"
	"github.com/ocelot-cloud/shared/assert"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestZipUnzip(t *testing.T) {
	dir := t.TempDir()
	defer RemoveDir(dir)
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

func TestGenerateCookie(t *testing.T) {
	cookie, err := GenerateCookie()
	assert.Nil(t, err)
	assert.NotNil(t, cookie)
	assert.Equal(t, "auth", cookie.Name)
	assert.True(t, len(cookie.Value) > 0)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.True(t, cookie.Expires.After(time.Now()))
	assert.True(t, cookie.Expires.Before(time.Now().Add(31*24*time.Hour)))
}

func TestFindDir(t *testing.T) {
	dir := FindDir("utils")
	assert.Equal(t, "utils", filepath.Base(dir))
}

func TestUnzipToDirLimit(t *testing.T) {
	tempDir := t.TempDir()
	defer RemoveDir(tempDir)

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	f, _ := zw.Create("file.txt")
	_, err := f.Write(bytes.Repeat([]byte("A"), 100))
	assert.Nil(t, err)
	Close(zw)
	
	err = unzipToDir(buf.Bytes(), tempDir, 101)
	assert.Nil(t, err)

	err = unzipToDir(buf.Bytes(), tempDir, 99)
	assert.NotNil(t, err)
	assert.Equal(t, "unpacked data exceeds limit", err.Error())
}
