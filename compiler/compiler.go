package compiler

import (
	"fmt"
	"maki/vm"
	"strconv"
)

type parser struct {
	*scanner
	current  *Token
	previous *Token
	hadError bool
}

func newParser(source string) *parser {
	return &parser{
		scanner:  newScanner(source),
		current:  nil,
		previous: nil,
		hadError: false,
	}
}

func (p *parser) advance() error {
	p.previous = p.current

	var err error
	if p.current, err = p.scanToken(); err != nil {
		return err
	}

	return nil
}

func (p *parser) consume(tt TokenType, msg string, a ...interface{}) error {
	if p.current.TokenType == tt {
		_ = p.advance()
		return nil
	}

	return fmt.Errorf(msg, a...)
}

type Compiler struct {
	*vm.PCode
	*parser
}

func NewCompiler() *Compiler {
	return &Compiler{
		PCode: vm.NewPCode(),
	}
}

type precedence uint8

const (
	PrecNone precedence = iota
	PrecAssignment
	PrecOr
	PrecAnd
	PrecEquality
	PrecComparison
	PrecTerm
	PrecFactor
	PrecUnary
	PrecCall
	PrecPrimary

)

type rule struct {
	prefix func(*Compiler) error
	infix func(*Compiler) error
	precedence
}

func getRule(tt TokenType) rule {
	var rules = map[TokenType]rule {
		False: { prefix: (*Compiler).literal, infix: nil, precedence: PrecNone },
		LeftParenthesis: { prefix: (*Compiler).grouping, infix: nil, precedence: PrecNone },
		Minus: { prefix: (*Compiler).unary, infix: (*Compiler).binary, precedence: PrecTerm },
		Nil: { prefix: (*Compiler).literal, infix: nil, precedence: PrecNone },
		Number: { prefix: (*Compiler).number, infix: nil, precedence: PrecNone },
		Plus: { prefix: nil, infix: (*Compiler).binary, precedence: PrecTerm },
		Slash: { prefix: nil, infix: (*Compiler).binary, precedence: PrecFactor },
		Star: { prefix: nil, infix: (*Compiler).binary, precedence: PrecFactor },
		True: { prefix: (*Compiler).literal, infix: nil, precedence: PrecNone },
	}

	if r, ok := rules[tt]; ok {
		return r
	}

	return rule{
		prefix:     nil,
		infix:      nil,
		precedence: PrecNone,
	}
}

func (c *Compiler) parsePrecedence(prec precedence) error {
	if err := c.advance(); err != nil {
		return err
	}

	prefix := getRule(c.previous.TokenType).prefix

	if prefix == nil {
		return fmt.Errorf("error at line %d: expected expression", c.current.Line)
	}

	if err := prefix(c); err != nil {
		return err
	}

	for prec < getRule(c.current.TokenType).precedence {
		if err := c.advance(); err != nil {
			return err
		}

		infix := getRule(c.previous.TokenType).infix

		if err := infix(c); err != nil {
			return err
		}
	}

	return nil
}

func (c *Compiler) Compile(source string) (*vm.PCode, error) {
	c.parser = newParser(source)

	if err := c.advance(); err != nil {
		c.hadError = true
		return nil, err
	}

	if err := c.expression(); err != nil {
		return nil, err
	}

	if err := c.consume(Eof, "error at line %d: expected EOF", c.current.Line); err != nil {
		c.hadError = true
		return nil, err
	}

	// Temporary: expression-only support
	c.emitByte(vm.OpReturn)

	return c.PCode, nil
}

func (c *Compiler) expression() error {
	if err := c.parsePrecedence(PrecAssignment); err != nil {
		return err
	}

	return nil
}

func (c *Compiler) literal() error {
	switch c.previous.TokenType {
	case False:
		{
			c.emitByte(vm.OpFalse)
		}
	case Nil:
		{
			c.emitByte(vm.OpNil)
		}
	case True:
		{
			c.emitByte(vm.OpTrue)
		}
	}

	return nil
}

func (c *Compiler) number()	error {
	n, err := strconv.ParseFloat(c.previous.Lexeme, 64)
	if err != nil {
		return err
	}

	v := vm.Value{
		ValueType: vm.Number,
		N: n,
	}

	c.emitConstant(v)
	return nil
}

func (c *Compiler) grouping() error {
	if err := c.expression(); err != nil {
		return err
	}
	return c.consume(RightParenthesis, "Expect ')' after expression")
}

func (c *Compiler) unary() error {
	tt := c.previous.TokenType

	if err := c.expression(); err != nil {
		return err
	}

	switch tt {
	case Minus:
		{
			c.emitByte(vm.OpMinus)
			break
		}
	}

	return nil
}

func (c *Compiler) binary()	error {
	tt := c.previous.TokenType
	if err := c.parsePrecedence(getRule(tt).precedence); err != nil {
		return err
	}

	switch tt {
	case Plus:
		{
			c.emitByte(vm.OpAdd)
			break
		}
	case Minus:
		{
			c.emitByte(vm.OpSubtract)
			break
		}
	case Star:
		{
			c.emitByte(vm.OpMultiply)
			break
		}
	case Slash:
		{
			c.emitByte(vm.OpDivide)
			break
		}
	}

	return nil
}

func (c *Compiler) emitByte(byte uint8) {
	c.Write(byte, c.current.Line)
}

func (c *Compiler) emitBytes(bytes ...uint8) {
	for _, b := range bytes {
		c.emitByte(b)
	}
}

func (c *Compiler) emitConstant(v vm.Value) {
	c.WriteConstant(v, c.current.Line)
}