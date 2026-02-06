package builtins

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/example/jsgo/runtime"
)

var DatePrototype *runtime.Object

func createDateConstructor(objProto *runtime.Object) (*runtime.Object, *runtime.Object) {
	proto := runtime.NewOrdinaryObject(objProto)
	DatePrototype = proto

	// Prototype methods
	setMethod(proto, "getTime", 0, dateGetTime)
	setMethod(proto, "getFullYear", 0, dateGetFullYear)
	setMethod(proto, "getMonth", 0, dateGetMonth)
	setMethod(proto, "getDate", 0, dateGetDate)
	setMethod(proto, "getHours", 0, dateGetHours)
	setMethod(proto, "getMinutes", 0, dateGetMinutes)
	setMethod(proto, "getSeconds", 0, dateGetSeconds)
	setMethod(proto, "getMilliseconds", 0, dateGetMilliseconds)
	setMethod(proto, "getTimezoneOffset", 0, dateGetTimezoneOffset)
	setMethod(proto, "toString", 0, dateToString)
	setMethod(proto, "toDateString", 0, dateToDateString)
	setMethod(proto, "toTimeString", 0, dateToTimeString)
	setMethod(proto, "toISOString", 0, dateToISOString)
	setMethod(proto, "toJSON", 1, dateToJSON)
	setMethod(proto, "toLocaleDateString", 0, dateToLocaleDateString)
	setMethod(proto, "toLocaleTimeString", 0, dateToLocaleTimeString)
	setMethod(proto, "toLocaleString", 0, dateToLocaleString)
	setMethod(proto, "valueOf", 0, dateValueOf)
	setMethod(proto, "getDay", 0, dateGetDay)
	setMethod(proto, "getUTCFullYear", 0, dateGetUTCFullYear)
	setMethod(proto, "getUTCMonth", 0, dateGetUTCMonth)
	setMethod(proto, "getUTCDate", 0, dateGetUTCDate)
	setMethod(proto, "getUTCHours", 0, dateGetUTCHours)
	setMethod(proto, "getUTCMinutes", 0, dateGetUTCMinutes)
	setMethod(proto, "getUTCSeconds", 0, dateGetUTCSeconds)
	setMethod(proto, "getUTCMilliseconds", 0, dateGetUTCMilliseconds)
	setMethod(proto, "getUTCDay", 0, dateGetUTCDay)
	setMethod(proto, "setTime", 1, dateSetTime)
	setMethod(proto, "setFullYear", 3, dateSetFullYear)
	setMethod(proto, "setMonth", 2, dateSetMonth)
	setMethod(proto, "setDate", 1, dateSetDate)
	setMethod(proto, "setHours", 4, dateSetHours)
	setMethod(proto, "setMinutes", 3, dateSetMinutes)
	setMethod(proto, "setSeconds", 2, dateSetSeconds)
	setMethod(proto, "setMilliseconds", 1, dateSetMilliseconds)
	setMethod(proto, "toUTCString", 0, dateToUTCString)

	// Annex B methods
	setMethod(proto, "getYear", 0, dateGetYear)
	setMethod(proto, "setYear", 1, dateSetYear)
	setMethod(proto, "toGMTString", 0, dateToUTCString) // toGMTString is an alias for toUTCString

	// Constructor: Date() as function returns string, new Date() creates object
	ctor := newFuncObject("Date", 7, dateCall)
	ctor.Constructor = dateConstruct

	// Static methods
	setMethod(ctor, "now", 0, dateNow)
	setMethod(ctor, "parse", 1, dateParse)
	setMethod(ctor, "UTC", 7, dateUTC)

	setDataProp(ctor, "prototype", runtime.NewObject(proto), false, false, false)
	setDataProp(proto, "constructor", runtime.NewObject(ctor), true, false, true)

	return ctor, proto
}

