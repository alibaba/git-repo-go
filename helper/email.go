package helper

import (
	"regexp"
)

var (
	UserEmailPattern = regexp.MustCompile(`^(.*?)\s*<([^@\s]+)@(\S*)>$`)
	EmailPattern     = regexp.MustCompile(`^<?([^@\s]+)@(\S+)>?$`)
)

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
