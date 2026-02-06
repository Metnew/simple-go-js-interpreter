package builtins

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

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

	// Well-known symbol methods
	if SymSplit != nil {
		splitFn := newFuncObject("[Symbol.split]", 2, regexpSymbolSplit)
		proto.Set(SymSplit.Key(), runtime.NewObject(splitFn))
	}
	if SymMatch != nil {
		matchFn := newFuncObject("[Symbol.match]", 1, regexpSymbolMatch)
		proto.Set(SymMatch.Key(), runtime.NewObject(matchFn))
	}

	ctor := newFuncObject("RegExp", 2, regexpConstructorCall)
	ctor.Constructor = regexpConstructorCall

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor, proto
}

func regexpConstructorCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	patternArg := argAt(args, 0)
	flagsArg := argAt(args, 1)

	pattern := ""
	flags := ""

	// Check if first arg is a RegExp object
	patternObj := toObject(patternArg)
	isRegExp := false
	if patternObj != nil {
		// Per spec: check Symbol.match to determine IsRegExp
		// This may trigger side effects (getter)
		if SymMatch != nil {
			matchVal := patternObj.GetSymbol(SymMatch)
			if matchVal != nil && matchVal != runtime.Undefined {
				isRegExp = true
			}
		}
		// Also check internal regexp slot
		if patternObj.Internal != nil && patternObj.Internal["regexp"] != nil {
			isRegExp = true
		}
	}

	if isRegExp {
		// Extract pattern and flags from the RegExp object
		srcVal := patternObj.Get("source")
		if srcVal != nil && srcVal != runtime.Undefined {
			pattern = srcVal.ToString()
		}
		if flagsArg.Type == runtime.TypeUndefined {
			flVal := patternObj.Get("flags")
			if flVal != nil && flVal != runtime.Undefined {
				flags = flVal.ToString()
			}
		} else {
			flags = flagsArg.ToString()
		}
	} else {
		if patternArg.Type != runtime.TypeUndefined {
			pattern = patternArg.ToString()
		}
		if flagsArg.Type != runtime.TypeUndefined {
			flags = flagsArg.ToString()
		}
	}

	return createRegExpObject(pattern, flags)
}

func validateRegExpFlags(flags string) error {
	valid := "gimsuy"
	seen := make(map[byte]bool)
	for i := 0; i < len(flags); i++ {
		c := flags[i]
		if !strings.ContainsRune(valid, rune(c)) {
			return fmt.Errorf("SyntaxError: Invalid flags supplied to RegExp constructor '%s'", flags)
		}
		if seen[c] {
			return fmt.Errorf("SyntaxError: Invalid flags supplied to RegExp constructor '%s'", flags)
		}
		seen[c] = true
	}
	return nil
}

