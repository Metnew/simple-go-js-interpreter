# jsgo

A JavaScript engine written in Go, implementing ES2015+ semantics with a tree-walking interpreter.

**100% Test262 pass rate** on supported features (925/925 tests passing).

## Build

```bash
make build
```

Or directly:

```bash
go build -o jsgo ./cmd/jsgo/
```

## Usage

Run a JavaScript file:

```bash
./jsgo script.js
```

Evaluate inline code:

```bash
./jsgo -e "console.log('hello')"
```

Dump the AST as JSON:

```bash
./jsgo -ast script.js
```

## Architecture

```
source code
    |
    v
 [Lexer]     lexer/       - Tokenization, UTF-8/UTF-16 handling
    |
    v
 [Parser]    parser/       - Recursive descent + Pratt expression parsing
    |
    v
  [AST]      ast/          - Abstract syntax tree node types
    |
    v
[Interpreter] interpreter/ - Tree-walking evaluator
    |
    v
 [Runtime]   runtime/      - Values, objects, environments, prototype chains
    |
    v
 [Builtins]  builtins/     - Standard library (Object, Array, String, RegExp, etc.)
```

### Key packages

| Package | Description |
|---------|-------------|
| `runtime/` | Core value types (`Value`, `Object`, `Environment`, `Symbol`), prototype chain, property descriptors |
| `interpreter/` | Statement/expression evaluation, hoisting, `eval()`, scope chain management |
| `builtins/` | Built-in constructors and prototype methods for all standard objects |
| `lexer/` | Tokenizer with full Unicode support, template literals, regex literal detection |
| `parser/` | ES2015+ parser: destructuring, arrow functions, classes, for-of, spread, computed properties |
| `ast/` | AST node definitions |
| `token/` | Token type definitions |
| `testrunner/` | Test262 conformance test runner |

## Supported Features

### Language

- Variable declarations (`var`, `let`, `const`) with proper hoisting and TDZ
- Functions (declarations, expressions, arrow functions, default/rest parameters)
- Classes (constructors, methods, static, getters/setters, `extends`, `super`)
- Destructuring (arrays, objects, nested, defaults, rest elements)
- Spread syntax (calls, arrays, objects)
- Template literals and tagged templates
- `for...of`, `for...in` loops
- `try`/`catch`/`finally` with optional catch binding
- Computed property names
- Shorthand methods and properties
- Symbols and well-known symbols (`Symbol.iterator`, `Symbol.toPrimitive`, `Symbol.hasInstance`, `Symbol.toStringTag`, `Symbol.match`, `Symbol.split`, `Symbol.search`, `Symbol.replace`, `Symbol.species`)
- Iterators and `Symbol.iterator` protocol
- `typeof`, `instanceof`, `in` operators
- Labeled statements, `break`/`continue` with labels
- `eval()` (direct and indirect) with proper scoping
- Strict mode
- Annex B compatibility (HTML comments, block-scoped functions, legacy Date/RegExp methods, octal escapes)

### Built-in Objects

- **Object**: `keys`, `values`, `entries`, `assign`, `create`, `defineProperty`, `defineProperties`, `getOwnPropertyDescriptor`, `getOwnPropertyNames`, `getPrototypeOf`, `setPrototypeOf`, `freeze`, `seal`, `is`, `preventExtensions`
- **Array**: `isArray`, `from`, `of`, `push`, `pop`, `shift`, `unshift`, `slice`, `splice`, `concat`, `join`, `reverse`, `sort`, `indexOf`, `lastIndexOf`, `includes`, `find`, `findIndex`, `every`, `some`, `filter`, `map`, `reduce`, `reduceRight`, `forEach`, `fill`, `copyWithin`, `flat`, `flatMap`, `keys`, `values`, `entries`
- **String**: `charAt`, `charCodeAt`, `codePointAt`, `includes`, `indexOf`, `lastIndexOf`, `startsWith`, `endsWith`, `slice`, `substring`, `trim`, `trimStart`, `trimEnd`, `padStart`, `padEnd`, `repeat`, `replace`, `split`, `match`, `search`, `toLowerCase`, `toUpperCase`, `concat`, `normalize`, `fromCharCode`, `fromCodePoint`, `raw`
- **Number**: `isFinite`, `isInteger`, `isNaN`, `isSafeInteger`, `parseInt`, `parseFloat`, `toFixed`, `toPrecision`, `toExponential`
- **Boolean**, **Math**, **Date**, **RegExp**, **Error** (TypeError, RangeError, SyntaxError, ReferenceError, URIError, EvalError)
- **JSON**: `parse`, `stringify`
- **Map**, **Set**, **WeakMap**, **WeakSet**
- **Promise** (basic)
- **Symbol**: `for`, `keyFor`, well-known symbols
- **Function**: `call`, `apply`, `bind`, `toString`
- **Proxy** (basic)
- Global functions: `parseInt`, `parseFloat`, `isNaN`, `isFinite`, `encodeURI`, `decodeURI`, `encodeURIComponent`, `decodeURIComponent`, `escape`, `unescape`, `eval`

### Not Yet Implemented

- Generators and `yield`
- `async`/`await`
- ES modules (`import`/`export`)
- `SharedArrayBuffer`, `Atomics`
- `WeakRef`, `FinalizationRegistry`
- `TypedArray`, `ArrayBuffer`, `DataView`
- `Intl` (internationalization)
- `Temporal`
- `Reflect.construct`
- Regexp lookbehind assertions, named groups, Unicode property escapes

## Test262 Conformance

The engine is validated against [Test262](https://github.com/nicams/test262), the official ECMAScript conformance test suite.

```bash
# Run full suite
make test262

# Run with limit
./test262runner -dir test262 -limit 100 -v

# Run filtered subset
./test262runner -dir test262 -filter "built-ins/Array" -v
```

Current results:

```
Total:   1000
Passed:  925
Failed:  0
Skipped: 75
Pass rate: 100.0% (925/925 excluding skipped)
```

The 75 skipped tests require features listed under "Not Yet Implemented" above.

## Unit Tests

```bash
go test ./...
```

## Project Structure

```
.
├── cmd/
│   ├── jsgo/            # Main CLI entry point
│   └── test262runner/   # Test262 runner CLI
├── ast/                 # AST node types
├── builtins/            # Built-in objects and methods
├── interpreter/         # Tree-walking interpreter
├── lexer/               # Tokenizer
├── parser/              # Recursive descent parser
├── runtime/             # Core types (Value, Object, Environment)
├── testrunner/          # Test262 harness
├── token/               # Token definitions
├── test262/             # Test262 test suite (submodule)
└── Makefile
```
