package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/benbanerjeerichards/lisp-calculator/types"
)

const (
	TokNumber   = "TokNumber"
	TokIdent    = "TokIdent"
	TokString   = "TokString"
	TokLBracket = "TokLBracket"
	TokRBracket = "TokRBracket"
)

// Taken from standard library (strings)
var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}
var eof uint8 = 0xFF
var identifierRegex, _ = regexp.Compile(`^[^0-9\s()][^()\s]*$`)

type Token struct {
	Kind  string
	Data  string
	Range types.FileRange
}

func (t Token) String() string {
	if len(t.Data) > 0 {
		return fmt.Sprintf("%s %s(%s)", t.Range, t.Kind, t.Data)
	}
	return fmt.Sprintf("%s %s", t.Range, t.Kind)
}

func isDigit(c uint8) bool {
	return (c >= 48 && c <= 57) || c == '.'
}

func isIdentifierChar(c uint8) bool {
	return ((c >= '!' && c <= '/') || (c >= ':' && c < eof)) && (c != '(' && c != ')')
}

type Tokeniser struct {
	input string
	index int
	line  int
	col   int
}

func (t *Tokeniser) New(input string) {
	t.index = 0
	// Line and col are 1-indexed as they are for display purpose only
	t.line = 1
	t.col = 1
	t.input = input
}

func (t Tokeniser) isEOF() bool {
	return t.Current() == eof
}

func (t *Tokeniser) nextChar() uint8 {
	t.index += 1
	t.col += 1
	return t.Current()
}

func (t *Tokeniser) consumeSpaces() {
	for !t.isEOF() && isSpace(t.Current()) {
		if t.Current() == '\n' {
			t.line += 1
			t.col = 1
		}
		t.nextChar()
	}
}

func (t Tokeniser) currentPos() types.FilePos {
	return types.FilePos{Line: t.line, Col: t.col, Position: t.index}
}

func (t Tokeniser) Current() uint8 {
	if t.index >= len(t.input) {
		return eof
	}
	return t.input[t.index]
}

// Peek ahead character stream, will return EOF if requested character exists after end of string
// Peek(0) = Current()
func (t Tokeniser) Peek(n int) uint8 {
	if t.index+n > len(t.input)-1 {
		return eof
	}
	return t.input[t.index+n]
}

func (t *Tokeniser) SeekAhead(amount int) {
	if t.index+amount > len(t.input) {
		amount = len(t.input) - 1
	}
	t.index += amount
}

func (t *Tokeniser) consumeWhile(condition func(uint8) bool) (string, types.FileRange) {
	start := t.currentPos()
	acc := ""
	for !t.isEOF() && condition(t.Current()) {
		acc += string(t.Current())
		t.nextChar()
	}
	return acc, types.FileRange{Start: start, End: t.currentPos()}
}

func (t *Tokeniser) nextToken() (Token, bool) {
	t.consumeSpaces()
	nextChar := t.Current()
	start := t.currentPos()
	if nextChar == '(' {
		t.nextChar()
		return Token{Kind: TokLBracket, Range: types.FileRange{Start: start, End: t.currentPos()}}, true
	}
	if nextChar == ')' {
		t.nextChar()
		return Token{Kind: TokRBracket, Range: types.FileRange{Start: start, End: t.currentPos()}}, true
	}
	// TODO should improve this, probably just use regexp
	if isDigit(nextChar) || (nextChar == '-' && isDigit(t.Peek(1))) {
		isNeg := false
		if nextChar == '-' {
			isNeg = true
			t.nextChar()
		}
		number, fRange := t.consumeWhile(isDigit)
		if isNeg {
			number = "-" + number
		}
		return Token{Kind: TokNumber, Data: number, Range: fRange}, true
	}
	if nextChar == '"' {
		var stringLit strings.Builder
		nextChar = t.nextChar()
		for nextChar != '"' {
			if nextChar == eof {
				panic("TODO proper error handling") //FIXME
			}
			if nextChar == '\\' && t.Peek(1) == '"' {
				stringLit.WriteByte('"')
				t.nextChar()
			} else {
				stringLit.WriteByte(nextChar)
			}
			nextChar = t.nextChar()
		}
		t.nextChar()
		return Token{Kind: TokString, Data: stringLit.String(), Range: types.FileRange{Start: start, End: t.currentPos()}}, true
	}

	// Now attempt to match an identifier
	// Scan all non-whitespace characters and then test using regex
	var identBuilder strings.Builder
	i := 0
	for !isSpace(nextChar) && nextChar != eof && nextChar != '(' && nextChar != ')' {
		identBuilder.WriteByte(nextChar)
		i += 1
		nextChar = t.Peek(i)
	}
	if identifierRegex.MatchString(identBuilder.String()) {
		t.SeekAhead(i)
		return Token{Kind: TokIdent, Data: identBuilder.String(),
			Range: types.FileRange{Start: start, End: t.currentPos()}}, true
	}

	if t.isEOF() {
		return Token{}, false
	}

	return Token{}, false
}

func (t *Tokeniser) doTokenise() []Token {
	tokens := make([]Token, 0)
	for {
		token, ok := t.nextToken()
		if !ok {
			return tokens
		}
		tokens = append(tokens, token)
	}
}

func Tokenise(input string) []Token {
	if len(input) == 0 {
		return []Token{}
	}
	tok := Tokeniser{}
	tok.New(input)
	return tok.doTokenise()
}

func isSpace(c uint8) bool {
	return asciiSpace[c] == 1
}
