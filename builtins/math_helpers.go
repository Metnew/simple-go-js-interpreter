package builtins

import (
	"math"
	"strconv"
	"strings"
)

func math_NaN() float64              { return math.NaN() }
func math_Inf(sign int) float64      { return math.Inf(sign) }
func isNaN(f float64) bool           { return math.IsNaN(f) }
func isInf(f float64, sign int) bool { return math.IsInf(f, sign) }
func math_Floor(f float64) float64   { return math.Floor(f) }
func math_Abs(f float64) float64     { return math.Abs(f) }
func math_Min(a, b float64) float64  { return math.Min(a, b) }
func math_Max(a, b float64) float64  { return math.Max(a, b) }

func parseStringToNumber(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if s == "Infinity" || s == "+Infinity" {
		return math.Inf(1)
	}
	if s == "-Infinity" {
		return math.Inf(-1)
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return math.NaN()
	}
	return f
}
