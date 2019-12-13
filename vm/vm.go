package vm

import "fmt"

type InterpretStatus uint8

const (
	OK InterpretStatus = iota
	RuntimeError
	CompileError
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

func (vm *VM) push(v Value) {
	vm.stack[vm.sp] = v
	vm.sp++
}

func (vm *VM) pop() Value {
	vm.sp--
	return vm.stack[vm.sp]
}

func (vm *VM) Run() InterpretStatus {
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
		case OpMinus:
			{
				vm.minus()
			}
		case OpMultiply:
			{
				vm.multiply()
			}
		case OpReturn:
			{
				fmt.Println(vm.pop())
				return OK
			}
		case OpSubtract:
			{
				vm.subtract()
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

func (vm *VM) minus() {
	rhs := vm.pop()

	v := Value{
		ValueType: Number,
		N: -rhs.N,
	}

	vm.push(v)
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

func (vm *VM) subtract() {
	rhs := vm.pop()
	lhs := vm.pop()

	v := Value{
		ValueType: Number,
		N: lhs.N - rhs.N,
	}

	vm.push(v)
}