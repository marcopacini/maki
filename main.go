package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"maki/compiler"
	"maki/vm"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		if err := repl(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	} else if len(os.Args) == 2 {
		if err := runFile(os.Args[1]); err != nil {
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

	for ;; {
		fmt.Print("> ")
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if err := interpret(line); err != nil {
			fmt.Println(err)
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

	return interpret(string(b))
}

func interpret(source string) error {
	pcode, err := compiler.NewCompiler().Compile(source)
	if err != nil {
		return err
	}

	vm.NewVM(pcode).Run()
	return nil
}