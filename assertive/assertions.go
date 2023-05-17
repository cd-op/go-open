// Package assertive provides two boolean assertion functions, both compatible
// with the standard testing.T type. They avoid repetitive boilerplate around
// T.Error and T.Fatal and augment those functions with extracted source code.
// See this package's tests for usage examples.
package assertive

import (
	"io/ioutil"
	"runtime"
	"strings"
)

// Want marks the test as failed if condition is false. Use this function
// when the test may continue even if the assertion fails.
func Want(t miniT, condition bool) {
	t.Helper()

	if !condition {
		line := getLine(t)
		t.Error(line)
	}
}

// Need marks the test as failed if condition is false and stops execution
// of the running test. Use this function when the test cannot continue after
// the current assertion fails, due to cascading failures or because it makes
// no sense to continue the execution for any other reason.
func Need(t miniT, condition bool) {
	t.Helper()

	if !condition {
		line := getLine(t)
		t.Fatal(line)
	}
}

type miniT interface {
	Error(...any)
	Fatal(...any)
	Helper()
}

func getLine(t miniT) string {
	t.Helper()

	// get the program counter for the caller of Want()/Need()
	// and convert the program counter to a runtime.Frame
	pc := make([]uintptr, 1)
	n := runtime.Callers(3, pc)

	if n < 1 {
		// this cannot happen under normal circumstances
		// at the very least we'd have 4 frames
		// main -> Want/Need -> getLine -> runtime.Callers
		return "[ERROR OBTAINING CALLER FRAME]"
	}

	frames := runtime.CallersFrames(pc)
	frame, _ := frames.Next()

	text, err := ioutil.ReadFile(frame.File)
	if err != nil {
		// this cannot happen under normal circumstances
		// frame.File is where the function corresponding
		// to the 4th stack frame is defined
		return "[ERROR READING SOURCE FILE]"
	}

	// isolate the line we want by asking for one extra split
	parts := strings.SplitN(string(text), "\n", frame.Line+1)

	// file lines are 1-indexed
	return parts[frame.Line-1]
}
