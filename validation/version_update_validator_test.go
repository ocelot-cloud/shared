package validation

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TODO post box: warn admin if auto update failed because unsafe

const (
	defaultSampleImage1          = "default-1.0"
	defaultSampleImage2          = "default-2.0"
	differentNameImage           = "different-image-name"
	unequalCount                 = "two-images"
	duplicateImages              = "duplicate-images"
	defaultSampleImageWithShaSum = "default-2.0-with-shasum"
	registryPortImage1           = "registry-port-1.0"
	registryPortImage2           = "registry-port-2.0"
	differentRegistryPort2       = "different-registry-port-2.0"
)

func TestIsComposeUpdateSafe(t *testing.T) {
	tests := []struct {
		name    string
		f1      string
		f2      string
		wantErr bool
	}{
		{"same_images_same_tags", defaultSampleImage1, defaultSampleImage1, false},
		{"same_images_diff_tags", defaultSampleImage1, defaultSampleImage2, false},
		{"image_name_diff", defaultSampleImage1, differentNameImage, true},
		{"unequal_count", defaultSampleImage1, unequalCount, true},
		{"duplicate_images", defaultSampleImage1, duplicateImages, false},
		{"digest_vs_tag", defaultSampleImage1, defaultSampleImageWithShaSum, false},
		{"custom_registry_port", registryPortImage1, registryPortImage2, false},
		{"registry_port_changes", registryPortImage1, differentRegistryPort2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFileDir := "./samples/update-validation-test/"
			c1, err := readCompose(testFileDir + tt.f1 + ".yml")
			assert.NoError(t, err)
			c2, err := readCompose(testFileDir + tt.f2 + ".yml")
			assert.NoError(t, err)

			err = CheckComposeUpdateSafety(c1, c2)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCollectImages(t *testing.T) {
	valid := map[string]interface{}{
		"services": map[string]interface{}{
			"web": map[string]interface{}{"image": "nginx:latest"},
			"db":  map[string]interface{}{"image": "mysql:8.0"},
		},
	}
	images, err := collectImages(valid)
	assert.NoError(t, err)
	assert.Len(t, images, 2)

	noServices := map[string]interface{}{}
	_, err = collectImages(noServices)
	assert.Error(t, err)

	badType := map[string]interface{}{"services": []int{1, 2}}
	_, err = collectImages(badType)
	assert.Error(t, err)

	notString := map[string]interface{}{
		"services": map[string]interface{}{
			"bad": map[string]interface{}{"image": 42},
		},
	}
	_, err = collectImages(notString)
	assert.Error(t, err)
}

func TestNormalizeImage(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"nginx:latest", "nginx"},
		{"repo/nginx@sha256:abcd", "repo/nginx"},
		{"registry:5000/repo:1.0", "registry:5000/repo"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.out, normalizeImage(tt.in))
	}
}
