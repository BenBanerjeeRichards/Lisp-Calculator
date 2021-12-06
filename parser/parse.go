package parser

import (
	"errors"
	"fmt"

	"github.com/benbanerjeerichards/lisp-calculator/types"
)

const (
	NumberNode     = "NumberNode"
	StringNode     = "StringNode"
	BoolNode       = "BoolNode"
	NullNode       = "NullNode"
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
	Range    types.FileRange
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
	case NullNode:
		return "null"
	case StringNode:
		return "\"" + node.Data + "\""
	case BoolNode:
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
		errRange := types.FileRange{Start: types.FilePos{Line: 0, Col: 0, Position: 0}, End: types.FilePos{Line: 0, Col: 0, Position: 0}}
		if len(p.tokens) > 0 {
			errRange = p.tokens[len(p.tokens)-1].Range
		}
		return Token{}, types.Error{Range: errRange, Simple: "Unexpected end of input"}
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

func (p *Parser) parseString() (Node, error) {
	token, err := p.currentToken()
	if err != nil {
		return Node{}, err
	}
	if token.Kind == TokString {
		p.nextToken()
		return Node{Kind: StringNode, Data: token.Data, Range: token.Range}, nil
	}
	return Node{}, errors.New("not a string")
}

func (p *Parser) parseLiteral() (Node, error) {
	token, err := p.currentToken()
	if err != nil {
		return Node{}, err
	}
	if token.Kind == TokIdent {
		p.nextToken()
		if token.Data == "true" || token.Data == "false" {
			return Node{Kind: BoolNode, Data: token.Data, Range: token.Range}, nil
		} else if token.Data == "null" {
			return Node{Kind: NullNode, Range: token.Range}, nil
		}
		return Node{Kind: LiteralNode, Data: token.Data, Range: token.Range}, nil
	}
	return Node{}, errors.New("not a string")
}

func (p *Parser) ParseExpression() (Node, error) {
	numNode, err := p.parserNumber()
	if err == nil {
		return Node{Kind: ExpressionNode, Children: []Node{numNode}, Range: numNode.Range}, nil
	}
	strNode, err := p.parseString()
	if err == nil {
		return Node{Kind: ExpressionNode, Children: []Node{strNode}, Range: strNode.Range}, nil
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
	if tokError != nil {
		return Node{}, tokError
	}
	if err != nil {
		return Node{}, err
	}
	p.nextToken()
	return Node{Kind: ExpressionNode, Children: childExpressions, Range: types.FileRange{Start: rangeStart.End, End: rangeEnd.End}}, nil
}

func (p *Parser) ParseProgram() (Node, error) {
	expressions := []Node{}
	for !p.isEndOfInput() {
		expr, err := p.ParseExpression()
		if err != nil {
			return Node{}, err
		}
		expressions = append(expressions, expr)
	}
	eRange := types.FileRange{}
	if len(expressions) == 0 {
		zeroPos := types.FilePos{Line: 0, Col: 0, Position: 0}
		eRange = types.FileRange{Start: zeroPos, End: zeroPos}
	} else {
		eRange = types.FileRange{Start: expressions[0].Range.Start, End: expressions[len(expressions)-1].Range.End}
	}
	prog := Node{Kind: ProgramNode, Children: expressions, Range: eRange}
	return prog, nil

}
