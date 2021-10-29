package parser

import (
	"errors"
	"fmt"
)

const (
	TokNumber   = "TokNumber"
	TokString   = "TokString"
	TokLBracket = "TokLBracket"
	TokRBracket = "TokRBracket"
)

const (
	NumberNode     = "NumberNode"
	LiteralNode    = "LiteralNode"
	ExpressionNode = "ExpressionNode"
	ProgramNode    = "ProgramNode"
)

// Taken from standard library (strings)
var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

type FilePos struct {
	Line     int
	Col      int
	Position int
}

// Range in a file from Start (inclusive) to End (Exclusive)
type FileRange struct {
	Start FilePos
	End   FilePos
}

type Token struct {
	Kind  string
	Data  string
	Range FileRange
}

type Node struct {
	Kind     string
	Data     string
	Children []Node
}

func (node Node) Label() string {
	switch node.Kind {
	case ExpressionNode:
		return "Expr"
	case ProgramNode:
		return "Prog"
	case NumberNode:
		return node.Data
	case LiteralNode:
		return fmt.Sprintf("'%s'", node.Data)
	default:
		return node.Kind
	}
}

func (node Node) ChildNodes() []Node {
	return node.Children
}

func isDigit(c uint8) bool {
	return (c >= 48 && c <= 57) || c == '.'
}

func isAlphaNumeric(c uint8) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

func takeWhile(input string, f func(uint8) bool) (string, string) {
	i := 0
	for i < len(input) && f(input[i]) {
		i++
	}
	return input[0:i], input[i:]
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
	return t.index >= len(t.input)-1
}

func (t *Tokeniser) nextChar() uint8 {
	if t.isEOF() {
		return t.input[len(t.input)-1]
	}
	t.index += 1
	t.col += 1
	return t.input[t.index]
}

func (t *Tokeniser) consumeSpaces() {
	for !t.isEOF() && asciiSpace[t.input[t.index]] == 1 {
		if t.input[t.index] == '\n' {
			t.line += 1
			t.col = 1
		}
		t.nextChar()
	}
}

func (t Tokeniser) currentPos() FilePos {
	return FilePos{Line: t.line, Col: t.col, Position: t.index}
}

func (t *Tokeniser) consumeWhile(condition func(uint8) bool) (string, FileRange) {
	start := t.currentPos()
	acc := ""
	for !t.isEOF() && condition(t.input[t.index]) {
		acc += string(t.input[t.index])
		t.nextChar()
	}
	return acc, FileRange{Start: start, End: t.currentPos()}
}

func (t *Tokeniser) nextToken() (Token, bool) {
	t.consumeSpaces()
	if t.isEOF() {
		return Token{}, false
	}
	nextChar := t.input[t.index]
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
	tok := Tokeniser{}
	tok.New(input)
	return tok.doTokenise()
}

type Parser struct {
	tokens    []Token
	currIndex int
}

func (p *Parser) New(tokens []Token) {
	p.currIndex = 0
	p.tokens = tokens
}

func (p Parser) currentToken() (Token, error) {
	if p.isEndOfInput() {
		return Token{}, errors.New("end of input")
	}
	return p.tokens[p.currIndex], nil
}

func (p Parser) isEndOfInput() bool {
	return p.currIndex >= len(p.tokens)
}

func (p *Parser) nextToken() (Token, error) {
	p.currIndex += 1
	return p.currentToken()
}

func (p *Parser) parserNumber() (Node, error) {
	token, err := p.currentToken()
	if err != nil {
		return Node{}, err
	}
	if token.Kind == TokNumber {
		p.nextToken()
		return Node{Kind: NumberNode, Data: token.Data}, nil
	}
	return Node{}, errors.New("not a number")
}

func (p *Parser) parseLiteral() (Node, error) {
	token, err := p.currentToken()
	if err != nil {
		return Node{}, err
	}
	if token.Kind == "TokString" {
		p.nextToken()
		return Node{Kind: LiteralNode, Data: token.Data}, nil
	}
	return Node{}, errors.New("not a string")
}

func (p *Parser) ParseExpression() (Node, error) {
	numNode, err := p.parserNumber()
	if err == nil {
		return Node{Kind: ExpressionNode, Children: []Node{numNode}}, nil
	}
	litNode, err := p.parseLiteral()
	if err == nil {
		return Node{Kind: ExpressionNode, Children: []Node{litNode}}, nil
	}
	token, err := p.currentToken()
	if err != nil || token.Kind != TokLBracket {
		return Node{}, errors.New("expected ( whilst parsing expression")
	}
	childExpressions := []Node{}
	token, tokError := p.nextToken()
	for tokError == nil && token.Kind != TokRBracket {
		expr, err := p.ParseExpression()
		if err != nil {
			return Node{}, err
		}
		childExpressions = append(childExpressions, expr)
		token, tokError = p.currentToken()
	}
	if err != nil {
		return Node{}, nil
	}
	p.nextToken()
	return Node{Kind: ExpressionNode, Children: childExpressions}, nil
}

func (p *Parser) ParseProgram() (Node, error) {
	expressions := []Node{}
	for {
		expr, err := p.ParseExpression()
		if err != nil {
			if p.isEndOfInput() {
				prog := Node{Kind: ProgramNode, Children: expressions}
				return prog, nil
			} else {
				// Unexpected end of input
				return Node{}, err
			}
		}
		expressions = append(expressions, expr)
	}
}
