package normalize

import (
	"regexp"
)

var spaces = regexp.MustCompile(`\s`)

func Bytes(b []byte) []byte {
	return []byte(String(string(b)))
}

func String(s string) string {
	return spaces.ReplaceAllLiteralString(s, "_")
}
