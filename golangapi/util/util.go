package util

import "strings"

func ParseZipCode(s string) string {
	return strings.Join(strings.Split(s, "-"), "")
}