func createRegExpObject(pattern, flags string) (*runtime.Value, error) {
	if err := validateRegExpFlags(flags); err != nil {
		return nil, err
	}
	flags = canonicalizeFlags(flags)
	if err := validateDuplicateNamedGroups(pattern); err != nil {
		return nil, err
	}
	goPattern := jsRegexpToGo(pattern)
	// Validate pattern with 'u' flag more strictly
	if strings.Contains(flags, "u") {
		// In unicode mode, lone { is invalid
		if strings.Contains(pattern, "{") && !strings.Contains(pattern, "\\{") {
			// Check for unbalanced braces
			depth := 0
			for _, c := range pattern {
				if c == '{' {
					depth++
				} else if c == '}' {
					depth--
				}
			}
			if depth != 0 {
				return nil, fmt.Errorf("SyntaxError: Invalid regular expression: /%s/: Lone quantifier brackets", pattern)
			}
		}
	}
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
	// Go's regexp doesn't support backreferences (\1, \2, etc.)
	// Count capture groups first to determine which \N are valid backreferences
	numGroups := countCaptureGroups(pattern)

	var result strings.Builder
	inCharClass := false
	i := 0
	for i < len(pattern) {
		if !inCharClass && pattern[i] == '[' {
			inCharClass = true
		} else if inCharClass && pattern[i] == ']' {
			inCharClass = false
		}
		if pattern[i] == '\\' && i+1 < len(pattern) {
			next := pattern[i+1]
			// Handle octal escapes and backreferences
			if next >= '0' && next <= '9' && !(next == '0' && (i+2 >= len(pattern) || pattern[i+2] < '0' || pattern[i+2] > '9')) {
				if next == '0' || inCharClass {
					// In character classes or \0, treat as octal
					val := 0
					j := i + 1
					for j < len(pattern) && pattern[j] >= '0' && pattern[j] <= '7' && val*8+int(pattern[j]-'0') <= 255 {
						val = val*8 + int(pattern[j]-'0')
						j++
						if j-i-1 >= 3 {
							break
						}
					}
					// Write as hex escape for Go
					result.WriteString(fmt.Sprintf("\\x%02x", val))
					i = j
					continue
				}
				// Outside character class: backreference
				groupNum := int(next - '0')
				if groupNum <= numGroups {
					// Valid backreference - Go doesn't support these
					// For now, replace with empty match (imperfect but handles many cases)
					result.WriteString("(?:)")
				} else {
					// Invalid backreference (> num groups) - spec says always fails
					result.WriteString("[^\\w\\W]")
				}
				i += 2
				continue
			}
			if next == 'x' {
				// \xHH - Go supports this
				// But \x without two hex digits should be identity escape (Annex B)
				if i+3 < len(pattern) && isHexDigit(pattern[i+2]) && i+4 <= len(pattern) && isHexDigit(pattern[i+3]) {
					result.WriteByte(pattern[i])
					result.WriteByte(pattern[i+1])
					i += 2
					continue
				}
				// Incomplete \x - treat as identity escape (match literal 'x')
				result.WriteByte('x')
				i += 2
				continue
			}
			if next == 'u' {
				// \uHHHH - check if valid
				if i+5 < len(pattern) && isHexDigit(pattern[i+2]) && isHexDigit(pattern[i+3]) && isHexDigit(pattern[i+4]) && isHexDigit(pattern[i+5]) {
					result.WriteByte(pattern[i])
					result.WriteByte(pattern[i+1])
					i += 2
					continue
				}
				// \u{HHHH} form
				if i+2 < len(pattern) && pattern[i+2] == '{' {
					result.WriteByte(pattern[i])
					result.WriteByte(pattern[i+1])
					i += 2
					continue
				}
				// Incomplete \u - treat as identity escape (match literal 'u')
				result.WriteByte('u')
				i += 2
				continue
			}
			// Special JS escapes not in Go
			if next == 'v' {
				// \v = vertical tab (0x0B) - Go doesn't support \v
				result.WriteString("\\x0b")
				i += 2
				continue
			}
			if next == 'c' && i+2 < len(pattern) && ((pattern[i+2] >= 'a' && pattern[i+2] <= 'z') || (pattern[i+2] >= 'A' && pattern[i+2] <= 'Z')) {
				// \cX = control character - Go doesn't support this
				ctrl := pattern[i+2] & 0x1F
				result.WriteString(fmt.Sprintf("\\x%02x", ctrl))
				i += 3
				continue
			}
			// Check if Go's regexp understands this escape
			if isGoRegexpEscape(next) {
				result.WriteByte(pattern[i])
				result.WriteByte(pattern[i+1])
			} else {
				// Identity escape: convert \X to literal match of X
				// Use \x{HH} for safety (avoids Go rejecting the escape)
				if next < 0x80 {
					if isRegexpSpecial(next) {
						result.WriteByte('\\')
						result.WriteByte(next)
					} else {
						result.WriteString(fmt.Sprintf("\\x%02x", next))
					}
				} else {
					// Multi-byte UTF-8: write the literal character bytes
					// Skip the backslash, copy the UTF-8 bytes for the rune
					_, size := utf8.DecodeRuneInString(pattern[i+1:])
					result.WriteString(pattern[i+1 : i+1+size])
					i += 1 + size
					continue
				}
			}
			i += 2
			continue
		}
		result.WriteByte(pattern[i])
		i++
	}
	return result.String()
}

