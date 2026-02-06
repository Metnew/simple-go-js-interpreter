package builtins

import (
	"github.com/example/jsgo/runtime"
)

func RegisterAll(env *runtime.Environment, globalObj *runtime.Object) {
	// 1. Object (foundational - other prototypes derive from it)
	objectCtor, objProto := createObjectConstructor()
	env.Declare("Object", "var", runtime.NewObject(objectCtor))

	// 2. Function
	functionCtor, _ := createFunctionConstructor(objProto)
	env.Declare("Function", "var", runtime.NewObject(functionCtor))

	// 3. Array
	arrayCtor, _ := createArrayConstructor(objProto)
	env.Declare("Array", "var", runtime.NewObject(arrayCtor))

	// 4. String
	stringCtor, _ := createStringConstructor(objProto)
	env.Declare("String", "var", runtime.NewObject(stringCtor))

	// 5. Number
	numberCtor, _ := createNumberConstructor(objProto)
	env.Declare("Number", "var", runtime.NewObject(numberCtor))

	// 6. Boolean
	booleanCtor, _ := createBooleanConstructor(objProto)
	env.Declare("Boolean", "var", runtime.NewObject(booleanCtor))

	// 7. Symbol
	symbolCtor := createSymbolConstructor(objProto)
	env.Declare("Symbol", "var", runtime.NewObject(symbolCtor))

	// 8. Error types
	errorCtor := createErrorConstructor(objProto)
	env.Declare("Error", "var", runtime.NewObject(errorCtor))

	typeErrorCtor := createErrorSubtype("TypeError", objProto, ErrorPrototype)
	env.Declare("TypeError", "var", runtime.NewObject(typeErrorCtor))

	refErrorCtor := createErrorSubtype("ReferenceError", objProto, ErrorPrototype)
	env.Declare("ReferenceError", "var", runtime.NewObject(refErrorCtor))

	syntaxErrorCtor := createErrorSubtype("SyntaxError", objProto, ErrorPrototype)
	env.Declare("SyntaxError", "var", runtime.NewObject(syntaxErrorCtor))

	rangeErrorCtor := createErrorSubtype("RangeError", objProto, ErrorPrototype)
	env.Declare("RangeError", "var", runtime.NewObject(rangeErrorCtor))

	uriErrorCtor := createErrorSubtype("URIError", objProto, ErrorPrototype)
	env.Declare("URIError", "var", runtime.NewObject(uriErrorCtor))

	evalErrorCtor := createErrorSubtype("EvalError", objProto, ErrorPrototype)
	env.Declare("EvalError", "var", runtime.NewObject(evalErrorCtor))

	// 9. RegExp
	regexpCtor, _ := createRegExpConstructor(objProto)
	env.Declare("RegExp", "var", runtime.NewObject(regexpCtor))

	// 10. Map, Set, WeakMap, WeakSet
	mapCtor, _ := createMapConstructor(objProto)
	env.Declare("Map", "var", runtime.NewObject(mapCtor))

	setCtor, _ := createSetConstructor(objProto)
	env.Declare("Set", "var", runtime.NewObject(setCtor))

	weakMapCtor := createWeakMapConstructor(objProto)
	env.Declare("WeakMap", "var", runtime.NewObject(weakMapCtor))

	weakSetCtor := createWeakSetConstructor(objProto)
	env.Declare("WeakSet", "var", runtime.NewObject(weakSetCtor))

	// 11. Promise
	promiseCtor, _ := createPromiseConstructor(objProto)
	env.Declare("Promise", "var", runtime.NewObject(promiseCtor))

	// 12. Proxy and Reflect
	proxyCtor := createProxyConstructor(objProto)
	env.Declare("Proxy", "var", runtime.NewObject(proxyCtor))

	reflectObj := createReflectObject(objProto)
	env.Declare("Reflect", "var", runtime.NewObject(reflectObj))

	// 13. Math
	mathObj := createMathObject(objProto)
	env.Declare("Math", "var", runtime.NewObject(mathObj))

	// 14. JSON
	jsonObj := createJSONObject(objProto)
	env.Declare("JSON", "var", runtime.NewObject(jsonObj))

	// 15. Console
	consoleObj := createConsoleObject(objProto)
	env.Declare("console", "var", runtime.NewObject(consoleObj))

	// 16. Global functions (parseInt, parseFloat, isNaN, etc.)
	registerGlobalFunctions(env)

	// 17. Set up global object properties if provided
	if globalObj != nil {
		globalObj.Prototype = objProto
	}
}
