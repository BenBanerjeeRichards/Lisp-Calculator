package parser

import "fmt"

const (
	TokNumber   = "TokNumber"
	TokString   = "TokString"
	TokLBracket = "TokLBracket"
	TokRBracket = "TokRBracket"
)

// Taken from standard library (strings)
var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}
var eof uint8 = 0xFF

type FilePos struct {
	Line     int
	Col      int
	Position int
}

func (f FilePos) String() string {
	return fmt.Sprintf("%d:%d", f.Line, f.Col)
}

// Range in a file from Start (inclusive) to End (Exclusive)
type FileRange struct {
	Start FilePos
	End   FilePos
}

func (f FileRange) String() string {
	return fmt.Sprintf("%s-%s", f.Start, f.End)
}

type Token struct {
	Kind  string
	Data  string
	Range FileRange
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

func isAlphaNumeric(c uint8) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
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
	for !t.isEOF() && asciiSpace[t.Current()] == 1 {
		if t.Current() == '\n' {
			t.line += 1
			t.col = 1
		}
		t.nextChar()
	}
}

func (t Tokeniser) currentPos() FilePos {
	return FilePos{Line: t.line, Col: t.col, Position: t.index}
}

func (t Tokeniser) Current() uint8 {
	if t.index >= len(t.input) {
		return eof
	}
	return t.input[t.index]
}

func (t *Tokeniser) consumeWhile(condition func(uint8) bool) (string, FileRange) {
	start := t.currentPos()
	acc := ""
	for !t.isEOF() && condition(t.Current()) {
		acc += string(t.Current())
		t.nextChar()
	}
	return acc, FileRange{Start: start, End: t.currentPos()}
}

func (t *Tokeniser) nextToken() (Token, bool) {
	t.consumeSpaces()
	nextChar := t.Current()
	start := t.currentPos()
	if nextChar == '(' {
		t.nextChar()
		return Token{Kind: TokLBracket, Range: FileRange{Start: start, End: t.currentPos()}}, true
	}
	if nextChar == ')' {
		t.nextChar()
		return Token{Kind: TokRBracket, Range: FileRange{Start: start, End: t.currentPos()}}, true
	}
	// TODO should improve this, probably just use regexp
	if isDigit(nextChar) || nextChar == '-' {
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
	if isAlphaNumeric(nextChar) {
		number, fRange := t.consumeWhile(isAlphaNumeric)
		return Token{Kind: TokString, Data: number, Range: fRange}, true
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
