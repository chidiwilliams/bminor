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

	Run(string(file), os.Stderr, nil)
}

func Run(source string, stdErr io.Writer, stdOut io.Writer) {
	scanner := NewScanner(source)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		_, _ = stdErr.Write([]byte(fmt.Sprintf("Error: %s\n", err.Error())))
		return
	}

	parser := NewParser(tokens, stdErr)
	statements, err := parser.Parse()
	if err != nil {
		_, _ = stdErr.Write([]byte(fmt.Sprintf("Error: %s\n", err.Error())))
		return
	}

	typeChecker := NewTypeChecker(statements)
	err = typeChecker.Check()
	if err != nil {
		_, _ = stdErr.Write([]byte(fmt.Sprintf("Error: %s\n", err.Error())))
		return
	}

	interpreter := NewInterpreter(&typeChecker, stdOut)
	err = interpreter.Interpret(statements)
	if err != nil {
		_, _ = stdErr.Write([]byte(fmt.Sprintf("Error: %s\n", err.Error())))
	}
}
