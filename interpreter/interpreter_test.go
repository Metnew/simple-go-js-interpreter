package interpreter

import (
	"math"
	"strings"
	"testing"

	"github.com/example/jsgo/runtime"
)

func evalExpect(t *testing.T, source string) *runtime.Value {
	t.Helper()
	interp := New()
	val, err := interp.Eval(source)
	if err != nil {
		t.Fatalf("Eval error for %q: %v", source, err)
	}
	return val
}

func evalExpectError(t *testing.T, source string) error {
	t.Helper()
	interp := New()
	_, err := interp.Eval(source)
	if err == nil {
		t.Fatalf("expected error for %q but got none", source)
	}
	return err
}

func expectNumber(t *testing.T, source string, expected float64) {
	t.Helper()
	val := evalExpect(t, source)
	if val.Type != runtime.TypeNumber {
		t.Fatalf("expected number for %q, got %v (type=%v)", source, val, val.Type)
	}
	if math.IsNaN(expected) {
		if !math.IsNaN(val.Number) {
			t.Fatalf("expected NaN for %q, got %v", source, val.Number)
		}
		return
	}
	if val.Number != expected {
		t.Fatalf("expected %v for %q, got %v", expected, source, val.Number)
	}
}

func expectString(t *testing.T, source string, expected string) {
	t.Helper()
	val := evalExpect(t, source)
	if val.Type != runtime.TypeString {
		t.Fatalf("expected string for %q, got type=%v val=%v", source, val.Type, val)
	}
	if val.Str != expected {
		t.Fatalf("expected %q for %q, got %q", expected, source, val.Str)
	}
}

func expectBool(t *testing.T, source string, expected bool) {
	t.Helper()
	val := evalExpect(t, source)
	if val.Type != runtime.TypeBoolean {
		t.Fatalf("expected boolean for %q, got type=%v", source, val.Type)
	}
	if val.Bool != expected {
		t.Fatalf("expected %v for %q, got %v", expected, source, val.Bool)
	}
}

func expectUndefined(t *testing.T, source string) {
	t.Helper()
	val := evalExpect(t, source)
	if val.Type != runtime.TypeUndefined {
		t.Fatalf("expected undefined for %q, got type=%v", source, val.Type)
	}
}

func expectNull(t *testing.T, source string) {
	t.Helper()
	val := evalExpect(t, source)
	if val.Type != runtime.TypeNull {
		t.Fatalf("expected null for %q, got type=%v", source, val.Type)
	}
}

// --- Literals ---

func TestLiterals(t *testing.T) {
	expectNumber(t, "42", 42)
	expectNumber(t, "3.14", 3.14)
	expectString(t, `"hello"`, "hello")
	expectString(t, "'world'", "world")
	expectBool(t, "true", true)
	expectBool(t, "false", false)
	expectNull(t, "null")
	expectUndefined(t, "undefined")
}

// --- Arithmetic ---

func TestArithmetic(t *testing.T) {
	expectNumber(t, "2 + 3", 5)
	expectNumber(t, "10 - 3", 7)
	expectNumber(t, "4 * 5", 20)
	expectNumber(t, "10 / 3", 10.0/3.0)
	expectNumber(t, "10 % 3", 1)
	expectNumber(t, "2 ** 10", 1024)
	expectNumber(t, "-5", -5)
	expectNumber(t, "+true", 1)
}

// --- String concatenation ---

func TestStringConcat(t *testing.T) {
	expectString(t, `"hello" + " " + "world"`, "hello world")
	expectString(t, `"num: " + 42`, "num: 42")
	expectString(t, `1 + "2"`, "12")
}

// --- Comparison operators ---

func TestComparisons(t *testing.T) {
	expectBool(t, "1 < 2", true)
	expectBool(t, "2 > 1", true)
	expectBool(t, "1 <= 1", true)
	expectBool(t, "1 >= 2", false)
	expectBool(t, "1 == 1", true)
	expectBool(t, "1 == '1'", true)
	expectBool(t, "1 === '1'", false)
	expectBool(t, "1 === 1", true)
	expectBool(t, "1 != 2", true)
	expectBool(t, "1 !== '1'", true)
	expectBool(t, "null == undefined", true)
	expectBool(t, "null === undefined", false)
}

// --- Logical operators ---