// dateCall is invoked when Date() is called as a function (without new)
func dateCall(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewString(time.Now().Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")), nil
}

// dateConstruct is invoked for new Date(...)
func dateConstruct(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	var t time.Time

	if len(args) == 0 {
		t = time.Now()
	} else if len(args) == 1 {
		arg := args[0]
		if arg.Type == runtime.TypeString {
			parsed, err := parseDate(arg.Str)
			if err != nil {
				t = time.Time{} // Invalid Date
				return makeDateObject(this, t, true), nil
			}
			t = parsed
		} else {
			ms := toNumber(arg)
			if math.IsNaN(ms) || math.IsInf(ms, 0) {
				return makeDateObject(this, time.Time{}, true), nil
			}
			t = time.UnixMilli(int64(ms))
		}
	} else {
		// new Date(year, month, day, hours, minutes, seconds, ms)
		year := int(toNumber(args[0]))
		month := time.Month(int(toNumber(argAt(args, 1))) + 1) // JS months 0-based
		day := 1
		if len(args) > 2 {
			day = int(toNumber(args[2]))
		}
		hour := 0
		if len(args) > 3 {
			hour = int(toNumber(args[3]))
		}
		min := 0
		if len(args) > 4 {
			min = int(toNumber(args[4]))
		}
		sec := 0
		if len(args) > 5 {
			sec = int(toNumber(args[5]))
		}
		msec := 0
		if len(args) > 6 {
			msec = int(toNumber(args[6]))
		}
		// If year is 0-99, add 1900
		if year >= 0 && year <= 99 {
			year += 1900
		}
		t = time.Date(year, month, day, hour, min, sec, msec*1e6, time.Local)
	}

	return makeDateObject(this, t, false), nil
}

func makeDateObject(this *runtime.Value, t time.Time, invalid bool) *runtime.Value {
	if this != nil && this.Type == runtime.TypeObject && this.Object != nil {
		if this.Object.Internal == nil {
			this.Object.Internal = make(map[string]interface{})
		}
		this.Object.Internal["DateValue"] = t
		this.Object.Internal["DateInvalid"] = invalid
		return this
	}
	obj := runtime.NewOrdinaryObject(DatePrototype)
	obj.Internal = map[string]interface{}{
		"DateValue":   t,
		"DateInvalid": invalid,
	}
	return runtime.NewObject(obj)
}

func getDateValue(this *runtime.Value) (time.Time, bool) {
	if this == nil || this.Type != runtime.TypeObject || this.Object == nil {
		return time.Time{}, true
	}
	inv, _ := this.Object.Internal["DateInvalid"].(bool)
	if inv {
		return time.Time{}, true
	}
	t, ok := this.Object.Internal["DateValue"].(time.Time)
	if !ok {
		return time.Time{}, true
	}
	return t, false
}

func getDateMs(this *runtime.Value) float64 {
	t, invalid := getDateValue(this)
	if invalid {
		return math.NaN()
	}
	return float64(t.UnixMilli())
}

// Static methods

func dateNow(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewNumber(float64(time.Now().UnixMilli())), nil
}

func dateParse(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	s := argAt(args, 0).ToString()
	t, err := parseDate(s)
	if err != nil {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.UnixMilli())), nil
}

func dateUTC(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	year := int(toNumber(argAt(args, 0)))
	month := time.Month(int(toNumber(argAt(args, 1))) + 1)
	day := 1
	if len(args) > 2 {
		day = int(toNumber(args[2]))
	}
	hour := 0
	if len(args) > 3 {
		hour = int(toNumber(args[3]))
	}
	min := 0
	if len(args) > 4 {
		min = int(toNumber(args[4]))
	}
	sec := 0
	if len(args) > 5 {
		sec = int(toNumber(args[5]))
	}
	msec := 0
	if len(args) > 6 {
		msec = int(toNumber(args[6]))
	}
	if year >= 0 && year <= 99 {
		year += 1900
	}
	t := time.Date(year, month, day, hour, min, sec, msec*1e6, time.UTC)
	return runtime.NewNumber(float64(t.UnixMilli())), nil
}

// Prototype methods - getters

func dateGetTime(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewNumber(getDateMs(this)), nil
}

