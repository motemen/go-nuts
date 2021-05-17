package stringstringmap

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func encodeToText(v interface{}) (string, error) {
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
		return fmt.Sprint(rv.Bool()), nil

	case reflect.Array:
	case reflect.Chan:
	case reflect.Complex128:
	case reflect.Complex64:
	case reflect.Float32:
	case reflect.Float64:
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

	return "", fmt.Errorf("unsupported: %v (type %T)", v, v)
}

func decodeFromText(s string, v interface{}) error {
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
		if s == "true" || s == "false" {
			rv.SetBool(s == "true")
			return nil
		} else {
			return fmt.Errorf("cannot parse to bool: %s", s)
		}

	case reflect.String:
		rv.SetString(s)
		return nil

	case reflect.Array:
	case reflect.Chan:
	case reflect.Complex128:
	case reflect.Complex64:
	case reflect.Float32:
	case reflect.Float64:
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

	return fmt.Errorf("unsupported: %T", v)
}

func Encode(v interface{}) (map[string]string, error) {
	m := map[string]string{}
	rv := reflect.ValueOf(v)
	return encodeToStringStringMap(rv, m)
}

func encodeToStringStringMap(rv reflect.Value, m map[string]string) (map[string]string, error) {
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
		_, err := encodeToStringStringMap(fv, m)
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
		if strings.Contains(tag, ",omitempty") && fv.IsZero() {
			continue
		}

		var err error
		m[field.Name], err = encodeToText(fv.Interface())
		if err != nil {
			return nil, fmt.Errorf("encodeToText: %w", err)
		}
	}

	return m, nil
}

func decodeFromStringStringMap(rv reflect.Value, m map[string]string) error {
	rt := rv.Type()

	for i, n := 0, rt.NumField(); i < n; i++ {
		fv := rv.Field(i)
		field := rt.Field(i)
		if field.Anonymous {
			err := decodeFromStringStringMap(fv, m)
			if err != nil {
				return fmt.Errorf("decoding embedded field %v: %w", rt.Field(i).Name, err)
			}
			continue
		}

		tag := field.Tag.Get("stringstringmap")
		if strings.Contains(tag, ",omitempty") && m[field.Name] == "" {
			continue
		}

		err := decodeFromText(m[field.Name], fv.Addr().Interface())
		if err != nil {
			return fmt.Errorf("decoding field %v: %w", field.Name, err)
		}
	}

	return nil
}

func Decode(m map[string]string, v interface{}) error {
	pv := reflect.ValueOf(v)
	// make a copy as decodeFromStringStringMap destroys it
	m2 := make(map[string]string, len(m))
	for k, v := range m {
		m2[k] = v
	}
	return decodeFromStringStringMap(pv.Elem(), m2)
}
