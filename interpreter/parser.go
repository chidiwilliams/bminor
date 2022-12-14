package main

import (
	"fmt"
	"io"
	"strconv"
)

type parseError struct {
	message string
	token   Token
}

func (p parseError) Error() string {
	if p.token.Lexeme != "" {
		return fmt.Sprintf("Error at '%s' on line %d: %s", p.token.Lexeme, p.token.Line+1, p.message)
	}
	return p.message
}

func NewParser(tokens []Token, stdErr io.Writer) *Parser {
	return &Parser{tokens: tokens, stdErr: stdErr}
}

/**
Parser grammar:
	program     => declaration* EOF
	declaration => printStmt | varDecl | exprStmt | blockStmt
	printStmt   => "print" expression ( "," expression )*
	varDecl     => IDENTIFIER ":" typeExpr ( "=" expression )? ";"
	typeExpr    => "integer" | "boolean" | "char" | "string"
	               | "array" "[" NUMBER "]" typeExpr
	exprStmt    => expression ";"
  expression  => assignment
	assignment  => subscript "=" assignment | or
	or          => and ( "||" and )*
	and         => comparison ( "&&" comparison )*
	comparison  => term ( ( "<" | ">" | ">=" | "<=" | "==" | "!=" ) term )*
	term        => factor ( ( "+" | "-" ) factor )*
	factor      => exponent ( ( "*" | "/" | "%" ) exponent )*
	exponent    => unary ( "^" exponent )*
	unary       => ( "-" | "!" ) unary | postfix
	postfix     => subscript ( "++" | "--" )?
	subscript   => primary ( "[" expression "]" )*
	primary     => IDENTIFIER | NUMBER | CHAR | STRING | "false" | "true"
	               | MAP
	MAP         => "{" expression ( "," expression )* "}"
*/

type Parser struct {
	tokens  []Token
	current int
	stdErr  io.Writer
	hadErr  bool
}

func (p *Parser) Parse() (statements []Stmt, err error) {
	defer func() {
		if recoveredErr := recover(); recoveredErr != nil {
			var ok bool
			if err, ok = recoveredErr.(parseError); !ok {
				panic(recoveredErr)
			}
		}
	}()

	for !p.isAtEnd() {
		stmt := p.declaration()
		statements = append(statements, stmt)
	}
	return statements, nil
}

func (p *Parser) declaration() Stmt {
	if p.match(TokenPrint) {
		return p.printStmt()
	}
	if p.match(TokenIf) {
		return p.ifStmt()
	}
	if p.match(TokenLeftBrace) {
		return p.blockStmt()
	}
	if p.match(TokenReturn) {
		return p.returnStmt()
	}
	if p.match(TokenFor) {
		return p.forStmt()
	}

	return p.exprStmt()
}

func (p *Parser) printStmt() Stmt {
	printToken := p.previous()
	var expressions []Expr
	expressions = append(expressions, p.expression())
	for p.match(TokenComma) {
		expressions = append(expressions, p.expression())
	}
	p.consume(TokenSemicolon, "expect semicolon after print statement")
	return &PrintStmt{Expressions: expressions, BeginLine: printToken.Line}
}

func (p *Parser) ifStmt() Stmt {
	ifToken := p.previous()
	p.consume(TokenLeftParen, "expect '(' after 'if'")
	condition := p.expression()
	p.consume(TokenRightParen, "expect ')' after if condition")
	consequent := p.declaration()

	var alternative Stmt
	if p.match(TokenElse) {
		alternative = p.declaration()
	}

	return &IfStmt{Condition: condition, Consequent: consequent, Alternative: alternative, BeginLine: ifToken.Line}
}

func (p *Parser) blockStmt() *BlockStmt {
	brace := p.previous()
	statements := make([]Stmt, 0)

	for !p.match(TokenRightBrace) && !p.isAtEnd() {
		stmt := p.declaration()
		statements = append(statements, stmt)
	}

	return &BlockStmt{Statements: statements, BeginLine: brace.Line}
}

