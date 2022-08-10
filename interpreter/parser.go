package main

import (
	"fmt"
	"io"
)

type parseError struct {
	message string
}

func (p parseError) Error() string {
	return p.message
}

func NewParser(tokens []Token, stdErr io.Writer) *Parser {
	return &Parser{tokens: tokens, stdErr: stdErr}
}

/**
Parser grammar:
	program     => declaration* EOF
	declaration => printStmt | varDecl | exprStmt
	printStmt   => "print" expression ( "," expression )*
	varDecl     => IDENTIFIER ":" typeExpr ( "=" expression )? ";"
	typeExpr    => "integer" | "boolean" | "char" | "string"
	               | "array" "[" NUMBER "]" typeExpr
	exprStmt    => expression ";"
  expression  => assignment
	assignment  => subscript "=" assignment
	subscript   => primary ("[" expression "]" )*
	primary     => IDENTIFIER | NUMBER | CHAR | STRING | "false" | "true"
	               | "{" expression ( "," expression )* "}"
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

	return p.exprStmt()
}

func (p *Parser) printStmt() Stmt {
	var expressions []Expr

	expressions = append(expressions, p.expression())

	for p.match(TokenComma) {
		expressions = append(expressions, p.expression())
	}

	p.consume(TokenSemicolon, "expect semicolon after print statement")
	return PrintStmt{Expressions: expressions}
}

func (p *Parser) exprStmt() Stmt {
	expr := p.expression()

	if expr, ok := expr.(VariableExpr); ok && p.match(TokenColon) {
		return p.varDecl(expr.Name)
	}

	p.consume(TokenSemicolon, "expect semicolon after variable declaration")
	return ExprStmt{Expr: expr}
}

func (p *Parser) varDecl(name Token) Stmt {
	typeExpr := p.typeExpr()

	var initializer Expr
	if p.match(TokenEqual) {
		initializer = p.expression()
	}

	p.consume(TokenSemicolon, "expect semicolon after variable declaration")
	return VarStmt{Name: name, Initializer: initializer, Type: typeExpr}
}

func (p *Parser) typeExpr() TypeExpr {
	typeExpr := p.consume(TokenTypeIdentifier, "expect type expression")
	switch typeExpr.Lexeme {
	case "integer":
		return AtomicTypeExpr{AtomicTypeInteger}
	case "string":
		return AtomicTypeExpr{AtomicTypeString}
	case "char":
		return AtomicTypeExpr{AtomicTypeChar}
	case "boolean":
		return AtomicTypeExpr{AtomicTypeBoolean}
	case "array":
		p.consume(TokenLeftSquareBracket, "expect '[' after array type expression")
		length, ok := p.consume(TokenNumber, "expect length of array type expression").Literal.(int)
		if !ok || length == 0 {
			panic(parseError{message: fmt.Sprintf("expect length of array type expression to be a positive integer")})
		}

		p.consume(TokenRightSquareBracket, "expect ']' after length of array type expression")
		elementType := p.typeExpr()
		return ArrayTypeExpr{Length: length, ElementType: elementType}
	case "map":
		keyType := p.typeExpr()
		valueType := p.typeExpr()
		return MapTypeExpr{KeyType: keyType, ValueType: valueType}
	default:
		panic(parseError{message: fmt.Sprintf("unexpected type expression: %s", p.previous().Lexeme)})
	}
}

func (p *Parser) expression() Expr {
	return p.assignment()
}

func (p *Parser) assignment() Expr {
	expr := p.subscript()

	if p.match(TokenEqual) {
		value := p.assignment()

		if expr, ok := expr.(VariableExpr); ok {
			return AssignExpr{Name: expr.Name, Value: value}
		}
		if expr, ok := expr.(GetExpr); ok {
			return SetExpr{Object: expr.Object, Name: expr.Name, Value: value}
		}
		panic(parseError{message: fmt.Sprintf("invalid assignment target")})
	}

	return expr
}

func (p *Parser) subscript() Expr {
	expr := p.primary()
	for p.match(TokenLeftSquareBracket) {
		index := p.expression()
		expr = GetExpr{Object: expr, Name: index}
		p.consume(TokenRightSquareBracket, "expect ']' after array subscript")
	}
	return expr
}

func (p *Parser) primary() Expr {
	switch {
	case p.match(TokenIdentifier):
		return VariableExpr{Name: p.previous()}
	case p.match(TokenNumber, TokenChar, TokenString):
		return p.literalExpr(p.previous().Literal)
	case p.match(TokenFalse):
		return p.literalExpr(false)
	case p.match(TokenTrue):
		return p.literalExpr(true)
	case p.match(TokenLeftBrace):
		if p.match(TokenRightBrace) {
			return MapExpr{}
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
			return MapExpr{Pairs: pairs}
		}

		elements := make([]Expr, 0)
		elements = append(elements, firstKeyOrElem)
		for p.match(TokenComma) {
			elements = append(elements, p.expression())
		}
		p.consume(TokenRightBrace, "expect '}' after array literal")
		return ArrayExpr{Elements: elements}
	}

	panic(parseError{message: fmt.Sprintf("expect expression, but got %s", p.peek().Lexeme)})
}

func (p *Parser) literalExpr(value interface{}) Expr {
	return LiteralExpr{Value: value}
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
	panic(parseError{message: fmt.Sprintf("Error at line %d: %s", p.peek().Line+1, message)})
}
