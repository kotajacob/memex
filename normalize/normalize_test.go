package normalize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	type test struct {
		input string
		want  string
	}
	tests := []test{
		{
			input: "hello world",
			want:  "hello_world",
		},
		{
			input: "hello\tworld",
			want:  "hello_world",
		},
		{
			input: "",
			want:  "",
		},
	}

	for _, tc := range tests {
		got := String(tc.input)
		assert.Equal(t, tc.want, got)
	}
}