func TestLogical(t *testing.T) {
	expectNumber(t, "1 && 2", 2)
	expectNumber(t, "0 && 2", 0)
	expectNumber(t, "1 || 2", 1)
	expectNumber(t, "0 || 2", 2)
	expectBool(t, "!true", false)
	expectBool(t, "!false", true)
	expectBool(t, "!0", true)
	expectBool(t, "!1", false)
}

// --- Nullish coalescing ---

func TestNullishCoalescing(t *testing.T) {
	expectNumber(t, "null ?? 42", 42)
	expectNumber(t, "undefined ?? 42", 42)
	expectNumber(t, "0 ?? 42", 0)
	expectString(t, `"" ?? "fallback"`, "")
}

// --- Variables ---

func TestVariables(t *testing.T) {
	expectNumber(t, "var x = 10; x", 10)
	expectNumber(t, "let x = 20; x", 20)
	expectNumber(t, "const x = 30; x", 30)
	expectNumber(t, "var x = 1; x = 2; x", 2)
}

func TestConstAssignment(t *testing.T) {
	err := evalExpectError(t, "const x = 1; x = 2")
	if !strings.Contains(err.Error(), "constant") {
		t.Fatalf("expected constant assignment error, got: %v", err)
	}
}

func TestLetBlockScoping(t *testing.T) {
	expectNumber(t, `
		let x = 1;
		{
			let x = 2;
			x;
		}
	`, 2)
	expectNumber(t, `
		let x = 1;
		{
			let x = 2;
		}
		x;
	`, 1)
}

func TestVarHoisting(t *testing.T) {
	expectUndefined(t, `
		var x;
		x;
	`)
	expectNumber(t, `
		x = 5;
		var x;
		x;
	`, 5)
}

// --- If/Else ---

func TestIfElse(t *testing.T) {
	expectNumber(t, "var x; if (true) { x = 1 } else { x = 2 } x", 1)
	expectNumber(t, "var x; if (false) { x = 1 } else { x = 2 } x", 2)
	expectNumber(t, `
		var x;
		if (false) { x = 1 }
		else if (true) { x = 2 }
		else { x = 3 }
		x
	`, 2)
}

// --- Ternary ---

func TestTernary(t *testing.T) {
	expectNumber(t, "true ? 1 : 2", 1)
	expectNumber(t, "false ? 1 : 2", 2)
	expectString(t, `1 > 0 ? "yes" : "no"`, "yes")
}

// --- While loop ---

func TestWhileLoop(t *testing.T) {
	expectNumber(t, `
		var i = 0;
		var sum = 0;
		while (i < 5) {
			sum = sum + i;
			i = i + 1;
		}
		sum;
	`, 10)
}

// --- Do-while loop ---

func TestDoWhileLoop(t *testing.T) {
	expectNumber(t, `
		var i = 0;
		do {
			i = i + 1;
		} while (i < 5);
		i;
	`, 5)
}

// --- For loop ---

func TestForLoop(t *testing.T) {
	expectNumber(t, `
		var sum = 0;
		for (var i = 0; i < 5; i++) {
			sum = sum + i;
		}
		sum;
	`, 10)
}

// --- Break and Continue ---

func TestBreakContinue(t *testing.T) {
	expectNumber(t, `
		var sum = 0;
		for (var i = 0; i < 10; i++) {
			if (i === 5) break;
			sum = sum + i;
		}
		sum;
	`, 10)

	expectNumber(t, `
		var sum = 0;
		for (var i = 0; i < 5; i++) {
			if (i === 2) continue;
			sum = sum + i;
		}
		sum;
	`, 8)
}

// --- Functions ---

func TestFunctionDeclaration(t *testing.T) {
	expectNumber(t, `
		function add(a, b) { return a + b; }
		add(3, 4);
	`, 7)
}

func TestFunctionExpression(t *testing.T) {
	expectNumber(t, `
		var mul = function(a, b) { return a * b; };
		mul(3, 4);
	`, 12)
}

func TestArrowFunction(t *testing.T) {
	expectNumber(t, `
		var add = (a, b) => a + b;
		add(3, 4);
	`, 7)

	expectNumber(t, `
		var square = (x) => { return x * x; };
		square(5);
	`, 25)
}

