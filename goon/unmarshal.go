package goon

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
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
				Name: field.Tag.Get("toon"),
				Pos:  i,
			}
		}
	}
	reader := bytes.NewReader(data)
	scanner := bufio.NewScanner(reader)

	r, err := regexp.Compile(`^(.*?)\[\s*([1-9]\d*)\s*\]`)
	if err != nil {
		return err
	}

	for scanner.Scan() {
		text := scanner.Text()

		strDoubleDot := strings.SplitN(text, ":", 2)

		// if is true is a mixed list
		if r.MatchString(strings.TrimSpace(strDoubleDot[0])) && strings.TrimSpace(strDoubleDot[1]) == "" {
			matches := r.FindStringSubmatch(strings.TrimSpace(strDoubleDot[0]))

			listLenght, err := strconv.Atoi(matches[2])
			if err != nil {
				return fmt.Errorf("goon: error parsing mixed list -> %s ", err.Error())
			}

			slice, err := multipleLineList(scanner, listLenght)
			if err != nil {
				return err
			}

			switch kind {
			case reflect.Struct:
				a := structMap[strings.TrimSpace(strings.Split(strDoubleDot[0], "[")[0])]

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

				if field.Kind() != slice.Kind() && fieldKind != reflect.Interface {
					return fmt.Errorf("goon: trying to assign %s to %s", slice.Kind(), fieldKind)
				}

				field.Set(slice)
			case reflect.Map:
				rv.SetMapIndex(reflect.ValueOf(strings.TrimSpace(strings.Split(strDoubleDot[0], "[")[0])), slice)
			}

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

func multipleLineList(scanner *bufio.Scanner, listLength int) (reflect.Value, error) {

	var elems []reflect.Value

	for i := 0; i < listLength; i++ {
		scanner.Scan()
		text := scanner.Text()
		trimmedLine := strings.TrimSpace(text)

		if !strings.HasPrefix(trimmedLine, "-") {
			i--
			continue
		}

		splited := strings.SplitN(trimmedLine, " ", 2)

		value, err := recognizeType(strings.TrimSpace(splited[1]))
		if err != nil {
			return reflect.ValueOf(nil), nil
		}

		elems = append(elems, value)
	}

	sliceType := reflect.SliceOf(reflect.TypeFor[any]())

	sliceValue := reflect.MakeSlice(sliceType, 0, len(elems))
	for _, e := range elems {
		sliceValue = reflect.Append(sliceValue, e)
	}

	return sliceValue, nil
}

func csvLike(scanner *bufio.Scanner, listLength int, orderList []string) ([]map[string]reflect.Value, error) {

	var maplists []map[string]reflect.Value

	for _ = range listLength {
		scanner.Scan()
		text := strings.TrimSpace(scanner.Text())

		splited := strings.Split(text, ",")

		maps := make(map[string]reflect.Value)
		for j, v := range splited {
			var err error
			maps[orderList[j]], err = recognizeType(v)
			if err != nil {
				return nil, err
			}
		}
		maplists = append(maplists, maps)
	}

	return maplists, nil
}
