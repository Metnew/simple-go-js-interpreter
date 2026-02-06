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
// IMPORTANT: escape operates on UTF-16 code units, not Unicode code points.
// Surrogate pairs are encoded as two separate %uXXXX sequences.
func globalEscape(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s, err := jsToString(argAt(args, 0))
	if err != nil {
		return nil, err
	}
	// Convert to UTF-16 code units to handle surrogate pairs correctly
	utf16Units := stringToUTF16(s)
	var sb strings.Builder
	for _, cu := range utf16Units {
		r := rune(cu)
		if isEscapeSafe(r) {
			sb.WriteRune(r)
		} else if cu <= 0xFF {
			sb.WriteString(fmt.Sprintf("%%%02X", cu))
		} else {
			sb.WriteString(fmt.Sprintf("%%u%04X", cu))
		}
	}
	return runtime.NewString(sb.String()), nil
}

// stringToUTF16 converts a Go string (UTF-8/WTF-8) to a slice of UTF-16 code units.
// Characters in the BMP are single code units; characters above U+FFFF become surrogate pairs.
// WTF-8 encoded surrogates (from JS \uD800-\uDFFF) are preserved as individual code units.
func stringToUTF16(s string) []uint16 {
	var result []uint16
	i := 0
	for i < len(s) {
		b := s[i]
		if b < 0x80 {
			result = append(result, uint16(b))
			i++
		} else if b < 0xC0 {
			// Continuation byte - shouldn't happen at start, treat as raw
			result = append(result, uint16(b))
			i++
		} else if b < 0xE0 {
			// 2-byte sequence
			if i+1 < len(s) {
				r := rune(b&0x1F)<<6 | rune(s[i+1]&0x3F)
				result = append(result, uint16(r))
				i += 2
			} else {
				result = append(result, uint16(b))
				i++
			}
		} else if b < 0xF0 {
			// 3-byte sequence (includes WTF-8 surrogates)
			if i+2 < len(s) {
				r := rune(b&0x0F)<<12 | rune(s[i+1]&0x3F)<<6 | rune(s[i+2]&0x3F)
				result = append(result, uint16(r))
				i += 3
			} else {
				result = append(result, uint16(b))
				i++
			}
		} else {
			// 4-byte sequence - encode as surrogate pair
			if i+3 < len(s) {
				r := rune(b&0x07)<<18 | rune(s[i+1]&0x3F)<<12 | rune(s[i+2]&0x3F)<<6 | rune(s[i+3]&0x3F)
				r -= 0x10000
				hi := uint16(0xD800 + (r>>10)&0x3FF)
				lo := uint16(0xDC00 + r&0x3FF)
				result = append(result, hi, lo)
				i += 4
			} else {
				result = append(result, uint16(b))
				i++
			}
		}
	}
	return result
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
// Uses UTF-16 code units to match JS behavior.
func globalUnescape(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s, err := jsToString(argAt(args, 0))
	if err != nil {
		return nil, err
	}
	// Collect as UTF-16 code units, then convert back to string
	var codeUnits []uint16
	i := 0
	for i < len(s) {
		if s[i] == '%' {
			if i+5 < len(s) && s[i+1] == 'u' {
				hex := s[i+2 : i+6]
				n, parseErr := strconv.ParseInt(hex, 16, 32)
				if parseErr == nil {
					codeUnits = append(codeUnits, uint16(n))
					i += 6
					continue
				}
			}
			if i+2 < len(s) {
				hex := s[i+1 : i+3]
				n, parseErr := strconv.ParseInt(hex, 16, 32)
				if parseErr == nil {
					codeUnits = append(codeUnits, uint16(n))
					i += 3
					continue
				}
			}
		}
		codeUnits = append(codeUnits, uint16(s[i]))
		i++
	}
	return runtime.NewString(utf16ToString(codeUnits)), nil
}

// utf16ToString converts UTF-16 code units to a Go string.
// Lone surrogates are preserved as replacement characters would lose information,
// so we use them directly.
func utf16ToString(units []uint16) string {
	var sb strings.Builder
	for i := 0; i < len(units); i++ {
		cu := units[i]
		if cu >= 0xD800 && cu <= 0xDBFF && i+1 < len(units) {
			lo := units[i+1]
			if lo >= 0xDC00 && lo <= 0xDFFF {
				// Surrogate pair - combine
				r := rune((uint32(cu)-0xD800)*0x400 + (uint32(lo) - 0xDC00) + 0x10000)
				sb.WriteRune(r)
				i++
				continue
			}
		}
		// For surrogates, use WTF-8 encoding (3-byte sequence) to preserve the value
		if cu >= 0xD800 && cu <= 0xDFFF {
			sb.WriteByte(byte(0xE0 | (cu >> 12)))
			sb.WriteByte(byte(0x80 | ((cu >> 6) & 0x3F)))
			sb.WriteByte(byte(0x80 | (cu & 0x3F)))
		} else {
			sb.WriteRune(rune(cu))
		}
	}
	return sb.String()
}
