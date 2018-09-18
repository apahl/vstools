package main

import "testing"

func TestLessF32(t *testing.T) {
	cases := []struct {
		a, b float32
		exp  bool
	}{
		{1.0, 2.0, true},
		{1.99, 2.0, true},
		{1.9999, 2.0, false},
		{-2.0001, -2.0, false},
	}
	for _, c := range cases {
		if LessF32(c.a, c.b) != c.exp {
			t.Errorf("Got unexpected result for %v", c)
		}
	}
}
