package vm

import "testing"

func TestArray_Write(t *testing.T) {
	tcs := []struct{
		name string
		init []float64
	} {
		{ name: "Happy Path", init: []float64{ 3.14} },
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			a := newArray()
			for _, v := range tc.init {
				a.Write(v)
			}

			for i, v := range a.values {
				if v != tc.init[i] {
					t.Errorf("want %f, got %f", v, tc.init[i])
				}
			}
		})
	}
}

func TestArray_At(t *testing.T) {
	tcs := []struct{
		name string
		init []float64
		input []int
		output []float64
		isErr bool
	}{
		{
			name:   "",
			init:   []float64{3.14},
			input:  []int{0},
			output: []float64{3.14},
			isErr:  false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c := newArray()
			for _, v := range tc.init {
				c.Write(v)
			}

			for i, _:= range tc.input {
				v := c.At(tc.input[i])
				if v != tc.output[i] {
					t.Errorf("got %f, want %f", v, tc.output[i])
				}
			}
		})
	}
}