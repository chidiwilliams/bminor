package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	filePathPtr := flag.String("file", "", "file")

	flag.Parse()

	file, err := os.ReadFile(*filePathPtr)
	if err != nil {
		log.Fatalf("input file not found")
	}

	Run(string(file), os.Stderr, os.Stdout)
}

func Run(source string, stdErr io.Writer, stdOut io.Writer) {
	scanner := NewScanner(source)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		_, _ = stdErr.Write([]byte(err.Error() + "\n"))
		return
	}

	parser := NewParser(tokens, stdErr)
	statements, err := parser.Parse()
	if err != nil {
		_, _ = stdErr.Write([]byte(err.Error() + "\n"))
		return
	}

	typeChecker := NewTypeChecker()
	err = typeChecker.Check(statements)
	if err != nil {
		_, _ = stdErr.Write([]byte(err.Error() + "\n"))
		return
	}

	interpreter := NewInterpreter(stdOut)
	err = interpreter.Interpret(statements)
	if err != nil {
		_, _ = stdErr.Write([]byte(fmt.Sprintf("Error: %s\n", err.Error())))
	}
}
