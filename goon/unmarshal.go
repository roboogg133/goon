package goon

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type posStruct struct {
	Name string
	Pos  int
}

func Unmarshal(data []byte, v any) error {

	rv := reflect.ValueOf(v)
	kind := rv.Type().Kind()

	if kind != reflect.Ptr || rv.IsNil() {
		return errors.New("goon: v must be a non-nil pointer")
	}

	rv = rv.Elem()
	kind = rv.Kind()

	structMap := make(map[string]posStruct)

	switch kind {
	case reflect.Struct:
		for i := 0; i < rv.NumField(); i++ {
			field := rv.Type().Field(i)
			structMap[field.Tag.Get("toon")] = posStruct{
				Name: field.Name,
				Pos:  i,
			}
		}
	}
	reader := bytes.NewReader(data)
	scanner := bufio.NewScanner(reader)

	r, _ := regexp.Compile(`^.*?(?=\[\s*([1-9]\d*)\s*\])`)

	for scanner.Scan() {
		text := scanner.Text()

		strDoubleDot := strings.SplitN(text, ":", 2)

		// if is true is a mixed list
		if r.MatchString(strings.TrimSpace(strDoubleDot[0])) && strings.TrimSpace(strDoubleDot[1]) == "" {

			//if true is a csv-style
		} else if strings.ContainsAny(strings.TrimSpace(strDoubleDot[0]), "[]") && strings.ContainsAny(strings.TrimSpace(strDoubleDot[0]), "{}") {

		}

		posVal, err := recognizeType(strings.TrimSpace(strDoubleDot[1]))
		if err != nil {
			return err
		}

		if !posVal.IsValid() {
			continue
		}

		switch kind {
		case reflect.Struct:
			a := structMap[strings.TrimSpace(strings.Split(strDoubleDot[0], "[")[0])]

			if err := signToStruct(rv, strings.TrimSpace(strDoubleDot[1]), a); err != nil {
				return err
			}
		case reflect.Map:
			rv.SetMapIndex(reflect.ValueOf(strings.TrimSpace(strDoubleDot[0])), posVal)
		}
		continue

	}

	return nil
}

func signToStruct(rv reflect.Value, rawValue string, a posStruct) error {

	posVal, err := recognizeType(rawValue)
	if err != nil {
		return err
	}
	kind := rv.Kind()

	if kind == reflect.Ptr {
		rv = rv.Elem()
		kind = rv.Kind()
	}

	field := rv.Field(a.Pos)
	fieldKind := field.Kind()

	if fieldKind == reflect.Ptr {
		field = field.Elem()
		fieldKind = field.Kind()
	}

	if field.Kind() != posVal.Kind() && fieldKind != reflect.Interface {
		return fmt.Errorf("goon: trying to assign %s to %s", posVal.Kind(), fieldKind)
	}

	field.Set(posVal)

	return nil
}
