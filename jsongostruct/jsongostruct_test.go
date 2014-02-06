package jsongostruct

import (
	"testing"
)

type CamelizeTest struct {
	in  string
	out string
}

func TestCamelize(t *testing.T) {
	specs := []CamelizeTest{
		{"Foo_Bar_Baz", "FooBarBaz"},
		{"Foo_Bar_BAZ", "FooBarBaz"},
		{"Foo_BarBAZ", "FooBarbaz"},
		{"Foo_1Bar", "Foo1bar"},
		{"foo_Bar", "FooBar"},
		{"_foo_bar", "FooBar"},
		{"1hoge_fuga_yap", "1hogeFugaYap"},
	}

	for _, spec := range specs {
		if s := camelize(spec.in); s != spec.out {
			t.Errorf("camelize(%v) should be %v. but got %v", spec.in, spec.out, s)
		}
	}
}
