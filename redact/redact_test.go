package redact

import (
	"reflect"
	"testing"
)

func TestReplace(t *testing.T) {
	type test struct {
		input []byte
		want  []byte
	}

	tests := []test{
		{
			input: []byte("abc"),
			want:  []byte("abc"),
		},
		{
			input: []byte("REDACT abc"),
			want:  []byte("██████ ███"),
		},
		{
			input: []byte("REDACT abc\nxyz"),
			want:  []byte("██████ ███\n███"),
		},
		{
			input: []byte("abc REDACT abc"),
			want:  []byte("abc ██████ ███"),
		},
		{
			input: []byte("abc REDACT abc UNREDACT"),
			want:  []byte("abc ██████ ███ ████████"),
		},
		{
			input: []byte("abc REDACT\nabc UNREDACT"),
			want:  []byte("abc ██████\n███ ████████"),
		},
		{
			input: []byte("abc REDACT abc UNREDACT abc"),
			want:  []byte("abc ██████ ███ ████████ abc"),
		},
		{
			input: []byte("abc REDACT\nabc UNREDACT abc"),
			want:  []byte("abc ██████\n███ ████████ abc"),
		},
		{
			input: []byte("abc REDACT\tabc UNREDACT abc"),
			want:  []byte("abc ██████\t███ ████████ abc"),
		},
	}

	for _, tc := range tests {
		got := Redact("", tc.input)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %s, got: %s", tc.want, got)
		}
	}
}
