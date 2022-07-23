package wiki

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceLinks(t *testing.T) {
	type test struct {
		input string
		want  string
	}

	tests := []test{
		{
			input: "[[basic]]",
			want:  "[basic](basic.html)",
		},
		{
			input: "This string says [[hello]]!\nHere's a second line!",
			want:  "This string says [hello](hello.html)!\nHere's a second line!",
		},
		{
			input: "[[dest|text]]",
			want:  "[text](dest.html)",
		},
		{
			input: "no links here lol",
			want:  "no links here lol",
		},
		{
			input: "double ]] without link",
			want:  "double ]] without link",
		},
		{
			input: "almost | a link ][",
			want:  "almost | a link ][",
		},
		{
			input: "[not a link]",
			want:  "[not a link]",
		},
		{
			input: "[[link with [[ inside]]",
			want:  "[link with [[ inside](link_with_[[_inside.html)",
		},
		{
			input: "[[link|with [[ inside label]]",
			want:  "[with [[ inside label](link.html)",
		},
		{
			input: "[[link with ] inside]]",
			want:  "[link with ] inside](link_with_]_inside.html)",
		},
		{
			input: "[[]]",
			want:  "[](.html)",
		},
		{
			input: "",
			want:  "",
		},
		{
			input: "[[this]] [[one]] [[has]] [[a]] [[lot]]",
			want:  "[this](this.html) [one](one.html) [has](has.html) [a](a.html) [lot](lot.html)",
		},
	}

	for _, tc := range tests {
		got := ReplaceLinks([]byte(tc.input))
		assert.Equal(t, tc.want, string(got))
	}
}