func TestClosure(t *testing.T) {
	expectNumber(t, `
		function makeCounter() {
			var count = 0;
			return function() {
				count = count + 1;
				return count;
			};
		}
		var counter = makeCounter();
		counter();
		counter();
		counter();
	`, 3)
}

func TestDefaultParams(t *testing.T) {
	expectNumber(t, `
		function greet(x, y = 10) {
			return x + y;
		}
		greet(5);
	`, 15)

	expectNumber(t, `
		function greet(x, y = 10) {
			return x + y;
		}
		greet(5, 20);
	`, 25)
}

func TestRestParams(t *testing.T) {
	expectNumber(t, `
		function sum(...nums) {
			var total = 0;
			for (var i = 0; i < nums.length; i++) {
				total = total + nums[i];
			}
			return total;
		}
		sum(1, 2, 3, 4, 5);
	`, 15)
}

func TestRecursion(t *testing.T) {
	expectNumber(t, `
		function fib(n) {
			if (n <= 1) return n;
			return fib(n - 1) + fib(n - 2);
		}
		fib(10);
	`, 55)

	expectNumber(t, `
		function fact(n) {
			if (n <= 1) return 1;
			return n * fact(n - 1);
		}
		fact(5);
	`, 120)
}

func TestFunctionHoisting(t *testing.T) {
	expectNumber(t, `
		var result = add(3, 4);
		function add(a, b) { return a + b; }
		result;
	`, 7)
}

// --- Arrow function lexical this ---

func TestArrowLexicalThis(t *testing.T) {
	expectNumber(t, `
		var obj = {
			x: 10,
			getX: function() {
				var inner = () => this.x;
				return inner();
			}
		};
		obj.getX();
	`, 10)
}

// --- Typeof ---

func TestTypeof(t *testing.T) {
	expectString(t, `typeof 42`, "number")
	expectString(t, `typeof "hello"`, "string")
	expectString(t, `typeof true`, "boolean")
	expectString(t, `typeof undefined`, "undefined")
	expectString(t, `typeof null`, "object")
	expectString(t, `typeof nonexistent`, "undefined")
	expectString(t, `typeof function(){}`, "function")
}

// --- Objects ---

func TestObjectLiteral(t *testing.T) {
	expectNumber(t, `
		var obj = { x: 1, y: 2 };
		obj.x + obj.y;
	`, 3)
}

func TestObjectPropertyAccess(t *testing.T) {
	expectNumber(t, `
		var obj = { x: 10 };
		obj["x"];
	`, 10)

	expectNumber(t, `
		var key = "x";
		var obj = { x: 42 };
		obj[key];
	`, 42)
}

func TestObjectMutation(t *testing.T) {
	expectNumber(t, `
		var obj = { x: 1 };
		obj.x = 10;
		obj.x;
	`, 10)

	expectNumber(t, `
		var obj = {};
		obj.y = 42;
		obj.y;
	`, 42)
}

func TestObjectShorthand(t *testing.T) {
	expectNumber(t, `
		var x = 10;
		var obj = { x };
		obj.x;
	`, 10)
}

func TestObjectMethod(t *testing.T) {
	expectNumber(t, `
		var obj = {
			x: 5,
			getX: function() { return this.x; }
		};
		obj.getX();
	`, 5)
}

// --- Arrays ---

func TestArrayLiteral(t *testing.T) {
	expectNumber(t, `
		var arr = [1, 2, 3];
		arr[0];
	`, 1)
	expectNumber(t, `
		var arr = [1, 2, 3];
		arr.length;
	`, 3)
}

func TestArrayMethods(t *testing.T) {
	expectNumber(t, `
		var arr = [1, 2, 3];
		arr.push(4);
		arr.length;
	`, 4)

	expectNumber(t, `
		var arr = [1, 2, 3];
		arr.pop();
	`, 3)

	expectString(t, `
		var arr = [1, 2, 3];
		arr.join("-");
	`, "1-2-3")

	expectNumber(t, `
		var arr = [1, 2, 3];
		arr.indexOf(2);
	`, 1)

	expectBool(t, `
		var arr = [1, 2, 3];
		arr.includes(2);
	`, true)
}

func TestArrayMap(t *testing.T) {
	expectString(t, `
		var arr = [1, 2, 3];
		var doubled = arr.map(function(x) { return x * 2; });
		doubled.join(",");
	`, "2,4,6")
}

