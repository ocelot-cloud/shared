package validation

import (
	"fmt"
	"strings"
)

// Known limitation: Digest (checksum) validation is not enforced. If an attacker compromises the image registry, it is considered outside the security scope of Ocelot-Cloud.
func CheckComposeUpdateSafety(composeYaml1, composeYaml2 map[string]interface{}) error {
	m1, err := collectImages(composeYaml1)
	if err != nil {
		return err
	}
	m2, err := collectImages(composeYaml2)
	if err != nil {
		return err
	}
	if len(m1) != len(m2) {
		return fmt.Errorf("different number of services")
	}
	for svc, img1 := range m1 {
		img2, ok := m2[svc]
		if !ok || img1 != img2 {
			return fmt.Errorf("service %s image differs", svc)
		}
	}
	return nil
}

func collectImages(compose map[string]interface{}) (map[string]string, error) {
	out := map[string]string{}
	servicesRaw, ok := compose["services"]
	if !ok {
		return nil, fmt.Errorf("no services section")
	}
	services, ok := servicesRaw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid services type")
	}
	for svcName, svcRaw := range services {
		svc, ok := svcRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid service definition")
		}
		imgRaw, ok := svc["image"]
		if !ok {
			return nil, fmt.Errorf("no image field in service")
		}
		img, ok := imgRaw.(string)
		if !ok {
			return out, fmt.Errorf("image must be string")
		}
		out[svcName] = normalizeImage(img)
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
