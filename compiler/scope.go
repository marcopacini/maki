package compiler

import "fmt"

const size = 256

type local struct {
	identifier string
	modifiable bool
	depth      int
}

type scope struct {
	globals map[string]bool // keep track of global constants
	locals  [size]local
	count   int
	depth   int
}

func newScope() *scope {
	return &scope{
		count: 0,
		depth: 0,
	}
}

func (s *scope) isEmpty() bool {
	return s.count == 0
}

func (s *scope) addLocal(identifier string, modifiable bool) error {
	if s.count >= size {
		return fmt.Errorf("compile error, too many variables in local scope")
	}

	// check redeclaration
	for i := s.count; i >= 0; i-- {
		local := s.locals[i]
		if local.depth != -1 && local.depth < s.depth {
			break
		}

		if local.identifier == identifier {
			return fmt.Errorf("compile error, variable '%s' is already defined in this scope", local.identifier)
		}
	}

	local := &s.locals[s.count]

	local.identifier = identifier
	local.modifiable = modifiable
	local.depth = s.depth

	s.count++
	return nil
}

func (s *scope) begin() {
	s.depth += 1
}

func (s *scope) end() {
	s.depth -= 1
}
