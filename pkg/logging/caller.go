package logging

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// This is the repo root
var root = func() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../..")
}()

type Caller struct {
	File string
	Line int
}

func (c Caller) LocationAbs() string {
	return fmt.Sprintf("%s:%d", c.File, c.Line)
}

func (c Caller) Location() string {
	return fmt.Sprintf("%s:%d", strings.Replace(c.File, root, ".", 1), c.Line)
}

func getCaller(depth int) Caller {
	pc := make([]uintptr, 10)
	if runtime.Callers(depth, pc) == 0 {
		return Caller{}
	}

	pc = pc[:1]
	frames := runtime.CallersFrames(pc)
	frame, _ := frames.Next()

	return Caller{
		File: frame.File,
		Line: frame.Line,
	}
}
