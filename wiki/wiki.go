package wiki

import (
	"bytes"

	"git.sr.ht/~kota/memex/normalize"
)

type linkLocation int

const (
	locNone linkLocation = iota
	locDest
	locLabel
)

// ReplaceLinks returns a copy of slice s with all instances of wiki style links
// replaced by markdown style links.
//
// A wiki style link is text surrounded with double square brackets [[example]].
// Normally, the link text and link destination are the same, but they can be
// specified separately using the pipe character like so: [[destination|label]].
//
// Destinations are modified slightly, any unicode characters are changed
// into similar ascii characters, spaces are replaced with underscores and
// ".html" is added to the end.
func ReplaceLinks(s []byte) []byte {
	var buf bytes.Buffer
	var prev byte
	var loc linkLocation
	var dest bytes.Buffer
	var label bytes.Buffer
	for _, c := range s {
		// A link could contain: [[, or ], or [, but not ]].
		switch c {
		case '[':
			switch loc {
			case locNone:
				if prev != '[' {
					// If we're not in a link and it's not preceeded by another
					// '[' we write it to normal cache and break.
					buf.WriteByte(c)
				} else {
					// Avoid printing current '[' and remove the last '[' from
					// the buffer since we've detected a link.
					buf.Truncate(buf.Len() - 1)
					loc = locDest
				}
			case locDest:
				dest.WriteByte(c)
			case locLabel:
				label.WriteByte(c)
			}
		case '|':
			if loc == locDest {
				loc = locLabel
			} else {
				buf.WriteByte(c)
			}
		case ']':
			if prev != ']' {
				switch loc {
				case locNone:
					buf.WriteByte(c)
				case locDest:
					dest.WriteByte(c)
				case locLabel:
					label.WriteByte(c)
				}
				break
			}
			if loc == locNone {
				// For example having: ]] by itself outside a link.
				buf.WriteByte(c)
				break
			}

			// We're now inside a link having just detected (and avoidded
			// printing) our second `]`. Let's remove previous ']' being careful
			// to do so for text or dest.
			if loc == locLabel {
				label.Truncate(label.Len() - 1)
				buf.WriteRune('[')
				buf.Write(label.Bytes())
			} else {
				dest.Truncate(dest.Len() - 1)
				buf.WriteRune('[')
				buf.Write(dest.Bytes())
			}
			buf.WriteRune(']')
			buf.WriteRune('(')
			if bytes.HasPrefix(dest.Bytes(), []byte("/")) {
				buf.WriteString("/m")
			}
			buf.Write(normalize.Bytes(dest.Bytes()))
			buf.WriteString(".html)")
			label.Reset()
			dest.Reset()
			loc = locNone
		default:
			switch loc {
			case locDest:
				dest.WriteByte(c)
			case locLabel:
				label.WriteByte(c)
			default:
				buf.WriteByte(c)
			}
		}
		prev = c
	}
	return buf.Bytes()
}
