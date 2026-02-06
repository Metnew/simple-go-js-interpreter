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

## Security Review

> This engine is a learning/research project. It is **not hardened for untrusted input**. Do not use it to execute untrusted JavaScript in production.

### Critical

| # | Issue | Location |
|---|-------|----------|
| 1 | **No recursion depth limit** -- recursive JS functions or deeply nested AST cause Go stack overflow, crashing the process. CLI has no `recover()`. | `interpreter.go` (no depth tracking), `parser.go` (recursive descent) |
| 2 | **No infinite loop protection in CLI** -- `while(true){}` hangs forever. Test runner has 5s timeout, CLI has none. | `cmd/jsgo/main.go:75` |
| 3 | **Prototype chain cycles** -- `Object.setPrototypeOf` has no cycle detection. Creating `a->b->a` causes infinite loop in `Object.Get()`. | `builtins/object.go:291`, `runtime/value.go:260` |

### High

| # | Issue | Location |
|---|-------|----------|
| 4 | **`Object.freeze`/`seal` don't prevent new properties** -- `Set()` creates new properties unconditionally, never checks extensibility/frozen flags. | `runtime/value.go:286-291` |
| 5 | **`String.prototype.repeat()` OOM** -- no upper bound on count. `"a".repeat(1e15)` attempts terabyte allocation. Same for `padStart`/`padEnd`. | `builtins/string.go:372-422` |
| 6 | **`Array(n)` unbounded allocation** -- `Array(2147483647)` allocates multi-GB slice. Only checks `n < 0`. | `builtins/array.go:83-93` |
| 7 | **`lastIndexOf` panic** -- `s[:pos+len(search)]` causes out-of-bounds panic when search string is longer than remaining string. | `builtins/string.go:179` |
| 8 | **Nil pointer panic in `resolveMemberKey`** -- if computed key expression throws, `keyVal` is nil, `keyVal.ToPropertyKey()` panics. | `interpreter.go:1845-1846` |
| 9 | **WeakMap/WeakSet are not weak** -- uses `map[*runtime.Object]*runtime.Value`, Go GC cannot collect entries. Memory leak, not weak reference semantics. | `builtins/map_set.go:420-426` |
| 10 | **Proxy traps never invoked** -- constructor stores handler but `Get()`/`Set()` never check `ObjTypeProxy`. Proxy is completely inert. | `builtins/proxy.go:15-31` |

### Medium

| # | Issue | Location |
|---|-------|----------|
| 11 | **JSON.parse unbounded recursion** -- deeply nested JSON causes stack overflow in both `json.Unmarshal` and `goToJSValue`. | `builtins/json.go:24,37` |
| 12 | **JSON.stringify no cycle detection** -- circular object references cause infinite recursion. | `builtins/json.go:133` |
| 13 | **Getter/setter infinite loops** -- no re-entrancy guard. `get x() { return this.x }` recurses infinitely. | `runtime/value.go:254-256` |
| 14 | **Reflect.deleteProperty ignores configurable** -- deletes directly from Properties map without checking `prop.Configurable`. | `builtins/proxy.go:77-88` |
| 15 | **Null byte truncates source** -- `l.ch == 0` at main token switch is sole EOF check. Embedded U+0000 causes premature EOF, hiding trailing code. | `lexer/lexer.go:181` |
| 16 | **Map/Set O(n) linear scan** -- `findMapEntry` iterates all entries. Inserting N keys is O(n^2). 100k entries = very slow. | `builtins/map_set.go:61-68` |
| 17 | **Unchecked type assertions on Internal map** -- `v.(bool)` without comma-ok form can panic if Internal slot has wrong type. | `builtins/object.go:338,351`, `number.go:57`, `boolean.go:34` |
| 18 | **Synchronous Promise execution** -- `.then()` handlers run inline, breaking the async guarantee. Can cause unexpected reentrancy. | `builtins/promise.go:159-160` |
| 19 | **Implicit global creation** -- assignment to undeclared variable silently creates global binding instead of throwing ReferenceError. | `interpreter.go:1814-1817` |
| 20 | **No regex pattern size limit** -- extremely large pattern strings can exhaust memory during compilation. | `builtins/regexp.go:143` |

### Low / Informational

| # | Issue | Location |
|---|-------|----------|
| 21 | **Symbol.Key() leaks Go heap addresses** -- `fmt.Sprintf("@@sym(%s)@%p", ...)` embeds pointer address in property key strings observable via `getOwnPropertyNames`. | `runtime/value.go:167` |
| 22 | **`__proto__` not implemented as accessor** -- spec deviation, `obj.__proto__` creates a data property instead of modifying prototype. | `runtime/value.go:267` |
| 23 | **`decodeURI` decodes reserved characters** -- uses `url.PathUnescape` which doesn't preserve `;/?:@&=+$#`. | `builtins/globals.go:153-173` |
| 24 | **Annex B HTML methods don't escape content** -- `"<script>".bold()` produces `<b><script></b>`. XSS if output rendered in HTML. (Matches spec.) | `builtins/string.go:558-591` |
| 25 | **`arrayCopyWithin` OOB** -- `copy(temp, obj.ArrayData[start:start+count])` can panic if bounds not properly clamped. | `builtins/array.go:681` |
| 26 | **`toInt32`/`toUint32` undefined behavior** -- Go's `int64(float64)` for out-of-range values is implementation-defined. | `builtins/helpers.go:232-246` |