func (p *Parser) exprStmt() Stmt {
	expr := p.expression()

	if expr, ok := expr.(*VariableExpr); ok && p.match(TokenColon) {
		return p.varDecl(expr.Name)
	}

	p.consume(TokenSemicolon, "expect semicolon after variable declaration")
	return &ExprStmt{Expr: expr}
}

func (p *Parser) returnStmt() Stmt {
	returnToken := p.previous()

	var value Expr
	if p.match(TokenColon) {
		value = nil
	} else {
		value = p.expression()
	}

	p.consume(TokenSemicolon, "expect semicolon after return statement")
	return &ReturnStmt{Value: value, BeginLine: returnToken.Line}
}

func (p *Parser) forStmt() Stmt {
	forToken := p.previous()
	p.consume(TokenLeftParen, "expect '(' after 'for'")

	var initializer Stmt
	if !p.match(TokenSemicolon) {
		initializer = p.exprStmt()
	}

	var condition Stmt
	if !p.match(TokenSemicolon) {
		condition = p.expression()
		p.consume(TokenSemicolon, "expect semicolon after condition clause")
	}

	var increment Expr
	if !p.match(TokenRightParen) {
		increment = p.expression()
		p.consume(TokenRightParen, "expect ')' after for clauses")
	}

	body := p.declaration()

	if increment != nil {
		body = &BlockStmt{Statements: []Stmt{body, &ExprStmt{Expr: increment}}}
	}

	if condition == nil {
		condition = &LiteralExpr{Value: BooleanValue(true)}
	}

	body = &WhileStmt{Condition: condition, Body: body}

	if initializer != nil {
		body = &BlockStmt{Statements: []Stmt{initializer, body}, BeginLine: forToken.Line}
	}

	return body
}

func (p *Parser) varDecl(name Token) Stmt {
	typeExpr := p.typeExpr(false)

	var initializer Expr
	isFnPrototype := false
	if p.match(TokenEqual) {
		if functionTypeExpr, ok := typeExpr.(FunctionTypeExpr); ok {
			return p.function(name, functionTypeExpr)
		} else {
			if p.match(TokenLeftBrace) {
				initializer = p.mapOrArrayExpr()
			} else {
				initializer = p.expression()
			}
		}
	} else {
		if _, ok := typeExpr.(FunctionTypeExpr); ok {
			isFnPrototype = true
		}
	}

	p.consume(TokenSemicolon, "expect semicolon after variable declaration")
	return &VarStmt{Name: name, Initializer: initializer, Type: typeExpr, IsFnPrototype: isFnPrototype, BeginLine: name.Line}
}

func (p *Parser) mapOrArrayExpr() Expr {
	brace := p.previous()
	if p.match(TokenRightBrace) {
		return &MapExpr{BeginLine: brace.Line}
	}

	firstKeyOrElem := p.expression()

	if p.match(TokenColon) {
		pairs := make([]Pair, 0)
		firstValue := p.expression()
		pairs = append(pairs, Pair{Key: firstKeyOrElem, Value: firstValue})
		for p.match(TokenComma) {
			key := p.expression()
			p.match(TokenColon)
			value := p.expression()
			pairs = append(pairs, Pair{Key: key, Value: value})
		}
		p.consume(TokenRightBrace, "expect '}' after map literal")
		return &MapExpr{Pairs: pairs, BeginLine: brace.Line}
	}

	elements := make([]Expr, 0)
	elements = append(elements, firstKeyOrElem)
	for p.match(TokenComma) {
		elements = append(elements, p.expression())
	}
	p.consume(TokenRightBrace, "expect '}' after array literal")
	return &ArrayExpr{Elements: elements, BeginLine: brace.Line}
}