func TestArrayFilter(t *testing.T) {
	expectString(t, `
		var arr = [1, 2, 3, 4, 5];
		var evens = arr.filter(function(x) { return x % 2 === 0; });
		evens.join(",");
	`, "2,4")
}

func TestArrayReduce(t *testing.T) {
	expectNumber(t, `
		var arr = [1, 2, 3, 4, 5];
		arr.reduce(function(acc, x) { return acc + x; }, 0);
	`, 15)
}

func TestArrayForEach(t *testing.T) {
	expectNumber(t, `
		var arr = [1, 2, 3];
		var sum = 0;
		arr.forEach(function(x) { sum = sum + x; });
		sum;
	`, 6)
}

func TestArraySpread(t *testing.T) {
	expectNumber(t, `
		var arr1 = [1, 2];
		var arr2 = [3, 4];
		var arr3 = [...arr1, ...arr2];
		arr3.length;
	`, 4)
}

// --- Switch ---

func TestSwitch(t *testing.T) {
	expectNumber(t, `
		var x = 2;
		var result;
		switch (x) {
			case 1: result = 10; break;
			case 2: result = 20; break;
			case 3: result = 30; break;
			default: result = 0;
		}
		result;
	`, 20)
}

func TestSwitchDefault(t *testing.T) {
	expectNumber(t, `
		var x = 99;
		var result;
		switch (x) {
			case 1: result = 10; break;
			default: result = 0;
		}
		result;
	`, 0)
}

func TestSwitchFallThrough(t *testing.T) {
	expectNumber(t, `
		var x = 1;
		var result = 0;
		switch (x) {
			case 1: result = result + 1;
			case 2: result = result + 2; break;
			case 3: result = result + 3;
		}
		result;
	`, 3)
}

// --- Try/Catch/Finally ---

func TestTryCatch(t *testing.T) {
	expectNumber(t, `
		var result;
		try {
			throw 42;
		} catch (e) {
			result = e;
		}
		result;
	`, 42)
}

func TestTryFinally(t *testing.T) {
	expectNumber(t, `
		var x = 0;
		try {
			x = 1;
		} finally {
			x = x + 10;
		}
		x;
	`, 11)
}

func TestTryCatchFinally(t *testing.T) {
	expectNumber(t, `
		var result = 0;
		try {
			throw "error";
		} catch (e) {
			result = 1;
		} finally {
			result = result + 10;
		}
		result;
	`, 11)
}

func TestThrowString(t *testing.T) {
	err := evalExpectError(t, `throw "my error"`)
	if !strings.Contains(err.Error(), "my error") {
		t.Fatalf("expected 'my error', got: %v", err)
	}
}

func TestNestedTryCatch(t *testing.T) {
	expectNumber(t, `
		var result = 0;
		try {
			try {
				throw 1;
			} catch (e) {
				result = e;
				throw 2;
			}
		} catch (e) {
			result = result + e;
		}
		result;
	`, 3)
}

// --- Typeof ---

func TestTypeofOperator(t *testing.T) {
	expectString(t, `typeof 42`, "number")
	expectString(t, `typeof "str"`, "string")
	expectString(t, `typeof true`, "boolean")
	expectString(t, `typeof undefined`, "undefined")
	expectString(t, `typeof null`, "object")
	expectString(t, `typeof {}`, "object")
	expectString(t, `typeof []`, "object")
	expectString(t, `typeof function(){}`, "function")
}

// --- Delete ---

func TestDelete(t *testing.T) {
	expectUndefined(t, `
		var obj = { x: 1, y: 2 };
		delete obj.x;
		obj.x;
	`)
}

// --- Void ---

func TestVoid(t *testing.T) {
	expectUndefined(t, `void 0`)
	expectUndefined(t, `void "hello"`)
}

// --- In operator ---

func TestInOperator(t *testing.T) {
	expectBool(t, `
		var obj = { x: 1, y: 2 };
		"x" in obj;
	`, true)

	expectBool(t, `
		var obj = { x: 1 };
		"z" in obj;
	`, false)
}

// --- Instanceof ---

func TestInstanceof(t *testing.T) {
	expectBool(t, `
		function Foo() {}
		var f = new Foo();
		f instanceof Foo;
	`, true)
}

