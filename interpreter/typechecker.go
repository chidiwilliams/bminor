package main

import (
	"fmt"
)

type typeError struct {
	message string
	line    int
}

func (e typeError) Error() string {
	return fmt.Sprintf("Error on line %d: %s", e.line+1, e.message)
}

var (
	typeInteger = newAtomicType[Integer]("integer")
	typeBoolean = newAtomicType[Boolean]("boolean")
	typeChar    = newAtomicType[Char]("char")
	typeString  = newAtomicType[String]("string")
	typeAny     = anyType{}
	typeVoid    = voidType{}
)

func NewTypeChecker(statements []Stmt) TypeChecker {
	return TypeChecker{
		statements: statements,
		env:        NewEnvironment[Type](nil),
	}
}

type TypeChecker struct {
	statements                 []Stmt
	env                        *Environment[Type]
	enclosingEnv               *Environment[Type]
	currentFunctionReturnType  Type
	hasCurrentFunctionReturned bool
}

func (c *TypeChecker) Check() error {
	for _, stmt := range c.statements {
		err := c.checkStmt(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *TypeChecker) checkStmt(stmt Stmt) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if recovered, ok := r.(typeError); ok {
				err = recovered
			} else if recovered, ok := r.(runtimeError); ok {
				err = recovered
			} else {
				panic(r)
			}
		}
	}()

	switch stmt := stmt.(type) {
	case *VarStmt:
		declaredType := c.getType(stmt.Type)
		if stmt.Initializer != nil {
			resolvedType := c.resolveExpr(stmt.Initializer)
			c.expectExpr(stmt.Initializer, resolvedType, declaredType)
		} else {
			stmt.Initializer = &LiteralExpr{Value: declaredType.ZeroValue()}
		}
		c.env.Define(stmt.Name.Lexeme, declaredType)
	case *PrintStmt:
		for _, expression := range stmt.Expressions {
			c.resolveExpr(expression)
		}
	case *ExprStmt:
		c.resolveExpr(stmt.Expr)
	case *IfStmt:
		conditionType := c.resolveExpr(stmt.Condition)
		c.expectExpr(stmt.Condition, conditionType, typeBoolean)
		if err := c.checkStmt(stmt.Body); err != nil {
			return err
		}
	case *BlockStmt:
		c.beginScope()
		for _, innerStmt := range stmt.Statements {
			if err := c.checkStmt(innerStmt); err != nil {
				return err
			}
		}
		c.endScope()
	case *FunctionStmt:
		if c.currentFunctionReturnType != nil {
			panic(c.error(stmt, "function definitions may not be nested"))
		}

		paramTypes := make([]ParamType, len(stmt.TypeExpr.Params))
		for i, param := range stmt.TypeExpr.Params {
			paramTypes[i] = ParamType{Name: param.Name, Type: c.getType(param.Type)}
		}

		returnType := c.getType(stmt.TypeExpr.ReturnType)

		c.env.Define(stmt.Name.Lexeme, newFunctionType(paramTypes, returnType))

		c.beginScope()
		c.hasCurrentFunctionReturned = false

		for _, param := range stmt.TypeExpr.Params {
			c.env.Define(param.Name.Lexeme, c.getType(param.Type))
		}

		c.currentFunctionReturnType = returnType

		for _, innerStmt := range stmt.Body {
			if err := c.checkStmt(innerStmt); err != nil {
				return err
			}
		}

		if c.currentFunctionReturnType != typeVoid && !c.hasCurrentFunctionReturned {
			return c.error(stmt, "expected function to return value of type '%s'", c.currentFunctionReturnType)
		}

		c.endScope()
	case *ReturnStmt:
		if stmt.Value != nil && c.currentFunctionReturnType == typeVoid {
			return c.error(stmt.Value, "not expecting any return value")
		}
		valueType := c.resolveExpr(stmt.Value)
		c.expectExpr(stmt.Value, valueType, c.currentFunctionReturnType)
		c.hasCurrentFunctionReturned = true
	case *ForStmt:
		conditionType := c.resolveExpr(stmt.Condition)
		c.expectExpr(stmt.Condition, conditionType, typeBoolean)
		if err := c.checkStmt(stmt.Body); err != nil {
			return err
		}
	default:
		return c.error(stmt, fmt.Sprintf("unexpected statement type: %v", stmt))
	}
	return nil
}