func dateGetFullYear(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.Year())), nil
}

func dateGetMonth(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.Month() - 1)), nil
}

func dateGetDate(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.Day())), nil
}

func dateGetDay(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.Weekday())), nil
}

func dateGetHours(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.Hour())), nil
}

func dateGetMinutes(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.Minute())), nil
}

func dateGetSeconds(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.Second())), nil
}

func dateGetMilliseconds(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.Nanosecond() / 1e6)), nil
}

func dateGetTimezoneOffset(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	_, offset := t.Zone()
	return runtime.NewNumber(float64(-offset / 60)), nil
}

// UTC getters

func dateGetUTCFullYear(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.UTC().Year())), nil
}

func dateGetUTCMonth(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.UTC().Month() - 1)), nil
}

func dateGetUTCDate(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.UTC().Day())), nil
}

func dateGetUTCHours(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.UTC().Hour())), nil
}

func dateGetUTCMinutes(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.UTC().Minute())), nil
}

func dateGetUTCSeconds(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.UTC().Second())), nil
}

func dateGetUTCMilliseconds(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.UTC().Nanosecond() / 1e6)), nil
}

func dateGetUTCDay(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.UTC().Weekday())), nil
}

// Setters

func dateSetTime(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	ms := toNumber(argAt(args, 0))
	if math.IsNaN(ms) {
		if this.Type == runtime.TypeObject && this.Object != nil {
			this.Object.Internal["DateInvalid"] = true
		}
		return runtime.NaN, nil
	}
	t := time.UnixMilli(int64(ms))
	if this.Type == runtime.TypeObject && this.Object != nil {
		this.Object.Internal["DateValue"] = t
		this.Object.Internal["DateInvalid"] = false
	}
	return runtime.NewNumber(float64(t.UnixMilli())), nil
}

func dateSetFullYear(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		t = time.Now()
	}
	year := int(toNumber(argAt(args, 0)))
	month := t.Month()
	if len(args) > 1 {
		month = time.Month(int(toNumber(args[1])) + 1)
	}
	day := t.Day()
	if len(args) > 2 {
		day = int(toNumber(args[2]))
	}
	t = time.Date(year, month, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	if this.Type == runtime.TypeObject && this.Object != nil {
		this.Object.Internal["DateValue"] = t
		this.Object.Internal["DateInvalid"] = false
	}
	return runtime.NewNumber(float64(t.UnixMilli())), nil
}

func dateSetMonth(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	month := time.Month(int(toNumber(argAt(args, 0))) + 1)
	day := t.Day()
	if len(args) > 1 {
		day = int(toNumber(args[1]))
	}
	t = time.Date(t.Year(), month, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	if this.Type == runtime.TypeObject && this.Object != nil {
		this.Object.Internal["DateValue"] = t
	}
	return runtime.NewNumber(float64(t.UnixMilli())), nil
}

func dateSetDate(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	day := int(toNumber(argAt(args, 0)))
	t = time.Date(t.Year(), t.Month(), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	if this.Type == runtime.TypeObject && this.Object != nil {
		this.Object.Internal["DateValue"] = t
	}
	return runtime.NewNumber(float64(t.UnixMilli())), nil
}

func dateSetHours(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	hour := int(toNumber(argAt(args, 0)))
	min := t.Minute()
	if len(args) > 1 {
		min = int(toNumber(args[1]))
	}
	sec := t.Second()
	if len(args) > 2 {
		sec = int(toNumber(args[2]))
	}
	ms := t.Nanosecond() / 1e6
	if len(args) > 3 {
		ms = int(toNumber(args[3]))
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), hour, min, sec, ms*1e6, t.Location())
	if this.Type == runtime.TypeObject && this.Object != nil {
		this.Object.Internal["DateValue"] = t
	}
	return runtime.NewNumber(float64(t.UnixMilli())), nil
}

func dateSetMinutes(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	min := int(toNumber(argAt(args, 0)))
	sec := t.Second()
	if len(args) > 1 {
		sec = int(toNumber(args[1]))
	}
	ms := t.Nanosecond() / 1e6
	if len(args) > 2 {
		ms = int(toNumber(args[2]))
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), min, sec, ms*1e6, t.Location())
	if this.Type == runtime.TypeObject && this.Object != nil {
		this.Object.Internal["DateValue"] = t
	}
	return runtime.NewNumber(float64(t.UnixMilli())), nil
}

