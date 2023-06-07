package redact

import (
	"bufio"
	"bytes"
	"os"
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
func Redact(path string, s []byte, denylist []string) []byte {
	if hidden(path, denylist) {
		s = append([]byte("REDACT\n"), s...)
	}
	return redactExp.ReplaceAllFunc(s, Blackout)
}

// hidden returns whether or not a path should be redacted by default.
func hidden(path string, denylist []string) bool {
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

func Load(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var denylist []string
	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		denylist = append(denylist, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return denylist, nil
}
