package vm

import (
	"fmt"
	"time"
)

type Native interface {
	Function(vs []Value) Value
}

type Println struct{}

func (p Println) Function(vs []Value) Value {
	for _, v := range vs {
		fmt.Print(v)
	}
	fmt.Println()
	return Value{ValueType: Nil}
}

type Clock struct{}

func (c Clock) Function(_ []Value) Value {
	return Value{
		ValueType: Number,
		Float:     float64(time.Now().Unix()),
	}
}
