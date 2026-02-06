package builtins

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/example/jsgo/runtime"
)

var StringPrototype *runtime.Object

func createStringConstructor(objProto *runtime.Object) (*runtime.Object, *runtime.Object) {
	proto := runtime.NewOrdinaryObject(objProto)
	StringPrototype = proto

	setMethod(proto, "charAt", 1, stringCharAt)
	setMethod(proto, "charCodeAt", 1, stringCharCodeAt)
	setMethod(proto, "codePointAt", 1, stringCodePointAt)
	setMethod(proto, "indexOf", 1, stringIndexOf)
	setMethod(proto, "lastIndexOf", 1, stringLastIndexOf)
	setMethod(proto, "includes", 1, stringIncludes)
	setMethod(proto, "startsWith", 1, stringStartsWith)
	setMethod(proto, "endsWith", 1, stringEndsWith)
	setMethod(proto, "slice", 2, stringSlice)
	setMethod(proto, "substring", 2, stringSubstring)
	setMethod(proto, "substr", 2, stringSubstr)
	setMethod(proto, "toUpperCase", 0, stringToUpperCase)
	setMethod(proto, "toLowerCase", 0, stringToLowerCase)
	setMethod(proto, "trim", 0, stringTrim)
	setMethod(proto, "trimStart", 0, stringTrimStart)
	setMethod(proto, "trimEnd", 0, stringTrimEnd)
	setMethod(proto, "repeat", 1, stringRepeat)
	setMethod(proto, "padStart", 1, stringPadStart)
	setMethod(proto, "padEnd", 1, stringPadEnd)
	setMethod(proto, "split", 1, stringSplit)
	setMethod(proto, "replace", 2, stringReplace)
	setMethod(proto, "match", 1, stringMatch)
	setMethod(proto, "search", 1, stringSearch)
	setMethod(proto, "concat", 1, stringConcat)
	setMethod(proto, "normalize", 0, stringNormalize)
	setMethod(proto, "toString", 0, stringToString)
	setMethod(proto, "valueOf", 0, stringValueOf)
	setMethod(proto, "at", 1, stringAt)
	// Annex B aliases - must be the SAME function object as the original
	trimStartProp := proto.Properties["trimStart"]
	proto.DefineProperty("trimLeft", trimStartProp)
	trimEndProp := proto.Properties["trimEnd"]
	proto.DefineProperty("trimRight", trimEndProp)
	// Annex B HTML methods
	setMethod(proto, "anchor", 1, makeHTMLWrapper("a", "name"))
	setMethod(proto, "big", 0, makeHTMLSimple("big"))
	setMethod(proto, "blink", 0, makeHTMLSimple("blink"))
	setMethod(proto, "bold", 0, makeHTMLSimple("b"))
	setMethod(proto, "fixed", 0, makeHTMLSimple("tt"))
	setMethod(proto, "fontcolor", 1, makeHTMLWrapper("font", "color"))
	setMethod(proto, "fontsize", 1, makeHTMLWrapper("font", "size"))
	setMethod(proto, "italics", 0, makeHTMLSimple("i"))
	setMethod(proto, "link", 1, makeHTMLWrapper("a", "href"))
	setMethod(proto, "small", 0, makeHTMLSimple("small"))
	setMethod(proto, "strike", 0, makeHTMLSimple("strike"))
	setMethod(proto, "sub", 0, makeHTMLSimple("sub"))
	setMethod(proto, "sup", 0, makeHTMLSimple("sup"))

	ctor := newFuncObject("String", 1, stringConstructorCall)
	ctor.Constructor = stringConstructorCall

	setMethod(ctor, "fromCharCode", 1, stringFromCharCode)
	setMethod(ctor, "fromCodePoint", 1, stringFromCodePoint)
	setMethod(ctor, "raw", 1, stringRaw)

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor, proto
}

func getStringValue(this *runtime.Value) string {
	s, _ := getStringValueErr(this)
	return s
}

