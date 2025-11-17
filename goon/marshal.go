package goon

import (
	"bufio"
	"fmt"
	"reflect"
	"slices"
	"strings"
)

func Marshal(v any) ([]byte, error) {

	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)

	s, err := marshalSolve(rv, rt)
	var data []byte
	return fmt.Append(data, s), err
}

func marshalSolve(rv reflect.Value, rt reflect.Type) (string, error) {

	kind := rt.Kind()
	if rt.Kind() == reflect.Interface {
		rv = rv.Elem()
		kind = rv.Kind()
	}

	switch kind {
	case reflect.Struct, reflect.Map:

		sttruc, err := marshalStruct(rv)
		return string(sttruc), err
	case reflect.String:
		s := rv.String()
		if strings.ContainsAny(s, "0123456789:,{}[]\"|\\-\t") {
			s = fmt.Sprintf("\"%s\"", s)
		} else if s == "" {
			s = "\"\""
		} else if s == "true" || s == "false" || s == "null" {
			s = fmt.Sprintf("\"%s\"", s)
		}
		if strings.HasPrefix(s, " ") || strings.HasSuffix(s, " ") {
			s = fmt.Sprintf("\"%s\"", s)
		}

		return s, nil

	case reflect.Int:
		return fmt.Sprint(rv.Int()), nil
	case reflect.Float64, reflect.Float32:
		return fmt.Sprint(rv.Float()), nil
	case reflect.Array, reflect.Slice:
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("[%d]", rv.Len()))
		s, error := arrayMarshal(rv)
		builder.WriteString(s)
		return builder.String(), error

	default:
		return "", fmt.Errorf("goon: invalid type for marshal: %s", rt.Kind())
	}
}

type entry struct {
	Name      string
	Value     reflect.Value
	OmitEmpty bool
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
			var omit bool
			if f.Tag.Get("omitempty") != "" {
				omit = true
			}

			out = append(out, entry{
				Name:      name,
				Value:     v.Field(i),
				OmitEmpty: omit,
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
			if value.IsNil() && !e.OmitEmpty {
				final.WriteString(fmt.Sprintf("%s : %s\n", e.Name, "null"))
			}
			value = value.Elem()
			valKind = value.Kind()
		}

		var a string

		switch valKind {
		case reflect.String:
			s := value.String()
			if strings.ContainsAny(s, "0123456789:,{}[]\"|\\-\t") {
				s = fmt.Sprintf("\"%s\"", s)
			} else if s == "" {
				s = "\"\""
			} else if s == "true" || s == "false" || s == "null" {
				s = fmt.Sprintf("\"%s\"", s)
			}
			if strings.HasPrefix(s, " ") || strings.HasSuffix(s, " ") {
				s = fmt.Sprintf("\"%s\"", s)
			}

			a = fmt.Sprintf("%s : %s\n", e.Name, s)
		case reflect.Int:
			a = fmt.Sprintf("%s : %s\n", e.Name, fmt.Sprint(value.Int()))
		case reflect.Float32, reflect.Float64:
			a = fmt.Sprintf("%s : %s\n", e.Name, fmt.Sprint(value.Float()))
		case reflect.Bool:
			a = fmt.Sprintf("%s : %s\n", e.Name, fmt.Sprint(value.Bool()))
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
			builder.WriteString(fmt.Sprintf("%s[%d]", e.Name, value.Len()))

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

	if value.Len() == 0 {
		builder.WriteString(": ")
	}

	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)
		valKind := elem.Kind()

		if valKind == reflect.Pointer {
			if value.IsNil() {
				if value.Len()-1 == i {
					builder.WriteString("null")
					builder.WriteByte('\n')
				} else {
					builder.WriteString("null" + ",")
				}
			}
			elem = elem.Elem()
			valKind = elem.Kind()
		}

		if valKind == reflect.Array || valKind == reflect.Slice || valKind == reflect.Interface || valKind == reflect.Map || valKind == reflect.Struct {
			return arrayMixMarshal(value)
		}
		switch valKind {
		case reflect.String:
			if i == 0 {
				builder.WriteString(": ")
			}
			s := elem.String()
			if strings.ContainsAny(s, "0123456789:,{}[]\"|\\-\t") {
				s = fmt.Sprintf("\"%s\"", s)
			} else if s == "" {
				s = "\"\""
			} else if s == "true" || s == "false" || s == "null" {
				s = fmt.Sprintf("\"%s\"", s)
			}
			if strings.HasPrefix(s, " ") || strings.HasSuffix(s, " ") {
				s = fmt.Sprintf("\"%s\"", s)
			}
			if value.Len()-1 == i {
				builder.WriteString(s)
				builder.WriteByte('\n')
			} else {
				builder.WriteString(s + ",")
			}

		case reflect.Int:
			if i == 0 {
				builder.WriteString(": ")
			}
			if value.Len()-1 == i {
				builder.WriteString(fmt.Sprint(elem.Int()))
				builder.WriteByte('\n')
			} else {
				builder.WriteString(fmt.Sprint(elem.Int()) + ",")
			}

		case reflect.Float32, reflect.Float64:
			if i == 0 {
				builder.WriteString(": ")
			}
			if value.Len()-1 == i {
				builder.WriteString(fmt.Sprint(elem.Float()))
			} else {
				builder.WriteString(fmt.Sprint(elem.Float()) + ",")
			}
		case reflect.Bool:
			if i == 0 {
				builder.WriteString(": ")
			}
			if value.Len()-1 == i {
				builder.WriteString(fmt.Sprint(elem.Bool()))
				builder.WriteByte('\n')
			} else {
				builder.WriteString(fmt.Sprint(elem.Bool()) + ",")
			}

		default:
			return "", fmt.Errorf("goon: unknown type kind %s", valKind)
		}
	}

	return builder.String(), nil

}

