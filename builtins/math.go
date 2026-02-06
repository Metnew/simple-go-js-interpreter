package builtins

import (
	"math"
	"math/rand"

	"github.com/example/jsgo/runtime"
)

func createMathObject(objProto *runtime.Object) *runtime.Object {
	m := runtime.NewOrdinaryObject(objProto)

	setConstant(m, "PI", runtime.NewNumber(math.Pi))
	setConstant(m, "E", runtime.NewNumber(math.E))
	setConstant(m, "LN2", runtime.NewNumber(math.Ln2))
	setConstant(m, "LN10", runtime.NewNumber(math.Log(10)))
	setConstant(m, "LOG2E", runtime.NewNumber(math.Log2E))
	setConstant(m, "LOG10E", runtime.NewNumber(math.Log10E))
	setConstant(m, "SQRT2", runtime.NewNumber(math.Sqrt2))
	setConstant(m, "SQRT1_2", runtime.NewNumber(1.0/math.Sqrt2))

	setMethod(m, "abs", 1, mathAbs)
	setMethod(m, "ceil", 1, mathCeil)
	setMethod(m, "floor", 1, mathFloor)
	setMethod(m, "round", 1, mathRound)
	setMethod(m, "trunc", 1, mathTrunc)
	setMethod(m, "sign", 1, mathSign)
	setMethod(m, "max", 2, mathMax)
	setMethod(m, "min", 2, mathMin)
	setMethod(m, "pow", 2, mathPow)
	setMethod(m, "sqrt", 1, mathSqrt)
	setMethod(m, "cbrt", 1, mathCbrt)
	setMethod(m, "hypot", 0, mathHypot)
	setMethod(m, "log", 1, mathLog)
	setMethod(m, "log2", 1, mathLog2)
	setMethod(m, "log10", 1, mathLog10)
	setMethod(m, "exp", 1, mathExp)
	setMethod(m, "expm1", 1, mathExpm1)
	setMethod(m, "log1p", 1, mathLog1p)
	setMethod(m, "sin", 1, mathSin)
	setMethod(m, "cos", 1, mathCos)
	setMethod(m, "tan", 1, mathTan)
	setMethod(m, "asin", 1, mathAsin)
	setMethod(m, "acos", 1, mathAcos)
	setMethod(m, "atan", 1, mathAtan)
	setMethod(m, "atan2", 2, mathAtan2)
	setMethod(m, "random", 0, mathRandom)
	setMethod(m, "fround", 1, mathFround)
	setMethod(m, "clz32", 1, mathClz32)
	setMethod(m, "imul", 2, mathImul)

	m.Set("@@toStringTag", runtime.NewString("Math"))
	return m
}

func mathUnary(args []*runtime.Value, fn func(float64) float64) (*runtime.Value, error) {
	n := toNumber(argAt(args, 0))
	return runtime.NewNumber(fn(n)), nil
}

func mathAbs(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Abs)
}

func mathCeil(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Ceil)
}

func mathFloor(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Floor)
}

func mathRound(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Round)
}

func mathTrunc(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Trunc)
}

func mathSign(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	n := toNumber(argAt(args, 0))
	if isNaN(n) {
		return runtime.NaN, nil
	}
	if n > 0 {
		return runtime.NewNumber(1), nil
	}
	if n < 0 {
		return runtime.NewNumber(-1), nil
	}
	return runtime.NewNumber(n), nil // preserves +0/-0
}

func mathMax(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	if len(args) == 0 {
		return runtime.NegInf, nil
	}
	result := math.Inf(-1)
	for _, a := range args {
		n := toNumber(a)
		if isNaN(n) {
			return runtime.NaN, nil
		}
		if n > result {
			result = n
		}
	}
	return runtime.NewNumber(result), nil
}

func mathMin(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	if len(args) == 0 {
		return runtime.PosInf, nil
	}
	result := math.Inf(1)
	for _, a := range args {
		n := toNumber(a)
		if isNaN(n) {
			return runtime.NaN, nil
		}
		if n < result {
			result = n
		}
	}
	return runtime.NewNumber(result), nil
}

func mathPow(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	base := toNumber(argAt(args, 0))
	exp := toNumber(argAt(args, 1))
	return runtime.NewNumber(math.Pow(base, exp)), nil
}

func mathSqrt(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Sqrt)
}

func mathCbrt(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Cbrt)
}

func mathHypot(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	if len(args) == 0 {
		return runtime.Zero, nil
	}
	sum := 0.0
	for _, a := range args {
		n := toNumber(a)
		if isNaN(n) {
			return runtime.NaN, nil
		}
		sum += n * n
	}
	return runtime.NewNumber(math.Sqrt(sum)), nil
}

func mathLog(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Log)
}

func mathLog2(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Log2)
}

func mathLog10(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Log10)
}

func mathExp(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Exp)
}

func mathExpm1(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Expm1)
}

func mathLog1p(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Log1p)
}

func mathSin(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Sin)
}

func mathCos(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Cos)
}

func mathTan(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Tan)
}

func mathAsin(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Asin)
}

func mathAcos(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Acos)
}

func mathAtan(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return mathUnary(args, math.Atan)
}

func mathAtan2(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	y := toNumber(argAt(args, 0))
	x := toNumber(argAt(args, 1))
	return runtime.NewNumber(math.Atan2(y, x)), nil
}

func mathRandom(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewNumber(rand.Float64()), nil
}

func mathFround(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	n := toNumber(argAt(args, 0))
	return runtime.NewNumber(float64(float32(n))), nil
}

func mathClz32(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	n := toUint32(argAt(args, 0))
	if n == 0 {
		return runtime.NewNumber(32), nil
	}
	count := 0
	for i := 31; i >= 0; i-- {
		if n&(1<<uint(i)) != 0 {
			break
		}
		count++
	}
	return runtime.NewNumber(float64(count)), nil
}

func mathImul(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	a := toInt32(argAt(args, 0))
	b := toInt32(argAt(args, 1))
	return runtime.NewNumber(float64(a * b)), nil
}
