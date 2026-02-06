# jsgo - JavaScript Engine in Go

## Project Overview

A tree-walking JavaScript interpreter written in Go, targeting ES2015+ compliance. Validated against the Test262 conformance suite with 100% pass rate on supported features (925/925 tests).

## Build & Test

```bash
# Build both binaries
make build

# Run unit tests
go test ./...

# Run Test262 conformance suite (full)
make test262

# Run Test262 with limit
./test262runner -dir test262 -limit 100 -v

# Run filtered tests
./test262runner -dir test262 -filter "built-ins/RegExp" -v
```

## Architecture

Lexer (`lexer/`) -> Parser (`parser/`) -> AST (`ast/`) -> Interpreter (`interpreter/`) -> Runtime (`runtime/`) + Builtins (`builtins/`)

- **runtime/value.go** - Core `Value` and `Object` types, prototype chain, property descriptors
- **runtime/environment.go** - Lexical scoping, scope chain, global object mirroring
- **interpreter/interpreter.go** - Main eval loop, statement/expression handlers, `Eval()` entry point
- **interpreter/hoisting.go** - Var/function hoisting, Annex B block-scoped function declarations
- **builtins/** - One file per built-in: `object.go`, `array.go`, `string.go`, `regexp.go`, `date.go`, `symbol.go`, `error.go`, `map.go`, `set.go`, `json.go`, `promise.go`, `proxy.go`, `function.go`, `globals.go`, `helpers.go`
- **testrunner/runner.go** - Test262 harness, metadata parsing, timeout handling

## Key Implementation Details

### Symbol Property Keys
Always use `ToPropertyKey()` (not `ToString()`) when resolving property names that could be Symbols. Symbol keys use the internal format `@@sym(desc)@<ptr>`.

### Global Object/Environment Duality
The global environment and global object are bidirectionally linked via `SetGlobalObject()`. Var declarations mirror to the global object. `Get()` falls back to global object properties at the root scope.

### Go Regexp Limitations
JS regex features unsupported by Go's `regexp` package are handled in `jsRegexpToGo()`:
- Identity escapes (`\C`, `\E`, etc.) converted to `\xHH`
- `\v` -> `\x0b`, `\cX` -> control chars
- Backreferences replaced with `(?:)` or `[^\w\W]`
- No lookbehind support

### Annex B Compatibility
Block-scoped function declarations are hoisted per Annex B.3.3. The hoisting respects lexical bindings in enclosing blocks, catch parameters, and the `arguments` name.

## Not Yet Implemented
Generators, async/await, ES modules, TypedArrays, SharedArrayBuffer, Intl, Temporal, regexp lookbehind/named groups/Unicode property escapes.
