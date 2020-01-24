package compiler

import "fmt"

type TokenType string

const (
	And              TokenType = "AND"
	Class                      = "CLASS"
	Comma                      = "COMMA"
	Dot                        = "DOT"
	Else                       = "ELSE"
	Eof                        = "EOF"
	Equal                      = "EQUAL"
	EqualEqual                 = "EQUAL_EQUAL"
	False                      = "FALSE"
	For                        = "FOR"
	Fun                        = "FUN"
	Greater                    = "GREATER"
	GreaterEqual               = "GREATER_EQUAL"
	Identifier                 = "IDENTIFIER"
	If                         = "IF"
	LeftBrace                  = "LEFT_BRACE"
	LeftParenthesis            = "LEFT_PARENTHESIS"
	LeftSquare                 = "LEFT_SQUARE"
	Less                       = "LESS"
	LessEqual                  = "LESS_EQUAL"
	Let                        = "LET"
	Minus                      = "MINUS"
	NewLine                    = "NEW_LINE"
	Nil                        = "NIL"
	Not                        = "NOT"
	NotEqual                   = "NOT_EQUAL"
	Number                     = "NUMBER"
	Or                         = "OR"
	Plus                       = "PLUS"
	Print                      = "PRINT"
	Return                     = "RETURN"
	RightBrace                 = "RIGHT_BRACE"
	RightParenthesis           = "RIGHT_PARENTHESIS"
	RightSquare                = "RIGHT_SQUARE"
	Semicolon                  = "SEMICOLON"
	Slash                      = "SLASH"
	Star                       = "STAR"
	String                     = "STRING"
	Super                      = "SUPER"
	This                       = "THIS"
	True                       = "TRUE"
	Var                        = "VAR"
	While                      = "WHILE"
)

var keywords = map[string]TokenType{
	"and":    And,
	"class":  Class,
	"else":   Else,
	"false":  False,
	"fun":    Fun,
	"for":    For,
	"if":     If,
	"nil":    Nil,
	"let":    Let,
	"or":     Or,
	"print":  Print,
	"return": Return,
	"super":  Super,
	"this":   This,
	"true":   True,
	"var":    Var,
	"while":  While,
}

type Token struct {
	TokenType
	Lexeme string
	Line   int
}

func (t Token) String() string {
	return fmt.Sprintf("%v %v %d", t.TokenType, t.Lexeme, t.Line)
}

type scanner struct {
	source  []rune
	start   int
	current int
	line    int
}

func newScanner(s string) *scanner {
	return &scanner{
		source:  []rune(s),
		start:   0,
		current: 0,
		line:    1,
	}
}

func (s *scanner) Scan() ([]Token, error) {
	ts := make([]Token, 0)

	for {
		t, err := s.scanToken()
		if err != nil {
			return nil, err
		}

		switch t.TokenType {
		case Eof:
			return append(ts, *t), nil
		default:
			ts = append(ts, *t)
		}
	}
}

