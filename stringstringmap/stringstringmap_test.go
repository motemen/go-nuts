package stringstringmap

import (
	"encoding"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

type caseReversedString string

var (
	_ encoding.TextMarshaler   = (*caseReversedString)(nil)
	_ encoding.TextUnmarshaler = (*caseReversedString)(nil)
)

func (s caseReversedString) MarshalText() ([]byte, error) {
	return []byte(strings.ToUpper(string(s))), nil
}

func (s *caseReversedString) UnmarshalText(b []byte) error {
	*s = caseReversedString(strings.ToLower(string(b)))
	return nil
}

type s struct {
	Int       int
	String    string
	Time      time.Time
	Bool      bool
	Omitempty int `stringstringmap:",omitempty"`
	R         caseReversedString
	Embedded
}

type Embedded struct {
	String2 string
}

func TestEncode(t *testing.T) {
	tests := []struct {
		v interface{}
		m map[string]string
	}{
		{
			s{
				Int:    99,
				String: "foo",
				Time:   time.Unix(0, 0).UTC(),
				Bool:   true,
				R:      "foo",
				Embedded: Embedded{
					String2: "bar",
				},
			},
			map[string]string{
				"Int":     "99",
				"String":  "foo",
				"Time":    "1970-01-01T00:00:00Z",
				"Bool":    "true",
				"R":       "FOO",
				"String2": "bar",
			},
		},
	}

	for _, test := range tests {
		m, err := Encode(test.v)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(test.m, m); diff != "" {
			t.Fatalf("got diff:\n%s", diff)
		}

		v2 := reflect.New(reflect.TypeOf(test.v)).Elem().Interface().(s)
		err = Decode(m, &v2)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(test.v, v2); diff != "" {
			t.Fatalf("got diff:\n%s", diff)
		}

	}
}