func getStringValueErr(this *runtime.Value) (string, error) {
	if this == nil {
		return "", nil
	}
	if this.Type == runtime.TypeString {
		return this.Str, nil
	}
	if this.Type == runtime.TypeObject && this.Object != nil {
		if iv, ok := this.Object.Internal["StringData"]; ok {
			return iv.(string), nil
		}
		return jsToString(this)
	}
	return this.ToString(), nil
}

func stringConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	if len(args) == 0 {
		return runtime.NewString(""), nil
	}
	return runtime.NewString(args[0].ToString()), nil
}

func stringCharAt(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	idx := 0
	if len(args) > 0 {
		idx = int(toInteger(args[0]))
	}
	runes := []rune(s)
	if idx < 0 || idx >= len(runes) {
		return runtime.NewString(""), nil
	}
	return runtime.NewString(string(runes[idx])), nil
}

func stringCharCodeAt(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	idx := 0
	if len(args) > 0 {
		idx = int(toInteger(args[0]))
	}
	runes := []rune(s)
	if idx < 0 || idx >= len(runes) {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(runes[idx])), nil
}

func stringCodePointAt(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	idx := 0
	if len(args) > 0 {
		idx = int(toInteger(args[0]))
	}
	runes := []rune(s)
	if idx < 0 || idx >= len(runes) {
		return runtime.Undefined, nil
	}
	return runtime.NewNumber(float64(runes[idx])), nil
}

func stringIndexOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	search := argAt(args, 0).ToString()
	pos := 0
	if len(args) > 1 {
		pos = int(toInteger(args[1]))
		if pos < 0 {
			pos = 0
		}
	}
	if pos > len(s) {
		return runtime.NewNumber(-1), nil
	}
	idx := strings.Index(s[pos:], search)
	if idx == -1 {
		return runtime.NewNumber(-1), nil
	}
	return runtime.NewNumber(float64(idx + pos)), nil
}

func stringLastIndexOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	search := argAt(args, 0).ToString()
	pos := len(s)
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		pos = int(toInteger(args[1]))
		if pos < 0 {
			pos = 0
		}
	}
	if pos > len(s) {
		pos = len(s)
	}
	idx := strings.LastIndex(s[:pos+len(search)], search)
	if idx == -1 {
		return runtime.NewNumber(-1), nil
	}
	if idx > pos {
		idx = -1
	}
	return runtime.NewNumber(float64(idx)), nil
}

func stringIncludes(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	search := argAt(args, 0).ToString()
	pos := 0
	if len(args) > 1 {
		pos = int(toInteger(args[1]))
		if pos < 0 {
			pos = 0
		}
	}
	if pos > len(s) {
		return runtime.False, nil
	}
	return runtime.NewBool(strings.Contains(s[pos:], search)), nil
}

func stringStartsWith(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	search := argAt(args, 0).ToString()
	pos := 0
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		pos = int(toInteger(args[1]))
	}
	if pos < 0 {
		pos = 0
	}
	if pos > len(s) {
		return runtime.False, nil
	}
	return runtime.NewBool(strings.HasPrefix(s[pos:], search)), nil
}

func stringEndsWith(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	search := argAt(args, 0).ToString()
	end := len(s)
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		end = int(toInteger(args[1]))
	}
	if end < 0 {
		end = 0
	}
	if end > len(s) {
		end = len(s)
	}
	return runtime.NewBool(strings.HasSuffix(s[:end], search)), nil
}

func stringSlice(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	runes := []rune(s)
	length := len(runes)
	start := 0
	end := length
	if len(args) > 0 {
		start = int(toInteger(args[0]))
		if start < 0 {
			start = length + start
			if start < 0 {
				start = 0
			}
		}
	}
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		end = int(toInteger(args[1]))
		if end < 0 {
			end = length + end
			if end < 0 {
				end = 0
			}
		}
	}
	if start > length {
		start = length
	}
	if end > length {
		end = length
	}
	if start >= end {
		return runtime.NewString(""), nil
	}
	return runtime.NewString(string(runes[start:end])), nil
}

