package vm

import (
	"fmt"
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
	Boolean bool
	Float   float64
	Ptr     interface{}
}

func (v Value) BoolValue() bool {
	if v.ValueType == Bool && v.Boolean {
		return true
	}
	return false
}

func (v Value) String() string {
	switch v.ValueType {
	case Bool:
		return strconv.FormatBool(v.Boolean)
	case Nil:
		return "Nil"
	case Number:
		return strconv.FormatFloat(v.Float, 'f', 0, 64)
	case Object:
		{
			switch value := v.Ptr.(type) {
			case string:
				return value
			case *Function:
				return value.Name
			case *Native:
				return "<native fun>"
			}
		}
	}
	return fmt.Sprintf("UnknownValue :: ValueType=%d", v.ValueType)
}
