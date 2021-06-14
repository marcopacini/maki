package vm

import (
	"fmt"
	"strconv"
)

type ValueType uint8

const (
	Array ValueType = iota
	Bool
	Nil
	Number
	Object
	Reference
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
	case Array:
		{
			values, ok := v.Ptr.([]Value)
			if !ok {
				return fmt.Sprintf("Invalid array content :: Value: %v", v.Ptr)
			}
			s := values[0].String()
			for _, v := range values[1:] {
				s += ", " + v.String()
			}
			return "[ " + s + " ]"
		}
	case Bool:
		return strconv.FormatBool(v.Boolean)
	case Nil:
		return "nil"
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
	case Reference:
		{
			v, _ := v.Ptr.(Value)
			return v.String()
		}
	}
	return fmt.Sprintf("UnknownValue :: ValueType=%d", v.ValueType)
}
