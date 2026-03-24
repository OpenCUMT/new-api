package common

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	RestrictedRegisterEmailDomain = "opencumt.org"
)

var restrictedRegisterEmailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@opencumt\.org$`)

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func IsRestrictedRegisterEmail(email string) bool {
	return restrictedRegisterEmailRegex.MatchString(NormalizeEmail(email))
}

func RestrictedRegisterEmailVerificationRequired() bool {
	return RestrictedRegisterEmailDomain != ""
}

func ValidateRestrictedRegisterEmail(email string) error {
	if !IsRestrictedRegisterEmail(email) {
		return fmt.Errorf("only @%s emails are allowed", RestrictedRegisterEmailDomain)
	}
	return nil
}
