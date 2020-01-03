package vm

import (
	"strconv"
)

type ValueType uint8

const (
	Bool ValueType = iota
	Nil
	Number
	Object
)

type Value struct {
	ValueType
	B   bool
	N   float64
	Ptr interface{}
}

func (v Value) String() string {
	switch v.ValueType {
	case Bool:
		return strconv.FormatBool(v.B)
	case Nil:
		return "Nil"
	case Number:
		return strconv.FormatFloat(v.N, 'f', 0, 64)
	case Object:
		{
			if s, ok := v.Ptr.(string); ok {
				return s
			}
		}
	}
	return "UNKNOWN"
}