// --- Bitwise operators ---

func TestBitwiseOps(t *testing.T) {
	expectNumber(t, "5 & 3", 1)
	expectNumber(t, "5 | 3", 7)
	expectNumber(t, "5 ^ 3", 6)
	expectNumber(t, "~5", -6)
	expectNumber(t, "1 << 3", 8)
	expectNumber(t, "8 >> 1", 4)
}

// --- Update expressions ---

func TestUpdateExpressions(t *testing.T) {
	expectNumber(t, `
		var x = 5;
		x++;
		x;
	`, 6)
	expectNumber(t, `
		var x = 5;
		x--;
		x;
	`, 4)
	expectNumber(t, `
		var x = 5;
		++x;
	`, 6)
	expectNumber(t, `
		var x = 5;
		--x;
	`, 4)
}

func TestPostfixReturnValue(t *testing.T) {
	expectNumber(t, `
		var x = 5;
		x++;
	`, 5)
}

// --- Compound assignment ---

func TestCompoundAssignment(t *testing.T) {
	expectNumber(t, "var x = 10; x += 5; x", 15)
	expectNumber(t, "var x = 10; x -= 3; x", 7)
	expectNumber(t, "var x = 4; x *= 3; x", 12)
	expectNumber(t, "var x = 10; x /= 2; x", 5)
	expectNumber(t, "var x = 10; x %= 3; x", 1)
}

// --- Sequence expression ---

func TestSequenceExpression(t *testing.T) {
	expectNumber(t, "(1, 2, 3)", 3)
}

// --- Template literals ---

func TestTemplateLiteral(t *testing.T) {
	expectString(t, "`hello world`", "hello world")
	expectString(t, "var x = 42; `the answer is ${x}`", "the answer is 42")
	expectString(t, "var a = 1; var b = 2; `${a} + ${b} = ${a + b}`", "1 + 2 = 3")
}

// --- For-of ---

func TestForOf(t *testing.T) {
	expectNumber(t, `
		var arr = [10, 20, 30];
		var sum = 0;
		for (var x of arr) {
			sum = sum + x;
		}
		sum;
	`, 60)
}

func TestForOfString(t *testing.T) {
	expectNumber(t, `
		var count = 0;
		for (var ch of "hello") {
			count = count + 1;
		}
		count;
	`, 5)
}

// --- For-in ---

func TestForIn(t *testing.T) {
	expectBool(t, `
		var obj = { a: 1, b: 2, c: 3 };
		var keys = [];
		for (var k in obj) {
			keys.push(k);
		}
		keys.includes("a") && keys.includes("b") && keys.includes("c");
	`, true)
}

// --- Destructuring ---

func TestArrayDestructuring(t *testing.T) {
	expectNumber(t, `
		var [a, b, c] = [1, 2, 3];
		a + b + c;
	`, 6)
}

func TestObjectDestructuring(t *testing.T) {
	expectNumber(t, `
		var { x, y } = { x: 10, y: 20 };
		x + y;
	`, 30)
}

func TestDestructuringDefaults(t *testing.T) {
	expectNumber(t, `
		var [a, b = 5] = [1];
		a + b;
	`, 6)
}

func TestDestructuringRest(t *testing.T) {
	expectNumber(t, `
		var [first, ...rest] = [1, 2, 3, 4];
		rest.length;
	`, 3)
}

// --- Classes ---

func TestClassBasic(t *testing.T) {
	expectNumber(t, `
		class Point {
			constructor(x, y) {
				this.x = x;
				this.y = y;
			}
			sum() {
				return this.x + this.y;
			}
		}
		var p = new Point(3, 4);
		p.sum();
	`, 7)
}

func TestClassInheritance(t *testing.T) {
	expectString(t, `
		class Animal {
			constructor(name) {
				this.name = name;
			}
			speak() {
				return this.name + " makes a sound";
			}
		}
		class Dog extends Animal {
			constructor(name) {
				super(name);
			}
			speak() {
				return this.name + " barks";
			}
		}
		var d = new Dog("Rex");
		d.speak();
	`, "Rex barks")
}

func TestClassStaticMethods(t *testing.T) {
	expectNumber(t, `
		class MathHelper {
			static add(a, b) {
				return a + b;
			}
		}
		MathHelper.add(3, 4);
	`, 7)
}

