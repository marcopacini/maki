package vm

import (
	"fmt"
)

const (
	StackSize = 256
)

type VM struct {
	*PCode
	ip int
	stack [StackSize]Value
	sp int
}

func NewVM(c *PCode) *VM {
	return &VM{
		PCode: c,
		ip:    0,
		sp:    0,
	}
}

func (vm *VM) top() *Value {
	return &vm.stack[0]
}

func (vm *VM) push(v Value) {
	vm.stack[vm.sp] = v
	vm.sp++
}

func (vm *VM) pop() Value {
	vm.sp--
	return vm.stack[vm.sp]
}

func (vm *VM) Run() error {
	for ;; {
		switch vm.Code[vm.ip] {
		case OpAdd:
			{
				if err := vm.add(); err != nil {
					return err
				}
			}
		case OpConstant:
			{
				vm.constant()
			}
		case OpDivide:
			{
				vm.divide()
			}
		case OpEqualEqual:
			{
				if err := vm.equalEqual(); err != nil {
					return err
				}
			}
		case OpGreater:
			{
				if err := vm.greater(); err != nil {
					return err
				}
			}
		case OpGreaterEqual:
			{
				if err := vm.greaterEqual(); err != nil {
					return err
				}
			}
		case OpLess:
			{
				if err := vm.less(); err != nil {
					return err
				}
			}
		case OpLessEqual:
			{
				if err := vm.lessEqual(); err != nil {
					return err
				}
			}
		case OpNil:
			{
				vm.nil()
			}
		case OpNot:
			{
				vm.not()
			}
		case OpNotEqual:
			{
				if err := vm.notEqual(); err != nil {
					return err
				}
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
				return nil
			}
		case OpSubtract:
			{
				vm.subtract()
			}
		default:
			{
				return fmt.Errorf("maki: runtime error, op code %04d not yet implemented", vm.Code[vm.ip])
			}
		}
		vm.ip++
	}
}

func (vm *VM) add() error {
	rhs, lhs := vm.getOperands()
	err := fmt.Errorf("maki: runtime error, invalid binary operands [line %d]", vm.getCurrentLine())

	if lhs.ValueType == Number && rhs.ValueType == Number {
		v := Value{ ValueType: Number, N: lhs.N + rhs.N }
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

		v := Value{ ValueType: Object, Ptr: ls + rs }
		vm.push(v)
		return nil
	}

	return err
}

func (vm *VM) constant() {
	vm.ip++
	addr := int(vm.Code[vm.ip])
	vm.push(vm.Constants.At(addr))
}

func (vm *VM) divide() {
	rhs := vm.pop()
	lhs := vm.pop()

	v := Value{ ValueType: Number, N: lhs.N / rhs.N }
	vm.push(v)
}

func (vm *VM) equalEqual() error {
	rhs, lhs := vm.getOperands()

	err := fmt.Errorf("maki: runtime error, invalid binary operands [line %d]", vm.getCurrentLine())

	if lhs.ValueType != rhs.ValueType{
		return err
	}

	v := Value{ ValueType: Bool, B: true }

	switch lhs.ValueType {
	case Bool: v.B = lhs.B == rhs.B
	case Number: v.B = lhs.N == lhs.N
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

			v.B = ls == rs
		}
	}

	vm.push(v)
	return nil
}

func (vm *VM) notEqual() error {
	rhs, lhs := vm.getOperands()

	if lhs.ValueType != rhs.ValueType{
		return fmt.Errorf("maki: runtime error, invalid binary operands [line %d]", vm.getCurrentLine())
	}

	v := Value{ ValueType: Bool, B: false }

	switch lhs.ValueType {
	case Bool: v.B = lhs.B != rhs.B
	case Number: v.B = lhs.N != lhs.N
	}

	vm.push(v)
	return nil
}

func (vm *VM) greater() error {
	rhs, lhs := vm.getOperands()

	if lhs.ValueType != Number || rhs.ValueType != Number {
		return fmt.Errorf("maki: runtime error, invalid binary operands [line %d]", vm.getCurrentLine())
	}

	v := Value{ ValueType: Bool, B: lhs.N > rhs.N }
	vm.push(v)

	return nil
}

func (vm *VM) greaterEqual() error {
	rhs, lhs := vm.getOperands()

	if lhs.ValueType != Number || rhs.ValueType != Number {
		return fmt.Errorf("maki: runtime error, invalid binary operands [line %d]", vm.getCurrentLine())
	}

	v := Value{ ValueType: Bool, B: lhs.N >= rhs.N }
	vm.push(v)

	return nil
}

func (vm *VM) less() error {
	rhs, lhs := vm.getOperands()

	if lhs.ValueType != Number || rhs.ValueType != Number {
		return fmt.Errorf("maki: runtime error, invalid binary operands [line %d]", vm.getCurrentLine())
	}

	v := Value{ ValueType: Bool, B: lhs.N < rhs.N }
	vm.push(v)

	return nil
}

func (vm *VM) lessEqual() error {
	rhs, lhs := vm.getOperands()

	if lhs.ValueType != Number || rhs.ValueType != Number {
		return fmt.Errorf("maki: runtime error, invalid binary operands [line %d]", vm.getCurrentLine())
	}

	v := Value{ ValueType: Bool, B: lhs.N <= rhs.N }
	vm.push(v)

	return nil
}

func (vm *VM) minus() error {
	v := vm.top()

	if v.ValueType != Number {
		return fmt.Errorf("maki : runtime error, operand must be a number [line %d]", vm.getCurrentLine())
	}

	v.N = -v.N
	return nil
}

func (vm *VM) multiply() {
	rhs, lhs := vm.getOperands()

	v := Value{ ValueType: Number, N: lhs.N * rhs.N }
	vm.push(v)
}

func (vm *VM) nil() {
	v := Value{ ValueType: Nil }
	vm.push(v)
}

func (vm *VM) not()	{
	lhs := vm.pop()

	switch lhs.ValueType {
	case Bool:
		{
			v := Value{ ValueType: Bool, B: !lhs.B }
			vm.push(v)
		}
	default:
		{
			v := Value{ ValueType: Bool, B: true }
			vm.push(v)
		}
	}
}

func (vm *VM) subtract() {
	rhs, lhs := vm.getOperands()

	v := Value{
		ValueType: Number,
		N: lhs.N - rhs.N,
	}

	vm.push(v)
}

func (vm *VM) getOperands()	(Value, Value) {
	return vm.pop(), vm.pop()
}

func (vm *VM) getCurrentLine() int {
	line, err := vm.Lines.At(vm.ip)
	if err != nil {
		panic(err)
	}

	return line
}