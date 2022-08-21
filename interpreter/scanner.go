package main

import (
	"fmt"
	"strconv"
)

type ScanError struct {
	line    int
	message string
}

func (e *ScanError) Error() string {
	return fmt.Sprintf("Error on line %d: %s", e.line+1, e.message)
}

func NewScanner(source string) *Scanner {
	return &Scanner{source: source}
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
	case ',':
		s.addToken(TokenComma)
	case '[':
		s.addToken(TokenLeftSquareBracket)
	case ']':
		s.addToken(TokenRightSquareBracket)
	case '{':
		s.addToken(TokenLeftBrace)
	case '}':
		s.addToken(TokenRightBrace)
	case '+':
		if s.match('+') {
			s.addToken(TokenPlusPlus)
		} else {
			s.addToken(TokenPlus)
		}
	case '-':
		if s.match('-') {
			s.addToken(TokenMinusMinus)
		} else {
			s.addToken(TokenMinus)
		}
	case '*':
		s.addToken(TokenStar)
	case '%':
		s.addToken(TokenPercent)
	case '(':
		s.addToken(TokenLeftParen)
	case ')':
		s.addToken(TokenRightParen)
	case '^':
		s.addToken(TokenCaret)

	case '|':
		if s.match('|') {
			s.addToken(TokenOr)
			break
		}
	case '&':
		if s.match('&') {
			s.addToken(TokenAnd)
			break
		}
	case '!':
		var nextToken TokenType
		if s.match('=') {
			nextToken = TokenBangEqual
		} else {
			nextToken = TokenBang
		}
		s.addToken(nextToken)
	case '=':
		var nextToken TokenType
		if s.match('=') {
			nextToken = TokenEqualEqual
		} else {
			nextToken = TokenEqual
		}
		s.addToken(nextToken)
	case '<':
		var nextToken TokenType
		if s.match('=') {
			nextToken = TokenLessEqual
		} else {
			nextToken = TokenLess
		}
		s.addToken(nextToken)
	case '>':
		var nextToken TokenType
		if s.match('=') {
			nextToken = TokenGreaterEqual
		} else {
			nextToken = TokenGreater
		}
		s.addToken(nextToken)

	case '\'':
		// TODO: unquote char
		s.advance()
		if s.isAtEnd() {
			return s.error("unterminated char")
		}
		s.advance() // advance past the closing "'"
		value := s.source[s.current-1]
		s.addTokenWithLiteral(TokenChar, CharValue(rune(value)))

	case '"':
		for s.peek() != '"' && !s.isAtEnd() {
			if s.peek() == '\n' {
				s.line++
			}
			s.advance()
		}
		if s.isAtEnd() {
			return s.error("unterminated string")
		}
		s.advance() // advance past the closing '"'
		rawString := s.source[s.start+1 : s.current-1]
		unquoted, err := s.unquote(rawString)
		if err != nil {
			return err
		}

		s.addTokenWithLiteral(TokenString, StringValue(unquoted))

	default:
		if s.isDigit(char) {
			err := s.number()
			if err != nil {
				return err
			}
		} else if s.isAlpha(char) {
			s.identifier()
		} else {
			return s.error(fmt.Sprintf("unexpected character: %s", strconv.QuoteRune(char)))
		}
	}

	return nil
}

func (s *Scanner) unquote(rawString string) (string, error) {
	return strconv.Unquote(`"` + rawString + `"`)
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

func (s *Scanner) number() error {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	val, err := strconv.Atoi(s.source[s.start:s.current])
	if err != nil {
		return err
	}

	s.addTokenWithLiteral(TokenNumber, IntegerValue(val))
	return nil
}

func (s *Scanner) isAlpha(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char == '_')
}

var keywords = map[string]TokenType{
	"else":   TokenElse,
	"false":  TokenFalse,
	"for":    TokenFor,
	"if":     TokenIf,
	"print":  TokenPrint,
	"return": TokenReturn,
	"true":   TokenTrue,
	"while":  TokenWhile,
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

func (s *Scanner) addTokenWithLiteral(tokenType TokenType, literal Value) {
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

func (s *Scanner) error(message string) error {
	return &ScanError{line: s.line, message: message}
}
