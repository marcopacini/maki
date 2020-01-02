package vm

import (
	"fmt"
	"strings"
)

type array struct {
	values []Value
}

func newArray() *array {
	return &array{
		values: make([]Value, 0, 8),
	}
}

func (a *array) Write(v Value) uint8 {
	a.values = append(a.values, v)
	return uint8(len(a.values) - 1)
}

func (a *array) At(i int) Value {
	return a.values[i]
}

type OpCode uint8

const (
	OpAdd OpCode = iota
	OpConstant
	OpDefineGlobal
	OpDivide
	OpEqual
	OpEqualEqual
	OpGetGlobal
	OpGreater
	OpGreaterEqual
	OpLess
	OpLessEqual
	OpMinus
	OpMultiply
	OpNil
	OpNot
	OpNotEqual
	OpPop
	OpPrint
	OpReturn
	OpSetGlobal
	OpSubtract
)

func (op OpCode) String() string {
	switch op {
	case OpAdd: return "OP_ADD"
	case OpConstant: return "OP_CONSTANT"
	case OpDefineGlobal: return "OP_DEFINE_GLOBAL"
	case OpGetGlobal: return "OP_GET_GLOBAL"
	case OpSetGlobal: return "OP_SET_GLOBAL"
	case OpMinus: return "OP_MINUS"
	case OpMultiply: return "OP_MULTIPLY"
	case OpPop: return "OP_POP"
	case OpReturn: return "OP_RETURN"
	case OpPrint: return "OP_PRINT"
	default: return "OP_UNKNOWN"
	}
}

type PCode struct {
	Code      []OpCode
	Constants array
	Lines     *RLE
}

func NewPCode() *PCode {
	return &PCode{
		Code:  make([]OpCode, 0, 8),
		Lines: NewRLE(),
	}
}

func (c *PCode) Write(op OpCode, line int) {
	c.Code = append(c.Code, op)
	c.Lines.Add(line)
}

func (c *PCode) WriteConstant(v Value, line int) {
	c.Write(OpConstant, line)
	address := c.Constants.Write(v)
	c.Write(OpCode(address), line)
}

func (c *PCode) WriteIdentifier(identifier string, line int) {
	v := Value{ ValueType: Object, Ptr: identifier }
	address := c.Constants.Write(v)
	c.Write(OpCode(address), line)
}

func (c PCode) String() string {
	var s strings.Builder

	line := -1
	for i := 0; i < len(c.Code); i++ {
		s.WriteString(fmt.Sprintf("%04d", i))

		// Print source code line
		l, _ := c.Lines.At(i)
		if l != line {
			s.WriteString(fmt.Sprintf("%4d ", l))
			line = l
		} else {
			s.WriteString("   | ")
		}

		s.WriteString(fmt.Sprintf("%v", c.Code[i]))

		// Skip next code
		switch c.Code[i] {
		case OpConstant, OpDefineGlobal, OpGetGlobal, OpSetGlobal:
			{
				i++
				addr := int(c.Code[i])
				s.WriteString(fmt.Sprintf(" '%v'", c.Constants.At(addr)))
			}
		}

		s.WriteString("\n")
	}

	return s.String()
}