func (p *Parser) typeExpr(isArrayReference bool) TypeExpr {
	typeExpr := p.consume(TokenIdentifier, "expect type expression")
	switch typeExpr.Lexeme {
	case "integer":
		return AtomicTypeExpr{Type: AtomicTypeInteger, BeginLine: typeExpr.Line}
	case "string":
		return AtomicTypeExpr{Type: AtomicTypeString, BeginLine: typeExpr.Line}
	case "char":
		return AtomicTypeExpr{Type: AtomicTypeChar, BeginLine: typeExpr.Line}
	case "boolean":
		return AtomicTypeExpr{Type: AtomicTypeBoolean, BeginLine: typeExpr.Line}
	case "void":
		return VoidTypeExpr{BeginLine: typeExpr.Line}
	case "array":
		p.consume(TokenLeftSquareBracket, "expect '[' after array type expression")
		var length IntegerValue
		if !isArrayReference {
			var ok bool
			length, ok = p.consume(TokenNumber, "expect length of array type expression").Literal.(IntegerValue)
			if !ok || length == 0 {
				panic(p.error("expect length of array type expression to be a positive integer"))
			}
		}

		p.consume(TokenRightSquareBracket, "expect ']' after length of array type expression")
		elementType := p.typeExpr(isArrayReference)
		return ArrayTypeExpr{Length: int(length), ElementType: elementType, BeginLine: typeExpr.Line, Reference: isArrayReference}
	case "map":
		keyType := p.typeExpr(isArrayReference)
		valueType := p.typeExpr(isArrayReference)
		return MapTypeExpr{KeyType: keyType, ValueType: valueType}
	case "function":
		returnType := p.typeExpr(isArrayReference)
		p.consume(TokenLeftParen, "expect '(' before function parameters")

		params := make([]ParamTypeExpr, 0)

		if !p.match(TokenRightParen) {
			for {
				name := p.consume(TokenIdentifier, "expect function parameter")
				p.consume(TokenColon, "expect ':' after function parameter")
				paramType := p.typeExpr(true)
				params = append(params, ParamTypeExpr{Name: name, Type: paramType})

				if !p.match(TokenComma) {
					break
				}
			}

			p.consume(TokenRightParen, "expect ')' after function parameters")
		}

		return FunctionTypeExpr{Params: params, ReturnType: returnType}
	default:
		panic(parseError{message: fmt.Sprintf("unexpected type expression: %s", p.previous().Lexeme)})
	}
}

func (p *Parser) function(name Token, functionTypeExpr FunctionTypeExpr) Stmt {
	fnToken := p.previous()
	p.consume(TokenLeftBrace, "expect '{' before function body")
	body := p.blockStmt()
	return &FunctionStmt{Name: name, TypeExpr: functionTypeExpr, Body: body, BeginLine: fnToken.Line}
}

func (p *Parser) expression() Expr {
	return p.assignment()
}

func (p *Parser) assignment() Expr {
	expr := p.or()

	if p.match(TokenEqual) {
		value := p.assignment()

		if expr, ok := expr.(*VariableExpr); ok {
			return &AssignExpr{Name: expr.Name, Value: value}
		}
		if expr, ok := expr.(*GetExpr); ok {
			return &SetExpr{Object: expr.Object, Name: expr.Name, Value: value}
		}
		panic(parseError{message: fmt.Sprintf("invalid assignment target")})
	}

	return expr
}

func (p *Parser) or() Expr {
	expr := p.and()
	for p.match(TokenOr) {
		operator := p.previous()
		right := p.and()
		expr = &LogicalExpr{Left: expr, Right: right, Operator: operator}
	}
	return expr
}

func (p *Parser) and() Expr {
	expr := p.comparison()
	for p.match(TokenAnd) {
		operator := p.previous()
		right := p.comparison()
		expr = &LogicalExpr{Left: expr, Operator: operator, Right: right}
	}
	return expr
}

func (p *Parser) comparison() Expr {
	expr := p.term()
	for p.match(TokenLess, TokenLessEqual, TokenGreater,
		TokenGreaterEqual, TokenEqualEqual, TokenBangEqual) {
		operator := p.previous()
		right := p.term()
		expr = &BinaryExpr{Left: expr, Right: right, Operator: operator}
	}
	return expr
}

