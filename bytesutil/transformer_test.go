package bytesutil

import "testing"

import (
	"bytes"
	"golang.org/x/text/transform"
	"io/ioutil"
)

func TestRemove(t *testing.T) {
	type test struct {
		remove string
		in     string
		out    string
	}
	tests := []test{
		{
			"\x00",
			"abcde",
			"abcde",
		},
		{
			"\x00",
			"abc\x00de",
			"abcde",
		},
		{
			"\x00",
			"\x00abc\x00de",
			"abcde",
		},
		{
			"\x00",
			"\x00abc\x00de\x00",
			"abcde",
		},
		{
			"\x00",
			"abc\x00\x00\x00de",
			"abcde",
		},
		{
			"\x00",
			"",
			"",
		},
		{
			"\x00",
			"\x00",
			"",
		},
		{
			"\x00\x01",
			"\x00\x01",
			"",
		},
		{
			"\x00\x01",
			"ab\x00cd\x01ef",
			"abcdef",
		},
		{
			"\xe3",
			"あ",
			"\x81\x82",
		},
	}

	for _, test := range tests {
		buf := bytes.NewBufferString(test.in)
		remover := Remove(In(test.remove))
		r := transform.NewReader(buf, remover)
		b, err := ioutil.ReadAll(r)
		if err != nil {
			t.Error(err)
			continue
		}

		if string(b) != test.out {
			t.Errorf("got %q != %q", string(b), test.out)
		}
	}
}
