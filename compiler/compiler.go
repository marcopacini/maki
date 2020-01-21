package compiler

import (
	"fmt"
	"maki/vm"
	"strconv"
)

type Compiler struct {
	*vm.PCode
	*parser
	*scope
}

func NewCompiler() *Compiler {
	return &Compiler{
		scope: newScope(),
	}
}

type precedence uint8

const (
	PrecNone       precedence = iota
	PrecAssignment            // =
	PrecOr                    // or
	PrecAnd                   // and
	PrecEquality              // == !=
	PrecComparison            // < > <= >=
	PrecTerm                  // + -
	PrecFactor                // * /
	PrecUnary                 // not !
	PrecCall                  // . ()
	PrecPrimary
)

type rule struct {
	prefix func(*Compiler, bool) error
	infix  func(*Compiler, bool) error
	precedence
}

func getRule(tt TokenType) rule {
	var rules = map[TokenType]rule{
		And:             {prefix: nil, infix: (*Compiler).and, precedence: PrecAnd},
		Equal:           {prefix: nil, infix: nil, precedence: PrecNone},
		EqualEqual:      {prefix: nil, infix: (*Compiler).binary, precedence: PrecEquality},
		False:           {prefix: (*Compiler).literal, infix: nil, precedence: PrecNone},
		Greater:         {prefix: nil, infix: (*Compiler).binary, precedence: PrecComparison},
		GreaterEqual:    {prefix: nil, infix: (*Compiler).binary, precedence: PrecComparison},
		Identifier:      {prefix: (*Compiler).identifier, infix: nil, precedence: PrecNone},
		LeftParenthesis: {prefix: (*Compiler).grouping, infix: nil, precedence: PrecNone},
		Less:            {prefix: nil, infix: (*Compiler).binary, precedence: PrecComparison},
		LessEqual:       {prefix: nil, infix: (*Compiler).binary, precedence: PrecComparison},
		Minus:           {prefix: (*Compiler).unary, infix: (*Compiler).binary, precedence: PrecTerm},
		Nil:             {prefix: (*Compiler).literal, infix: nil, precedence: PrecNone},
		Not:             {prefix: (*Compiler).unary, infix: nil, precedence: PrecNone},
		NotEqual:        {prefix: nil, infix: (*Compiler).binary, precedence: PrecEquality},
		Number:          {prefix: (*Compiler).number, infix: nil, precedence: PrecNone},
		Plus:            {prefix: nil, infix: (*Compiler).binary, precedence: PrecTerm},
		Slash:           {prefix: nil, infix: (*Compiler).binary, precedence: PrecFactor},
		Star:            {prefix: nil, infix: (*Compiler).binary, precedence: PrecFactor},
		String:          {prefix: (*Compiler).string, infix: nil, precedence: PrecNone},
		True:            {prefix: (*Compiler).literal, infix: nil, precedence: PrecNone},
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
	last := c.previous

	if err := c.advance(); err != nil {
		return err
	}

	prefix := getRule(c.previous.TokenType).prefix

	if prefix == nil {
		return fmt.Errorf("compile error, expected expression after '%s' [line %d]", last.Lexeme, c.current.Line)
	}

	assignable := prec <= PrecAssignment
	if err := prefix(c, assignable); err != nil {
		return err
	}

	for prec <= getRule(c.current.TokenType).precedence {
		if err := c.advance(); err != nil {
			return err
		}

		infix := getRule(c.previous.TokenType).infix

		if err := infix(c, false); err != nil {
			return err
		}
	}

	if assignable && c.current.TokenType == Equal {
		return fmt.Errorf("compile error, invalid assignment target [line %d]", c.current.Line)
	}

	return nil
}

func (c *Compiler) Compile(source string) (*vm.PCode, error) {
	c.PCode = vm.NewPCode()
	c.parser = newParser(source)

	if err := c.advance(); err != nil {
		return nil, err
	}

	for !c.match(Eof) {
		if err := c.declaration(false); err != nil {
			return nil, err
		}
	}

	if err := c.consume(Eof); err != nil {
		return nil, err
	}

	// temporary exit statement
	c.emitByte(vm.OpReturn)

	return c.PCode, nil
}

func (c *Compiler) and(_ bool) error {
	jump := c.emitJump(vm.OpJumpIfFalse)

	c.emitByte(vm.OpPop)
	if err := c.parsePrecedence(PrecAnd); err != nil {
		return err
	}

	c.applyPatch(jump)
	return nil
}

func (c *Compiler) binary(_ bool) error {
	tt := c.previous.TokenType
	if err := c.parsePrecedence(getRule(tt).precedence); err != nil {
		return err
	}

	switch tt {
	case EqualEqual:
		c.emitByte(vm.OpEqualEqual)
	case Greater:
		c.emitByte(vm.OpGreater)
	case GreaterEqual:
		c.emitByte(vm.OpGreaterEqual)
	case Less:
		c.emitByte(vm.OpLess)
	case LessEqual:
		c.emitByte(vm.OpLessEqual)
	case Minus:
		c.emitByte(vm.OpSubtract)
	case Plus:
		c.emitByte(vm.OpAdd)
	case Star:
		c.emitByte(vm.OpMultiply)
	case Slash:
		c.emitByte(vm.OpDivide)
	}

	return nil
}

func (c *Compiler) expression(_ bool) error {
	if err := c.parsePrecedence(PrecAssignment); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) declaration(_ bool) error {
	if c.match(Semicolon, NewLine) {
		return nil
	}

	return c.statement()
}

func (c *Compiler) grouping(_ bool) error {
	if err := c.expression(false); err != nil {
		return err
	}
	return c.consume(RightParenthesis)
}

func (c *Compiler) literal(_ bool) error {
	switch c.previous.TokenType {
	case False:
		{
			v := vm.Value{ValueType: vm.Bool, B: false}
			c.emitConstant(v)
		}
	case Nil:
		{
			v := vm.Value{ValueType: vm.Nil}
			c.emitConstant(v)
		}
	case True:
		{
			v := vm.Value{ValueType: vm.Bool, B: true}
			c.emitConstant(v)
		}
	}
	return nil
}

func (c *Compiler) number(_ bool) error {
	n, err := strconv.ParseFloat(c.previous.Lexeme, 64)
	if err != nil {
		return err
	}

	v := vm.Value{ValueType: vm.Number, N: n}
	c.emitConstant(v)
	return nil
}

func (c *Compiler) print() error {
	if err := c.expression(false); err != nil {
		return err
	}

	if !c.check(RightBrace) {
		if err := c.consume(Semicolon, NewLine); err != nil {
			return err
		}
	}

	c.emitByte(vm.OpPrint)
	return nil
}

func (c *Compiler) statement() error {
	if c.match(Print) {
		return c.print()
	}

	if c.match(LeftBrace) {
		return c.block()
	}

	if c.match(Var, Let) {
		return c.variable()
	}

	if c.match(If) {
		return c.ifStatement()
	}

	if err := c.expression(false); err != nil {
		return err
	}

	if err := c.consume(Semicolon, NewLine); err != nil {
		return err
	}
	c.emitByte(vm.OpPop)
	return nil
}

// block statements parser
func (c *Compiler) block() error {
	c.scope.begin()

	for !c.check(RightBrace) && !c.check(Eof) {
		if err := c.declaration(false); err != nil {
			return err
		}
	}
	if err := c.consume(RightBrace); err != nil {
		return err
	}

	c.scope.end()

	// clean variable out of scope
	for !c.scope.isEmpty() && c.scope.locals[c.scope.count-1].depth > c.scope.depth {
		c.emitByte(vm.OpPop)
		c.scope.count--
	}
	return nil
}

func (c *Compiler) string(_ bool) error {
	v := vm.Value{ValueType: vm.Object, Ptr: c.previous.Lexeme}
	c.emitConstant(v)
	return nil
}

func (c *Compiler) unary(_ bool) error {
	tt := c.previous.TokenType

	if err := c.parsePrecedence(PrecUnary); err != nil {
		return err
	}

	switch tt {
	case Not:
		c.emitByte(vm.OpNot)
	case Minus:
		c.emitByte(vm.OpMinus)
	}
	return nil
}

// variable declaration parser
func (c *Compiler) variable() error {
	modifiable := c.previous.TokenType == Var

	if err := c.consume(Identifier); err != nil {
		return err
	}
	identifier := c.previous.Lexeme

	// declare variable
	if c.scope.depth > 0 {
		if err := c.addLocal(identifier, modifiable); err != nil {
			return err
		}
	}

	// define variable
	if c.match(Equal) {
		if err := c.expression(false); err != nil {
			return err
		}
	} else {
		v := vm.Value{ValueType: vm.Nil}
		c.emitConstant(v)
	}
	if err := c.consume(Semicolon, NewLine); err != nil {
		return err
	}

	if c.scope.depth > 0 {
		// local scope
		return nil
	}

	// define variable as global
	c.emitByte(vm.OpDefineGlobal)
	c.WriteIdentifier(identifier, c.current.Line)
	c.scope.addGlobal(identifier, modifiable)
	return nil
}

// identifier parser
func (c *Compiler) identifier(assignable bool) error {
	identifier := c.previous.Lexeme

	var getOp, setOp vm.OpCode

	isLocal, addr, modifiable := c.resolveVar(identifier)

	if isLocal {
		getOp = vm.OpGetLocal
		setOp = vm.OpSetLocal
	} else {
		getOp = vm.OpGetGlobal
		setOp = vm.OpSetGlobal
	}

	if c.match(Equal) && assignable {
		if !modifiable {
			return fmt.Errorf("compile error, cannot assign expression to constant '%s' [line %d]", identifier, c.current.Line)
		}

		// assignment
		if err := c.expression(false); err != nil {
			return err
		}
		c.emitByte(setOp)
	} else {
		// reading identifier
		c.emitByte(getOp)
	}

	if isLocal {
		c.Write(vm.OpCode(addr), c.current.Line)
	} else {
		c.WriteIdentifier(identifier, c.current.Line)
	}

	return nil
}

func (c *Compiler) ifStatement() error {
	// condition
	if err := c.expression(false); err != nil {
		return err
	}

	thenJump := c.emitJump(vm.OpJumpIfFalse)
	c.emitByte(vm.OpPop)

	// then
	if err := c.statement(); err != nil {
		return err
	}

	elseJump := c.emitJump(vm.OpJump)
	c.emitByte(vm.OpPop)
	c.applyPatch(thenJump)

	if c.match(Else) {
		if err := c.statement(); err != nil {
			return err
		}
	}
	c.applyPatch(elseJump)

	return nil
}

func (c *Compiler) emitByte(byte vm.OpCode) {
	c.Write(byte, c.previous.Line)
}

func (c *Compiler) emitBytes(bytes ...vm.OpCode) {
	for _, b := range bytes {
		c.emitByte(b)
	}
}

func (c *Compiler) emitJump(op vm.OpCode) int {
	c.emitBytes(op, vm.OpCode(0))
	return len(c.Code) - 1
}

func (c *Compiler) applyPatch(patch int) {
	offset := len(c.Code) - patch
	c.Code[patch] = vm.OpCode(offset)
}

func (c *Compiler) emitConstant(v vm.Value) {
	c.WriteConstant(v, c.current.Line)
}