func (p *Parser) term() Expr {
	expr := p.factor()
	for p.match(TokenPlus, TokenMinus) {
		operator := p.previous()
		right := p.factor()
		expr = &BinaryExpr{Left: expr, Right: right, Operator: operator}
	}
	return expr
}

func (p *Parser) factor() Expr {
	expr := p.exponent()
	for p.match(TokenStar, TokenSlash, TokenPercent) {
		operator := p.previous()
		right := p.exponent()
		expr = &BinaryExpr{Left: expr, Right: right, Operator: operator}
	}
	return expr
}

func (p *Parser) exponent() Expr {
	expr := p.unary()
	if p.match(TokenCaret) {
		operator := p.previous()
		right := p.exponent()
		return &BinaryExpr{Left: expr, Right: right, Operator: operator}
	}
	return expr
}

func (p *Parser) unary() Expr {
	if p.match(TokenMinus, TokenBang) {
		operator := p.previous()
		right := p.unary()
		return &PrefixExpr{Operator: operator, Right: right}
	}
	return p.postfix()
}

func (p *Parser) postfix() Expr {
	expr := p.subscript()
	if p.match(TokenPlusPlus, TokenMinusMinus) {
		expr, ok := expr.(*VariableExpr)
		if !ok {
			panic(p.error("invalid left-hand side expression in postfix operation"))
		}
		operator := p.previous()
		return &PostfixExpr{Operator: operator, Left: expr}
	}
	return expr
}

func (p *Parser) subscript() Expr {
	expr := p.primary()
	for {
		if p.match(TokenLeftSquareBracket) {
			index := p.expression()
			expr = &GetExpr{Object: expr, Name: index}
			p.consume(TokenRightSquareBracket, "expect ']' after array subscript")
		} else if p.match(TokenLeftParen) {
			var paren Token
			args := make([]Expr, 0)
			if p.match(TokenRightParen) {
				paren = p.previous()
			} else {
				for {
					argument := p.expression()
					args = append(args, argument)
					if !p.match(TokenComma) {
						break
					}
				}
				paren = p.consume(TokenRightParen, "expect ')' after arguments")
			}
			return &CallExpr{Callee: expr, Paren: paren, Arguments: args}
		} else {
			break
		}
	}
	return expr
}

func (p *Parser) primary() Expr {
	switch {
	case p.match(TokenIdentifier):
		return &VariableExpr{Name: p.previous()}
	case p.match(TokenNumber, TokenChar, TokenString):
		return p.literalExpr(p.previous().Literal)
	case p.match(TokenFalse):
		return p.literalExpr(BooleanValue(false))
	case p.match(TokenTrue):
		return p.literalExpr(BooleanValue(true))
	case p.match(TokenLeftParen):
		expr := p.expression()
		p.consume(TokenRightParen, "expect ')' after expression")
		return expr
	}

	panic(p.error(fmt.Sprintf("expecting an expression, but found '%s'", p.peek().Lexeme)))
}

func (p *Parser) literalExpr(value Value) Expr {
	return &LiteralExpr{Value: value, BeginLine: p.previous().Line}
}

func (p *Parser) match(types ...TokenType) bool {
	for _, tokenType := range types {
		if p.check(tokenType) {
			p.advance()
			return true
		}
	}

	return false
}

func (p *Parser) check(tokenType TokenType) bool {
	if p.isAtEnd() {
		return false
	}

	return p.peek().TokenType == tokenType
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == TokenEof
}

func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

func (p *Parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p *Parser) consume(tokenType TokenType, message string) Token {
	if p.check(tokenType) {
		return p.advance()
	}
	panic(parseError{message: fmt.Sprintf("Error at %s on line %d: %s", strconv.Quote(p.peek().Lexeme), p.peek().Line+1, message)})
}

func (p *Parser) error(message string) error {
	return parseError{token: p.previous(), message: message}
}