func dateSetSeconds(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	sec := int(toNumber(argAt(args, 0)))
	ms := t.Nanosecond() / 1e6
	if len(args) > 1 {
		ms = int(toNumber(args[1]))
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), sec, ms*1e6, t.Location())
	if this.Type == runtime.TypeObject && this.Object != nil {
		this.Object.Internal["DateValue"] = t
	}
	return runtime.NewNumber(float64(t.UnixMilli())), nil
}

func dateSetMilliseconds(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	ms := int(toNumber(argAt(args, 0)))
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), ms*1e6, t.Location())
	if this.Type == runtime.TypeObject && this.Object != nil {
		this.Object.Internal["DateValue"] = t
	}
	return runtime.NewNumber(float64(t.UnixMilli())), nil
}

// Annex B methods

func dateGetYear(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NaN, nil
	}
	return runtime.NewNumber(float64(t.Year() - 1900)), nil
}

func dateSetYear(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		t = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.Local)
	}
	yearArg := toNumber(argAt(args, 0))
	if math.IsNaN(yearArg) {
		if this.Type == runtime.TypeObject && this.Object != nil {
			this.Object.Internal["DateInvalid"] = true
		}
		return runtime.NaN, nil
	}
	year := int(yearArg)
	if year >= 0 && year <= 99 {
		year += 1900
	}
	t = time.Date(year, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	if this.Type == runtime.TypeObject && this.Object != nil {
		this.Object.Internal["DateValue"] = t
		this.Object.Internal["DateInvalid"] = false
	}
	return runtime.NewNumber(float64(t.UnixMilli())), nil
}

// toString methods

func dateToString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NewString("Invalid Date"), nil
	}
	return runtime.NewString(formatDateString(t)), nil
}

func dateToDateString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NewString("Invalid Date"), nil
	}
	return runtime.NewString(t.Format("Mon Jan 02 2006")), nil
}

func dateToTimeString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NewString("Invalid Date"), nil
	}
	return runtime.NewString(t.Format("15:04:05 GMT-0700 (MST)")), nil
}

func dateToISOString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return nil, fmt.Errorf("RangeError: Invalid time value")
	}
	return runtime.NewString(t.UTC().Format("2006-01-02T15:04:05.000Z")), nil
}

func dateToJSON(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return dateToISOString(this, args)
}

func dateToLocaleDateString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NewString("Invalid Date"), nil
	}
	return runtime.NewString(t.Format("1/2/2006")), nil
}

func dateToLocaleTimeString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NewString("Invalid Date"), nil
	}
	return runtime.NewString(t.Format("3:04:05 PM")), nil
}

func dateToLocaleString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NewString("Invalid Date"), nil
	}
	return runtime.NewString(t.Format("1/2/2006, 3:04:05 PM")), nil
}

func dateToUTCString(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	t, inv := getDateValue(this)
	if inv {
		return runtime.NewString("Invalid Date"), nil
	}
	return runtime.NewString(t.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")), nil
}

func dateValueOf(this *runtime.Value, args []*runtime.Value) (*runtime.Value, error) {
	return runtime.NewNumber(getDateMs(this)), nil
}

// Helpers

func formatDateString(t time.Time) string {
	return t.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("invalid date")
	}

	formats := []string{
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000-07:00",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
		"Mon Jan 02 2006 15:04:05 GMT-0700 (MST)",
		"Mon Jan 02 2006 15:04:05 GMT-0700",
		"Mon, 02 Jan 2006 15:04:05 GMT",
		"January 2, 2006",
		"Jan 2, 2006",
		"1/2/2006",
		"2006/01/02",
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
	}

	for _, layout := range formats {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date: %s", s)
}
