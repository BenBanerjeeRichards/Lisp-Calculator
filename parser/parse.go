package parser

import (
	"errors"
	"fmt"
	"strings"
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

type Token struct {
	Kind string
	Data string
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

func isLower(c uint8) bool {
	return c >= 61 && c <= 122
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

func nextToken(input string) (Token, string) {
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return Token{}, ""
	}
	if input[0] == '(' {
		return Token{Kind: TokLBracket}, input[1:]
	}
	if input[0] == ')' {
		return Token{Kind: TokRBracket}, input[1:]
	}
	// TODO should improve this, probably just use regexp
	if isDigit(input[0]) || input[0] == '-' {
		isNeg := false
		if input[0] == '-' {
			isNeg = true
			input = input[1:]
		}
		number, remaining := takeWhile(input, isDigit)
		if isNeg {
			number = "-" + number
		}
		return Token{Kind: TokNumber, Data: number}, remaining
	}
	if isAlphaNumeric(input[0]) {
		number, remaining := takeWhile(input, isAlphaNumeric)
		return Token{Kind: TokString, Data: number}, remaining
	}

	return Token{}, ""
}

func Tokenise(input string) []Token {
	tokens := make([]Token, 0)
	for len(input) > 0 {
		token, remaining := nextToken(input)
		input = remaining
		tokens = append(tokens, token)
	}
	return tokens
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
