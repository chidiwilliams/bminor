package main

import (
	"fmt"
	"strconv"
)

func NewScanner(source string) Scanner {
	return Scanner{source: source}
}

type Scanner struct {
	source  string
	start   int
	current int
	line    int
	tokens  []Token
}

func (s *Scanner) ScanTokens() ([]Token, error) {
	for !s.isAtEnd() {
		// we're at the beginning of the next lexeme
		s.start = s.current
		err := s.scanToken()
		if err != nil {
			return nil, err
		}
	}

	s.tokens = append(s.tokens, Token{TokenType: TokenEof, Line: s.line})
	return s.tokens, nil
}

func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) scanToken() error {
	char := s.advance()

	switch char {
	case '/':
		// Handle "//" C-style comments
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else if s.match('*') { // Handle "/* ... */" C++-style comments
			for s.peek() != '*' && s.peekNext() != '/' && !s.isAtEnd() {
				s.advance()
			}

			// Advance past the closing "*/"
			s.advance()
			s.advance()
		} else {
			s.addToken(TokenSlash)
		}

	// Ignore whitespace
	case ' ', '\t', '\r':

	case '\n':
		s.line++

	case ':':
		s.addToken(TokenColon)

	case ';':
		s.addToken(TokenSemicolon)

	case '=':
		s.addToken(TokenEqual)

	case ',':
		s.addToken(TokenComma)

	case '\'':
		s.advance()
		if s.isAtEnd() {
			return fmt.Errorf("unterminated char")
		}
		s.advance() // advance past the closing "'"
		value := s.source[s.current-1]
		s.addTokenWithLiteral(TokenChar, value)

	case '"':
		for s.peek() != '"' && !s.isAtEnd() {
			if s.peek() == '\n' {
				s.line++
			}
			s.advance()
		}
		if s.isAtEnd() {
			return fmt.Errorf("unterminated string")
		}
		s.advance() // advance past the closing '"'
		value := s.source[s.start+1 : s.current-1]
		s.addTokenWithLiteral(TokenString, value)

	default:
		if s.isDigit(char) {
			s.number()
		} else if s.isAlpha(char) {
			s.identifier()
		} else {
			return fmt.Errorf("unexpected character: %s", strconv.QuoteRune(char))
		}
	}

	return nil
}

func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return '\000'
	}
	return rune(s.source[s.current])
}

func (s *Scanner) peekNext() rune {
	if s.isAtEnd() {
		return '\000'
	}
	return rune(s.source[s.current+1])
}

func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() {
		return false
	}

	if rune(s.source[s.current]) != expected {
		return false
	}

	s.current++
	return true
}

func (s *Scanner) isDigit(char rune) bool {
	return char >= '0' && char <= '9'
}

func (s *Scanner) advance() rune {
	curr := rune(s.source[s.current])
	s.current++
	return curr
}

func (s *Scanner) addToken(tokenType TokenType) {
	s.addTokenWithLiteral(tokenType, nil)
}

func (s *Scanner) number() {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	// look for a fractional part
	if s.peek() == '.' && s.isDigit(s.peekNext()) {
		s.advance()
		for s.isDigit(s.peek()) {
			s.advance()
		}
	}

	val, _ := strconv.ParseFloat(s.source[s.start:s.current], 64)
	s.addTokenWithLiteral(TokenNumber, val)
}

func (s *Scanner) isAlpha(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char == '_')
}

var keywords = map[string]TokenType{
	"array":    TokenTypeIdentifier,
	"boolean":  TokenTypeIdentifier,
	"char":     TokenTypeIdentifier,
	"else":     TokenElse,
	"false":    TokenFalse,
	"for":      TokenFor,
	"function": TokenFunction,
	"if":       TokenIf,
	"integer":  TokenTypeIdentifier,
	"map":      TokenTypeIdentifier,
	"print":    TokenPrint,
	"return":   TokenReturn,
	"string":   TokenTypeIdentifier,
	"true":     TokenTrue,
	"void":     TokenVoid,
	"while":    TokenWhile,
}

func (s *Scanner) identifier() {
	for s.isAlphaNumeric(s.peek()) {
		s.advance()
	}

	text := s.source[s.start:s.current]
	tokenType, found := keywords[text]
	if !found {
		tokenType = TokenIdentifier
	}
	s.addToken(tokenType)

}

func (s *Scanner) addTokenWithLiteral(tokenType TokenType, literal interface{}) {
	text := s.source[s.start:s.current]
	token := Token{
		TokenType: tokenType,
		Lexeme:    text,
		Literal:   literal,
		Line:      s.line,
		Start:     s.start,
	}
	s.tokens = append(s.tokens, token)

}

func (s *Scanner) isAlphaNumeric(char rune) bool {
	return s.isAlpha(char) || s.isDigit(char)
}
