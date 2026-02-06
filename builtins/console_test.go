package builtins

import (
	"bytes"
	"strings"
	"testing"

	"github.com/example/jsgo/runtime"
)

func TestConsoleLog(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := stdout
	stdout = &buf
	defer func() { stdout = oldStdout }()

	consoleLog(runtime.Undefined, []*runtime.Value{runtime.NewString("hello"), runtime.NewNumber(42)})
	got := strings.TrimSpace(buf.String())
	if got != "hello 42" {
		t.Errorf("console.log: got %q, want %q", got, "hello 42")
	}
}

func TestConsoleError(t *testing.T) {
	var buf bytes.Buffer
	oldStderr := stderr
	stderr = &buf
	defer func() { stderr = oldStderr }()

	consoleError(runtime.Undefined, []*runtime.Value{runtime.NewString("error!")})
	got := strings.TrimSpace(buf.String())
	if got != "error!" {
		t.Errorf("console.error: got %q, want %q", got, "error!")
	}
}

func TestConsoleLogArray(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := stdout
	stdout = &buf
	defer func() { stdout = oldStdout }()

	arr := newArray([]*runtime.Value{runtime.NewNumber(1), runtime.NewNumber(2), runtime.NewNumber(3)})
	consoleLog(runtime.Undefined, []*runtime.Value{runtime.NewObject(arr)})
	got := strings.TrimSpace(buf.String())
	if got != "[ 1, 2, 3 ]" {
		t.Errorf("console.log array: got %q, want %q", got, "[ 1, 2, 3 ]")
	}
}
