package parser

import (
	"errors"
	"fmt"
)

const (
	NumberNode     = "NumberNode"
	LiteralNode    = "LiteralNode"
	ExpressionNode = "ExpressionNode"
	ProgramNode    = "ProgramNode"
)

type Parser struct {
	tokens    []Token
	currIndex int
}

type Node struct {
	Kind     string
	Data     string
	Range    FileRange
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
		return Node{Kind: NumberNode, Data: token.Data, Range: token.Range}, nil
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
		return Node{Kind: LiteralNode, Data: token.Data, Range: token.Range}, nil
	}
	return Node{}, errors.New("not a string")
}

func (p *Parser) ParseExpression() (Node, error) {
	numNode, err := p.parserNumber()
	if err == nil {
		return Node{Kind: ExpressionNode, Children: []Node{numNode}, Range: numNode.Range}, nil
	}
	litNode, err := p.parseLiteral()
	if err == nil {
		return Node{Kind: ExpressionNode, Children: []Node{litNode}, Range: litNode.Range}, nil
	}
	token, err := p.currentToken()
	if err != nil || token.Kind != TokLBracket {
		return Node{}, errors.New("expected ( whilst parsing expression")
	}
	// Start and End point to '(' and ')' respectivly
	rangeStart := token.Range
	rangeEnd := token.Range
	childExpressions := []Node{}
	token, tokError := p.nextToken()
	for tokError == nil && token.Kind != TokRBracket {
		expr, err := p.ParseExpression()
		if err != nil {
			return Node{}, err
		}
		childExpressions = append(childExpressions, expr)
		token, tokError = p.currentToken()
		rangeEnd = token.Range
	}
	if err != nil {
		return Node{}, nil
	}
	p.nextToken()
	return Node{Kind: ExpressionNode, Children: childExpressions, Range: FileRange{Start: rangeStart.End, End: rangeEnd.End}}, nil
}

func (p *Parser) ParseProgram() (Node, error) {
	expressions := []Node{}
	for {
		expr, err := p.ParseExpression()
		if err != nil {
			if p.isEndOfInput() {
				eRange := FileRange{}
				if len(expressions) == 0 {
					zeroPos := FilePos{Line: 0, Col: 0, Position: 0}
					eRange = FileRange{Start: zeroPos, End: zeroPos}
				} else {
					eRange = FileRange{Start: expressions[0].Range.Start, End: expressions[len(expressions)-1].Range.End}
				}
				prog := Node{Kind: ProgramNode, Children: expressions, Range: eRange}
				return prog, nil
			} else {
				// Unexpected end of input
				return Node{}, err
			}
		}
		expressions = append(expressions, expr)
	}
}
