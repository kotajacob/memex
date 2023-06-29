package links

import (
	"bytes"
	"regexp"
	"strings"

	"git.sr.ht/~kota/memex/normalize"
)

type linkForm uint8

const (
	wikiLink linkForm = iota
	mdLink
	wikiImgLink
	mdImgLink
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
// 2. All non-image wiki style links have .html appended to the destination.
//
// A link may be prefixed with ! to indicate it is an image.
func Modify(
	s []byte,
	linkMap map[string]map[string]struct{},
) []byte {
	m := modifier{
		linkMap: linkMap,
	}
	return matchLinks.ReplaceAllFunc(s, m.replace)
}

type modifier struct {
	linkMap map[string]map[string]struct{}
}

func Map(
	s []byte,
	name string,
	linkMap map[string]map[string]struct{},
) map[string]map[string]struct{} {
	matches := matchLinks.FindAll(s, -1)
	for _, match := range matches {
		dest, _, form := parseLink(match)
		if form != wikiLink {
			continue
		}
		d := strings.TrimPrefix(string(dest), "/")
		if _, ok := linkMap[d]; !ok {
			continue
		}
		linkMap[d][name] = struct{}{}
	}
	return linkMap
}

func (m *modifier) replace(src []byte) []byte {
	var label []byte
	var form linkForm
	dest, label, form := parseLink(src)

	// Remove dead internal links.
	if form == wikiLink {
		d := strings.TrimPrefix(string(dest), "/")
		if _, ok := m.linkMap[d]; !ok {
			return dest
		}
	}

	// Write markdown style link.
	var buf bytes.Buffer
	if form == wikiImgLink || form == mdImgLink {
		buf.WriteRune('!')
	}
	buf.WriteRune('[')
	buf.Write(label)
	buf.WriteRune(']')
	buf.WriteRune('(')
	buf.Write(normalize.Bytes(dest))
	if form == wikiLink {
		buf.WriteString(".html")
	}
	buf.WriteRune(')')
	return buf.Bytes()
}

func parseLink(s []byte) ([]byte, []byte, linkForm) {
	// Check if link is an image.
	var image bool
	if bytes.HasPrefix(s, []byte("!")) {
		s = bytes.TrimPrefix(s, []byte("!"))
		image = true
	}

	// Parse wiki style links.
	dest, label, found := parseWikiLink(s)
	if found {
		if image {
			return dest, label, wikiImgLink
		}
		return dest, label, wikiLink
	}

	// Parse markdown style links.
	dest, label, _ = parseMDLink(s)
	if image {
		return dest, label, mdImgLink
	}
	return dest, label, mdLink
}

func parseWikiLink(s []byte) ([]byte, []byte, bool) {
	if !bytes.HasPrefix(s, []byte("[[")) {
		return nil, nil, false
	}
	dest := bytes.TrimPrefix(s, []byte("[["))
	dest = bytes.TrimSuffix(dest, []byte("]]"))

	var label []byte
	var found bool
	dest, label, found = bytes.Cut(dest, []byte("|"))
	if !found {
		label = dest
	}
	return dest, label, true
}

func parseMDLink(s []byte) ([]byte, []byte, bool) {
	if !bytes.HasPrefix(s, []byte("[")) {
		return nil, nil, false
	}
	dest := bytes.TrimPrefix(s, []byte("["))
	dest = bytes.TrimSuffix(dest, []byte(")"))

	var label []byte
	var found bool
	label, dest, found = bytes.Cut(dest, []byte("]("))
	if !found {
		panic(`matched markdown link which did not contain "]("`)
	}
	return dest, label, true
}

// // Update the linkMap.
// m.linkMap[d] = append(m.linkMap[d], m.name)
