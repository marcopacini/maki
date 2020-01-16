package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"maki/compiler"
	"maki/vm"
	"os"
)

var debug bool

func main() {
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		if err := repl(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	} else if len(args) == 1 {
		if err := runFile(args[0]); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: maki [path]\n")
		os.Exit(64)
	}
}

func repl() error {
	r := bufio.NewReader(os.Stdin)

	replCompiler := compiler.NewCompiler()
	replVM := vm.NewVM()

	for {
		fmt.Print("> ")
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err := interpret(replCompiler, replVM, line); err != nil {
			fmt.Println("maki ::", err)
		}
	}

	return nil
}

func runFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			_, _ = fmt.Fprint(os.Stderr, err.Error())
		}
	}()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	return interpret(compiler.NewCompiler(), vm.NewVM(), string(b))
}

func interpret(c *compiler.Compiler, vm *vm.VM, source string) error {
	pcode, err := c.Compile(source)
	if err != nil {
		return err
	}

	if debug {
		fmt.Print(pcode)
	}

	return vm.Run(pcode)
}
