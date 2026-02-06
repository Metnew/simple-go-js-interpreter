package builtins

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/example/jsgo/runtime"
)

func registerGlobalFunctions(env *runtime.Environment) {
	declareFunc(env, "parseInt", 2, globalParseInt)
	declareFunc(env, "parseFloat", 1, globalParseFloat)
	declareFunc(env, "isNaN", 1, globalIsNaN)
	declareFunc(env, "isFinite", 1, globalIsFinite)
	declareFunc(env, "encodeURI", 1, globalEncodeURI)
	declareFunc(env, "decodeURI", 1, globalDecodeURI)
	declareFunc(env, "encodeURIComponent", 1, globalEncodeURIComponent)
	declareFunc(env, "decodeURIComponent", 1, globalDecodeURIComponent)
	declareFunc(env, "eval", 1, globalEval)
	declareFunc(env, "escape", 1, globalEscape)
	declareFunc(env, "unescape", 1, globalUnescape)

	env.Declare("undefined", "var", runtime.Undefined)
	env.Declare("NaN", "var", runtime.NaN)
	env.Declare("Infinity", "var", runtime.PosInf)
}

func declareFunc(env *runtime.Environment, name string, length int, fn runtime.CallableFunc) {
	obj := newFuncObject(name, length, fn)
	env.Declare(name, "var", runtime.NewObject(obj))
}

func globalParseInt(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := strings.TrimSpace(argAt(args, 0).ToString())
	radix := 10
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		radix = int(toInteger(args[1]))
	}
	if radix == 0 {
		radix = 10
	}
	if radix < 2 || radix > 36 {
		return runtime.NaN, nil
	}
	if s == "" {
		return runtime.NaN, nil
	}
	// Handle hex prefix
	if radix == 16 || (radix == 10 && (strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X"))) {
		radix = 16
		s = strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
	}
	// Parse as many valid characters as possible
	validChars := "0123456789abcdefghijklmnopqrstuvwxyz"[:radix]
	validCharsUpper := strings.ToUpper(validChars)
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	} else if len(s) > 0 && s[0] == '+' {
		s = s[1:]
	}
	end := 0
	for _, c := range s {
		cs := string(c)
		if !strings.Contains(validChars, cs) && !strings.Contains(validCharsUpper, cs) {
			break
		}
		end++
	}
	if end == 0 {
		return runtime.NaN, nil
	}
	n, err := strconv.ParseInt(s[:end], radix, 64)
	if err != nil {
		return runtime.NaN, nil
	}
	if neg {
		n = -n
	}
	return runtime.NewNumber(float64(n)), nil
}

func globalParseFloat(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := strings.TrimSpace(argAt(args, 0).ToString())
	if s == "" {
		return runtime.NaN, nil
	}
	// Find the longest prefix that is a valid float
	end := 0
	hasDecimal := false
	hasE := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			end = i + 1
			continue
		}
		if c == '.' && !hasDecimal && !hasE {
			hasDecimal = true
			end = i + 1
			continue
		}
		if (c == 'e' || c == 'E') && !hasE && end > 0 {
			hasE = true
			if i+1 < len(s) && (s[i+1] == '+' || s[i+1] == '-') {
				i++
			}
			continue
		}
		if (c == '+' || c == '-') && i == 0 {
			continue
		}
		break
	}
	if s == "Infinity" || s == "+Infinity" {
		return runtime.PosInf, nil
	}
	if s == "-Infinity" {
		return runtime.NegInf, nil
	}
	if end == 0 {
		return runtime.NaN, nil
	}
	f, err := strconv.ParseFloat(s[:end], 64)
	if err != nil {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(f), nil
}

func globalIsNaN(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	n := toNumber(argAt(args, 0))
	return runtime.NewBool(math.IsNaN(n)), nil
}

func globalIsFinite(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	n := toNumber(argAt(args, 0))
	return runtime.NewBool(!math.IsNaN(n) && !math.IsInf(n, 0)), nil
}

func globalEncodeURI(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := argAt(args, 0).ToString()
	// encodeURI does not encode: ; , / ? : @ & = + $ - _ . ! ~ * ' ( ) # and alphanumeric
	result := encodeURIHelper(s, ";,/?:@&=+$-_.!~*'()#")
	return runtime.NewString(result), nil
}

func globalDecodeURI(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := argAt(args, 0).ToString()
	decoded, err := url.PathUnescape(s)
	if err != nil {
		return nil, fmt.Errorf("URIError: URI malformed")
	}
	return runtime.NewString(decoded), nil
}

func globalEncodeURIComponent(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := argAt(args, 0).ToString()
	result := encodeURIHelper(s, "-_.!~*'()")
	return runtime.NewString(result), nil
}

func globalDecodeURIComponent(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := argAt(args, 0).ToString()
	decoded, err := url.PathUnescape(s)
	if err != nil {
		return nil, fmt.Errorf("URIError: URI malformed")
	}
	return runtime.NewString(decoded), nil
}

func globalEval(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return nil, fmt.Errorf("EvalError: eval is not supported")
}

func encodeURIHelper(s string, safe string) string {
	var sb strings.Builder
	for _, r := range s {
		c := string(r)
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || strings.ContainsRune(safe, r) {
			sb.WriteString(c)
		} else {
			encoded := url.PathEscape(c)
			sb.WriteString(encoded)
		}
	}
	return sb.String()
}

// escape encodes a string using the legacy escape encoding.
// Characters not encoded: A-Z a-z 0-9 @ * _ + - . /
func globalEscape(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := argAt(args, 0).ToString()
	var sb strings.Builder
	for _, r := range s {
		if isEscapeSafe(r) {
			sb.WriteRune(r)
		} else if r <= 0xFF {
			sb.WriteString(fmt.Sprintf("%%%02X", r))
		} else {
			sb.WriteString(fmt.Sprintf("%%u%04X", r))
		}
	}
	return runtime.NewString(sb.String()), nil
}

func isEscapeSafe(r rune) bool {
	if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
		return true
	}
	switch r {
	case '@', '*', '_', '+', '-', '.', '/':
		return true
	}
	return false
}

// unescape decodes a string produced by escape.
func globalUnescape(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := argAt(args, 0).ToString()
	var sb strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '%' {
			if i+5 < len(s) && s[i+1] == 'u' {
				hex := s[i+2 : i+6]
				n, err := strconv.ParseInt(hex, 16, 32)
				if err == nil {
					sb.WriteRune(rune(n))
					i += 6
					continue
				}
			}
			if i+2 < len(s) {
				hex := s[i+1 : i+3]
				n, err := strconv.ParseInt(hex, 16, 32)
				if err == nil {
					sb.WriteRune(rune(n))
					i += 3
					continue
				}
			}
		}
		sb.WriteByte(s[i])
		i++
	}
	return runtime.NewString(sb.String()), nil
}
