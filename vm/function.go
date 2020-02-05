package vm

import (
	"fmt"
	"strings"
)

type Function struct {
	Name  string
	Arity int
	*PCode
}

func NewFunction(n string) *Function {
	return &Function{
		Name:  n,
		Arity: 0,
		PCode: NewPCode(),
	}
}

func (f Function) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("__%s__\n", f.Name))
	builder.WriteString(f.PCode.String())

	return builder.String()
}