func TestClassGetterSetter(t *testing.T) {
	expectNumber(t, `
		class Circle {
			constructor(r) {
				this._r = r;
			}
			get radius() {
				return this._r;
			}
			set radius(val) {
				this._r = val;
			}
		}
		var c = new Circle(5);
		c.radius = 10;
		c.radius;
	`, 10)
}

func TestClassInstanceof(t *testing.T) {
	expectBool(t, `
		class Foo {}
		class Bar extends Foo {}
		var b = new Bar();
		b instanceof Foo;
	`, true)
}

// --- New operator ---

func TestNewOperator(t *testing.T) {
	expectNumber(t, `
		function Point(x, y) {
			this.x = x;
			this.y = y;
		}
		var p = new Point(3, 4);
		p.x + p.y;
	`, 7)
}

// --- Native functions ---

func TestRegisterNative(t *testing.T) {
	interp := New()
	interp.RegisterNative("add", func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		if len(args) < 2 {
			return runtime.NewNumber(0), nil
		}
		return runtime.NewNumber(args[0].ToNumber() + args[1].ToNumber()), nil
	})
	val, err := interp.Eval("add(3, 4)")
	if err != nil {
		t.Fatal(err)
	}
	if val.Number != 7 {
		t.Fatalf("expected 7, got %v", val.Number)
	}
}

// --- String methods ---

func TestStringMethods(t *testing.T) {
	expectString(t, `"hello".toUpperCase()`, "HELLO")
	expectString(t, `"HELLO".toLowerCase()`, "hello")
	expectNumber(t, `"hello".indexOf("ll")`, 2)
	expectString(t, `"hello".slice(1, 3)`, "el")
	expectString(t, `"  hello  ".trim()`, "hello")
	expectBool(t, `"hello world".includes("world")`, true)
	expectBool(t, `"hello".startsWith("hel")`, true)
	expectBool(t, `"hello".endsWith("llo")`, true)
	expectString(t, `"ha".repeat(3)`, "hahaha")
	expectString(t, `"hello world".replace("world", "earth")`, "hello earth")
	expectNumber(t, `"hello".length`, 5)
}

func TestStringSplit(t *testing.T) {
	expectNumber(t, `
		var parts = "a,b,c".split(",");
		parts.length;
	`, 3)
}

// --- This binding ---

func TestThisInMethod(t *testing.T) {
	expectNumber(t, `
		var obj = {
			value: 42,
			getValue: function() { return this.value; }
		};
		obj.getValue();
	`, 42)
}

// --- Higher-order functions ---

func TestHigherOrderFunctions(t *testing.T) {
	expectNumber(t, `
		function apply(fn, x) {
			return fn(x);
		}
		function double(n) { return n * 2; }
		apply(double, 5);
	`, 10)
}

// --- IIFE ---

func TestIIFE(t *testing.T) {
	expectNumber(t, `
		var result = (function() { return 42; })();
		result;
	`, 42)
}

// --- Complex scenarios ---

func TestFibonacciWithClosure(t *testing.T) {
	expectNumber(t, `
		function memoize(fn) {
			var cache = {};
			return function(n) {
				var key = "" + n;
				if (cache[key] !== undefined) return cache[key];
				var result = fn(n);
				cache[key] = result;
				return result;
			};
		}
		var fib = memoize(function(n) {
			if (n <= 1) return n;
			return fib(n - 1) + fib(n - 2);
		});
		fib(10);
	`, 55)
}

func TestArraySortWithComparator(t *testing.T) {
	expectString(t, `
		var arr = [3, 1, 4, 1, 5, 9];
		arr.sort(function(a, b) { return a - b; });
		arr.join(",");
	`, "1,1,3,4,5,9")
}

func TestNestedFunctions(t *testing.T) {
	expectNumber(t, `
		function outer(x) {
			function inner(y) {
				return x + y;
			}
			return inner(10);
		}
		outer(5);
	`, 15)
}

func TestComplexControlFlow(t *testing.T) {
	expectNumber(t, `
		function calculate(n) {
			var result = 0;
			for (var i = 0; i < n; i++) {
				if (i % 3 === 0) continue;
				if (i > 7) break;
				result = result + i;
			}
			return result;
		}
		calculate(10);
	`, 1+2+4+5+7)
}

