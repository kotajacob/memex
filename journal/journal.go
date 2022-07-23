package journal

import (
	"bytes"
	"regexp"
)

var journalLink = regexp.MustCompile(`\d\d\d\d-\d\d-\d\d`)

// ReplaceLinks returns a copy of slice s with all instances of journal style
// links replaced by markdown style links.
//
// A journal style link is a date in the format YYYY-MM-DD without any other
// markup needed.
func ReplaceLinks(s []byte) []byte {
	return journalLink.ReplaceAllFunc(s, Surround)
}

// Surround slice s in brackets, and then again in parenthesis to form a
// markdown link. Finally, place a backslash at the end to indicate a hard-line
// break.
func Surround(s []byte) []byte {
	var buf bytes.Buffer
	buf.WriteRune('[')
	buf.Write(s)
	buf.WriteRune(']')
	buf.WriteString("(journal/")
	buf.Write(s)
	buf.WriteString(".html)")
	return buf.Bytes()
}