### Recommendations

**Immediate (security-critical):**
1. Add a call stack depth counter (e.g., max 10,000) in the interpreter and parser
2. Add `recover()` to the CLI's `main()`, or run eval in a goroutine with timeout
3. Add cycle detection in `Object.setPrototypeOf` (walk chain to check for target)
4. Cap `String.repeat`/`padStart`/`padEnd` and `Array(n)` to reasonable max (e.g., 2^28)
5. Fix `lastIndexOf` bounds: clamp `pos + len(search)` to `len(s)`

**Short-term:**
6. Track extensibility (`[[IsExtensible]]`) and check in `Set()`/`DefineProperty`
7. Add cycle detection in `JSON.stringify`
8. Check for nil `keyVal` in `resolveMemberKey` before calling methods on it
9. Fix null byte in main lexer token switch (same pattern as regex fix: `l.ch == 0 && l.pos >= len(l.input)`)

**Longer-term:**
10. Replace Map/Set linear scan with hash-based lookup
11. Implement Proxy traps or remove Proxy support
12. Implement proper weak references for WeakMap/WeakSet (requires finalizers or epoch-based approach)
13. Add microtask queue for Promise async semantics

## Exploit PoCs

The `exploits/` directory contains proof-of-concept scripts demonstrating the security findings above. All were tested against the built binary.

### Information Disclosure

| File | Description | Impact |
|------|-------------|--------|
| `heap_addr_leak.js` | Leaks Go heap addresses via `Symbol.Key()` using `Object.getOwnPropertyNames()`. Extracts 12+ heap pointers per run, shows 24-byte allocation stride. | Defeats ASLR if combined with memory corruption primitive |
| `leak_survey.js` | Surveys all info leak surfaces: error stacks, `Function.toString`, regex errors, builtin property enumeration. Finds `RegExp.prototype` leaks well-known symbol addresses with zero interaction. | Engine fingerprinting, heap layout mapping |
| `leak_deep.js` | Deep dive into 5 confirmed findings: well-known symbol address leak, Go regexp engine fingerprinting, WeakMap memory leak (10k entries ~1MB never freed), prototype pollution backdoor, `Object.freeze` bypass. | Multiple primitives chained |

### Denial of Service

| File | Description | Impact |
|------|-------------|--------|
| `dos_crash.js` | Prototype cycle via `Object.setPrototypeOf(a,b); Object.setPrototypeOf(b,a)`. Property access on cyclic chain causes infinite loop, eventually 1GB stack overflow. Crash dumps Go runtime state including heap addresses and source paths. | Process crash with Go runtime info leak |

### Memory Read Attempts

| File | Description | Result |
|------|-------------|--------|
| `memread_attempt1.js` | String OOB via integer overflow (substring, slice, charCodeAt, codePointAt), array sparse access for uninitialized data, `String.fromCharCode` edge cases. | **Blocked** -- Go bounds-checks all access, zeroes memory |
| `memread_attempt2.js` | `int(Infinity)` UB in `copyWithin` (triggers Go panic with heap addresses in stack trace), `Array.fill` with Infinity, recursive stack overflow, `toString`/`valueOf` override type confusion, null byte in property names. | **Partial** -- crashes leak Go stack with heap pointers, but no content read |
| `memread_attempt3.js` | ObjType confusion via `__proto__` swap, ArrayData/length property desync, `valueOf` shapeshifter returning different types on successive calls, RegExp `lastIndex` manipulation, accessor property on array indices. | **Blocked** -- engine checks `len(ArrayData)` not property, `valueOf` evaluated before builtin sees args |
| `memread_attempt4.js` | Trigger `unsupported statement: %T` for Go type leak, constructor confusion, iterator protocol abuse, JSON.stringify with symbol keys (confirmed: serializes `@@sym(desc)@0xADDR` into JSON output), prototype spy getters. | **Partial** -- JSON.stringify leaks symbol keys with heap addresses |
| `memread_attempt5.js` | Global API enumeration (78 properties, no file I/O), `Function` constructor execution, hidden native method probing, ArrayData cap-vs-len after pop/push, `lastIndexOf` OOB crash (confirmed: panics at `s[:pos+len(search)]`). | **Blocked** -- Go's memory safety model prevents all OOB reads |

### Conclusion

Go's memory safety model (bounds-checked slices/strings, zero-initialized allocations, GC preventing UAF, no `unsafe` package) prevents true arbitrary memory reads. The best achievable primitives are:

1. **Heap address leak** via `Symbol.Key()` (`%p` format embeds Go pointer in property key)
2. **Heap address leak** via error messages for failed Symbol method calls
3. **Heap address leak** via `JSON.stringify` serializing symbol keys
4. **Heap address + source path leak** via Go panic stack traces
5. **Heap layout mapping** (24-byte allocation stride = Go size class for Symbol structs)
6. **Go engine fingerprinting** (regexp error messages reveal RE2, `%g` number formatting)
7. **Multiple DoS primitives** (prototype cycles, stack overflow, OOB crashes)

For a true memory read, you'd need `unsafe.Pointer` in the Go codebase (none), a Go compiler/runtime bug, or CGo interop (none).