func TestComplexObject(t *testing.T) {
	expectNumber(t, `
		var person = {
			name: "Alice",
			age: 30,
			greet: function() {
				return this.age;
			}
		};
		person.greet();
	`, 30)
}

func TestArrayEvery(t *testing.T) {
	expectBool(t, `
		var arr = [2, 4, 6];
		arr.every(function(x) { return x % 2 === 0; });
	`, true)
}

func TestArraySome(t *testing.T) {
	expectBool(t, `
		var arr = [1, 3, 5, 6];
		arr.some(function(x) { return x % 2 === 0; });
	`, true)
}

func TestArrayFind(t *testing.T) {
	expectNumber(t, `
		var arr = [1, 2, 3, 4];
		arr.find(function(x) { return x > 2; });
	`, 3)
}

func TestArrayFindIndex(t *testing.T) {
	expectNumber(t, `
		var arr = [1, 2, 3, 4];
		arr.findIndex(function(x) { return x > 2; });
	`, 2)
}

func TestArrayConcat(t *testing.T) {
	expectNumber(t, `
		var a = [1, 2];
		var b = a.concat([3, 4], [5]);
		b.length;
	`, 5)
}

func TestArraySlice(t *testing.T) {
	expectString(t, `
		var arr = [1, 2, 3, 4, 5];
		arr.slice(1, 3).join(",");
	`, "2,3")
}

func TestArraySplice(t *testing.T) {
	expectString(t, `
		var arr = [1, 2, 3, 4, 5];
		arr.splice(1, 2);
		arr.join(",");
	`, "1,4,5")
}

func TestArrayReverse(t *testing.T) {
	expectString(t, `
		var arr = [1, 2, 3];
		arr.reverse();
		arr.join(",");
	`, "3,2,1")
}

// --- Edge cases ---

func TestDivisionByZero(t *testing.T) {
	val := evalExpect(t, "1 / 0")
	if !math.IsInf(val.Number, 1) {
		t.Fatalf("expected Infinity, got %v", val.Number)
	}
}

func TestNaNComparison(t *testing.T) {
	expectBool(t, "NaN === NaN", false)
	expectBool(t, "NaN !== NaN", true)
}

func TestEmptyReturn(t *testing.T) {
	expectUndefined(t, `
		function f() { return; }
		f();
	`)
}

func TestUndefinedAccess(t *testing.T) {
	expectUndefined(t, `
		var obj = {};
		obj.nonexistent;
	`)
}

func TestMultipleArgs(t *testing.T) {
	expectNumber(t, `
		function sum(a, b, c, d, e) {
			return a + b + c + d + e;
		}
		sum(1, 2, 3, 4, 5);
	`, 15)
}

// --- Labeled statements ---

func TestLabeledBreak(t *testing.T) {
	expectNumber(t, `
		var result = 0;
		outer: for (var i = 0; i < 3; i++) {
			for (var j = 0; j < 3; j++) {
				if (i === 1 && j === 1) break outer;
				result = result + 1;
			}
		}
		result;
	`, 4)
}

// --- Spread in function call ---

func TestSpreadInCall(t *testing.T) {
	expectNumber(t, `
		function add(a, b, c) { return a + b + c; }
		var args = [1, 2, 3];
		add(...args);
	`, 6)
}

// --- Array assignment ---

func TestArrayAssignment(t *testing.T) {
	expectNumber(t, `
		var arr = [1, 2, 3];
		arr[1] = 20;
		arr[1];
	`, 20)
}

// --- Conditional chaining of method calls ---

func TestMethodChaining(t *testing.T) {
	expectString(t, `
		var arr = [3, 1, 2];
		arr.sort(function(a, b) { return a - b; }).join("-");
	`, "1-2-3")
}

// --- Complex class ---

func TestComplexClassHierarchy(t *testing.T) {
	expectString(t, `
		class Shape {
			constructor(type) {
				this.type = type;
			}
			describe() {
				return "I am a " + this.type;
			}
		}
		class Circle extends Shape {
			constructor(radius) {
				super("circle");
				this.radius = radius;
			}
			area() {
				return 3.14159 * this.radius * this.radius;
			}
		}
		var c = new Circle(5);
		c.describe();
	`, "I am a circle")
}

