# Maki

maki-lang or just Maki is a programming language built for **fun** completely written in Go. It is not **absolutely** 
production ready and ~~probably~~ will never be.

```
fun hello(name) {
  print "Meowww, " + name +  "!"
}
hello("World") // print Meowww, World!
```
Maki ~~was written~~ is being written keeping in mind that is should be easy to read. So the implementation choices
are very affected by KISS principle.

## Run
###### Build
```
go build
```
###### REPL
```
./maki
```
###### File
```
./maki program.maki
```

## To Do

- Array (builtin data structure)
- Closure
- Class
- Module

#### Nice to have

- `switch` statement
- `break` statement
- `continue` statement

## Credits

This project owe much indeed to @munificient's book: Crafting Interpreters. In fact Maki is deeply inspired by Lox.
