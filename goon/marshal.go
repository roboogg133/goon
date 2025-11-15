package goon

import (
	"bufio"
	"fmt"
	"reflect"
	"strings"
)

func Marshal(v any) ([]byte, error) {

	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)

	if rt.Kind() != reflect.Struct && rt.Kind() != reflect.Map {
		return nil, fmt.Errorf("Marshal only accepts struct types")
	}

	return marshalStruct(rv)
}

type entry struct {
	Name  string
	Value reflect.Value
}

func normalize(v reflect.Value) ([]entry, error) {
	t := v.Type()

	switch v.Kind() {

	case reflect.Struct:
		var out []entry
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := f.Tag.Get("toon")
			if name == "" {
				continue
			}
			out = append(out, entry{
				Name:  name,
				Value: v.Field(i),
			})
		}
		return out, nil

	case reflect.Map:
		var out []entry
		for _, key := range v.MapKeys() {
			k := key.Interface()
			name := fmt.Sprint(k)
			out = append(out, entry{
				Name:  name,
				Value: v.MapIndex(key),
			})
		}
		return out, nil

	default:
		return nil, fmt.Errorf("goon: unsupported kind %s", v.Kind())
	}
}

func marshalStruct(v reflect.Value) ([]byte, error) {
	entries, err := normalize(v)
	if err != nil {
		return nil, err
	}

	var final strings.Builder

	for _, e := range entries {

		value := e.Value
		valKind := value.Kind()

		if valKind == reflect.Interface {
			value = value.Elem()
			valKind = value.Kind()
		}

		if valKind == reflect.Pointer {
			if value.IsNil() {
				continue
			}
			value = value.Elem()
			valKind = value.Kind()
		}

		var a string

		switch valKind {
		case reflect.String:
			s := fmt.Sprint(value.Interface().(string))

			if strings.ContainsAny(s, "0123456789:") {
				s = fmt.Sprintf("\"%s\"", s)
			} else if s == "" {
				s = "\"\""
			} else if s == "true" {
				s = "\"true\""
			} else if s == "false" {
				s = "\"false\""
			}

			a = fmt.Sprintf("%s : %s\n", e.Name, s)
		case reflect.Int:
			a = fmt.Sprintf("%s : %s\n", e.Name, fmt.Sprint(value.Interface().(int)))
		case reflect.Float32:
			a = fmt.Sprintf("%s : %s\n", e.Name, fmt.Sprint(value.Interface().(float32)))
		case reflect.Float64:
			a = fmt.Sprintf("%s : %s\n", e.Name, fmt.Sprint(value.Interface().(float64)))
		case reflect.Bool:
			a = fmt.Sprintf("%s : %s\n", e.Name, fmt.Sprint(value.Interface().(bool)))
		case reflect.Struct, reflect.Map:
			content, err := marshalStruct(value)
			if err != nil {
				return nil, err
			}

			str := string(content)

			var builder strings.Builder
			scanner := bufio.NewScanner(strings.NewReader(str))
			for scanner.Scan() {
				builder.WriteString("  ")
				builder.WriteString(scanner.Text())
				builder.WriteByte('\n')
			}

			a = fmt.Sprintf("%s :\n%s", e.Name, builder.String())

		case reflect.Array, reflect.Slice:
			var builder strings.Builder
			builder.WriteString(fmt.Sprintf("%s[%d]: ", e.Name, value.Len()))

			s, err := arrayMarshal(value)
			if err != nil {
				return nil, err
			}
			builder.WriteString(s)

			final.WriteString(builder.String())

		default:
			return nil, fmt.Errorf("goon: unknown type kind %s", valKind)
		}

		final.WriteString(a)
	}

	return []byte(final.String()), nil
}

