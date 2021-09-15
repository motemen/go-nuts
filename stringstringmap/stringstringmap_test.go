package stringstringmap

import (
	"encoding"
	"encoding/base64"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type base64EncodedString string

var (
	_ encoding.TextMarshaler   = (*base64EncodedString)(nil)
	_ encoding.TextUnmarshaler = (*base64EncodedString)(nil)
)

func (s base64EncodedString) MarshalText() ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString([]byte(s))), nil
}

func (s *base64EncodedString) UnmarshalText(b []byte) error {
	b, err := base64.StdEncoding.DecodeString(string(b))
	*s = base64EncodedString(string(b))
	return err
}

type s struct {
	Int       int
	Uint      uint
	Float     float64
	String    string
	Time      time.Time
	Bool      bool
	Omitempty int `stringstringmap:",omitempty"`
	Base64    base64EncodedString
	Embedded
	Enum  enum
	Skip  string `stringstringmap:"-"`
	Named string `stringstringmap:"customname"`
	//lint:ignore U1000 testing purpose
	unexported int
}

type enum int

const (
	e0 enum = iota
	e1
	e2
	e3
)

type Embedded struct {
	String2 string
}

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		marshaled   map[string]string
		unmarshaled interface{}
		error       string
	}{
		{
			name: "complicated struct",
			value: s{
				Int:    -99,
				Uint:   100,
				Float:  3.14,
				String: "foo",
				Time:   time.Unix(12345, 0).UTC(),
				Bool:   true,
				Base64: "hello",
				Embedded: Embedded{
					String2: "bar",
				},
				Enum:  e1,
				Skip:  "skipthis",
				Named: "named",
			},
			unmarshaled: &s{},
			marshaled: map[string]string{
				"Int":        "-99",
				"Uint":       "100",
				"Float":      "3.14",
				"String":     "foo",
				"Time":       "12345",
				"Bool":       "true",
				"Base64":     "aGVsbG8=",
				"String2":    "bar",
				"Enum":       "1",
				"customname": "named",
			},
		},
		{
			name: "unsupported type",
			value: struct {
				C chan struct{}
			}{
				make(chan struct{}),
			},
			error: "encoding field C: unsupported type chan struct {}",
		},
		{
			name: "marshalling pointer",
			value: &struct {
				S string
			}{
				S: "a",
			},
			unmarshaled: &struct{ S string }{},
			marshaled: map[string]string{
				"S": "a",
			},
		},
	}

	e := Encoder{
		OverrideEncode: func(v interface{}, field reflect.StructField) (string, error) {
			if t, ok := v.(time.Time); ok {
				n := t.Unix()
				return strconv.Itoa(int(n)), nil
			}

			return "", ErrSkipOverride
		},
	}

	d := Decoder{
		OverrideDecode: func(s string, v interface{}, field reflect.StructField) error {
			if t, ok := v.(*time.Time); ok {
				n, err := strconv.Atoi(s)
				if err != nil {
					return err
				}
				*t = time.Unix(int64(n), 0)
				return nil
			}

			return ErrSkipOverride
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.error != "" {
				_, err := e.Encode(test.value)
				if err == nil {
					t.Error("should err")
				} else if test.error != err.Error() {
					t.Errorf("expected error %s but got %s", test.error, err.Error())
				}
				return
			}

			m, err := e.Encode(test.value)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.marshaled, m); diff != "" {
				t.Fatalf("comparing marshaled got diff:\n%s", diff)
			}

			unmarshalled := test.unmarshaled
			err = d.Decode(m, unmarshalled)
			if err != nil {
				t.Fatal(err)
			}

			value := test.value
			if reflect.ValueOf(value).Kind() == reflect.Ptr {
				value = indirect(value)
			}
			if diff := cmp.Diff(
				value,
				indirect(unmarshalled),
				cmp.FilterPath(func(p cmp.Path) bool { return p.String() == "Skip" }, cmp.Ignore()),
				cmpopts.IgnoreUnexported(value),
			); diff != "" {
				t.Fatalf("comparing unmarshaled got diff:\n%s", diff)
			}
		})
	}
}

func indirect(v interface{}) interface{} {
	return reflect.Indirect(reflect.ValueOf(v)).Interface()
}

func TestEncodeDecode_Omitempty(t *testing.T) {
	e := Encoder{
		Omitempty: true,
	}
	d := Decoder{
		Omitempty: true,
	}

	v := struct {
		Int    int
		String string
	}{
		0,
		"",
	}

	m, err := e.Encode(v)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	err = d.Decode(m, &v)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
}