func stringSubstring(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	runes := []rune(s)
	length := len(runes)
	start := 0
	end := length
	if len(args) > 0 {
		start = int(toInteger(args[0]))
	}
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		end = int(toInteger(args[1]))
	}
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if start > length {
		start = length
	}
	if end > length {
		end = length
	}
	if start > end {
		start, end = end, start
	}
	return runtime.NewString(string(runes[start:end])), nil
}

func stringSubstr(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	if this == nil || this.Type == runtime.TypeUndefined || this.Type == runtime.TypeNull {
		return nil, fmt.Errorf("TypeError: Cannot read properties of %s", this.ToString())
	}
	s, err := getStringValueErr(this)
	if err != nil {
		return nil, err
	}
	// Use UTF-16 code units (not code points) per JS spec
	units := stringToUTF16(s)
	length := len(units)
	intStart, err2 := toIntegerErr(argAt(args, 0))
	if err2 != nil {
		return nil, err2
	}
	if math.IsInf(intStart, -1) {
		intStart = 0
	} else if intStart < 0 {
		intStart = math.Max(float64(length)+intStart, 0)
	} else if intStart > float64(length) {
		intStart = float64(length)
	}
	start := int(intStart)
	if start < 0 {
		start = 0
	}
	if start > length {
		start = length
	}
	var intLength float64
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		intLength, err2 = toIntegerErr(args[1])
		if err2 != nil {
			return nil, err2
		}
	} else {
		intLength = float64(length)
	}
	count := int(math.Min(math.Max(intLength, 0), float64(length)))
	if start >= length || count == 0 {
		return runtime.NewString(""), nil
	}
	end := start + count
	if end > length {
		end = length
	}
	return runtime.NewString(utf16ToString(units[start:end])), nil
}

func stringToUpperCase(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewString(strings.ToUpper(getStringValue(this))), nil
}

func stringToLowerCase(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewString(strings.ToLower(getStringValue(this))), nil
}

func stringTrim(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewString(strings.TrimSpace(getStringValue(this))), nil
}

func stringTrimStart(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewString(strings.TrimLeft(getStringValue(this), " \t\n\r\v\f")), nil
}

func stringTrimEnd(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewString(strings.TrimRight(getStringValue(this), " \t\n\r\v\f")), nil
}

func stringRepeat(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	count := 0
	if len(args) > 0 {
		count = int(toInteger(args[0]))
	}
	if count < 0 {
		return nil, fmt.Errorf("RangeError: Invalid count value")
	}
	return runtime.NewString(strings.Repeat(s, count)), nil
}

func stringPadStart(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	targetLen := 0
	if len(args) > 0 {
		targetLen = int(toInteger(args[0]))
	}
	padStr := " "
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		padStr = args[1].ToString()
	}
	runes := []rune(s)
	if len(runes) >= targetLen || padStr == "" {
		return runtime.NewString(s), nil
	}
	needed := targetLen - len(runes)
	padding := strings.Repeat(padStr, (needed/utf8.RuneCountInString(padStr))+1)
	padRunes := []rune(padding)
	return runtime.NewString(string(padRunes[:needed]) + s), nil
}

func stringPadEnd(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	targetLen := 0
	if len(args) > 0 {
		targetLen = int(toInteger(args[0]))
	}
	padStr := " "
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		padStr = args[1].ToString()
	}
	runes := []rune(s)
	if len(runes) >= targetLen || padStr == "" {
		return runtime.NewString(s), nil
	}
	needed := targetLen - len(runes)
	padding := strings.Repeat(padStr, (needed/utf8.RuneCountInString(padStr))+1)
	padRunes := []rune(padding)
	return runtime.NewString(s + string(padRunes[:needed])), nil
}

func stringSplit(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	if len(args) == 0 || args[0].Type == runtime.TypeUndefined {
		return createStringArray([]string{s}), nil
	}
	sep := args[0].ToString()
	limit := -1
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		limit = int(toUint32(args[1]))
	}
	var parts []string
	if sep == "" {
		runes := []rune(s)
		parts = make([]string, len(runes))
		for i, r := range runes {
			parts[i] = string(r)
		}
	} else {
		parts = strings.Split(s, sep)
	}
	if limit >= 0 && len(parts) > limit {
		parts = parts[:limit]
	}
	return createStringArray(parts), nil
}

