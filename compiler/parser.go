package compiler

import (
	"fmt"
	"strings"
)

type parser struct {
	*scanner
	current  *Token
	previous *Token
}

func newParser(source string) *parser {
	return &parser{
		scanner:  newScanner(source),
		current:  nil,
		previous: nil,
	}
}

func (p *parser) advance() error {
	p.previous = p.current

	var err error
	if p.current, err = p.scanToken(); err != nil {
		return err
	}

	return nil
}

func (p *parser) check(tt TokenType) bool {
	return p.current.TokenType == tt
}

func (p *parser) match(tts ...TokenType) bool {
	for _, tt := range tts {
		if p.current.TokenType == tt {
			_ = p.advance()
			return true
		}
	}
	return false
}

func (p *parser) consume(tts ...TokenType) error {
	for _, tt := range tts {
		if p.current.TokenType == tt {
			_ = p.advance()
			return nil
		}
	}

	beautify := func(tts []TokenType) string {
		var l []string
		for _, tt := range tts {
			l = append(l, string(tt))
		}

		return strings.Join(l, " or ")
	}

	return fmt.Errorf("compile error, expected '%s' [line %d]", beautify(tts), p.current.Line)
}
