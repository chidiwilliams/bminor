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
	declaration => printStmt | varDecl
	printStmt   => "print" primary ( "," primary )*
	varDecl     => IDENTIFIER ":" typeDecl ( "=" primary )? ";"
	primary     => IDENTIFIER | NUMBER
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
	return p.varDecl()
}

func (p *Parser) printStmt() Stmt {
	var expressions []Expr

	expressions = append(expressions, p.primary())

	for p.match(TokenComma) {
		expressions = append(expressions, p.primary())
	}

	p.consume(TokenSemicolon, "expect semicolon after print statement")
	return PrintStmt{Expressions: expressions}
}

func (p *Parser) varDecl() Stmt {
	name := p.consume(TokenIdentifier, "expect name")

	p.consume(TokenColon, "expect colon after name")
	typeDecl := p.consume(TokenTypeIdentifier, "expect type declaration after name")

	var initializer Expr
	if p.match(TokenEqual) {
		initializer = p.primary()
	}

	p.consume(TokenSemicolon, "expect semicolon after variable declaration")

	return VarStmt{Name: name, Initializer: initializer, TypeDecl: typeDecl}
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
