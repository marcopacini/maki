package vm

import (
	"fmt"
	"strings"
)

type Function struct {
	name  string
	arity int
	*PCode
}

func NewFunction(n string) *Function {
	return &Function{
		name:  n,
		arity: 0,
		PCode: NewPCode(),
	}
}

func (f Function) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("__%s__\n", f.name))
	builder.WriteString(f.PCode.String())

	return builder.String()
}
