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
		globals: make(map[string]bool),
		count:   0,
		depth:   0,
	}
}

func (s *scope) isEmpty() bool {
	return s.count == 0
}

func (s *scope) addGlobal(identifier string, modifiable bool) {
	s.globals[identifier] = modifiable
}

func (s *scope) addLocal(t Token, modifiable bool) error {
	if s.count >= size {
		return fmt.Errorf("compile error, too many variables in local scope")
	}

	// check redeclaration
	for i := s.count; i > 0; i-- {
		local := s.locals[i-1]
		if local.depth != -1 && local.depth < s.depth {
			break
		}

		if local.identifier == t.Lexeme {
			return fmt.Errorf("compile error, variable '%s' is already defined in this scope [line %d]", t.Lexeme, t.Line)
		}
	}

	local := &s.locals[s.count]

	local.identifier = t.Lexeme
	local.modifiable = modifiable
	local.depth = s.depth

	s.count++
	return nil
}

func (s scope) resolveVar(identifier string) (bool, int, bool) {
	for i := s.count - 1; i >= 0; i-- {
		local := s.locals[i]
		if local.identifier == identifier {
			return true, i, local.modifiable
		}
	}

	return false, -1, s.globals[identifier]
}

func (s *scope) begin() {
	s.depth += 1
}

func (s *scope) end(cancel func()) {
	s.depth -= 1

	// clean variable out of scope
	for !s.isEmpty() && s.locals[s.count-1].depth > s.depth {
		cancel()
		s.count--
	}
}
