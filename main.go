package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/benbanerjeerichards/lisp-calculator/parser"
)

func evalLiteral(node parser.Node) (string, error) {
	switch node.Kind {
	case parser.LiteralNode:
		return node.Data, nil
	case parser.ExpressionNode:
		if len(node.Children) != 1 {
			return "", errors.New("can not obtain literal from n-ary expression")
		}
		return evalLiteral(node.Children[0])
	default:
		return "", fmt.Errorf("can not obtain literal from not type %s", node.Kind)
	}
}

func Eval(node parser.Node) (float64, error) {
	switch node.Kind {
	case parser.NumberNode:
		f, err := strconv.ParseFloat(node.Data, 64)
		if err != nil {
			return 0, errors.New("failed to parse as float")
		}
		return f, nil
	case parser.LiteralNode:
		// TODO implement definitions
		return 0, errors.New("variables not yet implemented")
	case parser.ExpressionNode:
		if len(node.Children) == 0 {
			return 0, errors.New("invalid expression - must have non-zero child expressions")
		}
		if len(node.Children) == 1 {
			return Eval(node.Children[0])
		}
		literal, err := evalLiteral(node.Children[0])
		if err != nil {
			return 0, errors.New("first argument to expression operation should be a literal")
		}
		switch literal {
		case "add":
			if len(node.Children) != 3 {
				return 0, errors.New("expected two operands to binary operation add")
			}
			lhs, err := Eval(node.Children[1])
			if err != nil {
				return 0, err
			}
			rhs, err := Eval(node.Children[2])
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

func RunRepl() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("calc> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from stdin: ", err)
			break
		}
		tokens := parser.Tokenise(text)
		parser := parser.Parser{}
		parser.New(tokens)
		expr, err := parser.ParseExpression()
		if err != nil {
			fmt.Println("Parse Error: ", err.Error())
			continue
		}
		val, err := Eval(expr)
		if err != nil {
			fmt.Println("Eval Error: ", err.Error())
			continue
		}
		fmt.Println(val)
	}
}

func main() {
	RunRepl()
	tokens := parser.Tokenise("(add (add 14 10) 2)")
	// tokens := tokenise("(define r (add 55 23))(print (plus 10 r))")

	parser := parser.Parser{}
	parser.New(tokens)
	node, err := parser.ParseExpression()
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
}