func (s *scanner) scanToken() (*Token, error) {
	s.start = s.current
	s.trim()

	if s.isEnd() {
		eof := &Token{
			TokenType: Eof,
			Lexeme:    "",
			Line:      s.line,
		}
		return eof, nil
	}

	r := s.advance()
	switch r {
	case '\n':
		{
			t := s.makeToken(NewLine)
			s.line++
			return t, nil
		}
	case '(':
		{
			return s.makeToken(LeftParenthesis), nil
		}
	case ')':
		{
			return s.makeToken(RightParenthesis), nil
		}
	case '{':
		{
			return s.makeToken(LeftBrace), nil
		}
	case '}':
		{
			return s.makeToken(RightBrace), nil
		}
	case '[':
		{
			return s.makeToken(LeftSquare), nil
		}
	case ']':
		{
			return s.makeToken(RightSquare), nil
		}
	case ';':
		{
			return s.makeToken(Semicolon), nil
		}
	case ',':
		{
			return s.makeToken(Comma), nil
		}
	case '.':
		{
			return s.makeToken(Dot), nil
		}
	case '+':
		{
			return s.makeToken(Plus), nil
		}
	case '-':
		{
			return s.makeToken(Minus), nil
		}
	case '*':
		{
			return s.makeToken(Star), nil
		}
	// Multi-character lexeme
	case '!':
		{
			if s.isNext('=') {
				return s.makeToken(NotEqual), nil
			}
			return s.makeToken(Not), nil
		}
	case '=':
		{
			if s.isNext('=') {
				return s.makeToken(EqualEqual), nil
			}
			return s.makeToken(Equal), nil
		}
	case '>':
		{
			if s.isNext('=') {
				return s.makeToken(GreaterEqual), nil
			}
			return s.makeToken(Greater), nil
		}
	case '<':
		{
			if s.isNext('=') {
				return s.makeToken(LessEqual), nil
			}
			return s.makeToken(Less), nil
		}
	case '/':
		{
			return s.scanComment()
		}
	case '"':
		{
			return s.scanString()
		}
	default:
		{
			if isDigit(r) {
				return s.scanDigit()
			}
			if isLetter(r) {
				return s.scanIdentifier()
			}
			return nil, fmt.Errorf("scanner error, unknown character '%v' [line %d]", string(r), s.line)
		}
	}
}

func (s *scanner) scanComment() (*Token, error) {
	// Single-line comment
	if s.isNext('/') {
		for s.peek() != '\n' && !s.isEnd() {
			_ = s.advance()
		}
		return s.scanToken()
	}

	// Multi-line comment
	if s.isNext('*') {
		for {
			if s.isEnd() {
				return nil, fmt.Errorf("scanner error, comment not terminated [line %d]", s.line)
			}

			r := s.advance()
			switch r {
			case '\n':
				{
					s.line++
					break
				}
			case '*':
				{
					if s.peek() == '/' {
						_ = s.advance()
						return s.scanToken()
					}
				}
			}
		}
	}

	return s.makeToken(Slash), nil
}

func (s *scanner) scanString() (*Token, error) {
	for s.peek() != '"' && !s.isEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		_ = s.advance()
	}

	if s.isEnd() {
		return nil, fmt.Errorf("scanner error, unterminated string [line %d]", s.line)
	}
	_ = s.advance()

	t := &Token{
		TokenType: String,
		Lexeme:    string(s.source[s.start+1 : s.current-1]),
		Line:      s.line,
	}
	return t, nil
}

func (s *scanner) scanDigit() (*Token, error) {
	for isDigit(s.peek()) {
		_ = s.advance()
	}

	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance()

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	return s.makeToken(Number), nil
}

func (s *scanner) scanIdentifier() (*Token, error) {
	for isLetter(s.peek()) || isDigit(s.peek()) {
		s.advance()
	}

	if t, ok := keywords[string(s.source[s.start:s.current])]; ok {
		return s.makeToken(t), nil
	}

	return s.makeToken(Identifier), nil
}

func (s *scanner) isEnd() bool {
	return s.current >= len(s.source)
}

func (s *scanner) peek() rune {
	if s.isEnd() {
		return '\x00'
	}
	return s.source[s.current]
}

func (s *scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return '\x00'
	}
	return s.source[s.current-1]
}

func (s *scanner) advance() rune {
	s.current++
	return s.source[s.current-1]
}

func (s *scanner) isNext(r rune) bool {
	if s.isEnd() || s.source[s.current] != r {
		return false
	}

	s.current++
	return true
}

func (s *scanner) trim() {
	for !s.isEnd() {
		r := s.peek()
		switch r {
		case ' ', '\t', '\r':
			{
				s.start++
				s.current++
				break
			}
		default:
			return
		}
	}
}

func isDigit(r rune) bool {
	if r >= '0' && r <= '9' {
		return true
	}
	return false
}

func isLetter(r rune) bool {
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return true
	}

	return false
}

func (s *scanner) makeToken(tt TokenType) *Token {
	return &Token{
		TokenType: tt,
		Lexeme:    string(s.source[s.start:s.current]),
		Line:      s.line,
	}
}
