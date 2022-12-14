package main

type TokenType int

const (
	TokenEof TokenType = iota
	TokenSlash
	TokenNumber
	TokenIdentifier
	TokenSemicolon
	TokenColon
	TokenComma
	TokenChar
	TokenElse
	TokenFalse
	TokenFor
	TokenIf
	TokenPrint
	TokenReturn
	TokenString
	TokenTrue
	TokenWhile

	TokenPlus
	TokenMinus
	TokenStar
	TokenPercent
	TokenCaret

	TokenPlusPlus
	TokenMinusMinus

	TokenOr
	TokenAnd

	TokenLeftParen
	TokenRightParen
	TokenLeftSquareBracket
	TokenRightSquareBracket
	TokenLeftBrace
	TokenRightBrace

	TokenEqual
	TokenEqualEqual
	TokenLess
	TokenLessEqual
	TokenGreater
	TokenGreaterEqual
	TokenBang
	TokenBangEqual
)

type Token struct {
	TokenType TokenType
	Line      int
	Lexeme    string
	Start     int
	Literal   Value
}
