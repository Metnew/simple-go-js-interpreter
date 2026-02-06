package testrunner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/example/jsgo/builtins"
	"github.com/example/jsgo/interpreter"
	"github.com/example/jsgo/runtime"
)

type Result int

const (
	Pass Result = iota
	Fail
	Skip
	Error
)

func (r Result) String() string {
	switch r {
	case Pass:
		return "PASS"
	case Fail:
		return "FAIL"
	case Skip:
		return "SKIP"
	case Error:
		return "ERROR"
	}
	return "UNKNOWN"
}

type TestResult struct {
	Path    string
	Result  Result
	Message string
	Elapsed time.Duration
}

type Summary struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
	Errors  int
	Elapsed time.Duration
}

type Config struct {
	Test262Dir string
	Filter     string
	Limit      int
	Verbose    bool
}

// Run discovers and runs Test262 tests, returning results and a summary.
func Run(cfg Config) ([]TestResult, Summary) {
	testDir := filepath.Join(cfg.Test262Dir, "test")
	harnessDir := filepath.Join(cfg.Test262Dir, "harness")

	// Load harness files
	staJS := loadFile(filepath.Join(harnessDir, "sta.js"))
	assertJS := loadFile(filepath.Join(harnessDir, "assert.js"))
	harness := staJS + "\n" + assertJS + "\n"

	// Discover test files
	var testFiles []string
	filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".js") {
			return nil
		}
		// Apply filter
		if cfg.Filter != "" {
			rel, _ := filepath.Rel(testDir, path)
			if !strings.Contains(rel, cfg.Filter) {
				return nil
			}
		}
		testFiles = append(testFiles, path)
		return nil
	})

	// Apply limit
	if cfg.Limit > 0 && len(testFiles) > cfg.Limit {
		testFiles = testFiles[:cfg.Limit]
	}

	start := time.Now()
	var results []TestResult
	var summary Summary
	summary.Total = len(testFiles)

	for _, path := range testFiles {
		rel, _ := filepath.Rel(cfg.Test262Dir, path)
		tr := runSingleTest(path, rel, harness, harnessDir)
		results = append(results, tr)

		switch tr.Result {
		case Pass:
			summary.Passed++
		case Fail:
			summary.Failed++
		case Skip:
			summary.Skipped++
		case Error:
			summary.Errors++
		}

		if cfg.Verbose {
			msg := ""
			if tr.Message != "" {
				msg = " " + tr.Message
			}
			fmt.Printf("%s %s%s\n", tr.Result, rel, msg)
		}
	}

	summary.Elapsed = time.Since(start)
	return results, summary
}

func runSingleTest(path, rel, baseHarness, harnessDir string) TestResult {
	source, err := os.ReadFile(path)
	if err != nil {
		return TestResult{Path: rel, Result: Error, Message: "read error: " + err.Error()}
	}

	meta := parseMetadata(string(source))

	// Skip tests with unsupported features
	for _, feat := range meta.Features {
		if isUnsupportedFeature(feat) {
			return TestResult{Path: rel, Result: Skip, Message: "unsupported feature: " + feat}
		}
	}

	// Skip module tests
	for _, flag := range meta.Flags {
		if flag == "module" {
			return TestResult{Path: rel, Result: Skip, Message: "module test"}
		}
	}

	// Build harness: load any additional includes
	harness := baseHarness
	for _, inc := range meta.Includes {
		incPath := filepath.Join(harnessDir, inc)
		incSrc := loadFile(incPath)
		harness += incSrc + "\n"
	}

	// Build full source
	fullSource := harness + string(source)

	// Run with timeout
	start := time.Now()
	interp := interpreter.New()
	builtins.RegisterAll(interp.GlobalEnv(), nil)
	registerTestNatives(interp)

	resultCh := make(chan evalResult, 1)
	go func() {
		val, err := interp.Eval(fullSource)
		resultCh <- evalResult{val: val, err: err}
	}()

	var evalRes evalResult
	select {
	case evalRes = <-resultCh:
	case <-time.After(5 * time.Second):
		return TestResult{
			Path:    rel,
			Result:  Error,
			Message: "timeout (5s)",
			Elapsed: time.Since(start),
		}
	}

	elapsed := time.Since(start)

	// Handle negative tests (expected to throw)
	if meta.Negative.Phase != "" {
		if evalRes.err != nil {
			// Test expected an error and got one
			return TestResult{Path: rel, Result: Pass, Elapsed: elapsed}
		}
		return TestResult{
			Path:    rel,
			Result:  Fail,
			Message: fmt.Sprintf("expected %s error in %s phase", meta.Negative.Type, meta.Negative.Phase),
			Elapsed: elapsed,
		}
	}

	// Normal test: error means failure
	if evalRes.err != nil {
		return TestResult{
			Path:    rel,
			Result:  Fail,
			Message: evalRes.err.Error(),
			Elapsed: elapsed,
		}
	}

	return TestResult{Path: rel, Result: Pass, Elapsed: elapsed}
}

type evalResult struct {
	val *runtime.Value
	err error
}

