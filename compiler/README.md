# compiler package

*tl;dr* The _compiler_ package contains the source code used to compile Maki source code to _pcode_.

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

## Scope

The `scope` struct implements _lexical scope_ or _static scope_. Considering the following code:
```
{
    // block A
    var a
    {
        // block B
        var b
        {
            // block C
            var c
        }
    }
}

```
The variable `b` can be accessed by _block B_ and _block C_. From _block B_ the variable `a` can be 
accessed as well but not the variable `c`. In general, every inner level can access its outer levels variables. 
Local variables are stored using the `local` struct, that keep track of the variable name
(identifier), if it is modifiable and its _depth_, that is the _block level_. For example in the code above
variable `a`, `b` and `c` are being stored in `locals` field of `scope` struct like this:
```
[
    {
        identifier: "a",
        modifiable: true,
        depth:      0,       
    },
    {
        identifier: "b",
        modifiable: true,
        depth:      1,       
    },
    {
        identifier: "c",
        modifiable: true,
        depth:      2,       
    }
]
```
Now suppose to want resolve variable `b` from _block C_. The `resolveVar()` method iterate over `locals`
array from last element until it found the first variable that match the identifier passed to. If no variable
is found then it is supposed to be a global variable. For simplicity global variable are implemented using
the `map` builtin data structure. Regarding adding a new variable it is enough add a new `local` struct to
the end of `locals` array. But before to add it is checked if a variable with the same identifier is declared
in the same scope, that is the same depth. 

## Compiler
_TBD_
