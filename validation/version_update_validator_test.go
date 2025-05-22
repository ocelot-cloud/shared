package validation

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

/* TODO cases
unequal number of distinct images → false, error
same images, different tags → true, nil
same images, same tags → true, nil
same count, at least one image name differs → false, error
duplicate images within a file but identical unique sets → true, nil
custom registry with port (registry:5000/repo:tag) vs different tag → true, nil
digest vs tag (repo@sha256:… vs repo:tag) → true, nil
compose without services section → false, error
service image value not a string → false, error
service lacking image field in one file causes set mismatch → false, error
no images contained in a file

post box: warn admin if auto update failed because unsafe
handle special sign like "@", ":" and "/"
*/

// TODO post box: warn admin if auto update failed because unsafe

const (
	defaultSampleImage1          = "default-1.0"
	defaultSampleImage2          = "default-2.0"
	differentNameImage           = "different-image-name"
	unequalCount                 = "two-images"
	duplicateImages              = "duplicate-images"
	defaultSampleImageWithShaSum = "default-2.0-with-shasum"
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
		/*
			{"custom_registry_port", "port_tag_1", "port_tag_2", false},
			{"custom_registry_port", "port_tag_1", "port_tag_2", false},
			{"registry_port_changes", "port_tag_1", "port_tag_2", true},
			{"no_services_section", "no_services", "same_tags_1", true},
			{"image_not_string", "not_string", "same_tags_1", true},
			{"no_images", "no_images", "no_images", false},
		*/
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
