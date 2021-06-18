package vm

import (
	"fmt"
)

const (
	FrameSize  = 128
	GlobalSize = 1024
	StackSize  = 4096
)

type Frame struct {
	*Function
	rp     int
	locals int
}

func newFrame(fun *Function) Frame {
	return Frame{
		Function: fun,
	}
}

type VM struct {
	ip      int // instruction pointer
	sp      int // stack pointer
	fp      int // frame pointer
	stack   [StackSize]Value
	frames  [FrameSize]Frame
	globals map[string]Value
}

func NewVM() *VM {
	vm := &VM{
		globals: make(map[string]Value, GlobalSize),
	}

	vm.defineNative("println", Println{})
	vm.defineNative("clock", Clock{})

	return vm
}

func (vm *VM) initPointers() {
	vm.ip = 0
	vm.sp = 0
	vm.fp = 0
}

func (vm *VM) defineNative(name string, native Native) {
	vm.globals[name] = Value{
		ValueType: Object,
		Ptr:       native,
	}
}

func (vm *VM) top() *Value {
	return &vm.stack[vm.sp-1]
}

func (vm *VM) push(v Value) {
	vm.stack[vm.sp] = v
	vm.sp++
}

func (vm *VM) pop() Value {
	vm.sp--
	return vm.stack[vm.sp]
}

func (vm *VM) peekFrame() Frame {
	return vm.frames[vm.fp-1]
}

func (vm *VM) pushFrame(frame Frame) {
	vm.frames[vm.fp] = frame
	vm.fp++
}

func (vm *VM) popFrame() {
	vm.sp = vm.peekFrame().locals
	vm.ip = vm.peekFrame().rp
	vm.fp--
}

func (vm *VM) peekByte() OpCode {
	return vm.peekFrame().Code[vm.ip]
}

func (vm *VM) readByte() OpCode {
	vm.ip++
	return vm.peekFrame().Code[vm.ip-1]
}

func (vm *VM) Run(fun *Function) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("maki :: panic at ip %04d\n\n", vm.ip)
			panic(err)
		}
	}()

	vm.initPointers()
	vm.pushFrame(newFrame(fun))

	for {
		switch op := vm.readByte(); op {
		case OpAdd:
			{
				if err := vm.add(); err != nil {
					return err
				}
			}
		case OpArray:
			{
				vm.array()
			}
		case OpAssert:
			{
				if err := vm.assert(); err != nil {
					return err
				}
			}
		case OpCall:
			{
				if err := vm.call(); err != nil {
					return err
				}
			}
		case OpValue:
			{
				vm.constant()
			}
		case OpDefineGlobal:
			{
				vm.defineGlobal()
			}
		case OpDivide:
			{
				vm.divide()
			}
		case OpEqualEqual, OpNotEqual:
			{
				if err := vm.equality(op); err != nil {
					return err
				}
			}
		case OpGetGlobal:
			{
				if err := vm.getGlobal(false); err != nil {
					return err
				}
			}
		case OpGetGlobalIndex:
			{
				if err := vm.getGlobal(true); err != nil {
					return err
				}
			}
		case OpGetLocal:
			{
				if err := vm.getLocal(false); err != nil {
					return err
				}
			}
		case OpGetLocalIndex:
			{
				if err := vm.getLocal(true); err != nil {
					return err
				}
			}
		case OpGreater, OpGreaterEqual, OpLess, OpLessEqual:
			{
				if err := vm.comparison(op); err != nil {
					return err
				}
			}
		case OpJump:
			{
				vm.jump()
			}
		case OpJumpIfFalse:
			{
				vm.jumpIfFalse()
			}
		case OpLoop:
			{
				vm.loop()
			}
		case OpNil:
			{
				vm.nil()
			}
		case OpNot:
			{
				vm.not()
			}
		case OpMinus:
			{
				if err := vm.minus(); err != nil {
					return err
				}
			}
		case OpMultiply:
			{
				vm.multiply()
			}
		case OpPop:
			{
				_ = vm.pop()
			}
		case OpPrint:
			{
				fmt.Printf("%+v\n", vm.pop())
			}
		case OpReturn:
			{
				vm.callReturn()
			}
		case OpSetGlobal:
			{
				if err := vm.setGlobal(false); err != nil {
					return err
				}
			}
		case OpSetGlobalIndex:
			{
				if err := vm.setGlobal(true); err != nil {
					return err
				}
			}
		case OpSetLocal:
			{
				if err := vm.setLocal(false); err != nil {
					return err
				}
			}
		case OpSetLocalIndex:
			{
				if err := vm.setLocal(true); err != nil {
					return err
				}
			}
		case OpSubtract:
			{
				vm.subtract()
			}
		case OpTerminate:
			{
				return nil
			}
		default:
			{
				return fmt.Errorf("maki :: runtime error, op code %04d not yet implemented", vm.peekByte())
			}
		}
	}
}

