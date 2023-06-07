package links

import (
	"bytes"
	"regexp"
	"strings"

	"git.sr.ht/~kota/memex/normalize"
)

var matchLinks = regexp.MustCompile(`(?U)!?\[.+\]\(.+\)|!?\[\[.+\]\]`)

// Modify returns a copy of slice s after making several modifications to the
// links found within.
//
// 1. "Wiki-style" links are all replaced with markdown style links.
//
// A wiki style link is text surrounded with double square brackets [[example]].
// Normally, the link text and link destination are the same, but they can be
// specified separately using the pipe character like so: [[destination|label]].
//
// 2. All non-image links have .html appended to the destination.
//
// A link may be prefixed with ! to indicate it is an image.
func Modify(s []byte, inputSet map[string]struct{}) []byte {
	m := modifier{
		has: inputSet,
	}
	return matchLinks.ReplaceAllFunc(s, m.replace)
}

type modifier struct {
	has map[string]struct{}
}

func (m modifier) replace(src []byte) []byte {
	if len(src) < 4 {
		// This should never happen! Not long enough to be a link.
		panic(`matched link shorter than 4 characters`)
	}

	// Check if link is an image.
	dest := src
	var image bool
	if bytes.HasPrefix(dest, []byte("!")) {
		dest = bytes.TrimPrefix(dest, []byte("!"))
		image = true
	}

	// Parse wiki or markdown style links.
	var label []byte
	if bytes.HasPrefix(dest, []byte("[[")) {
		dest = bytes.TrimPrefix(dest, []byte("[["))
		dest = bytes.TrimSuffix(dest, []byte("]]"))

		var found bool
		dest, label, found = bytes.Cut(dest, []byte("|"))
		if !found {
			label = dest
		}

		// Check if the link leads anywhere.
		d := strings.TrimPrefix(string(dest), "/")
		if _, ok := m.has[d]; !ok {
			return dest
		}
	} else {
		dest = bytes.TrimPrefix(dest, []byte("["))
		dest = bytes.TrimSuffix(dest, []byte(")"))

		var found bool
		label, dest, found = bytes.Cut(dest, []byte("]("))
		if !found {
			panic(`matched markdown link which did not contain "]("`)
		}
	}

	// Write markdown style link.
	var buf bytes.Buffer
	if image {
		buf.WriteRune('!')
	}
	buf.WriteRune('[')
	buf.Write(label)
	buf.WriteRune(']')
	buf.WriteRune('(')
	buf.Write(normalize.Bytes(dest))
	if !image {
		buf.WriteString(".html")
	}
	buf.WriteRune(')')
	return buf.Bytes()
}
