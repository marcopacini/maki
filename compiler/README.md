# compiler package

*tl;dr* the _compiler_ package contains the source code used to compile Maki source code to _pcode_.

## `scanner.go`

The `scanner` struct contains the source code as an array of rune (aka character) and a few counter
for keeping track where we are. The `scanToken()` consume a token at time and is used by `Scan()` method
for scanning all the source code. The `Token` type is a struct that store the token type, the lexeme and
the line.

```go
    source := `
		fun hello(name) {
			print "Hello, " + name + "!"
		}`
	
	if tokens, err := newScanner(source).Scan(); err == nil {
		for _, t := range tokens {
			println(t)
		}
	}
```

The code print for each token the type, the lexeme and the line.

```
NEW_LINE 1
FUN fun 2
IDENTIFIER hello 2
LEFT_PARENTHESIS ( 2
IDENTIFIER name 2
RIGHT_PARENTHESIS ) 2
LEFT_BRACE { 2
NEW_LINE 2
PRINT print 3
STRING Hello,  3
PLUS + 3
IDENTIFIER name 3
PLUS + 3
STRING ! 3
NEW_LINE 3
RIGHT_BRACE } 4
EOF  4
```

## Compiler
_TBD_

## Scope
_TBD_
