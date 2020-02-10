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
	OpDefineGlobal
	OpDivide
	OpCall
	OpEqualEqual
	OpGetGlobal
	OpGetLocal
	OpGreater
	OpGreaterEqual
	OpJump
	OpJumpIfFalse
	OpLess
	OpLessEqual
	OpLoop
	OpMinus
	OpMultiply
	OpNil
	OpNot
	OpNotEqual
	OpPop
	OpPrint
	OpReturn
	OpSetGlobal
	OpSetLocal
	OpSubtract
	OpTerminate
	OpValue
)

func (op OpCode) String() string {
	switch op {
	case OpAdd:
		return "OP_ADD"
	case OpCall:
		return "OP_CALL"
	case OpDefineGlobal:
		return "OP_DEFINE_GLOBAL"
	case OpEqualEqual:
		return "OP_EQUAL_EQUAL"
	case OpGetGlobal:
		return "OP_GET_GLOBAL"
	case OpGetLocal:
		return "OP_GET_LOCAL"
	case OpGreater:
		return "OP_GREATER"
	case OpGreaterEqual:
		return "OP_GREATER_EQUAL"
	case OpJump:
		return "OP_JUMP"
	case OpJumpIfFalse:
		return "OP_JUMP_IF_FALSE"
	case OpLess:
		return "OP_LESS"
	case OpLessEqual:
		return "OP_LESS_EQUAL"
	case OpLoop:
		return "OP_LOOP"
	case OpMinus:
		return "OP_MINUS"
	case OpMultiply:
		return "OP_MULTIPLY"
	case OpNotEqual:
		return "OP_NOT_EQUAL"
	case OpSetGlobal:
		return "OP_SET_GLOBAL"
	case OpSetLocal:
		return "OP_SET_LOCAL"
	case OpSubtract:
		return "OP_SUBTRACT"
	case OpPop:
		return "OP_POP"
	case OpPrint:
		return "OP_PRINT"
	case OpReturn:
		return "OP_RETURN"
	case OpTerminate:
		return "OP_TERMINATE"
	case OpValue:
		return "OP_VALUE"
	default:
		return "OP_UNKNOWN"
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
	c.Write(OpValue, line)
	address := c.Constants.Write(v)
	c.Write(OpCode(address), line)
}

func (c *PCode) WriteIdentifier(identifier string, line int) {
	v := Value{ValueType: Object, Ptr: identifier}
	address := c.Constants.Write(v)
	c.Write(OpCode(address), line)
}

func (c PCode) String() string {
	var s strings.Builder
	fs := make([]*Function, 0)

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
		case OpCall:
			{
				i++
				s.WriteString(fmt.Sprintf(" #%d", int(c.Code[i])))
			}
		case OpValue, OpDefineGlobal, OpGetGlobal, OpSetGlobal:
			{
				i++
				v := c.Constants.At(int(c.Code[i]))
				switch v.ValueType {
				case Object:
					{
						s.WriteString(fmt.Sprintf(" '%s'", v))
						if f, ok := v.Ptr.(*Function); ok {
							s.WriteString(" __fun__")
							fs = append(fs, f)
						}
					}
				default:
					{
						s.WriteString(fmt.Sprintf(" '%s'", v))
					}
				}
			}
		case OpGetLocal, OpSetLocal:
			{
				i++ // ignore depth level
				s.WriteString(fmt.Sprintf(" at %d", c.Code[i]))
			}
		case OpJump, OpJumpIfFalse:
			{
				i++
				offset := int(c.Code[i])
				s.WriteString(fmt.Sprintf(" %d -> %d", offset, i+offset-1))
			}
		case OpLoop:
			{
				i++
				offset := int(c.Code[i])
				s.WriteString(fmt.Sprintf(" %d -> %d", offset, i-offset-1))
			}
		}

		s.WriteString("\n")
	}

	for _, f := range fs {
		s.WriteString("\n__" + f.Name + "__\n")
		s.WriteString(f.PCode.String())
	}

	return s.String()
}