func countCaptureGroups(pattern string) int {
	count := 0
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '\\' {
			i++ // skip escaped character
			continue
		}
		if pattern[i] == '[' {
			// Skip character class
			for i++; i < len(pattern); i++ {
				if pattern[i] == '\\' {
					i++
					continue
				}
				if pattern[i] == ']' {
					break
				}
			}
			continue
		}
		if pattern[i] == '(' {
			// Check if it's a non-capturing group (?:...) or (?=...) etc
			if i+1 < len(pattern) && pattern[i+1] == '?' {
				continue // non-capturing group - don't count
			}
			count++
		}
	}
	return count
}

// validateDuplicateNamedGroups checks for duplicate named capturing groups
// in the same alternative. (?<x>a)(?<x>b) is invalid but (?<x>a)|(?<x>b) is valid.
func validateDuplicateNamedGroups(pattern string) error {
	// Track named groups per alternative at each nesting level.
	// We use a stack of sets: each entry is the set of names in the current alternative
	// at that group nesting level.
	type level struct {
		names map[string]bool
	}
	stack := []level{{names: make(map[string]bool)}}

	i := 0
	for i < len(pattern) {
		switch pattern[i] {
		case '\\':
			i += 2 // skip escaped char
			continue
		case '[':
			// skip character class
			i++
			for i < len(pattern) {
				if pattern[i] == '\\' {
					i += 2
					continue
				}
				if pattern[i] == ']' {
					break
				}
				i++
			}
		case '(':
			// Push a new level
			stack = append(stack, level{names: make(map[string]bool)})
			// Check for named group (?<name>...)
			if i+3 < len(pattern) && pattern[i+1] == '?' && pattern[i+2] == '<' && pattern[i+3] != '=' && pattern[i+3] != '!' {
				// Extract the name
				nameStart := i + 3
				nameEnd := nameStart
				for nameEnd < len(pattern) && pattern[nameEnd] != '>' {
					nameEnd++
				}
				if nameEnd < len(pattern) {
					name := pattern[nameStart:nameEnd]
					// Check in parent level (the alternative containing this group)
					parent := &stack[len(stack)-2]
					if parent.names[name] {
						return fmt.Errorf("SyntaxError: Invalid regular expression: /%s/: Duplicate capture group name", pattern)
					}
					parent.names[name] = true
				}
			}
		case ')':
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		case '|':
			// Reset names at current nesting level - new alternative
			stack[len(stack)-1].names = make(map[string]bool)
		}
		i++
	}
	return nil
}

// isGoRegexpEscape returns true if \c is a valid Go regexp escape sequence
// that has the SAME meaning in both JS and Go regexps.
func isGoRegexpEscape(c byte) bool {
	switch c {
	case 'd', 'D', 'w', 'W', 's', 'S': // character classes
		return true
	case 'b', 'B': // word boundary
		return true
	case 'f', 'n', 'r', 't': // common escapes Go accepts
		return true
	}
	// Regex metacharacters escaped: \. \* \+ \? \( \) \[ \] \{ \} \| \^ \$ \\
	if isRegexpSpecial(c) {
		return true
	}
	return false
}

