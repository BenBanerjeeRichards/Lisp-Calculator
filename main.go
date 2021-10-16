package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
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
	kind string
	data string
}

type Node struct {
	kind     string
	data     string
	children []Node
}

func (node Node) Label() string {
	switch node.kind {
	case ExpressionNode:
		return "Expr"
	case ProgramNode:
		return "Prog"
	case NumberNode:
		return node.data
	case LiteralNode:
		return fmt.Sprintf("'%s'", node.data)
	default:
		return node.kind
	}
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

func writeToFile(path string, contents string) {
	file, _ := os.Create(path)
	defer file.Close()
	file.WriteString(contents)
}

func parseTreeToDot(node Node) string {
	var dotBuilder strings.Builder
	index := 0
	dotBuilder.WriteString("digraph parse {\n")
	dotBuilder.WriteString(fmt.Sprintf("\t%d[label=\"%s\"]\n", 0, node.Label()))
	doParseTreeToDot(node, &dotBuilder, &index)
	dotBuilder.WriteString("}\n")
	return dotBuilder.String()
}

func doParseTreeToDot(node Node, builder *strings.Builder, i *int) {
	rootIndex := *i
	for _, child := range node.children {
		(*i) += 1
		builder.WriteString(fmt.Sprintf("\t%d -> %d\n", rootIndex, *i))
		builder.WriteString(fmt.Sprintf("\t%d[label=\"%s\"]\n", *i, child.Label()))
		doParseTreeToDot(child, builder, i)
	}
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

func (p *Parser) parseLiteral() (Node, error) {
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

func (p *Parser) parseExpression() (Node, error) {
	numNode, err := p.ParserNumber()
	if err == nil {
		return Node{kind: ExpressionNode, children: []Node{numNode}}, nil
	}
	litNode, err := p.parseLiteral()
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
		expr, err := p.parseExpression()
		if err != nil {
			return Node{}, err
		}
		childExpressions = append(childExpressions, expr)
		token, tokError = p.CurrentToken()
	}
	if err != nil {
		return Node{}, nil
	}
	p.NextToken()
	return Node{kind: ExpressionNode, children: childExpressions}, nil
}

func (p *Parser) parseProgram() (Node, error) {
	expressions := []Node{}
	for {
		expr, err := p.parseExpression()
		if err != nil {
			if p.IsEndOfInput() {
				prog := Node{kind: ProgramNode, children: expressions}
				return prog, nil
			} else {
				// Unexpected end of input
				return Node{}, err
			}
		}
		expressions = append(expressions, expr)
	}
}

func evalLiteral(node Node) (string, error) {
	switch node.kind {
	case LiteralNode:
		return node.data, nil
	case ExpressionNode:
		if len(node.children) != 1 {
			return "", errors.New("can not obtain literal from n-ary expression")
		}
		return evalLiteral(node.children[0]);
	default:
		return "", fmt.Errorf("can not obtain literal from not type %s", node.kind)
	}
}

func Eval(node Node) (float64, error) {
	switch node.kind {
	case NumberNode:
		f, err := strconv.ParseFloat(node.data, 64)
		if err != nil {
			return 0, errors.New("failed to parse as float")
		}
		return f, nil
	case LiteralNode:
		// TODO implement definitions
		return 0, errors.New("variables not yet implemented")
	case ExpressionNode:
		if len(node.children) == 1 {
			return Eval(node.children[0])
		}
		literal, err := evalLiteral(node.children[0])
		if err != nil {
			return 0, errors.New("first argument to expression operation should be a literal")
		}
		switch literal {
		case "add":
			if len(node.children) != 3 {
				return 0, errors.New("expected two operands to binary operation add")
			}
			lhs, err := Eval(node.children[1])
			if err != nil {
				return 0, err
			}
			rhs, err := Eval(node.children[2])
			if err != nil {
				return 0, err
			}
			return lhs + rhs, nil
		}
	default:
		return 0, errors.New("unknown node")
	}
	return 0, nil
}

func main() {
	tokens := tokenise("(add (add 14 10) 2)")
	// tokens := tokenise("(define r (add 55 23))(print (plus 10 r))")

	parser := Parser{}
	parser.New(tokens)
	node, err := parser.parseExpression()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(node)
	}
	val, err := Eval(node)
	if err != nil {
		fmt.Println("Eval error occured", err.Error())
	} else {
		fmt.Println(val)
	}

	writeToFile("syntax.dot", parseTreeToDot(node))
}