func arrayMarshal(value reflect.Value) (string, error) {
	var builder strings.Builder

	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)
		valKind := elem.Kind()

		if valKind == reflect.Pointer {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
			valKind = elem.Kind()
		}

		if valKind == reflect.Interface {
			return arrayMixMarshal(value)
		}

		switch valKind {
		case reflect.String:
			s := elem.String()
			if strings.ContainsAny(s, "0123456789:") {
				s = fmt.Sprintf("\"%s\"", s)
			} else if s == "" {
				s = "\"\""
			} else if s == "true" {
				s = "\"true\""
			} else if s == "false" {
				s = "\"false\""
			}
			if value.Len()-1 == i {
				builder.WriteString(s)
				builder.WriteByte('\n')
			} else {
				builder.WriteString(s + ",")
			}

		case reflect.Int:
			if value.Len()-1 == i {
				builder.WriteString(fmt.Sprint(elem.Int()))
				builder.WriteByte('\n')
			} else {
				builder.WriteString(fmt.Sprint(elem.Int()) + ",")
			}

		case reflect.Float32, reflect.Float64:
			if value.Len()-1 == i {
				builder.WriteString(fmt.Sprint(elem.Float()))
			} else {
				builder.WriteString(fmt.Sprint(elem.Float()) + ",")
			}
		case reflect.Bool:
			if value.Len()-1 == i {
				builder.WriteString(fmt.Sprint(elem.Bool()))
				builder.WriteByte('\n')
			} else {
				builder.WriteString(fmt.Sprint(elem.Bool()) + ",")
			}

		case reflect.Struct, reflect.Map:
			content, err := marshalStruct(elem)
			if err != nil {
				return "", err
			}

			str := string(content)

			var indented strings.Builder
			scanner := bufio.NewScanner(strings.NewReader(str))
			lock := true
			for scanner.Scan() {
				if lock {
					lock = false
				} else {
					indented.WriteString("    ")
				}
				indented.WriteString(scanner.Text())
				indented.WriteByte('\n')
			}

			builder.WriteString("  - ")
			builder.WriteString(indented.String())

		default:
			return "", fmt.Errorf("goon: unknown type kind %s", valKind)
		}
	}

	return builder.String(), nil

}

func arrayMixMarshal(value reflect.Value) (string, error) {
	var builder strings.Builder

	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)
		valKind := elem.Kind()

		if valKind == reflect.Pointer {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
			valKind = elem.Kind()
		}

		if valKind == reflect.Interface {
			elem = elem.Elem()
			valKind = elem.Kind()
		}

		switch valKind {
		case reflect.String:
			s := elem.String()
			if strings.ContainsAny(s, "0123456789:") {
				s = fmt.Sprintf("\"%s\"", s)
			} else if s == "" {
				s = "\"\""
			} else if s == "true" {
				s = "\"true\""
			} else if s == "false" {
				s = "\"false\""
			}
			if i == 0 {
				builder.WriteByte('\n')
			}
			builder.WriteString("  - " + s + "\n")

		case reflect.Int:
			if i == 0 {
				builder.WriteByte('\n')
			}
			builder.WriteString("  - " + fmt.Sprint(elem.Int()) + "\n")

		case reflect.Float32, reflect.Float64:
			if i == 0 {
				builder.WriteByte('\n')
			}
			builder.WriteString("  - " + fmt.Sprint(elem.Float()) + "\n")

		case reflect.Bool:
			if i == 0 {
				builder.WriteByte('\n')
			}
			builder.WriteString("  - " + fmt.Sprint(elem.Bool()) + "\n")

		case reflect.Struct, reflect.Map:
			content, err := marshalStruct(elem)
			if err != nil {
				return "", err
			}

			str := string(content)

			var indented strings.Builder
			scanner := bufio.NewScanner(strings.NewReader(str))
			lock := true
			for scanner.Scan() {
				if lock {
					lock = false
				} else {
					indented.WriteString("    ")
				}
				indented.WriteString(scanner.Text())
				indented.WriteByte('\n')
			}

			if i == 0 {
				builder.WriteByte('\n')
			}
			builder.WriteString("  - ")
			builder.WriteString(indented.String())

		default:
			return "", fmt.Errorf("goon: unknown type kind %s", valKind)
		}
	}

	return builder.String(), nil
}