func registerTestNatives(interp *interpreter.Interpreter) {
	interp.RegisterNative("print", func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.Undefined, nil
	})
	interp.RegisterNative("_print", func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.Undefined, nil
	})
	interp.RegisterNative("_printErr", func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		return runtime.Undefined, nil
	})

	// Register $262 test harness object with stub methods
	register262Object(interp)
}

func register262Object(interp *interpreter.Interpreter) {
	obj := &runtime.Object{
		OType:      runtime.ObjTypeOrdinary,
		Properties: make(map[string]*runtime.Property),
	}

	stubFn := func(name string) *runtime.Value {
		fn := &runtime.Object{
			OType:      runtime.ObjTypeFunction,
			Properties: make(map[string]*runtime.Property),
			Callable: func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
				return runtime.Undefined, nil
			},
		}
		fn.Set("name", runtime.NewString(name))
		return runtime.NewObject(fn)
	}

	// createRealm - stub that returns a new object with an evalScript method
	createRealmFn := &runtime.Object{
		OType:      runtime.ObjTypeFunction,
		Properties: make(map[string]*runtime.Property),
		Callable: func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
			realm := &runtime.Object{
				OType:      runtime.ObjTypeOrdinary,
				Properties: make(map[string]*runtime.Property),
			}
			evalScriptFn := &runtime.Object{
				OType:      runtime.ObjTypeFunction,
				Properties: make(map[string]*runtime.Property),
				Callable: func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
					return runtime.Undefined, nil
				},
			}
			realm.Set("evalScript", runtime.NewObject(evalScriptFn))
			realm.Set("global", runtime.Undefined)
			return runtime.NewObject(realm), nil
		},
	}
	obj.Set("createRealm", runtime.NewObject(createRealmFn))
	obj.Set("detachArrayBuffer", stubFn("detachArrayBuffer"))
	obj.Set("gc", stubFn("gc"))
	obj.Set("global", runtime.Undefined)
	obj.Set("evalScript", stubFn("evalScript"))

	interp.GlobalEnv().Declare("$262", "var", runtime.NewObject(obj))
}

func loadFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// Metadata from Test262 YAML frontmatter
type TestMetadata struct {
	Description string
	Features    []string
	Flags       []string
	Includes    []string
	Negative    NegativeExpectation
}

type NegativeExpectation struct {
	Phase string // "parse", "resolution", "runtime"
	Type  string // "SyntaxError", "TypeError", etc.
}

func parseMetadata(source string) TestMetadata {
	var meta TestMetadata

	// Find YAML frontmatter between /*--- and ---*/
	startIdx := strings.Index(source, "/*---")
	if startIdx < 0 {
		return meta
	}
	endIdx := strings.Index(source[startIdx:], "---*/")
	if endIdx < 0 {
		return meta
	}

	yaml := source[startIdx+5 : startIdx+endIdx]
	lines := strings.Split(yaml, "\n")

	var currentKey string
	var inList bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for list items
		if strings.HasPrefix(trimmed, "- ") && inList {
			val := strings.TrimPrefix(trimmed, "- ")
			val = strings.TrimSpace(val)
			switch currentKey {
			case "features":
				meta.Features = append(meta.Features, val)
			case "flags":
				meta.Flags = append(meta.Flags, val)
			case "includes":
				meta.Includes = append(meta.Includes, val)
			}
			continue
		}

		// Check for key: value
		if idx := strings.Index(trimmed, ":"); idx > 0 {
			key := strings.TrimSpace(trimmed[:idx])
			value := strings.TrimSpace(trimmed[idx+1:])

			switch key {
			case "description":
				meta.Description = value
				inList = false
				currentKey = ""
			case "features":
				currentKey = "features"
				inList = true
				// Handle inline list [feat1, feat2]
				if strings.HasPrefix(value, "[") {
					meta.Features = parseInlineList(value)
					inList = false
				}
			case "flags":
				currentKey = "flags"
				inList = true
				if strings.HasPrefix(value, "[") {
					meta.Flags = parseInlineList(value)
					inList = false
				}
			case "includes":
				currentKey = "includes"
				inList = true
				if strings.HasPrefix(value, "[") {
					meta.Includes = parseInlineList(value)
					inList = false
				}
			case "phase":
				meta.Negative.Phase = value
			case "type":
				if currentKey == "negative" {
					meta.Negative.Type = value
				}
			case "negative":
				currentKey = "negative"
				inList = false
			default:
				if currentKey != "negative" {
					currentKey = key
					inList = false
				}
			}
		}
	}

	return meta
}

func parseInlineList(s string) []string {
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func isUnsupportedFeature(feat string) bool {
	unsupported := map[string]bool{
		"SharedArrayBuffer":      true,
		"Atomics":                true,
		"WeakRef":                true,
		"FinalizationRegistry":   true,
		"import.meta":            true,
		"dynamic-import":         true,
		"top-level-await":        true,
		"regexp-lookbehind":      true,
		"regexp-named-groups":    true,
		"regexp-unicode-property-escapes": true,
		"Intl":                   true,
		"Temporal":               true,
		"decorators":             true,
		"import-assertions":      true,
		"json-modules":           true,
		"IsHTMLDDA":              true,
	}
	return unsupported[feat]
}
