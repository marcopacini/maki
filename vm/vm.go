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
				vm.add()
			}
		case OpConstant:
			{
				vm.constant()
			}
		case OpDivide:
			{
				vm.divide()
			}
		case OpFalse:
			{
				vm.false()
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
		case OpReturn:
			{
				fmt.Printf("%+v\n", vm.pop())
				return nil
			}
		case OpSubtract:
			{
				vm.subtract()
			}
		case OpTrue:
			{
				vm.true()
			}
		}
		vm.ip++
	}
}

func (vm *VM) add() {
	rhs := vm.pop()
	lhs := vm.pop()

	v := Value{
		ValueType: Number,
		N: lhs.N + rhs.N,
	}

	vm.push(v)
}

func (vm *VM) constant() {
	vm.ip++
	addr := int(vm.Code[vm.ip])
	vm.push(vm.Constants.At(addr))
}

func (vm *VM) divide() {
	rhs := vm.pop()
	lhs := vm.pop()

	v := Value{
		ValueType: Number,
		N: lhs.N / rhs.N,
	}

	vm.push(v)
}

func (vm *VM) false() {
	v := Value{
		ValueType: Bool,
		B: false,
	}

	vm.push(v)
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
	rhs := vm.pop()
	lhs := vm.pop()

	v := Value{
		ValueType: Number,
		N: lhs.N * rhs.N,
	}

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
	rhs := vm.pop()
	lhs := vm.pop()

	v := Value{
		ValueType: Number,
		N: lhs.N - rhs.N,
	}

	vm.push(v)
}

func (vm *VM) true() {
	v := Value{
		ValueType: Bool,
		B: true,
	}

	vm.push(v)
}

func (vm *VM) getCurrentLine() int {
	line, err := vm.Lines.At(vm.ip)
	if err != nil {
		panic(err)
	}

	return line
}