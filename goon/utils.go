package goon

import (
	"reflect"
	"strconv"
	"strings"
)

func recognizeType(s string) (reflect.Value, error) {
	switch {
	case s == "true":
		return reflect.ValueOf(true), nil
	case s == "false":
		return reflect.ValueOf(false), nil
	case s == "null":
		return reflect.ValueOf(nil), nil
	case strings.ContainsAny(s, "1234567890") && !strings.ContainsAny(s, "\"."):
		i, err := strconv.Atoi(s)
		return reflect.ValueOf(i), err
	case strings.ContainsAny(s, "1234567890") && !strings.Contains(s, "\"") && strings.ContainsAny(s, "."):
		f, err := strconv.ParseFloat(s, 64)
		return reflect.ValueOf(f), err
	default:
		s, _ = strings.CutPrefix(s, "\"")
		s, _ = strings.CutSuffix(s, "\"")
		return reflect.ValueOf(s), nil
	}
}
