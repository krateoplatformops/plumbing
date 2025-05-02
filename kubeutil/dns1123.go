package kubeutil

import (
	"regexp"
	"strings"
)

const (
	DNS1123NameMaximumLength         = 63
	DNS1123NotAllowedCharacters      = "[^-a-z0-9]"
	DNS1123NotAllowedStartCharacters = "^[^a-z0-9]+"
)

// Most resource types require a name that can
// be used as a DNS subdomain name as defined in RFC 1123.
//
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names
func MakeDNS1123Compatible(name string) string {
	name = strings.ToLower(name)

	nameNotAllowedChars := regexp.MustCompile(DNS1123NotAllowedCharacters)
	name = nameNotAllowedChars.ReplaceAllString(name, "")

	nameNotAllowedStartChars := regexp.MustCompile(DNS1123NotAllowedStartCharacters)
	name = nameNotAllowedStartChars.ReplaceAllString(name, "")

	if len(name) > DNS1123NameMaximumLength {
		name = name[0:DNS1123NameMaximumLength]
	}

	return name
}
