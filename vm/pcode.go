package vm

type array struct {
	values []float64
}

func newArray() *array {
	return &array{
		values: make([]float64, 0, 8),
	}
}

func (a *array) Write(val float64) uint8 {
	a.values = append(a.values, val)
	return uint8(len(a.values) - 1)
}

func (a *array) At(i int) float64 {
	return a.values[i]
}

const (
	OpAdd  uint8 = iota
	OpConstant
	OpDivide
	OpMinus
	OpMultiply
	OpReturn
	OpSubtract
)

type PCode struct {
	code []uint8
	constants array
	lines *RLE
}

func NewPCode() *PCode {
	return &PCode{
		code: make([]uint8, 0, 8),
		lines: NewRLE(),
	}
}

func (c *PCode) Write(op uint8, line int) {
	c.code = append(c.code, op)
	c.lines.Add(line)
}

func (c *PCode) WriteConstant(value float64, line int) {
	c.Write(OpConstant, line)
	address := c.constants.Write(value)
	c.Write(address, line)
}

