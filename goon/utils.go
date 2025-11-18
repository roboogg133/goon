package goon

import (
	"reflect"
	"strconv"
	"strings"
)

func recognizeType(s string) (reflect.Value, error) {
	switch {
	case strings.ContainsAny(s, ",") && !strings.HasPrefix(s, "\""):

		split := strings.Split(s, ",")
		var elems []reflect.Value

		var typeToSet reflect.Type
		reflectAny := reflect.TypeOf((*any)(nil)).Elem()
		for _, v := range split {
			v = strings.TrimSpace(v)
			v, _ = strings.CutPrefix(v, "\"")
			v, _ = strings.CutSuffix(v, "\"")
			elem, err := recognizeType(v)
			if err != nil {
				return reflect.Value{}, err
			}
			if typeToSet != reflectAny {
				if typeToSet != elem.Type() {

					typeToSet = reflectAny
				} else {

					typeToSet = elem.Type()
				}
			}
			elems = append(elems, elem)
		}

		sliceType := reflect.SliceOf(typeToSet)

		sliceValue := reflect.MakeSlice(sliceType, 0, len(elems))

		for _, e := range elems {
			sliceValue = reflect.Append(sliceValue, e)
		}

		return sliceValue, nil
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

	case s == "":
		return reflect.ValueOf(nil), nil
	default:
		s, _ = strings.CutPrefix(s, "\"")
		s, _ = strings.CutSuffix(s, "\"")
		return reflect.ValueOf(s), nil
	}
}