// --- FizzBuzz ---

func TestFizzBuzz(t *testing.T) {
	expectString(t, `
		function fizzbuzz(n) {
			if (n % 15 === 0) return "FizzBuzz";
			if (n % 3 === 0) return "Fizz";
			if (n % 5 === 0) return "Buzz";
			return "" + n;
		}
		fizzbuzz(15);
	`, "FizzBuzz")
}

// --- Error propagation across function calls ---

func TestErrorPropagation(t *testing.T) {
	expectNumber(t, `
		function thrower() {
			throw 42;
		}
		var result;
		try {
			thrower();
		} catch (e) {
			result = e;
		}
		result;
	`, 42)
}

// --- String charAt ---

func TestStringCharAt(t *testing.T) {
	expectString(t, `"hello".charAt(0)`, "h")
	expectString(t, `"hello".charAt(4)`, "o")
}

// --- Nullish assignment ---

func TestNullishAssignment(t *testing.T) {
	expectNumber(t, `
		var x = null;
		x ??= 42;
		x;
	`, 42)
	expectNumber(t, `
		var x = 10;
		x ??= 42;
		x;
	`, 10)
}

// --- Logical assignment ---

func TestLogicalAssignment(t *testing.T) {
	expectNumber(t, `
		var x = 0;
		x ||= 42;
		x;
	`, 42)
	expectNumber(t, `
		var x = 10;
		x &&= 42;
		x;
	`, 42)
}

// --- Block-scoped function declarations (Annex B) ---

func TestBlockFunctionHoisting(t *testing.T) {
	// Block function hoists name as undefined, assigns value when block executes
	expectString(t, `
		(function() {
			var before = typeof f;
			{ function f() { return 42; } }
			var after = typeof f;
			return before + " / " + after;
		})();
	`, "undefined / function")

	// Block function value accessible after block
	expectNumber(t, `
		(function() {
			{ function f() { return 42; } }
			return f();
		})();
	`, 42)

	// Multiple blocks with same function name
	expectString(t, `
		(function() {
			{ function f() { return 1; } }
			var r1 = f();
			{ function f() { return 2; } }
			var r2 = f();
			return r1 + " / " + r2;
		})();
	`, "1 / 2")

	// Only executed branch propagates
	expectString(t, `
		(function() {
			if (true) { function f() { return 'yes'; } }
			else { function f() { return 'no'; } }
			return f();
		})();
	`, "yes")
}

func TestBlockFunctionMutable(t *testing.T) {
	// Block function binding is mutable (not const)
	expectString(t, `
		(function() {
			{ function f() { return 1; } f = 'reassigned'; }
			return "ok";
		})();
	`, "ok")

	// Function declaration name is mutable inside function body
	expectNumber(t, `
		(function() {
			function f() { f = 123; return f; }
			return f();
		})();
	`, 123)
}

func TestBlockFunctionVarCoexistence(t *testing.T) {
	// var and block function with same name should not conflict
	expectString(t, `
		(function() {
			var f;
			{ function f() { return 1; } }
			return typeof f;
		})();
	`, "function")

	// var with value, block function overwrites after block
	expectString(t, `
		(function() {
			var f = 1;
			{ function f() { return 2; } }
			return typeof f;
		})();
	`, "function")
}

func TestBlockFunctionLexicalConflict(t *testing.T) {
	// let in function scope prevents Annex B hoisting
	expectNumber(t, `
		(function() {
			let f = 123;
			{ function f() {} }
			return f;
		})();
	`, 123)
}

func TestSwitchFunctionDeclaration(t *testing.T) {
	// Function in switch case hoists to function scope
	expectString(t, `
		(function() {
			switch (1) { default: function f() { return 'decl'; } }
			return typeof f;
		})();
	`, "function")

	// Function in switch, body can assign to name
	expectNumber(t, `
		(function() {
			switch (1) { default: function f() { f = 123; return f; } }
			return f();
		})();
	`, 123)
}

func TestNamedFunctionExpressionImmutable(t *testing.T) {
	// Named function expression has immutable self-reference
	expectString(t, `
		var g = function f() {
			try { f = 123; } catch(e) {}
			return typeof f;
		};
		g();
	`, "function")
}
