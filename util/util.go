package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/benbanerjeerichards/lisp-calculator/parser"
)

func ReadFile(path string) (string, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(f), nil
}

func WriteToFile(path string, contents string) {
	file, _ := os.Create(path)
	defer file.Close()
	file.WriteString(contents)
}

func ParseTreeToDot(node parser.Node) string {
	var dotBuilder strings.Builder
	index := 0
	dotBuilder.WriteString("digraph parse {\n")
	dotBuilder.WriteString(fmt.Sprintf("\t%d[label=\"%s\"]\n", 0, node.Label()))
	doParseTreeToDot(node, &dotBuilder, &index)
	dotBuilder.WriteString("}\n")
	return dotBuilder.String()
}

func doParseTreeToDot(node parser.Node, builder *strings.Builder, i *int) {
	rootIndex := *i
	for _, child := range node.Children {
		(*i) += 1
		builder.WriteString(fmt.Sprintf("\t%d -> %d\n", rootIndex, *i))
		builder.WriteString(fmt.Sprintf("\t%d[label=\"%s\"]\n", *i, child.Label()))
		doParseTreeToDot(child, builder, i)
	}
}

func ParseTreeToString(node parser.Node) string {
	var dotBuilder strings.Builder
	doParseTreeToString(node, &dotBuilder, 0)
	return dotBuilder.String()
}

func doParseTreeToString(node parser.Node, builder *strings.Builder, i int) {
	for _, child := range node.Children {
		for n := 1; n < i; n++ {
			builder.WriteString("\t")
		}
		builder.WriteString(child.Kind)
		if len(child.Data) > 0 {
			builder.WriteString(fmt.Sprintf("(%s)", child.Data))
		}
		builder.WriteString("\n")
		doParseTreeToString(child, builder, i+1)
	}
}

func FileExists(filename string) bool {
	// https://stackoverflow.com/a/57791506/6404474
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func SubHomeDir(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		lower := len(home) + 1
		if lower > len(path) {
			lower = len(path)
		}
		return "~/" + path[lower:]
	}
	return path
}
