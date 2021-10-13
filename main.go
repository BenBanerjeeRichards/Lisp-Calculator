package main

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
)

type Token struct {
	kind string
	data string
}

type Node struct {
	kind     string
	data     string
	children []Node
}

func isDigit(c uint8) bool {
	return c >= 48 && c <= 57
}

func isLower(c uint8) bool {
	return c >= 61 && c <= 122
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
		return Token{kind: TokLBracket}, input[1:]
	}
	if input[0] == ')' {
		return Token{kind: TokRBracket}, input[1:]
	}
	if isDigit(input[0]) {
		number, remaining := takeWhile(input, isDigit)
		return Token{kind: TokNumber, data: number}, remaining
	}
	if isLower(input[0]) {
		number, remaining := takeWhile(input, isLower)
		return Token{kind: TokString, data: number}, remaining
	}

	return Token{}, ""
}

func tokenise(input string) []Token {
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

func (p Parser) CurrentToken() (Token, error) {
	if p.IsEndOfInput() {
		return Token{}, errors.New("end of input")
	}
	return p.tokens[p.currIndex], nil
}

func (p Parser) IsEndOfInput() bool {
	return p.currIndex >= len(p.tokens)
}

func (p *Parser) PeekToken() (Token, error) {
	p.currIndex += 1
	if p.IsEndOfInput() {
		return Token{}, errors.New("end of input")
	}
	token, _ := p.CurrentToken()
	p.currIndex -= 1
	return token, nil
}

func (p *Parser) NextToken() (Token, error) {
	p.currIndex += 1
	return p.CurrentToken()
}

func (p *Parser) ParserNumber() (Node, error) {
	token, err := p.CurrentToken()
	if err != nil {
		return Node{}, err
	}
	if token.kind == TokNumber {
		p.NextToken()
		return Node{kind: NumberNode, data: token.data}, nil
	}
	return Node{}, errors.New("not a number")
}

func (p *Parser) ParseLiteral() (Node, error) {
	token, err := p.CurrentToken()
	if err != nil {
		return Node{}, err
	}
	if token.kind == "TokString" {
		p.NextToken()
		return Node{kind: LiteralNode, data: token.data}, nil
	}
	return Node{}, errors.New("not a string")
}

func (p *Parser) ParseExpression() (Node, error) {
	numNode, err := p.ParserNumber()
	if err == nil {
		return Node{kind: ExpressionNode, children: []Node{numNode}}, nil
	}
	litNode, err := p.ParseLiteral()
	if err == nil {
		return Node{kind: ExpressionNode, children: []Node{litNode}}, nil
	}
	token, err := p.CurrentToken()
	if err != nil || token.kind != TokLBracket {
		return Node{}, errors.New("expected ( whilst parsing expression")
	}
	childExpressions := []Node{}
	token, tokError := p.NextToken()
	for tokError == nil && token.kind != TokRBracket {
		fmt.Println(token.kind)
		expr, err := p.ParseExpression()
		if err != nil {
			return Node{}, err
		}
		childExpressions = append(childExpressions, expr)
		token, tokError = p.CurrentToken()
	}
	if err != nil {
		return Node{}, nil
	}
	return Node{kind: ExpressionNode, children: childExpressions}, nil
}

func main() {
	// tokens := tokenise("hello 34")
	tokens := tokenise("(define r (add 5 23))(print (plus 10 r))")
	fmt.Printf("Tokens:%v\n", tokens)

	parser := Parser{}
	parser.New(tokens)
	node, err := parser.ParseExpression()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(node)
		fmt.Println(err)
	}
}