func (vm *VM) add() error {
	rhs, lhs := vm.getOperands()
	err := fmt.Errorf("maki :: runtime error, invalid binary operands [line %d]", vm.getCurrentLine())

	if lhs.ValueType == Number && rhs.ValueType == Number {
		v := Value{ValueType: Number, Float: lhs.Float + rhs.Float}
		vm.push(v)
		return nil
	}

	if lhs.ValueType == Object && rhs.ValueType == Object {
		ls, ok := lhs.Ptr.(string)
		if !ok {
			return err
		}

		rs, ok := rhs.Ptr.(string)
		if !ok {
			return err
		}

		v := Value{ValueType: Object, Ptr: ls + rs}
		vm.push(v)
		return nil
	}

	return err
}

func (vm *VM) call() error {
	count := int(vm.readByte())
	args := make([]Value, count)
	for i := count - 1; i >= 0; i-- {
		args[i] = vm.pop()
	}
	v := vm.pop()

	if v.ValueType != Object {
		return fmt.Errorf("maki :: runtime error, %s is not callable [line %d]", v.String(), vm.getCurrentLine())
	}

	if v.ValueType == Object {
		switch f := v.Ptr.(type) {
		case *Function:
			{
				vm.pushFrame(Frame{
					Function: f,
					rp:       vm.ip,
					locals:   vm.sp,
				})
				for i := 0; i < count; i++ {
					vm.push(args[i])
				}
				vm.ip = 0
			}
		case Native:
			{
				v := f.Function(args)
				vm.push(v)
			}
		default:
			{
				return fmt.Errorf("maki :: runtime error, %s is not callable [line %d]", v.String(), vm.getCurrentLine())
			}
		}
	}

	return nil
}

func (vm *VM) callReturn() {
	returnValue := vm.pop()
	vm.popFrame()
	vm.push(returnValue)
}

func (vm *VM) constant() {
	address := int(vm.readByte())
	vm.push(vm.peekFrame().Constants.At(address))
}

func (vm *VM) defineGlobal() {
	address := int(vm.readByte())
	identifier, _ := vm.peekFrame().Constants.At(address).Ptr.(string)
	vm.globals[identifier] = vm.pop()
}

func (vm *VM) divide() {
	rhs, lhs := vm.getOperands()
	v := Value{ValueType: Number, Float: lhs.Float / rhs.Float}
	vm.push(v)
}

func (vm *VM) equality(op OpCode) error {
	rhs, lhs := vm.getOperands()

	err := fmt.Errorf("maki :: runtime error, invalid binary operands [line %d]", vm.getCurrentLine())

	v := Value{ValueType: Bool, Boolean: true}

	if lhs.ValueType != rhs.ValueType {
		if lhs.ValueType == Nil || rhs.ValueType == Nil {
			if op == OpEqualEqual {
				v.Boolean = false
			}
			vm.push(v)
			return nil
		} else {
			return err
		}
	}

	switch lhs.ValueType {
	case Bool:
		v.Boolean = lhs.Boolean == rhs.Boolean
	case Number:
		v.Boolean = lhs.Float == rhs.Float
	case Object:
		{
			ls, ok := lhs.Ptr.(string)
			if !ok {
				return err
			}

			rs, ok := rhs.Ptr.(string)
			if !ok {
				return err
			}

			v.Boolean = ls == rs
		}
	}

	if op == OpNotEqual {
		v.Boolean = !v.Boolean
	}

	vm.push(v)
	return nil
}

func (vm *VM) getIndex() (int, error) {
	v := vm.pop()
	if v.ValueType != Number {
		return 0, fmt.Errorf("maki :: runtime error, invalid type for indexing: %v", v.ValueType)
	}
	return int(v.Float), nil
}

func (vm *VM) getGlobal(isIndexed bool) error {
	address := int(vm.readByte())
	identifier, _ := vm.peekFrame().Constants.At(address).Ptr.(string)

	variable, ok := vm.globals[identifier]
	if !ok {
		return fmt.Errorf("maki :: runtime error, variable '%s' not defined [line %d]", identifier, vm.getCurrentLine())
	}

	if variable.ValueType == Array {
		if isIndexed {
			index, err := vm.getIndex()
			if err != nil {
				return err
			}
			array, ok := variable.Ptr.([]Value)
			if !ok {
				return fmt.Errorf("maki :: runtime error, variable '%s' is not a valid array", identifier)
			}
			if index >= len(array) {
				return fmt.Errorf("maki :: runtume error, index out of range with length %d: %s[%d]", len(array), identifier, index)
			}
			variable = array[index]
		} else {
			variable = Value{ValueType: Reference, Ptr: variable}
		}
	}

	vm.push(variable)
	return nil
}

func (vm *VM) getLocal(isIndexed bool) error {
	address := int(vm.readByte())
	variable := vm.stack[vm.peekFrame().locals+address]

	if variable.ValueType == Array {
		if isIndexed {
			index, err := vm.getIndex()
			if err != nil {
				return err
			}
			array, ok := variable.Ptr.([]Value)
			if !ok {
				return fmt.Errorf("maki :: runtime error, invalid array")
			}
			if index >= len(array) {
				return fmt.Errorf("maki :: runtume error, index out of range with length %d: [%d]", len(array), index)
			}
			variable = array[index]
		} else {
			variable = Value{ValueType: Reference, Ptr: variable}
		}
	}

	vm.push(variable)
	return nil
}

