package compiler

import "testing"

func listTokenTypes(tokens []Token) []TokenType {
	tt := make([]TokenType, len(tokens))
	for i, t := range tokens {
		tt[i] = t.TokenType
	}
	return tt
}

func TestScanner_Scan(t *testing.T) {
	tcs := []struct {
		name string
		in  string
		out []TokenType
	}{
		{
			name: "Parenthesis Tokens",
			in: "(){}[]",
			out: []TokenType{ LeftParenthesis, RightParenthesis, LeftBrace, RightBrace, LeftSquare, RightSquare, Eof },
		},
		{
			name: "Single Character Tokens",
			in: "+ - * / , ; ! > <",
			out: []TokenType{ Plus, Minus, Star, Slash, Comma, Semicolon, Not, Greater, Less, Eof },
		},
		{
			name: "Multi Characters Tokens",
			in: "== != >= <=",
			out: []TokenType{ EqualEqual, NotEqual, GreaterEqual, LessEqual, Eof },
		},
		{
			name: "Single-line Comment Token",
			in: "// This text have to be ignored",
			out: []TokenType{ Eof },
		},
		{
			name: "Multi-line comment token",
			in: "/* This text have to be ignored */",
			out: []TokenType{ Eof },
		},
		{
			name: "String Token",
			in: "\"This is a string!\"",
			out: []TokenType{ String, Eof },
		},
		{
			name: "Number Token",
			in: "1 12 12.3",
			out: []TokenType{ Number, Number, Number, Eof },
		},
		{
			name: "Logic Keywords Tokens",
			in: "and or true false",
			out: []TokenType{ And, Or, True, False, Eof },
		},
		{
			name: "Conditional Keywords Tokens",
			in: "if else for while",
			out: []TokenType{ If, Else, For, While, Eof },
		},
		{
			name: "Function Keywords Tokens",
			in: "fun return",
			out: []TokenType{ Fun, Return, Eof },
		},
		{
			name: "Type Keywords Tokens",
			in: "class var let nil",
			out: []TokenType{ Class, Var, Let, Nil, Eof },
		},
		{
			name: "Print and Identifier Tokens",
			in: "print x",
			out: []TokenType{ Print, Identifier, Eof },
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			scanner := newScanner(tc.in)
			tokens, _ := scanner.Scan()

			if len(tokens) != len(tc.out) {
				t.Errorf("got %d tokens, want %d: %v != %v", len(tokens), len(tc.out), tc.out, listTokenTypes(tokens))
			} else {
				for i := range tokens {
					if tokens[i].TokenType != tc.out[i] {
						t.Errorf("got %v, want %v: %v != %v", tokens[i].TokenType, tc.out[i], tc.out, listTokenTypes(tokens))
					}
				}
			}
		})
	}
}