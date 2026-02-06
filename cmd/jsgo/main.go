package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/example/jsgo/builtins"
	"github.com/example/jsgo/interpreter"
	"github.com/example/jsgo/parser"
	"github.com/example/jsgo/runtime"
)

// consoleShim creates a console object using the registered _print/_printErr natives.
const consoleShim = `var console = {
	log: function() { var i = 0; var s = ""; while (i < arguments.length) { if (i > 0) s = s + " "; s = s + arguments[i]; i++; } _print(s); },
	warn: function() { var i = 0; var s = ""; while (i < arguments.length) { if (i > 0) s = s + " "; s = s + arguments[i]; i++; } _printErr(s); },
	error: function() { var i = 0; var s = ""; while (i < arguments.length) { if (i > 0) s = s + " "; s = s + arguments[i]; i++; } _printErr(s); },
	info: function() { var i = 0; var s = ""; while (i < arguments.length) { if (i > 0) s = s + " "; s = s + arguments[i]; i++; } _print(s); }
};
`

func main() {
	evalCode := flag.String("e", "", "evaluate inline JavaScript code")
	dumpAST := flag.Bool("ast", false, "dump the AST as JSON")
	flag.Parse()

	var source string

	if *evalCode != "" {
		source = *evalCode
	} else if flag.NArg() > 0 {
		filename := flag.Arg(0)
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		source = string(data)
	} else {
		fmt.Fprintf(os.Stderr, "Usage: jsgo [options] <file.js>\n")
		fmt.Fprintf(os.Stderr, "       jsgo -e \"code\"\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// AST dump mode: parse and print JSON
	if *dumpAST {
		p := parser.New(source)
		program, errs := p.ParseProgram()
		if len(errs) > 0 {
			for _, err := range errs {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			os.Exit(1)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(program); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding AST: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create interpreter, register builtins and native print functions
	interp := interpreter.New()
	builtins.RegisterAll(interp.GlobalEnv(), nil)
	registerNatives(interp)

	// Prepend console shim + user source and evaluate
	fullSource := consoleShim + source
	result, err := interp.Eval(fullSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if result != nil && result.Type != runtime.TypeUndefined {
		fmt.Println(result.ToString())
	}
}

func registerNatives(interp *interpreter.Interpreter) {
	interp.RegisterNative("_print", func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		if len(args) > 0 {
			fmt.Println(args[0].ToString())
		} else {
			fmt.Println()
		}
		return runtime.Undefined, nil
	})
	interp.RegisterNative("_printErr", func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		if len(args) > 0 {
			fmt.Fprintln(os.Stderr, args[0].ToString())
		} else {
			fmt.Fprintln(os.Stderr)
		}
		return runtime.Undefined, nil
	})
}

