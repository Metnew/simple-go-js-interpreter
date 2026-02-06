       Security Review: jsgo builtins package

       1. JSON.parse injection (/Users/v6r/v/c-compiler/builtins/json.go)

       1a. MEDIUM -- Unbounded recursion/allocation via Go's encoding/json

       File: /Users/v6r/v/c-compiler/builtins/json.go:24
       if err := json.Unmarshal([]byte(text), &raw); err != nil {
       json.Unmarshal into interface{} will happily allocate arbitrarily deep nested structures. A payload like [[[[[[...]]]]]] (thousands of nesting levels) will cause Go's
       json.Unmarshal to recurse deeply on the Go call stack and may OOM or hit Go's goroutine stack limit. There is no depth limit enforced.

       Additionally, goToJSValue at line 37 is recursive with no depth guard:
       func goToJSValue(v interface{}) *runtime.Value {
           ...
           case []interface{}:
               data := make([]*runtime.Value, len(val))
               for i, item := range val {
                   data[i] = goToJSValue(item)     // <-- unbounded recursion
       A deeply nested JSON array/object will stack overflow here.

       1b. MEDIUM -- Unbounded allocation via large JSON arrays

       File: /Users/v6r/v/c-compiler/builtins/json.go:49
       data := make([]*runtime.Value, len(val))
       A JSON payload like [0,0,0,...,0] with millions of elements parses into a Go []interface{} and then allocates a correspondingly massive []*runtime.Value slice. No limit
       on the number of elements.

       1c. LOW -- Reviver function error swallowed

       File: /Users/v6r/v/c-compiler/builtins/json.go:88-92
       result, err := reviver(runtime.Undefined, []*runtime.Value{key, val})
       if err != nil {
           return val  // error is silently discarded
       }
       If the reviver function throws, the error is swallowed and the original value is returned. Per the spec, the reviver's error should propagate.

       1d. LOW -- JSON.stringify has no cycle detection

       File: /Users/v6r/v/c-compiler/builtins/json.go:133
       stringifyValue is recursive and has no cycle detection. A circular object reference (obj.self = obj) will cause infinite recursion and a stack overflow:
       func stringifyValue(val *runtime.Value, ...) string {
           ...
           return stringifyObject(val.Object, ...)
               -> stringifyValue(v, ...)  // recurse into property values

       1e. LOW -- toJSON can trigger arbitrary code, no recursion guard

       File: /Users/v6r/v/c-compiler/builtins/json.go:138-147
       toJSON := val.Object.Get("toJSON")
       if toJSON != runtime.Undefined {
           fn := getCallable(toJSON)
           if fn != nil {
               result, err := fn(val, nil)
       A malicious toJSON method that returns another object with a toJSON method can cause repeated re-entry through stringifyValue, though this alone is bounded by the natural
        recursion unless it creates cycles.

       ---
       2. Proxy/Reflect (/Users/v6r/v/c-compiler/builtins/proxy.go)

       2a. HIGH -- Proxy traps are never invoked; proxy is inert

       File: /Users/v6r/v/c-compiler/builtins/proxy.go:21-30
       The proxy constructor stores target and handler in Internal slots, but the runtime's Object.Get() and Object.Set() methods (in
       /Users/v6r/v/c-compiler/runtime/value.go:252-298) have zero awareness of ObjTypeProxy. They never check if an object is a proxy and never invoke handler traps. The proxy
       is just a plain object that stores references.

       This means:
       - Handler traps (get, set, has, deleteProperty, etc.) are completely ignored -- accessing a property on a proxy goes through the normal prototype chain, not through the
       handler.
       - This is a correctness bug, not a security vulnerability per se, but it means anyone using Proxy expecting trap enforcement (e.g., access control, validation) gets zero
       protection.

       2b. LOW -- No revocation support

       There is no Proxy.revocable() implementation. Once a proxy is created, there is no way to revoke it.

       2c. LOW -- Reflect.deleteProperty bypasses non-configurable check

       File: /Users/v6r/v/c-compiler/builtins/proxy.go:77-88
       func reflectDeleteProperty(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
           ...
           _, ok := target.Properties[key]
           if ok {
               delete(target.Properties, key)   // <-- deletes regardless of configurable flag
           }
       This directly deletes from the Properties map without checking prop.Configurable. Non-configurable properties can be deleted via Reflect.deleteProperty().

       ---
       3. Object.defineProperty abuse (/Users/v6r/v/c-compiler/builtins/object.go)

       3a. HIGH -- Object.freeze/seal don't prevent new property addition

       File: /Users/v6r/v/c-compiler/builtins/object.go:298-314 and /Users/v6r/v/c-compiler/runtime/value.go:266-297

       objectFreeze sets an Internal["frozen"] flag and marks all existing properties as Configurable: false, Writable: false, but Object.Set() in the runtime never checks the
       frozen or sealed internal flag. At line 286 of value.go:
       o.Properties[name] = &Property{
           Value: val, Writable: true, Enumerable: true, Configurable: true,
       }
       When a property does not already exist, Set() creates a new one unconditionally. This means:
       - New properties can be added to frozen/sealed objects.
       - Object.defineProperty can add new properties to frozen/sealed objects -- validateDefineOwnProperty only checks existing properties (line 590: if !exists { return nil
       }).

       3b. MEDIUM -- Getter/setter infinite loops are possible

       There is no re-entrancy guard in Object.Get() (/Users/v6r/v/c-compiler/runtime/value.go:252-264):
       if prop.IsAccessor && prop.Getter != nil {
           val, _ := prop.Getter.Object.Callable(NewObject(o), nil)
           return val
       }
       A getter that accesses the same property on the same object will recurse infinitely:
       Object.defineProperty(obj, 'x', { get() { return this.x; } });
       obj.x; // infinite recursion -> stack overflow
       Same for setters at line 270.

       3c. MEDIUM -- Object.setPrototypeOf allows prototype cycles

       File: /Users/v6r/v/c-compiler/builtins/object.go:282-296
       func objectSetPrototypeOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
           ...
           obj.Prototype = proto.Object   // <-- no cycle check
       There is no check that proto does not already have obj in its own prototype chain. This allows:
       let a = {}, b = {};
       Object.setPrototypeOf(a, b);
       Object.setPrototypeOf(b, a);  // creates a cycle
       a.anything; // infinite loop in Object.Get() walking the prototype chain
       The Get() method at /Users/v6r/v/c-compiler/runtime/value.go:260-263 walks the prototype chain with no cycle detection:
       if o.Prototype != nil {
           return o.Prototype.Get(name)  // unbounded recursion if cycle exists
       }
       Similarly, HasProperty at line 326 and isPrototypeOf at line 116 of object.go will loop infinitely.

       3d. LOW -- validateDefineOwnProperty has a logic gap at step 9

       File: /Users/v6r/v/c-compiler/builtins/object.go:616-621
       if desc.IsAccessor != current.IsAccessor {
           if desc.IsAccessor || (desc.HasValue || desc.HasWritable) {
               return fmt.Errorf("TypeError: Cannot redefine property: %s", name)
           }
       }
       When desc.IsAccessor == false and current.IsAccessor == true, but desc has neither HasValue nor HasWritable, the check falls through and allows the conversion from
       accessor to data descriptor on a non-configurable property. This should always reject changing descriptor type on non-configurable.

       ---
       4. Array methods (/Users/v6r/v/c-compiler/builtins/array.go)

       4a. HIGH -- Array(n) has no upper bound on length

       File: /Users/v6r/v/c-compiler/builtins/array.go:83-93
       if len(args) == 1 && args[0].Type == runtime.TypeNumber {
           n := int(args[0].Number)
           if n < 0 {
               return nil, fmt.Errorf("RangeError: Invalid array length")
           }
           data := make([]*runtime.Value, n)
       - Only checks n < 0. An n of 2^31-1 (or even larger float64 values that fit in int) will attempt to allocate a multi-gigabyte slice, OOMing the process.
       - Also, int(args[0].Number) where args[0].Number is NaN or Infinity converts to 0 or architecture-dependent values respectively in Go -- it does not produce a RangeError.

       4b. MEDIUM -- Integer truncation in float64-to-int conversion

       File: /Users/v6r/v/c-compiler/builtins/array.go:84
       n := int(args[0].Number)
       float64 to int conversion in Go is implementation-defined for values outside the int range. For example, Array(4294967296) (2^32) will wrap on 32-bit systems. On 64-bit,
       Array(1e18) converts to a huge int that will OOM on allocation.

       4c. MEDIUM -- flattenArray unbounded recursion

       File: /Users/v6r/v/c-compiler/builtins/array.go:827-837
       func flattenArray(data []*runtime.Value, depth int) []*runtime.Value {
           for _, v := range data {
               if depth > 0 && v.Type == runtime.TypeObject && ... {
                   result = append(result, flattenArray(v.Object.ArrayData, depth-1)...)
       While depth is decremented, flat(Infinity) converts Infinity to int:
       depth = int(args[0].Number)   // line 785
       In Go, int(math.Inf(1)) is implementation-defined but on amd64 it produces -9223372036854775808 (math.MinInt64), which is negative, so the depth > 0 check would actually
       prevent recursion. However, int(1e18) produces a valid huge positive int allowing effective infinite recursion on deeply nested arrays, consuming stack.

       4d. LOW -- arrayConcat has no length limit

       File: /Users/v6r/v/c-compiler/builtins/array.go:280-295
       result = append(result, a.Object.ArrayData...)
       Concatenating many large arrays has no limit, allowing unbounded memory growth.

       4e. LOW -- arrayCopyWithin out-of-bounds potential

       File: /Users/v6r/v/c-compiler/builtins/array.go:681
       copy(temp, obj.ArrayData[start:start+count])
       If start + count > len(obj.ArrayData), this will panic with an index out of range. The bounds checking above does not properly clamp start + count to length.

       ---
       5. Promise (/Users/v6r/v/c-compiler/builtins/promise.go)

       5a. MEDIUM -- Synchronous execution model means no microtask queue

       The entire Promise implementation is synchronous. resolvePromise and rejectPromise (lines 70-100) immediately invoke all pending handlers in a loop. There is no microtask
        queue, no event loop integration.

       This means .then() handlers on an already-resolved promise execute synchronously inline (line 159-160):
       case promiseFulfilled:
           handleFulfill(pd.result)  // runs inline, right now
       This breaks the JS spec guarantee that .then() callbacks are always asynchronous, which can cause unexpected reentrancy issues in user code.

       5b. MEDIUM -- Unbounded chain growth with no cycle protection

       File: /Users/v6r/v/c-compiler/builtins/promise.go:123-176
       Each .then() creates a new promise and attaches handlers to the parent. In a loop like:
       let p = Promise.resolve(1);
       for (let i = 0; i < 1000000; i++) p = p.then(x => x);
       This builds a chain of 1M promise objects. Since resolution is synchronous, this also means 1M nested function calls when the chain resolves.

       5c. LOW -- Promise.resolve does not unwrap thenables

       File: /Users/v6r/v/c-compiler/builtins/promise.go:199-207
       func promiseResolve(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
           val := argAt(args, 0)
           if val.Type == runtime.TypeObject && val.Object != nil && val.Object.OType == runtime.ObjTypePromise {
               return val, nil
           }
           obj, pd := newPromiseObject()
           resolvePromise(pd, val)
       Only native Promise objects are recognized. Thenables (objects with a .then() method) are not unwrapped. This is a spec deviation.

       5d. LOW -- Unhandled rejection silently dropped

       File: /Users/v6r/v/c-compiler/builtins/promise.go:86-100
       When a promise is rejected and has no onReject handlers, the rejection value is stored but nothing signals an unhandled rejection. This can mask bugs.

       ---
       6. Map/Set/WeakMap/WeakSet (/Users/v6r/v/c-compiler/builtins/map_set.go)

       6a. HIGH -- WeakMap/WeakSet are NOT weak (memory leak)

       File: /Users/v6r/v/c-compiler/builtins/map_set.go:420-426
       func getWeakMapStore(obj *runtime.Object) map[*runtime.Object]*runtime.Value {
           ...
           store, _ := obj.Internal["store"].(map[*runtime.Object]*runtime.Value)
           return store
       }
       The WeakMap uses a map[*runtime.Object]*runtime.Value -- a regular Go map with pointer keys. Go's garbage collector cannot collect entries in this map even if the key
       object is no longer referenced elsewhere. The Go map holds a strong reference to the key pointer, preventing GC.

       This is the exact opposite of WeakMap semantics. Every entry added to a WeakMap will be retained forever (until explicitly deleted), causing memory leaks. Same issue for
       WeakSet at line 508-514.

       6b. MEDIUM -- Map/Set use O(n) linear scan for lookup

       File: /Users/v6r/v/c-compiler/builtins/map_set.go:61-68
       func findMapEntry(entries []*mapEntry, key *runtime.Value) int {
           for i, e := range entries {
               if sameValueZero(e.key, key) {
                   return i
               }
           }
           return -1
       }
       Map operations (get, set, has, delete) all use linear search. For a Map with N entries, every operation is O(N). While not a hash collision DoS (there is no hash function
        to collide), it effectively gives an attacker a quadratic time attack: inserting N keys takes O(N^2) time. A Map with 100k entries will be extremely slow.

       Same applies to Set's findSetItem at line 268-275.

       6c. LOW -- Map/Set delete is O(n) slice manipulation

       File: /Users/v6r/v/c-compiler/builtins/map_set.go:137
       entries = append(entries[:idx], entries[idx+1:]...)
       Deletion copies the entire slice tail, making it O(n) for each delete.

       ---
       7. Error objects (/Users/v6r/v/c-compiler/builtins/error.go)

       7a. MEDIUM -- Error stack traces do not capture actual Go call stack

       File: /Users/v6r/v/c-compiler/builtins/error.go:62
       obj.Set("stack", runtime.NewString(fmt.Sprintf("%s: %s", name, msg)))
       The stack property is just "ErrorType: message" -- it does not contain any Go stack trace information, file paths, or memory addresses. This is actually good from a
       security perspective -- no Go internal state leaks through error objects.

       7b. LOW -- Symbol Key() leaks Go memory addresses

       File: /Users/v6r/v/c-compiler/runtime/value.go:167
       func (s *Symbol) Key() string {
           return fmt.Sprintf("@@sym(%s)@%p", s.Description, s)
       }
       While not in error.go, the %p format verb embeds the Go pointer address of the Symbol object into the property key string. If JS code can observe these keys (e.g., via
       Object.getOwnPropertyNames on an object with symbol properties), it can learn Go heap addresses. This could be useful for heap layout inference in an exploit chain.

       ---
