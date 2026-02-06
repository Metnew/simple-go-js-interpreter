package builtins

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/example/jsgo/runtime"
)

var (
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

func createConsoleObject(proto *runtime.Object) *runtime.Object {
	console := runtime.NewOrdinaryObject(proto)

	setMethod(console, "log", 0, consoleLog)
	setMethod(console, "error", 0, consoleError)
	setMethod(console, "warn", 0, consoleWarn)
	setMethod(console, "info", 0, consoleLog)
	setMethod(console, "debug", 0, consoleLog)

	return console
}

func formatArgs(args []*runtime.Value) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = formatValue(a)
	}
	return strings.Join(parts, " ")
}

func formatValue(v *runtime.Value) string {
	if v == nil {
		return "undefined"
	}
	switch v.Type {
	case runtime.TypeObject:
		if v.Object == nil {
			return "[object Object]"
		}
		if v.Object.OType == runtime.ObjTypeArray {
			return formatArray(v.Object)
		}
		return "[object Object]"
	default:
		return v.ToString()
	}
}

func formatArray(obj *runtime.Object) string {
	parts := make([]string, len(obj.ArrayData))
	for i, v := range obj.ArrayData {
		if v == nil || v.Type == runtime.TypeUndefined {
			parts[i] = ""
		} else {
			parts[i] = formatValue(v)
		}
	}
	return "[ " + strings.Join(parts, ", ") + " ]"
}

func consoleLog(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	fmt.Fprintln(stdout, formatArgs(args))
	return runtime.Undefined, nil
}

func consoleError(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	fmt.Fprintln(stderr, formatArgs(args))
	return runtime.Undefined, nil
}

func consoleWarn(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	fmt.Fprintln(stderr, formatArgs(args))
	return runtime.Undefined, nil
}
