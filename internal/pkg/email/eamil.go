package email

import (
	"regexp"
	"strings"
)

var emailRegexp = regexp.MustCompile(
	`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`,
)

// CanonicalEmail 1. Trim spaces 2. Convert to lower case
func CanonicalEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func IsValidEmail(email string) bool {
	return emailRegexp.MatchString(email)
}
