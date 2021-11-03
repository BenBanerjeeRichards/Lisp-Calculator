package types

import (
	"fmt"
)

// Types that are used throughout the program
// Errors etc

type Error struct {
	Range  FileRange
	Simple string
	Detail string
}

func (a Error) Error() string {
	return fmt.Sprintf("[%d:%d-%d:%d]: %s (%s)", a.Range.Start.Line, a.Range.Start.Col, a.Range.End.Line, a.Range.End.Col, a.Simple, a.Detail)
}

type FilePos struct {
	Line     int
	Col      int
	Position int
}

func (f FilePos) String() string {
	return fmt.Sprintf("%d:%d", f.Line, f.Col)
}

// Range in a file from Start (inclusive) to End (Exclusive)
type FileRange struct {
	Start FilePos
	End   FilePos
}

func (f FileRange) String() string {
	return fmt.Sprintf("%s-%s", f.Start, f.End)
}
