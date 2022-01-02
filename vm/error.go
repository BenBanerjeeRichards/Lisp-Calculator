package vm

import "fmt"

type RuntimeError struct {
	Line   int
	Simple string
	Detail string
}

func (a RuntimeError) Error() string {
	return fmt.Sprintf("%d: %s (%s)", a.Line, a.Simple, a.Detail)
}
