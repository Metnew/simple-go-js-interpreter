package builtins

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/example/jsgo/runtime"
)

var RegExpPrototype *runtime.Object

func createRegExpConstructor(objProto *runtime.Object) (*runtime.Object, *runtime.Object) {
	proto := runtime.NewOrdinaryObject(objProto)
	proto.OType = runtime.ObjTypeRegExp
	RegExpPrototype = proto

	setMethod(proto, "test", 1, regexpTest)
	setMethod(proto, "exec", 1, regexpExec)
	setMethod(proto, "toString", 0, regexpToString)
	setMethod(proto, "compile", 2, regexpCompile)

	ctor := newFuncObject("RegExp", 2, regexpConstructorCall)
	ctor.Constructor = regexpConstructorCall
	ctor.Prototype = proto

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor, proto
}

func regexpConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	pattern := ""
	flags := ""
	if len(args) > 0 && args[0].Type != runtime.TypeUndefined {
		pattern = args[0].ToString()
	}
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		flags = args[1].ToString()
	}
	return createRegExpObject(pattern, flags)
}

func createRegExpObject(pattern, flags string) (*runtime.Value, error) {
	goPattern := jsRegexpToGo(pattern)
	if strings.Contains(flags, "i") {
		goPattern = "(?i)" + goPattern
	}
	if strings.Contains(flags, "s") {
		goPattern = "(?s)" + goPattern
	}
	re, err := regexp.Compile(goPattern)
	if err != nil {
		return nil, fmt.Errorf("SyntaxError: Invalid regular expression: %s", err)
	}
	obj := &runtime.Object{
		OType:      runtime.ObjTypeRegExp,
		Properties: make(map[string]*runtime.Property),
		Prototype:  RegExpPrototype,
		Internal:   map[string]interface{}{"regexp": re, "pattern": pattern, "flags": flags},
	}
	setDataProp(obj, "source", runtime.NewString(pattern), false, false, true)
	setDataProp(obj, "flags", runtime.NewString(flags), false, false, true)
	setDataProp(obj, "global", runtime.NewBool(strings.Contains(flags, "g")), false, false, true)
	setDataProp(obj, "ignoreCase", runtime.NewBool(strings.Contains(flags, "i")), false, false, true)
	setDataProp(obj, "multiline", runtime.NewBool(strings.Contains(flags, "m")), false, false, true)
	setDataProp(obj, "sticky", runtime.NewBool(strings.Contains(flags, "y")), false, false, true)
	setDataProp(obj, "unicode", runtime.NewBool(strings.Contains(flags, "u")), false, false, true)
	obj.Set("lastIndex", runtime.NewNumber(0))
	return runtime.NewObject(obj), nil
}

func jsRegexpToGo(pattern string) string {
	// Go's regexp2 isn't available, so we do basic translation
	// This is a simplified mapping; a full engine would need more
	return pattern
}

func getRegExp(this *runtime.Value) *regexp.Regexp {
	obj := toObject(this)
	if obj == nil || obj.Internal == nil {
		return nil
	}
	re, ok := obj.Internal["regexp"].(*regexp.Regexp)
	if !ok {
		return nil
	}
	return re
}

func regexpTest(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	re := getRegExp(this)
	if re == nil {
		return runtime.False, nil
	}
	s := argAt(args, 0).ToString()
	return runtime.NewBool(re.MatchString(s)), nil
}

func regexpExec(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	re := getRegExp(this)
	if re == nil {
		return runtime.Null, nil
	}
	s := argAt(args, 0).ToString()
	match := re.FindStringSubmatchIndex(s)
	if match == nil {
		return runtime.Null, nil
	}
	groups := make([]*runtime.Value, 0)
	for i := 0; i < len(match); i += 2 {
		if match[i] == -1 {
			groups = append(groups, runtime.Undefined)
		} else {
			groups = append(groups, runtime.NewString(s[match[i]:match[i+1]]))
		}
	}
	result := newArray(groups)
	result.Set("index", runtime.NewNumber(float64(match[0])))
	result.Set("input", runtime.NewString(s))
	return runtime.NewObject(result), nil
}

func regexpCompile(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return nil, fmt.Errorf("TypeError: RegExp.prototype.compile called on incompatible receiver")
	}

	pattern := ""
	flags := ""
	if len(args) > 0 && args[0].Type != runtime.TypeUndefined {
		pattern = args[0].ToString()
	}
	if len(args) > 1 && args[1].Type != runtime.TypeUndefined {
		flags = args[1].ToString()
	}

	goPattern := jsRegexpToGo(pattern)
	if strings.Contains(flags, "i") {
		goPattern = "(?i)" + goPattern
	}
	if strings.Contains(flags, "s") {
		goPattern = "(?s)" + goPattern
	}
	re, err := regexp.Compile(goPattern)
	if err != nil {
		return nil, fmt.Errorf("SyntaxError: Invalid regular expression: %s", err)
	}

	if obj.Internal == nil {
		obj.Internal = make(map[string]interface{})
	}
	obj.Internal["regexp"] = re
	obj.Internal["pattern"] = pattern
	obj.Internal["flags"] = flags

	obj.Set("source", runtime.NewString(pattern))
	obj.Set("flags", runtime.NewString(flags))
	obj.Set("global", runtime.NewBool(strings.Contains(flags, "g")))
	obj.Set("ignoreCase", runtime.NewBool(strings.Contains(flags, "i")))
	obj.Set("multiline", runtime.NewBool(strings.Contains(flags, "m")))
	obj.Set("sticky", runtime.NewBool(strings.Contains(flags, "y")))
	obj.Set("unicode", runtime.NewBool(strings.Contains(flags, "u")))
	obj.Set("lastIndex", runtime.NewNumber(0))

	return this, nil
}

func regexpToString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil {
		return runtime.NewString("/(?:)/"), nil
	}
	source := obj.Get("source").ToString()
	flags := obj.Get("flags").ToString()
	return runtime.NewString("/" + source + "/" + flags), nil
}
