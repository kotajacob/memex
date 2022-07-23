package journal

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
			input: "1917-03-08",
			want:  "[1917-03-08](1917-03-08)",
		},
		{
			input: "1917-03-08 - with comment",
			want:  "[1917-03-08](1917-03-08) - with comment",
		},
		{
			input: "",
			want:  "",
		},
		{
			input: "hello, no links here!",
			want:  "hello, no links here!",
		},
		{
			input: "2020",
			want:  "2020",
		},
		{
			input: "20-20-20",
			want:  "20-20-20",
		},
	}

	for _, tc := range tests {
		got := ReplaceLinks([]byte(tc.input))
		assert.Equal(t, tc.want, string(got))
	}
}
