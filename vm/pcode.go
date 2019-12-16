package vm

type ValueType string

const (
	Bool ValueType = "BOOL"
	Nil = "NIL"
	Number = "NUMBER"
)

type Value struct {
	ValueType
	B bool
	N float64
}

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

const (
	OpAdd  uint8 = iota
	OpConstant
	OpDivide
	OpEqual
	OpEqualEqual
	OpFalse
	OpGreater
	OpGreaterEqual
	OpLess
	OpLessEqual
	OpMinus
	OpMultiply
	OpNil
	OpNot
	OpNotEqual
	OpReturn
	OpSubtract
	OpTrue
)

type PCode struct {
	Code      []uint8
	Constants array
	Lines     *RLE
}

func NewPCode() *PCode {
	return &PCode{
		Code:  make([]uint8, 0, 8),
		Lines: NewRLE(),
	}
}

func (c *PCode) Write(op uint8, line int) {
	c.Code = append(c.Code, op)
	c.Lines.Add(line)
}

func (c *PCode) WriteConstant(v Value, line int) {
	c.Write(OpConstant, line)
	address := c.Constants.Write(v)
	c.Write(address, line)
}