func stringReplace(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	if len(args) < 2 {
		return runtime.NewString(s), nil
	}
	search := args[0].ToString()
	replacement := args[1].ToString()
	result := strings.Replace(s, search, replacement, 1)
	return runtime.NewString(result), nil
}

func stringMatch(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	// Simplified: just returns null for now without RegExp integration
	return runtime.Null, nil
}

func stringSearch(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	if len(args) == 0 {
		return runtime.NewNumber(0), nil
	}
	search := args[0].ToString()
	idx := strings.Index(s, search)
	return runtime.NewNumber(float64(idx)), nil
}

func stringConcat(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	var sb strings.Builder
	sb.WriteString(getStringValue(this))
	for _, a := range args {
		sb.WriteString(a.ToString())
	}
	return runtime.NewString(sb.String()), nil
}

func stringNormalize(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewString(getStringValue(this)), nil
}

func stringToString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewString(getStringValue(this)), nil
}

func stringValueOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewString(getStringValue(this)), nil
}

func stringAt(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := getStringValue(this)
	runes := []rune(s)
	idx := 0
	if len(args) > 0 {
		idx = int(toInteger(args[0]))
	}
	if idx < 0 {
		idx = len(runes) + idx
	}
	if idx < 0 || idx >= len(runes) {
		return runtime.Undefined, nil
	}
	return runtime.NewString(string(runes[idx])), nil
}

func stringFromCharCode(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	runes := make([]rune, len(args))
	for i, a := range args {
		runes[i] = rune(toUint32(a) & 0xFFFF)
	}
	return runtime.NewString(string(runes)), nil
}

func stringFromCodePoint(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	runes := make([]rune, len(args))
	for i, a := range args {
		cp := toNumber(a)
		if cp < 0 || cp > 0x10FFFF || cp != float64(int(cp)) {
			return nil, fmt.Errorf("RangeError: Invalid code point %v", cp)
		}
		runes[i] = rune(int(cp))
	}
	return runtime.NewString(string(runes)), nil
}

func stringRaw(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	if len(args) == 0 {
		return runtime.NewString(""), nil
	}
	template := toObject(args[0])
	if template == nil {
		return runtime.NewString(""), nil
	}
	rawProp := template.Get("raw")
	rawObj := toObject(rawProp)
	if rawObj == nil || rawObj.ArrayData == nil {
		return runtime.NewString(""), nil
	}
	var sb strings.Builder
	subs := args[1:]
	for i, v := range rawObj.ArrayData {
		sb.WriteString(v.ToString())
		if i < len(subs) {
			sb.WriteString(subs[i].ToString())
		}
	}
	return runtime.NewString(sb.String()), nil
}

// Annex B HTML string methods
func makeHTMLSimple(tag string) runtime.CallableFunc {
	return func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		if this == nil || this.Type == runtime.TypeUndefined || this.Type == runtime.TypeNull {
			return nil, fmt.Errorf("TypeError: String.prototype.%s requires that 'this' not be undefined or null", tag)
		}
		s, err := jsToString(this)
		if err != nil {
			return nil, err
		}
		return runtime.NewString("<" + tag + ">" + s + "</" + tag + ">"), nil
	}
}

func makeHTMLWrapper(tag, attr string) runtime.CallableFunc {
	return func(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
		if this == nil || this.Type == runtime.TypeUndefined || this.Type == runtime.TypeNull {
			return nil, fmt.Errorf("TypeError: String.prototype.%s requires that 'this' not be undefined or null", tag)
		}
		s, err := jsToString(this)
		if err != nil {
			return nil, err
		}
		val := ""
		if len(args) > 0 {
			v, verr := jsToString(args[0])
			if verr != nil {
				return nil, verr
			}
			val = v
		}
		// Escape double quotes in attribute value
		val = strings.ReplaceAll(val, "\"", "&quot;")
		return runtime.NewString("<" + tag + " " + attr + "=\"" + val + "\">" + s + "</" + tag + ">"), nil
	}
}
