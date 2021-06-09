package matchers

import (
	"fmt"
	"reflect"

	"github.com/golang/mock/gomock"
)

type String string

var _ gomock.Matcher = (*String)(nil)

func (s String) Matches(x interface{}) bool {
	return fmt.Sprint(x) == string(s)
}

func (s String) String() string {
	return string(s)
}

type functionSpec struct {
	rv reflect.Value
}

var _ gomock.Matcher = (*functionSpec)(nil)

func Function(f interface{}) gomock.Matcher {
	rv := reflect.ValueOf(f)
	rt := rv.Type()
	if rt.Kind() != reflect.Func || rt.NumIn() != 1 {
		panic(fmt.Errorf("must be a function with 1 arg: %v", f))
	}
	return functionSpec{
		rv: reflect.ValueOf(f),
	}
}

func (f functionSpec) Matches(x interface{}) bool {
	rv := reflect.ValueOf(x)
	if !rv.Type().AssignableTo(f.rv.Type().In(0)) {
		return false
	}
	return f.rv.Call([]reflect.Value{rv})[0].Bool()
}

func (f functionSpec) String() string {
	return fmt.Sprintf("%v matching some predicates", f.rv.Type().In(0))
}
