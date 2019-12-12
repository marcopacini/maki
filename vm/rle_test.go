package vm

import "testing"

func TestRLE_Add(t *testing.T) {
	tcs := []struct {
		name string
		input []int
	} {
		{
			name: "Happy Path",
			input: []int{ 1, 2, 2, 3, 3, 3 },
		},
		{
			name: "No duplication",
			input: []int{ 1, 2, 3, 4, 5 },
		},
		{
			name: "All equal",
			input: []int{ 42, 42, 42 },
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			rle := NewRLE()
			for _, v := range tc.input {
				rle.Add(v)
			}

			for i, v := range tc.input {
				val, err := rle.At(i)
				if err != nil {
					t.Errorf("got %v, want nil", err.Error())
				}

				if val != v {
					t.Errorf("got %d, want %d", val, v)
				}
			}
		})
	}
}

func TestRLE_At(t *testing.T) {
	tcs := []struct{
		name string
		init []int
		input []int
		output []int
		isErr bool
	} {
		{
			name: "Happy Path",
			init: []int{ 1, 1, 1, 2, 2 },
			input: []int { 2 },
			output: []int { 1 },
			isErr: false,
		},
		{
			name: "Out of range",
			init: []int{},
			input: []int { 0, 1, 5 },
			output: []int{},
			isErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			rle := NewRLE()
			for _, v := range tc.init {
				rle.Add(v)
			}

			for i, _ := range tc.input {
				v, err := rle.At(tc.input[i])
				if tc.isErr {
					if err == nil {
						t.Errorf("got nil, want out of range")
					}
				} else {
					if err != nil {
						t.Errorf("got %v, want nil", err.Error())
					}

					if v != tc.output[i] {
						t.Errorf("got %d, want %d", v, tc.output)
					}
				}
			}
		})
	}
}