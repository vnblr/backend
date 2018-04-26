package commute

import (
	"testing"
)

func TestDummy(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"one", "SomeStr:one"},
		{"two", "SomeStr:two"},
		{"three", "SomeStr:three"},
	}
	for _, c := range cases {
		got := dummyFun(c.in)
		if got != c.want {
			t.Errorf("dummyFun(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
