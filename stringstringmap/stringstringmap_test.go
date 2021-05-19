package stringstringmap

import (
	"encoding"
	"encoding/base64"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
	Enum enum
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

func TestEncode(t *testing.T) {
	tests := []struct {
		v     interface{}
		m     map[string]string
		error string
	}{
		{
			v: s{
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
				Enum: e1,
			},
			m: map[string]string{
				"Int":     "-99",
				"Uint":    "100",
				"Float":   "3.14",
				"String":  "foo",
				"Time":    "12345",
				"Bool":    "true",
				"Base64":  "aGVsbG8=",
				"String2": "bar",
				"Enum":    "1",
			},
		},
		{
			v: struct {
				C chan struct{}
			}{
				make(chan struct{}),
			},
			error: "encoding field C: unsupported type chan struct {}",
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
		if test.error != "" {
			_, err := e.Encode(test.v)
			if err == nil {
				t.Error("should err")
			} else if test.error != err.Error() {
				t.Errorf("expected error %s but got %s", test.error, err.Error())
			}
			continue
		}

		m, err := e.Encode(test.v)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(test.m, m); diff != "" {
			t.Fatalf("got diff:\n%s", diff)
		}

		v2 := reflect.New(reflect.TypeOf(test.v)).Elem().Interface().(s)
		err = d.Decode(m, &v2)
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(test.v, v2); diff != "" {
			t.Fatalf("got diff:\n%s", diff)
		}
	}
}
