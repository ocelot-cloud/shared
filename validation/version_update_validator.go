package validation

import (
	"fmt"
	"strings"
)

// Known limitation: Digest (checksum) validation is not enforced. If an attacker compromises the image registry, it is considered outside the security scope of Ocelot-Cloud.
func CheckComposeUpdateSafety(composeYaml1, composeYaml2 map[string]interface{}) error {
	set1, err := collectImages(composeYaml1)
	if err != nil {
		return err
	}
	set2, err := collectImages(composeYaml2)
	if err != nil {
		return err
	}
	if len(set1) != len(set2) {
		return fmt.Errorf("different number of images")
	}
	for img := range set1 {
		if _, ok := set2[img]; !ok {
			return fmt.Errorf("image %s differs", img)
		}
	}
	return nil
}

func collectImages(compose map[string]interface{}) (map[string]struct{}, error) {
	out := map[string]struct{}{}
	servicesRaw, ok := compose["services"]
	if !ok {
		return nil, fmt.Errorf("no services section")
	}
	services, ok := servicesRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid services type")
	}
	for _, svcRaw := range services {
		svc, ok := svcRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid service definition")
		}
		imgRaw, ok := svc["image"]
		if !ok {
			continue
		}
		img, ok := imgRaw.(string)
		if !ok {
			return out, fmt.Errorf("image must be string")
		}
		out[normalizeImage(img)] = struct{}{}
	}
	return out, nil
}

func normalizeImage(imageString string) string {
	if i := strings.Index(imageString, "@"); i != -1 {
		imageString = imageString[:i]
	}
	lastSlash := strings.LastIndex(imageString, "/")
	colon := strings.LastIndex(imageString, ":")
	if colon > lastSlash {
		imageString = imageString[:colon]
	}
	return imageString
}
