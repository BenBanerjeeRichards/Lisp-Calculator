package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/benbanerjeerichards/lisp-calculator/parser"
)

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