func (vm *VM) jump() {
	jump := int(vm.readByte())
	// n.b.
	// -1 because readByte() advance the ip counter
	// -1 because after jump the ip counter will be incremented
	vm.ip += jump - 2
}

func (vm *VM) jumpIfFalse() {
	if !vm.top().BoolValue() {
		vm.jump()
	} else {
		_ = vm.readByte() // skip jump address instruction
	}
}

func (vm *VM) loop() {
	jump := int(vm.readByte())
	vm.ip -= jump + 2
}

func (vm *VM) array() {
	count := int(vm.readByte())
	values := make([]Value, count)
	for i := count - 1; i >= 0; i-- {
		values[i] = vm.pop()
	}
	vm.push(Value{ValueType: Array, Ptr: values})
}

func (vm *VM) assert() error {
	value := vm.pop()
	if !value.Boolean {
		return fmt.Errorf("maki :: assertion failed at line %d", vm.getCurrentLine())
	}
	return nil
}

func (vm *VM) setGlobal(isIndexed bool) error {
	value := vm.pop()
	address := int(vm.readByte())
	identifier, _ := vm.peekFrame().Constants.At(address).Ptr.(string)

	if _, ok := vm.globals[identifier]; !ok {
		return fmt.Errorf("maki :: runtime error, variable '%s' not defined [line %d]", identifier, vm.getCurrentLine())
	}
	variable := vm.globals[identifier]

	if isIndexed && variable.ValueType == Array {
		index, err := vm.getIndex()
		if err != nil {
			return err
		}
		array, ok := variable.Ptr.([]Value)
		if !ok {
			return fmt.Errorf("maki :: runtime error, variable '%s' is not a valid array")
		}
		if index >= len(array) {
			return fmt.Errorf("maki :: runtume error, index out of range with length %d: %s[%d]", len(array), identifier, index)
		}
		array[index] = value
	} else {
		vm.globals[identifier] = value
	}
	vm.push(value)
	return nil
}

func (vm *VM) setLocal(isIndexed bool) error {
	value := vm.pop()
	address := int(vm.readByte())
	variable := vm.stack[address]

	if isIndexed && value.ValueType == Array {
		index, err := vm.getIndex()
		if err != nil {
			return err
		}
		array, ok := variable.Ptr.([]Value)
		if !ok {
			return fmt.Errorf("maki :: runtime error, invalid array")
		}
		if index >= len(array) {
			return fmt.Errorf("maki :: runtume error, index out of range with length %d: [%d]", len(array), index)
		}
		array[index] = value
	} else {
		vm.stack[address] = value
	}
	vm.push(value)
	return nil
}

func (vm *VM) comparison(op OpCode) error {
	rhs, lhs := vm.getOperands()

	if lhs.ValueType != Number || rhs.ValueType != Number {
		return fmt.Errorf("maki :: runtime error, invalid binary operands [line %d]", vm.getCurrentLine())
	}

	v := Value{ValueType: Bool}
	switch op {
	case OpGreater:
		v.Boolean = lhs.Float > rhs.Float
	case OpGreaterEqual:
		v.Boolean = lhs.Float >= rhs.Float
	case OpLess:
		v.Boolean = lhs.Float < rhs.Float
	case OpLessEqual:
		v.Boolean = lhs.Float <= rhs.Float
	}

	vm.push(v)
	return nil
}

func (vm *VM) minus() error {
	v := vm.top()

	if v.ValueType != Number {
		return fmt.Errorf("maki :: runtime error, operand must be a number [line %d]", vm.getCurrentLine())
	}

	v.Float = -v.Float
	return nil
}

func (vm *VM) multiply() {
	rhs, lhs := vm.getOperands()
	v := Value{ValueType: Number, Float: lhs.Float * rhs.Float}
	vm.push(v)
}

func (vm *VM) nil() {
	v := Value{ValueType: Nil}
	vm.push(v)
}

func (vm *VM) not() {
	lhs := vm.pop()

	switch lhs.ValueType {
	case Bool:
		{
			v := Value{ValueType: Bool, Boolean: !lhs.Boolean}
			vm.push(v)
		}
	default:
		{
			v := Value{ValueType: Bool, Boolean: true}
			vm.push(v)
		}
	}
}

func (vm *VM) subtract() {
	rhs, lhs := vm.getOperands()
	v := Value{
		ValueType: Number,
		Float:     lhs.Float - rhs.Float,
	}
	vm.push(v)
}

func (vm *VM) getOperands() (Value, Value) {
	return vm.pop(), vm.pop()
}

func (vm *VM) getCurrentLine() int {
	line, err := vm.peekFrame().Lines.At(vm.ip)
	if err != nil {
		panic(err.Error())
	}
	return line
}
