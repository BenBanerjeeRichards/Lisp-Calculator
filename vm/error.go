package vm

import "fmt"

type TraceFrame struct {
	FilePath   string
	LineNumber int
}

type RuntimeError struct {
	Line       int
	Simple     string
	Detail     string
	FilePath   string
	StackTrace []TraceFrame
}

func (r *RuntimeError) AddStackTrace(filePath string, lineNumber int) {
	if r.StackTrace == nil {
		r.StackTrace = make([]TraceFrame, 0)
	}
	r.StackTrace = append(r.StackTrace, TraceFrame{FilePath: filePath, LineNumber: lineNumber})
}

func (a RuntimeError) Error() string {
	out := fmt.Sprintf("%s:%d: %s (%s)", a.FilePath, a.Line, a.Simple, a.Detail)
	if a.StackTrace == nil {
		return out
	}
	for _, trace := range a.StackTrace {
		out += fmt.Sprintf("\n\tat %s:%d", trace.FilePath, trace.LineNumber)
	}
	return out
}
