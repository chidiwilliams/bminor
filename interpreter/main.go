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

	RunAndReportError(string(file), os.Stderr, nil)
}

func RunAndReportError(source string, stdErr io.Writer, stdOut io.Writer) {
	Run(source, stdErr, stdOut)
}

func Run(source string, stdErr io.Writer, stdOut io.Writer) error {
	scanner := NewScanner(source)
	tokens, err := scanner.ScanTokens()
	if err != nil {
		_, _ = stdErr.Write([]byte(fmt.Sprintf("Error: %s\n", err.Error())))
		return nil
	}

	parser := NewParser(tokens, stdErr)
	statements, err := parser.Parse()
	if err != nil {
		_, _ = stdErr.Write([]byte(fmt.Sprintf("Error: %s\n", err.Error())))
		return nil
	}

	fmt.Println(statements)
	return nil

	// typeChecker := newTypeChecker(statements)
	// err = typeChecker.Check()
	// if err != nil {
	// 	return err
	// }
	//
	// interpreter := newInterpreter(statements)
	// err = interpreter.Interpret()
	// return err
}
