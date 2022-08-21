package main

func puts(interpreter *Interpreter, args []Value) Value {
	_, _ = interpreter.stdOut.Write([]byte(args[0].String() + "\n"))
	return IntegerValue(0)
}
