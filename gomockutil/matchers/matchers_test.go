package matchers

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestMatchers(t *testing.T) {
	tests := []struct {
		name    string
		matcher gomock.Matcher
		yes, no []interface{}
	}{
		{"String()", String("123"), []interface{}{123, "123", fmt.Errorf("123")}, []interface{}{123.4, "x"}},
		{"Function()", Function(func(x interface{}) bool {
			rv := reflect.ValueOf(x)
			return (rv.Kind() == reflect.Slice || rv.Kind() == reflect.String) && rv.Len() == 3
		}), []interface{}{"xxx", []int{6, 6, 6}}, []interface{}{1, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, v := range tt.yes {
				if !tt.matcher.Matches(v) {
					t.Errorf("%#v should match %T", v, tt.matcher)
				}
			}
			for _, v := range tt.no {
				if tt.matcher.Matches(v) {
					t.Errorf("%#v should not match %T", v, tt.matcher)
				}
			}
		})
	}
}
