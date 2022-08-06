package main

type TokenType int

const (
	TokenEof TokenType = iota
	TokenSlash
	TokenNumber
	TokenIdentifier
	TokenSemicolon
	TokenColon
	TokenEqual
	TokenComma
	TokenChar
	TokenElse
	TokenFalse
	TokenFor
	TokenFunction
	TokenIf
	TokenPrint
	TokenReturn
	TokenString
	TokenTrue
	TokenVoid
	TokenWhile
	TokenTypeIdentifier
)

type Token struct {
	TokenType TokenType
	Line      int
	Lexeme    string
	Start     int
	Literal   interface{}
}
