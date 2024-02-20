package parser

import (
	"errors"
	"fmt"

	"github.com/benbanerjeerichards/lisp-calculator/types"
)

const (
	NumberNode            = "NumberNode"
	StringNode            = "StringNode"
	BoolNode              = "BoolNode"
	NullNode              = "NullNode"
	LiteralNode           = "LiteralNode"
	QualifiedLiteralNode  = "QualifiedLiteral"
	ExpressionNode        = "ExpressionNode"
	ProgramNode           = "ProgramNode"
	AccessorNode          = "AccessorNode"
	AccessorOperationNode = "AccessorOperationNode"
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
	case QualifiedLiteralNode:
		return fmt.Sprintf("'%s'", node.Data)
	case AccessorNode:
		return "Accessor"
	case AccessorOperationNode:
		return "AccessorExpression"
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

func (p *Parser) backtrack() {
	p.currIndex -= 1
	if p.currIndex < 0 {
		p.currIndex = 0
	}
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

func (p *Parser) parseQualifiedLiteral() (Node, error) {
	startIdx := p.currIndex
	qualifierNode, err := p.parseLiteral()
	if err == nil {
		tok, err := p.currentToken()
		if err == nil && tok.Kind == TokDot {
			p.nextToken()
			nameNode, err := p.parseLiteral()
			if err == nil {
				nodeRange := qualifierNode.Range
				nodeRange.End = nameNode.Range.End
				qu := Node{Range: nodeRange, Kind: QualifiedLiteralNode, Children: []Node{qualifierNode, nameNode}}
				return Node{Range: nodeRange, Kind: ExpressionNode, Children: []Node{qu}}, nil
			}
		}
	}
	p.currIndex = startIdx
	return Node{}, errors.New("not a qualified literal")
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

	// TODO what about struct access on qualified literal? (e.g. qual.st:field)
	qual, err := p.parseQualifiedLiteral()
	if err == nil {
		return qual, nil
	}
	litNode, err := p.parseLiteral()
	if err == nil {
		// Could be a literal accessor (<literal>:<literal>)
		currToken, err := p.currentToken()
		if err == nil && currToken.Kind == TokColon {
			p.nextToken()
			accessorRhs, err := p.parseLiteral()
			if err != nil {
				return Node{}, types.Error{Simple: "Invalid struct accessor format - RHS of colon must be an identifier", Range: currToken.Range}
			}
			accessorRange := litNode.Range
			accessorRange.End = accessorRhs.Range.End
			accessorNode := Node{Kind: AccessorNode, Range: accessorRange, Children: []Node{litNode, accessorRhs}}
			return Node{Kind: ExpressionNode, Children: []Node{accessorNode}}, nil
		}
		return Node{Kind: ExpressionNode, Children: []Node{litNode}, Range: litNode.Range}, nil
	}

	token, err := p.currentToken()
	if err != nil || token.Kind != TokLBracket {
		return Node{}, types.Error{Simple: fmt.Sprintf("Expected `(` whilst parsing expression, got %s", token.Kind), Range: token.Range}
	}

	peekToken, err := p.nextToken()
	if peekToken.Kind == TokColon && err == nil {
		// Must be struct accessor (:name person)
		p.nextToken()
		literal, err := p.parseLiteral()
		if err != nil {
			return Node{}, types.Error{Range: peekToken.Range, Simple: "Invalid accessor syntax - expected literal after :"}
		}
		structParse, err := p.ParseExpression()
		if err != nil {
			return Node{}, err
		}
		endBracket, err := p.currentToken()
		if err != nil || endBracket.Kind != TokRBracket {
			return Node{}, types.Error{Simple: "Expected ) after accessor", Range: literal.Range}
		}
		p.nextToken()
		return Node{Kind: AccessorOperationNode, Children: []Node{literal, structParse}}, nil
	} else {
		p.backtrack()
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
