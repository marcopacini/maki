package compiler

import (
	"fmt"
	"maki/vm"
	"strconv"
)

type Compiler struct {
	*vm.Function
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
	rules := map[TokenType]rule{
		And:             {prefix: nil, infix: (*Compiler).and, precedence: PrecAnd},
		Equal:           {prefix: nil, infix: nil, precedence: PrecNone},
		EqualEqual:      {prefix: nil, infix: (*Compiler).binary, precedence: PrecEquality},
		False:           {prefix: (*Compiler).literal, infix: nil, precedence: PrecNone},
		Greater:         {prefix: nil, infix: (*Compiler).binary, precedence: PrecComparison},
		GreaterEqual:    {prefix: nil, infix: (*Compiler).binary, precedence: PrecComparison},
		Identifier:      {prefix: (*Compiler).identifier, infix: nil, precedence: PrecNone},
		LeftParenthesis: {prefix: (*Compiler).grouping, infix: (*Compiler).call, precedence: PrecCall},
		LeftSquare:      {prefix: (*Compiler).array, infix: nil, precedence: PrecNone},
		Less:            {prefix: nil, infix: (*Compiler).binary, precedence: PrecComparison},
		LessEqual:       {prefix: nil, infix: (*Compiler).binary, precedence: PrecComparison},
		Minus:           {prefix: (*Compiler).unary, infix: (*Compiler).binary, precedence: PrecTerm},
		Nil:             {prefix: (*Compiler).literal, infix: nil, precedence: PrecNone},
		Not:             {prefix: (*Compiler).unary, infix: nil, precedence: PrecNone},
		NotEqual:        {prefix: nil, infix: (*Compiler).binary, precedence: PrecEquality},
		Number:          {prefix: (*Compiler).number, infix: nil, precedence: PrecNone},
		Plus:            {prefix: nil, infix: (*Compiler).binary, precedence: PrecTerm},
		Or:              {prefix: nil, infix: (*Compiler).or, precedence: PrecOr},
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

func (c *Compiler) Compile(source string) (*vm.Function, error) {
	c.Function = vm.NewFunction("MAIN")
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

	c.emitByte(vm.OpTerminate)

	return c.Function, nil
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

func (c *Compiler) or(_ bool) error {
	elseJump := c.emitJump(vm.OpJumpIfFalse)
	thenJump := c.emitJump(vm.OpJump)

	c.applyPatch(elseJump)
	c.emitByte(vm.OpPop)

	if err := c.parsePrecedence(PrecOr); err != nil {
		return err
	}
	c.applyPatch(thenJump)

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
	case NotEqual:
		c.emitByte(vm.OpNotEqual)
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
	default:
		return fmt.Errorf("compile error, invalid binary operator '%s' [line %d]", tt, c.previous.Line)
	}

	return nil
}

func (c *Compiler) call(_ bool) error {
	count, err := c.arguments()
	if err != nil {
		return err
	}
	c.emitBytes(vm.OpCall, vm.OpCode(count))
	return nil
}

func (c *Compiler) arguments() (int, error) {
	var t TokenType
	switch c.previous.TokenType {
	case LeftParenthesis:
		t = RightParenthesis
	case LeftSquare:
		t = RightSquare
	default:
		return 0, fmt.Errorf("compiler error, unpexcted parenthesis: %s", c.current.Lexeme)
	}

	count := 0
	for c.current.TokenType != t {
		count++
		if err := c.expression(false); err != nil {
			return count, err
		}
		c.trim(Comma)
	}
	if err := c.consume(t); err != nil {
		return count, err
	}
	return count, nil
}

func (c *Compiler) expression(_ bool) error {
	if err := c.parsePrecedence(PrecAssignment); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) declaration(_ bool) error {
	if c.match(Semicolon, NewLine, Eof) {
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
			v := vm.Value{ValueType: vm.Bool, Boolean: false}
			c.emitConstant(v)
		}
	case Nil:
		{
			v := vm.Value{ValueType: vm.Nil}
			c.emitConstant(v)
		}
	case True:
		{
			v := vm.Value{ValueType: vm.Bool, Boolean: true}
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

	v := vm.Value{ValueType: vm.Number, Float: n}
	c.emitConstant(v)
	return nil
}

func (c *Compiler) array(_ bool) error {
	count, err := c.arguments()
	if err != nil {
		return err
	}
	c.emitBytes(vm.OpArray, vm.OpCode(count))
	return nil
}

func (c *Compiler) assert() error {
	if err := c.expression(false); err != nil {
		return err
	}
	c.emitByte(vm.OpAssert)
	return nil
}

func (c *Compiler) print() error {
	if err := c.expression(false); err != nil {
		return err
	}
	c.emitByte(vm.OpPrint)
	return nil
}

func (c *Compiler) statement() error {
	switch c.current.TokenType {
	case Assert:
		{
			_ = c.advance()
			if err := c.assert(); err != nil {
				return err
			}
		}
	case For:
		{
			_ = c.advance()
			if err := c.forStatement(); err != nil {
				return err
			}
		}
	case Fun:
		{
			_ = c.advance()
			if err := c.funStatement(); err != nil {
				return err
			}
		}
	case If:
		{
			_ = c.advance()
			if err := c.ifStatement(); err != nil {
				return err
			}
		}

	case LeftBrace:
		{
			_ = c.advance()
			if err := c.block(); err != nil {
				return err
			}
		}
	case Print:
		{
			_ = c.advance()
			if err := c.print(); err != nil {
				return err
			}
		}
	case Return:
		{
			_ = c.advance()
			if err := c.returnStatement(); err != nil {
				return err
			}
		}
	case Var, Let:
		{
			_ = c.advance()
			if err := c.listVariables(); err != nil {
				return err
			}
		}
	case While:
		{
			_ = c.advance()
			if err := c.whileStatement(); err != nil {
				return err
			}
		}
	default:
		{
			if err := c.expression(false); err != nil {
				return err
			}

			if c.current.TokenType != RightBrace {
				if err := c.consume(Semicolon, NewLine, Eof); err != nil {
					return err
				}
			}

			c.emitByte(vm.OpPop)
			return nil
		}
	}

	c.trim(Semicolon, NewLine)
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

	c.scope.end(func() { c.emitByte(vm.OpPop) })

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

// variable declarations parser
func (c *Compiler) listVariables() error {
	modifiable := c.previous.TokenType == Var

	// multiple declarations
	if c.match(LeftBrace) {
		c.trim(NewLine)
		for c.current.TokenType != RightBrace {
			if err := c.variable(false, modifiable); err != nil {
				return err
			}
			c.trim(NewLine)
		}
		return c.consume(RightBrace)
	}

	// single declaration
	return c.variable(false, modifiable)
}

func (c *Compiler) variable(isArgument bool, modifiable bool) error {
	identifier := c.current

	if err := c.consume(Identifier); err != nil {
		return err
	}

	// declare variable
	if c.scope.depth > 0 {
		if err := c.addLocal(*identifier, modifiable); err != nil {
			return err
		}
	}

	// define variable
	if c.match(Equal) {
		if err := c.expression(false); err != nil {
			return err
		}
	} else {
		if !isArgument {
			v := vm.Value{ValueType: vm.Nil}
			c.emitConstant(v)
		}
	}

	if c.scope.depth > 0 {
		// local scope
		return nil
	}

	// define variable as global
	if _, ok := c.scope.globals[identifier.Lexeme]; ok {
		return fmt.Errorf("compile error, variable '%s' is already defined in global scope [line %d]", identifier.Lexeme, identifier.Line)
	}

	c.emitByte(vm.OpDefineGlobal)
	c.WriteIdentifier(identifier.Lexeme, identifier.Line)
	c.scope.addGlobal(identifier.Lexeme, modifiable)

	return nil
}

func (c *Compiler) indexing() error {
	if c.match(Number) {
		n, err := strconv.ParseInt(c.previous.Lexeme, 10, 64)
		if err != nil {
			return err
		}
		if n < 0 {
			return fmt.Errorf("compile error, index cannot be negative: %d", n)
		}
		if err := c.consume(RightSquare); err != nil {
			return err
		}
		v := vm.Value{ValueType: vm.Number, Float: float64(n)}
		c.emitConstant(v)
		return nil
	}
	if err := c.expression(false); err != nil {
		return err
	}
	return c.consume(RightSquare)
}

// identifier parser
func (c *Compiler) identifier(assignable bool) error {
	identifier := c.previous.Lexeme
	isLocal, addr, modifiable := c.resolveVar(identifier)
	isIndexed := c.match(LeftSquare)

	var getOp, setOp vm.OpCode
	if isLocal {
		if isIndexed {
			getOp = vm.OpGetLocalIndex
			setOp = vm.OpSetLocalIndex
		} else {
			getOp = vm.OpGetLocal
			setOp = vm.OpSetLocal
		}
	} else {
		if isIndexed {
			getOp = vm.OpGetGlobalIndex
			setOp = vm.OpSetGlobalIndex
		} else {
			getOp = vm.OpGetGlobal
			setOp = vm.OpSetGlobal
		}
	}

	if isIndexed {
		if err := c.indexing(); err != nil {
			return err
		}
	}
	if c.match(Equal) && assignable {
		if !modifiable {
			return fmt.Errorf("compile error, cannot assign expression to constant '%s' [line %d]", identifier, c.current.Line)
		}

		// assignment
		if err := c.expression(false); err != nil {
			return err
		}

		c.emitBytes(setOp)
	} else {
		// reading identifier
		c.emitBytes(getOp)
	}
	if isLocal {
		c.Write(vm.OpCode(addr), c.current.Line)
	} else {
		c.WriteIdentifier(identifier, c.current.Line)
	}

	return nil
}

func (c *Compiler) funStatement() error {
	fun := c.Function
	t := c.current

	c.Function = vm.NewFunction(t.Lexeme)
	c.begin()

	if err := c.consume(Identifier); err != nil {
		return err
	}

	if err := c.consume(LeftParenthesis); err != nil {
		return err
	}

	for c.current.TokenType != RightParenthesis {
		c.Arity++
		c.trim(Var)

		if err := c.variable(true, true); err != nil {
			return err
		}

		c.trim(Comma)
	}

	if err := c.consume(RightParenthesis); err != nil {
		return err
	}

	if err := c.consume(LeftBrace); err != nil {
		return err
	}

	if err := c.block(); err != nil {
		return err
	}
	c.end(func() { c.emitByte(vm.OpPop) })
	c.WriteConstant(vm.Value{ValueType: vm.Nil}, c.current.Line)
	c.emitByte(vm.OpReturn)

	v := vm.Value{
		ValueType: vm.Object,
		Ptr:       c.Function,
	}

	c.Function = fun

	c.emitConstant(v)
	c.emitByte(vm.OpDefineGlobal)
	c.WriteIdentifier(t.Lexeme, t.Line)
	c.scope.addGlobal(t.Lexeme, false)

	return nil
}

func (c *Compiler) returnStatement() error {
	if err := c.expression(false); err != nil {
		return err
	}

	c.emitByte(vm.OpReturn)
	return nil
}

func (c *Compiler) ifStatement() error {
	// condition
	if err := c.expression(false); err != nil {
		return err
	}

	thenJump := c.emitJump(vm.OpJumpIfFalse)
	c.emitByte(vm.OpPop) // pop condition in then branch

	// then
	if err := c.consume(LeftBrace); err != nil {
		return err
	}
	if err := c.block(); err != nil {
		return err
	}

	elseJump := c.emitJump(vm.OpJump)
	c.applyPatch(thenJump)
	c.emitByte(vm.OpPop) // pop condition in else branch

	if c.match(Else) {
		if err := c.consume(LeftBrace); err != nil {
			return err
		}
		if err := c.block(); err != nil {
			return err
		}
	}
	c.applyPatch(elseJump)

	return nil
}

func (c *Compiler) whileStatement() error {
	// condition
	loopStart := c.getCurrentAddress()
	if err := c.expression(false); err != nil {
		return err
	}
	exitJump := c.emitJump(vm.OpJumpIfFalse)

	// body
	c.emitByte(vm.OpPop)
	if err := c.statement(); err != nil {
		return err
	}
	c.emitLoop(loopStart)
	c.applyPatch(exitJump)

	return nil
}

func (c *Compiler) forStatement() error {
	c.scope.begin()

	// initializer
	if err := c.declaration(false); err != nil {
		return err
	}

	// condition
	conditionLoop := c.getCurrentAddress()
	if err := c.expression(false); err != nil {
		return err
	}
	if err := c.consume(Semicolon); err != nil {
		return err
	}
	exitJump := c.emitJump(vm.OpJumpIfFalse)
	c.emitByte(vm.OpPop) // pop condition value
	bodyJump := c.emitJump(vm.OpJump)

	// increment
	incrementLoop := c.getCurrentAddress()
	if err := c.expression(false); err != nil {
		return err
	}
	c.emitByte(vm.OpPop)
	c.emitLoop(conditionLoop)

	// body
	c.applyPatch(bodyJump)
	if err := c.statement(); err != nil {
		return err
	}
	c.emitLoop(incrementLoop)

	c.applyPatch(exitJump)
	c.emitByte(vm.OpPop) // pop condition value
	c.scope.end(func() { c.emitByte(vm.OpPop) })

	return nil
}

func (c Compiler) getCurrentAddress() int {
	return len(c.Code)
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
	return c.getCurrentAddress() - 1
}

func (c *Compiler) applyPatch(patch int) {
	offset := len(c.Code) - patch + 1
	c.Code[patch] = vm.OpCode(offset)
}

func (c *Compiler) emitLoop(startLoop int) {
	c.emitByte(vm.OpLoop)
	offset := c.getCurrentAddress() - startLoop - 1
	c.emitByte(vm.OpCode(offset))
}

func (c *Compiler) emitConstant(v vm.Value) {
	c.WriteConstant(v, c.current.Line)
}
