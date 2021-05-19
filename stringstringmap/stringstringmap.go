package stringstringmap

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var ErrSkipOverride = fmt.Errorf("skip this override")

type Encoder struct {
	OverrideEncode func(v interface{}, field reflect.StructField) (string, error)
	Omitempty      bool
}

type Decoder struct {
	OverrideDecode func(s string, v interface{}, field reflect.StructField) error
	Omitempty      bool
}

func (e Encoder) encodeToText(v interface{}, field reflect.StructField) (string, error) {
	if e.OverrideEncode != nil {
		s, err := e.OverrideEncode(v, field)
		if err == nil {
			return s, nil
		} else if err != ErrSkipOverride {
			return "", err
		}
	}

	if m, ok := v.(encoding.TextMarshaler); ok {
		b, err := m.MarshalText()
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10), nil

	case reflect.String:
		return rv.String(), nil

	case reflect.Bool:
		return strconv.FormatBool(rv.Bool()), nil

	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64), nil

	case reflect.Array:
	case reflect.Chan:
	case reflect.Complex128:
	case reflect.Complex64:
	case reflect.Func:
	case reflect.Interface:
	case reflect.Invalid:
	case reflect.Map:
	case reflect.Ptr:
	case reflect.Slice:
	case reflect.Struct:
	case reflect.Uintptr:
	case reflect.UnsafePointer:
	}

	return "", fmt.Errorf("unsupported type %T", v)
}

func (d Decoder) decodeFromText(s string, v interface{}, field reflect.StructField) error {
	if d.OverrideDecode != nil {
		err := d.OverrideDecode(s, v, field)
		if err == nil {
			return nil
		} else if err != ErrSkipOverride {
			return err
		}
	}

	if u, ok := v.(encoding.TextUnmarshaler); ok {
		return u.UnmarshalText([]byte(s))
	}

	pv := reflect.ValueOf(v)
	if pv.Kind() != reflect.Ptr {
		return fmt.Errorf("want pointer; got %v (%T)", v, v)
	}

	rv := pv.Elem()
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			return err
		}

		rv.SetInt(i)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 10, 0)
		if err != nil {
			return err
		}

		rv.SetUint(i)
		return nil

	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		rv.SetBool(b)
		return nil

	case reflect.String:
		rv.SetString(s)
		return nil

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		rv.SetFloat(f)
		return nil

	case reflect.Array:
	case reflect.Chan:
	case reflect.Complex128:
	case reflect.Complex64:
	case reflect.Func:
	case reflect.Interface:
	case reflect.Invalid:
	case reflect.Map:
	case reflect.Ptr:
	case reflect.Slice:
	case reflect.Struct:
	case reflect.Uintptr:
	case reflect.UnsafePointer:
	}

	return fmt.Errorf("unsupported type: %s", rv.Type())
}

func (e Encoder) Encode(v interface{}) (map[string]string, error) {
	m := map[string]string{}
	rv := reflect.ValueOf(v)
	return e.encodeToStringStringMap(rv, m)
}

func (e Encoder) encodeToStringStringMap(rv reflect.Value, m map[string]string) (map[string]string, error) {
	rt := rv.Type()

	embeddedIdx := []int{}
	for i, n := 0, rt.NumField(); i < n; i++ {
		if rt.Field(i).Anonymous {
			embeddedIdx = append(embeddedIdx, i)
		}
	}

	// process embedded fields first because they have lower priority
	// TODO(motemen): how about decoding?
	for _, i := range embeddedIdx {
		fv := rv.Field(i)
		_, err := e.encodeToStringStringMap(fv, m)
		if err != nil {
			return nil, fmt.Errorf("encodeToStringStringMap: %w", err)
		}
		continue
	}

	for i, n := 0, rt.NumField(); i < n; i++ {
		fv := rv.Field(i)
		field := rt.Field(i)
		if field.Anonymous {
			continue
		}

		tag := field.Tag.Get("stringstringmap")
		if (e.Omitempty || strings.Contains(tag, ",omitempty")) && fv.IsZero() {
			continue
		}

		var err error
		m[field.Name], err = e.encodeToText(fv.Interface(), field)
		if err != nil {
			return nil, fmt.Errorf("encoding field %s: %w", field.Name, err)
		}
	}

	return m, nil
}

func (d Decoder) decodeFromStringStringMap(rv reflect.Value, m map[string]string) error {
	rt := rv.Type()

	for i, n := 0, rt.NumField(); i < n; i++ {
		fv := rv.Field(i)
		field := rt.Field(i)
		if field.Anonymous {
			err := d.decodeFromStringStringMap(fv, m)
			if err != nil {
				return fmt.Errorf("decoding embedded field %v: %w", rt.Field(i).Name, err)
			}
			continue
		}

		tag := field.Tag.Get("stringstringmap")
		if (d.Omitempty || strings.Contains(tag, ",omitempty")) && m[field.Name] == "" {
			continue
		}

		err := d.decodeFromText(m[field.Name], fv.Addr().Interface(), field)
		if err != nil {
			return fmt.Errorf("decoding field %v: %w", field.Name, err)
		}
	}

	return nil
}

func (d Decoder) Decode(m map[string]string, v interface{}) error {
	pv := reflect.ValueOf(v)
	// make a copy as decodeFromStringStringMap destroys it
	m2 := make(map[string]string, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return d.decodeFromStringStringMap(pv.Elem(), m2)
}

func Marshal(v interface{}) (map[string]string, error) {
	return Encoder{}.Encode(v)
}

func Unmarshal(m map[string]string, v interface{}) error {
	return Decoder{}.Decode(m, v)
}