func arrayMixMarshal(value reflect.Value) (string, error) {
	var builder strings.Builder

	if value.Len() == 0 {
		builder.WriteString(":\n")
	}

	for i := 0; i < value.Len(); i++ {
		elem := value.Index(i)
		valKind := elem.Kind()

		if valKind == reflect.Pointer {
			if value.IsNil() {
				builder.WriteString("  - null\n")
			}
			elem = elem.Elem()
			valKind = elem.Kind()
		}

		if valKind == reflect.Interface {
			elem = elem.Elem()
			valKind = elem.Kind()
		}

		if valKind == reflect.Struct || valKind == reflect.Map {
			str, err := doTheCSVThingORNothing(value)
			if err == nil {
				return str, nil
			}
		}
		switch valKind {
		case reflect.String:
			s := elem.String()
			if strings.ContainsAny(s, "0123456789:,{}[]\"|\\-\t") {
				s = fmt.Sprintf("\"%s\"", s)
			} else if s == "" {
				s = "\"\""
			} else if s == "true" || s == "false" || s == "null" {
				s = fmt.Sprintf("\"%s\"", s)
			}
			if strings.HasPrefix(s, " ") || strings.HasSuffix(s, " ") {
				s = fmt.Sprintf("\"%s\"", s)
			}

			if i == 0 {
				builder.WriteRune(':')
				builder.WriteByte('\n')
			}

			builder.WriteString("  - " + s + "\n")

		case reflect.Int:
			if i == 0 {
				builder.WriteRune(':')
				builder.WriteByte('\n')
			}
			builder.WriteString("  - " + fmt.Sprint(elem.Int()) + "\n")

		case reflect.Float32, reflect.Float64:
			if i == 0 {
				builder.WriteRune(':')
				builder.WriteByte('\n')
			}
			builder.WriteString("  - " + fmt.Sprint(elem.Float()) + "\n")

		case reflect.Bool:
			if i == 0 {
				builder.WriteRune(':')
				builder.WriteByte('\n')
			}
			builder.WriteString("  - " + fmt.Sprint(elem.Bool()) + "\n")

		case reflect.Struct, reflect.Map:
			if i == 0 {
				entrys, err := normalize(elem)
				if err != nil {
					return "", err
				}
				if value.Index(i).Type().Kind() != reflect.Interface {
					builder.WriteRune('{')
					total := len(entrys) - 1
					for i, e := range entrys {
						builder.WriteString(e.Name)
						if i == total {
							continue
						}
						builder.WriteString(",")
					}
					builder.WriteRune('}')
					builder.WriteRune(':')
					builder.WriteByte('\n')
				} else {
					builder.WriteRune(':')
					builder.WriteByte('\n')
				}
			}

			entrys, err := normalize(elem)
			if err != nil {
				return "", err
			}
			if value.Index(i).Type().Kind() != reflect.Interface {
				total := len(entrys) - 1
				builder.WriteString("  ")
				for i, e := range entrys {

					s, err := marshalSolve(e.Value, e.Value.Type())
					if err != nil {
						return "", err
					}
					builder.WriteString(s)
					if i != total {
						builder.WriteString(",")
					}
				}
				builder.WriteByte('\n')
			} else {
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
			}

		case reflect.Array, reflect.Slice:
			var build strings.Builder

			if i == 0 {
				builder.WriteRune(':')
				builder.WriteByte('\n')
			}
			build.WriteString(fmt.Sprintf("  - [%d]", elem.Len()))

			s, err := arrayMarshal(elem)
			if err != nil {
				return "", err
			}
			build.WriteString(s)
			builder.WriteString(build.String())

		default:
			return "", fmt.Errorf("goon: unknown type kind %s", valKind)
		}
	}

	return builder.String(), nil
}

func doTheCSVThingORNothing(rv reflect.Value) (string, error) {
	var builder strings.Builder

	var allnames []string

	var allEntrys []map[string]reflect.Value

	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		elemKind := elem.Kind()

		if elemKind == reflect.Interface {
			elem = elem.Elem()
			elemKind = elem.Kind()
		}

		if elemKind != reflect.Map && elemKind != reflect.Struct {
			return "", fmt.Errorf("goon: can't do the csv thing ")
		}
		entrys, err := normalize(elem)
		if err != nil {
			return "", err
		}

		entryM := make(map[string]reflect.Value)
		for _, e := range entrys {
			if !slices.Contains(allnames, e.Name) {
				allnames = append(allnames, e.Name)
			}
			entryM[e.Name] = e.Value
		}
		allEntrys = append(allEntrys, entryM)
	}

	builder.WriteRune('{')
	total := len(allnames) - 1
	for i, v := range allnames {
		builder.WriteString(v)
		if i != total {
			builder.WriteRune(',')
		}
	}
	builder.WriteString("}:\n")

	for _, entrys := range allEntrys {
		total = len(entrys) - 1
		builder.WriteString("  ")
		j := 0
		total := len(allnames) - 1
		for _, name := range allnames {
			v, exists := entrys[name]
			if !exists {
				builder.WriteString("null")
			} else {
				s, err := marshalSolve(v, v.Type())
				if err != nil {
					return "", err
				}
				builder.WriteString(s)
			}
			if j != total {
				builder.WriteString(",")
			}
			j++
		}
		builder.WriteByte('\n')
	}

	return builder.String(), nil

}