// isRegexpSpecial returns true if the byte is a regexp metacharacter.
func isRegexpSpecial(c byte) bool {
	switch c {
	case '.', '*', '+', '?', '(', ')', '[', ']', '{', '}', '|', '^', '$', '\\':
		return true
	}
	return false
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
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

func canonicalizeFlags(flags string) string {
	order := "dgimsuy"
	var result strings.Builder
	for _, c := range order {
		if strings.ContainsRune(flags, c) {
			result.WriteRune(c)
		}
	}
	return result.String()
}

func regexpCompile(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	obj := toObject(this)
	if obj == nil || obj.Internal == nil || obj.Internal["regexp"] == nil {
		return nil, fmt.Errorf("TypeError: Method RegExp.prototype.compile called on incompatible receiver")
	}

	patternArg := argAt(args, 0)
	flagsArg := argAt(args, 1)

	var pattern, flags string

	patternObj := toObject(patternArg)
	isRegExp := patternObj != nil && patternObj.Internal != nil && patternObj.Internal["regexp"] != nil

	if isRegExp {
		if flagsArg.Type != runtime.TypeUndefined {
			return nil, fmt.Errorf("TypeError: Cannot supply flags when constructing one RegExp from another")
		}
		if p, ok := patternObj.Internal["pattern"].(string); ok {
			pattern = p
		}
		if f, ok := patternObj.Internal["flags"].(string); ok {
			flags = f
		}
	} else {
		if patternArg.Type == runtime.TypeUndefined {
			pattern = ""
		} else {
			if patternArg.Type == runtime.TypeSymbol {
				return nil, fmt.Errorf("TypeError: Cannot convert a Symbol value to a string")
			}
			s, err := jsToString(patternArg)
			if err != nil {
				return nil, err
			}
			pattern = s
		}
		if flagsArg.Type == runtime.TypeUndefined {
			flags = ""
		} else {
			if flagsArg.Type == runtime.TypeSymbol {
				return nil, fmt.Errorf("TypeError: Cannot convert a Symbol value to a string")
			}
			s, err := jsToString(flagsArg)
			if err != nil {
				return nil, err
			}
			flags = s
		}
	}

	if err := validateRegExpFlags(flags); err != nil {
		return nil, err
	}

	flags = canonicalizeFlags(flags)

	if err := validateDuplicateNamedGroups(pattern); err != nil {
		return nil, err
	}

	// Validate pattern in unicode mode
	if strings.Contains(flags, "u") {
		// In unicode mode, lone { and backreferences like \2 are invalid
		for i := 0; i < len(pattern); i++ {
			if pattern[i] == '{' {
				return nil, fmt.Errorf("SyntaxError: Invalid regular expression: /%s/u: Lone quantifier brackets", pattern)
			}
			if pattern[i] == '\\' && i+1 < len(pattern) && pattern[i+1] >= '1' && pattern[i+1] <= '9' {
				return nil, fmt.Errorf("SyntaxError: Invalid regular expression: /%s/u: Invalid escape", pattern)
			}
			if pattern[i] == '\\' && i+1 < len(pattern) {
				i++ // skip escaped character
			}
		}
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

	setDataProp(obj, "source", runtime.NewString(pattern), false, false, true)
	setDataProp(obj, "flags", runtime.NewString(flags), false, false, true)
	setDataProp(obj, "global", runtime.NewBool(strings.Contains(flags, "g")), false, false, true)
	setDataProp(obj, "ignoreCase", runtime.NewBool(strings.Contains(flags, "i")), false, false, true)
	setDataProp(obj, "multiline", runtime.NewBool(strings.Contains(flags, "m")), false, false, true)
	setDataProp(obj, "sticky", runtime.NewBool(strings.Contains(flags, "y")), false, false, true)
	setDataProp(obj, "unicode", runtime.NewBool(strings.Contains(flags, "u")), false, false, true)

	if prop, ok := obj.Properties["lastIndex"]; ok && !prop.Writable {
		return nil, fmt.Errorf("TypeError: Cannot assign to read only property 'lastIndex'")
	}
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

// regexpSymbolSplit implements RegExp.prototype[@@split].
// Simplified implementation per spec 22.2.5.13.
func regexpSymbolSplit(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	rx := toObject(this)
	if rx == nil {
		return nil, fmt.Errorf("TypeError: RegExp.prototype[@@split] called on incompatible receiver")
	}

	s := argAt(args, 0).ToString()

	// Step 4: Get constructor C
	// Per spec: C = SpeciesConstructor(rx, %RegExp%)
	// For now, always use RegExp constructor

	// Step 6: Get flags
	flagsVal := rx.Get("flags")
	flags := ""
	if flagsVal != nil && flagsVal != runtime.Undefined {
		flags = flagsVal.ToString()
	}

	// Step 7-8: Add 'y' (sticky) flag if not present
	newFlags := flags
	if !strings.Contains(newFlags, "y") {
		newFlags += "y"
	}

	// Step 10: Construct splitter via RegExp constructor
	// Per spec: Construct(C, « rx, newFlags »)
	// This goes through the RegExp constructor, which calls IsRegExp(pattern)
	// reading Symbol.match - triggering side effects like compile("b")
	splitter, err := regexpConstructorCall(runtime.Undefined, []*runtime.Value{this, runtime.NewString(newFlags)})
	if err != nil {
		return nil, err
	}
	splitterObj := toObject(splitter)
	if splitterObj == nil {
		return nil, fmt.Errorf("TypeError: RegExp constructor returned non-object")
	}

	// Step 13: Handle limit
	var lim uint32
	limitArg := argAt(args, 1)
	if limitArg.Type == runtime.TypeUndefined {
		lim = 0xFFFFFFFF
	} else {
		// ToUint32 - this triggers valueOf which is the second test's side-effect
		limF, err := toNumberErr(limitArg)
		if err != nil {
			return nil, err
		}
		lim = uint32(limF)
	}

	if lim == 0 {
		return runtime.NewObject(newArray(nil)), nil
	}

	// Re-read the regexp from splitter (may have been recompiled by side effects)
	re := getRegExp(splitter)
	if re == nil {
		return runtime.NewObject(newArray(nil)), nil
	}

	if len(s) == 0 {
		// If string is empty, test if it matches
		match := re.FindStringIndex(s)
		if match != nil {
			return runtime.NewObject(newArray(nil)), nil
		}
		return runtime.NewObject(newArray([]*runtime.Value{runtime.NewString("")})), nil
	}

	// Split using the compiled regex
	var result []*runtime.Value
	p := 0 // end of last match
	for p <= len(s) {
		// Find match starting at position p or later
		loc := re.FindStringIndex(s[p:])
		if loc == nil {
			break
		}
		matchStart := p + loc[0]
		matchEnd := p + loc[1]

		if matchEnd == p {
			// Zero-length match at current position, advance
			p++
			continue
		}

		// Add substring before match
		result = append(result, runtime.NewString(s[p:matchStart]))
		if uint32(len(result)) >= lim {
			return runtime.NewObject(newArray(result)), nil
		}

		// Add capture groups
		submatches := re.FindStringSubmatch(s[p:])
		for i := 1; i < len(submatches); i++ {
			if submatches[i] == "" && i < len(submatches) {
				result = append(result, runtime.Undefined)
			} else {
				result = append(result, runtime.NewString(submatches[i]))
			}
			if uint32(len(result)) >= lim {
				return runtime.NewObject(newArray(result)), nil
			}
		}

		p = matchEnd
	}

	// Add remaining string
	result = append(result, runtime.NewString(s[p:]))

	return runtime.NewObject(newArray(result)), nil
}

// regexpSymbolMatch implements RegExp.prototype[@@match].
// Simplified implementation.
func regexpSymbolMatch(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	rx := toObject(this)
	if rx == nil {
		return nil, fmt.Errorf("TypeError: RegExp.prototype[@@match] called on incompatible receiver")
	}

	s := argAt(args, 0).ToString()
	re := getRegExp(this)
	if re == nil {
		return runtime.Null, nil
	}

	global := rx.Get("global")
	if global != nil && global.Bool {
		// Global match: find all matches
		var matches []*runtime.Value
		lastIndex := 0
		for {
			loc := re.FindStringIndex(s[lastIndex:])
			if loc == nil {
				break
			}
			matchStr := s[lastIndex+loc[0] : lastIndex+loc[1]]
			matches = append(matches, runtime.NewString(matchStr))
			if loc[0] == loc[1] {
				lastIndex += loc[1] + 1
			} else {
				lastIndex += loc[1]
			}
			if lastIndex > len(s) {
				break
			}
		}
		if len(matches) == 0 {
			return runtime.Null, nil
		}
		return runtime.NewObject(newArray(matches)), nil
	}

	// Non-global: same as exec
	return regexpExec(this, args)
}
