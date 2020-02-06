package vm

import "testing"

func makeValue(i interface{}) Value {
	value := Value{}

	switch v := i.(type) {
	case bool:
		{
			value.ValueType = Bool
			value.Boolean = v
		}
	case float64:
		{
			value.ValueType = Number
			value.Float = v
		}
	}

	return value
}

func TestArray_Write(t *testing.T) {
	tcs := []struct {
		name string
		init []interface{}
	}{
		{
			name: "Happy Path",
			init: []interface{}{3.14, true, false},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// Build a slice of values
			vs := make([]Value, 0, len(tc.init))
			for _, v := range tc.init {
				vs = append(vs, makeValue(v))
			}

			// Write values to array
			a := newArray()
			for _, v := range vs {
				a.Write(v)
			}

			// Test Write method
			for i, v := range a.values {
				if v != vs[i] {
					t.Errorf("want %+v, got %+v", v, tc.init[i])
				}
			}
		})
	}
}

func TestArray_At(t *testing.T) {
	tcs := []struct {
		name   string
		init   []interface{}
		input  []int
		output []interface{}
		isErr  bool
	}{
		{
			name:   "Happy Path",
			init:   []interface{}{3.14, true, false},
			input:  []int{0, 1, 2},
			output: []interface{}{3.14, true, false},
			isErr:  false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// Build a slice of values
			vs := make([]Value, 0, len(tc.init))
			for _, v := range tc.init {
				vs = append(vs, makeValue(v))
			}

			// Write values to array
			c := newArray()
			for _, v := range vs {
				c.Write(v)
			}

			// Test At method
			for i := range tc.input {
				v := c.At(tc.input[i])
				if v != vs[i] {
					t.Errorf("got %+v, want %+v", v, tc.output[i])
				}
			}
		})
	}
}
