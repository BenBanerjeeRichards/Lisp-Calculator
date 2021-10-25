package main

import (
	"github.com/benbanerjeerichards/lisp-calculator/eval"
	"github.com/benbanerjeerichards/lisp-calculator/test"
)

func main() {
	test.Run()
	eval.RunRepl()
}
