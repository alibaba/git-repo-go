package helper

import (
	"regexp"
)

// Email patterns
var (
	UserEmailPattern = regexp.MustCompile(`^(.*?)\s*<([^@\s]+)@(\S*)>$`)
	EmailPattern     = regexp.MustCompile(`^<?([^@\s]+)@(\S+)>?$`)
)

// GetLoginFromEmail gets login name from email address.
func GetLoginFromEmail(email string) string {
	m := UserEmailPattern.FindStringSubmatch(email)
	if m != nil {
		return m[2]
	}
	m = EmailPattern.FindStringSubmatch(email)
	if m == nil {
		return ""
	}
	return m[1]
}