// resolveExpr infers the type of an expression from its (possibly nested)
// literal values and operations while checking for type errors within the expression
func (c *TypeChecker) resolveExpr(expr Expr) Type {
	switch expr := expr.(type) {
	case *LiteralExpr:
		switch expr.Value.(type) {
		case Integer:
			return typeInteger
		case Boolean:
			return typeBoolean
		case Char:
			return typeChar
		case String:
			return typeString
		}
	case *VariableExpr:
		return c.env.Get(expr.Name.Lexeme)
	case *ArrayExpr:
		firstElementType := c.resolveExpr(expr.Elements[0])
		for _, element := range expr.Elements[1:] {
			elementType := c.resolveExpr(element)
			c.expectExpr(element, elementType, firstElementType)
		}
		return newArrayType(firstElementType, len(expr.Elements), false)
	case *MapExpr:
		if len(expr.Pairs) == 0 {
			return newMapType(typeAny, typeAny)
		}

		firstPair := expr.Pairs[0]
		firstKeyType := c.resolveExpr(firstPair.Key)
		firstValueType := c.resolveExpr(firstPair.Value)

		for _, pair := range expr.Pairs[1:] {
			keyType := c.resolveExpr(pair.Key)
			c.expectExpr(pair.Key, keyType, firstKeyType)
			valueType := c.resolveExpr(pair.Value)
			c.expectExpr(pair.Value, valueType, firstValueType)
		}

		return newMapType(firstKeyType, firstValueType)
	case *GetExpr:
		return c.resolveLookup(expr.Object, expr.Name)
	case *SetExpr:
		expectedValueType := c.resolveLookup(expr.Object, expr.Name)
		valueType := c.resolveExpr(expr.Value)
		c.expectExpr(expr.Value, valueType, expectedValueType)
		return expectedValueType
	case *BinaryExpr:
		left := c.resolveExpr(expr.Left)
		right := c.resolveExpr(expr.Right)
		if left.Equals(typeBoolean) || right.Equals(typeBoolean) {
			panic(c.error(expr, "cannot perform operation on boolean type"))
		}

		if !left.Equals(right) {
			panic(c.error(expr.Left, "'%s' and '%s' are of different types", expr.Left, expr.Right))
		}

		switch expr.Operator.TokenType {
		case TokenLess, TokenGreater, TokenLessEqual,
			TokenGreaterEqual, TokenEqualEqual, TokenBangEqual:
			return typeBoolean
		default:
			if expr.Operator.TokenType == TokenPlus && left.Equals(typeString) {
				panic(c.error(expr.Left, "cannot perform operation on string type"))
			}

			return typeInteger
		}
	case *PrefixExpr:
		rightType := c.resolveExpr(expr.Right)
		switch expr.Operator.TokenType {
		case TokenMinus:
			c.expectExpr(expr.Right, rightType, typeInteger)
			return typeInteger
		case TokenBang:
			c.expectExpr(expr.Right, rightType, typeBoolean)
			return typeBoolean
		}
	case *PostfixExpr:
		leftType := c.resolveExpr(expr.Left)
		c.expectExpr(expr.Left, leftType, typeInteger)
		return typeInteger
	case *LogicalExpr:
		leftType := c.resolveExpr(expr.Left)
		c.expectExpr(expr.Left, leftType, typeBoolean)
		rightType := c.resolveExpr(expr.Right)
		c.expectExpr(expr.Right, rightType, typeBoolean)
		return typeBoolean
	case *AssignExpr:
		valueType := c.resolveExpr(expr.Value)
		nameType := c.env.Get(expr.Name.Lexeme)
		c.expectExpr(expr.Value, valueType, nameType)
		return valueType
	case *CallExpr:
		calleeType, ok := c.resolveExpr(expr.Callee).(functionType)
		if !ok {
			panic(c.error(expr.Callee, "%s is not a function", expr.Callee))
		}

		for i, arg := range expr.Arguments {
			argType := c.resolveExpr(arg)
			expectedType := calleeType.paramTypes[i].Type
			c.expectExpr(arg, argType, expectedType)
		}

		return calleeType.returnType
	}
	panic(c.error(expr, "unexpected expression type: %s", expr))
}

func (c *TypeChecker) resolveLookup(object, name Expr) Type {
	objectType := c.resolveExpr(object)
	switch objectType := objectType.(type) {
	case arrayType:
		nameType := c.resolveExpr(name)
		c.expectExpr(name, nameType, typeInteger)
		return objectType.elementType
	case mapType:
		nameType := c.resolveExpr(name)
		c.expectExpr(name, nameType, objectType.keyType)
		return objectType.valueType
	default:
		panic(c.error(object, "can only index maps and arrays"))
	}
}

func (c *TypeChecker) getType(typeExpr TypeExpr) Type {
	switch typeExpr := typeExpr.(type) {
	case AtomicTypeExpr:
		switch typeExpr.Type {
		case AtomicTypeBoolean:
			return typeBoolean
		case AtomicTypeString:
			return typeString
		case AtomicTypeChar:
			return typeChar
		case AtomicTypeInteger:
			return typeInteger
		}
	case VoidTypeExpr:
		return voidType{}
	case ArrayTypeExpr:
		elementType := c.getType(typeExpr.ElementType)
		return newArrayType(elementType, typeExpr.Length, typeExpr.Reference)
	case MapTypeExpr:
		keyType := c.getType(typeExpr.KeyType)
		valueType := c.getType(typeExpr.ValueType)
		return newMapType(keyType, valueType)
	}
	panic(c.error(typeExpr, "unexpected type: %s", typeExpr))
}

func (c *TypeChecker) equals(resolvedType Type, declaredType Type) bool {
	return resolvedType.Equals(declaredType)
}

// expectExpr panics with a typeError if the given expression type is not equal to the expected type
func (c *TypeChecker) expectExpr(expr Expr, exprType, expectedType Type) {
	if !exprType.Equals(expectedType) {
		panic(c.error(expr, "expected '%s' to be of type '%s', but got '%s'", expr, expectedType, exprType))
	}
}

// error returns a typeError at the given line
func (c *TypeChecker) error(node Node, format string, any ...any) error {
	return typeError{line: node.StartLine(), message: fmt.Sprintf(format, any...)}
}

func (c *TypeChecker) beginScope() {
	c.enclosingEnv = c.env
	c.env = NewEnvironment[Type](c.enclosingEnv)
}

func (c *TypeChecker) endScope() {
	c.env = c.enclosingEnv
	c.enclosingEnv = c.env.enclosing
}
