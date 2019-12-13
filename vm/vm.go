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
	stack [StackSize]float64
	sp int
}

func NewVM(c *PCode) *VM {
	return &VM{
		PCode: c,
		ip:    0,
		sp:    0,
	}
}

func (vm *VM) push(val float64) {
	vm.stack[vm.sp] = val
	vm.sp++
}

func (vm *VM) pop() float64 {
	vm.sp--
	return vm.stack[vm.sp]
}

func (vm *VM) Run() InterpretStatus {
	for ;; {
		switch vm.Code[vm.ip] {
		case OpAdd:
			{
				vm.push(vm.pop() + vm.pop())
				break
			}
		case OpConstant:
			{
				vm.ip++
				addr := vm.Code[vm.ip]
				vm.push(vm.Constants.At(int(addr)))
				break
			}
		case OpDivide:
			{
				vm.push(1. / vm.pop() * vm.pop())
				break
			}
		case OpMinus:
			{
				vm.push(-vm.pop())
				break
			}
		case OpMultiply:
			{
				vm.push(vm.pop() * vm.pop())
			}
		case OpReturn:
			{
				fmt.Println(vm.pop())
				return OK
			}
		case OpSubtract:
			{
				vm.push(-vm.pop() + vm.pop())
				break
			}
		}
		vm.ip++
	}
}