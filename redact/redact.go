package redact

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
)

var redactExp = regexp.MustCompile(`(?s)REDACT(.*(?:UNREDACT)|.*)`)

// Redact will blackout any text that appears between the keywords REDACT and
// UNREDACT. If the UNREDACT is missing the text will be redacted until the end.
//
// Some specific file paths (the whole log folder for example) "default" to
// being redacted. That just means we prepend the REDACT keyword on those files.
func Redact(path string, s []byte) []byte {
	if hidden(path) {
		s = append([]byte("REDACT\n"), s...)
	}
	return redactExp.ReplaceAllFunc(s, Blackout)
}

// hidden returns whether or not a path should be redacted by default.
func hidden(path string) bool {
	denylist := []string{
		"kota.md",
		"jazzi.md",
		"mom.md",
		"brian.md",
		"paul.md",
		"kyrin.md",
		"max.md",
		"leon.md",
		"mary.md",
		"nik.md",
		"henry.md",
		"amanda.md",
		"matthew.md",
		"lucas.md",
		"portainer",
		"todo",
		"log",
		"/journal/",
	}
	for _, v := range denylist {
		if strings.Contains(path, v) {
			return true
		}
	}
	return false
}

// Blackout all non-whitespace characters.
func Blackout(s []byte) []byte {
	var buf bytes.Buffer
	for _, c := range string(s) {
		if unicode.IsSpace(c) {
			buf.WriteRune(c)
		} else {
			buf.WriteString("█")
		}
	}
	return buf.Bytes()
}
