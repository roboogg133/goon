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

const IndentationRune = ' '

func Unmarshal(data []byte, v any) error {

	rv := reflect.ValueOf(v)
	kind := rv.Type().Kind()

	if kind != reflect.Pointer || rv.IsNil() {
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

	r, _ := regexp.Compile(`^(.*?)\[\s*([1-9]\d*)\s*\]`)
	csvl, _ := regexp.Compile(`^(.*?)\[\s*([1-9]\d*)([|\t]?)\s*\]\{([^}]+)\}`)

	var lastIdentationN int
	var saveName string
	var builder4Nested strings.Builder
	for scanner.Scan() {
		text := scanner.Text()

		strDoubleDot := strings.SplitN(text, ":", 2)
		if strings.TrimSpace(strDoubleDot[1]) == "" {
			saveName = strings.TrimSpace(strDoubleDot[0])
		}
		s := strings.Replace(strDoubleDot[0], strings.TrimSpace(strDoubleDot[0]), "", 1)
		indentationN := strings.Count(s, string(IndentationRune))
		if indentationN > lastIdentationN {
			lastIdentationN = indentationN
			builder4Nested.WriteString(text)
			continue
		}

		// if is true is a csv like list
		if csvl.MatchString(strings.TrimSpace(strDoubleDot[0])) && strings.TrimSpace(strDoubleDot[1]) == "" {
			matches := csvl.FindStringSubmatch(strings.TrimSpace(strDoubleDot[0]))
			if matches[3] == "" {
				matches[3] = ","
			}
			length, err := strconv.Atoi(matches[2])
			if err != nil {
				return err
			}
			sliceObject, err := csvLike(scanner, length, strings.Split(matches[4], matches[3]), matches[3])
			if err != nil {
				return err
			}

			switch kind {
			case reflect.Struct:
				a, exists := structMap[matches[1]]
				if !exists {
					continue
				}
				field := rv.Field(a.Pos)

				if field.Kind() != reflect.Slice {
					return fmt.Errorf("invalid type can only unmarshal a csv like object on slices")
				}

				sliceType := field.Type().Elem()
				switch sliceType.Kind() {
				case reflect.Struct:

					elemType := field.Type().Elem()

					innerMap := make(map[string]posStruct)

					for i := 0; i < elemType.NumField(); i++ {
						var a posStruct
						a.Name = elemType.Field(i).Tag.Get("toon")

						if a.Name == "" {
							continue
						}
						a.Pos = i
						innerMap[a.Name] = a
					}

					newSlice := reflect.MakeSlice(field.Type(), 0, len(sliceObject))

					for _, structs := range sliceObject {
						newElement := reflect.New(elemType).Elem()
						for j, v := range structs {
							b, exists := innerMap[j]
							if !exists {
								continue
							}
							if !v.IsValid() {
								continue
							}
							newElement.Field(b.Pos).Set(v)
						}
						newSlice = reflect.Append(newSlice, newElement)
					}

					field.Set(newSlice)
				case reflect.Map:
					var result []map[string]any

					new := reflect.MakeSlice(reflect.TypeOf(result), len(sliceObject), len(sliceObject))

					for i, v := range sliceObject {
						temp := make(map[string]any)
						for i, v2 := range v {
							if !v2.IsValid() {
								continue
							}
							temp[i] = v2.Interface()
						}
						result = append(result, temp)
						new.Index(i).Set(reflect.ValueOf(temp))
					}

					field.Set(new)
				}
				continue
			case reflect.Map:

				var result []map[string]any

				new := reflect.MakeSlice(reflect.TypeOf(result), len(sliceObject), len(sliceObject))

				for i, v := range sliceObject {
					temp := make(map[string]any)
					for i, v2 := range v {
						if !v2.IsValid() {
							continue
						}
						temp[i] = v2.Interface()
					}
					result = append(result, temp)
					new.Index(i).Set(reflect.ValueOf(temp))
				}

				rv.SetMapIndex(reflect.ValueOf(matches[1]), new)
			}
			continue
		} else if r.MatchString(strings.TrimSpace(strDoubleDot[0])) && strings.TrimSpace(strDoubleDot[1]) == "" {
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

				if kind == reflect.Pointer {
					rv = rv.Elem()
					kind = rv.Kind()
				}

				field := rv.Field(a.Pos)
				fieldKind := field.Kind()

				if fieldKind == reflect.Pointer {
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
			continue
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

			if indentationN < lastIdentationN {
				if kind == reflect.Pointer {
					rv = rv.Elem()
					kind = rv.Kind()
				}
				if err := Unmarshal([]byte(builder4Nested.String()), &rv); err != nil {
					return err
				}
				continue
			}

			if err := signToStruct(rv, strings.TrimSpace(strDoubleDot[1]), a); err != nil {
				return err
			}

		case reflect.Map:
			fmt.Printf("now : %d before: %d\n", indentationN, lastIdentationN)
			if indentationN < lastIdentationN {
				if kind == reflect.Pointer {
					rv = rv.Elem()
					kind = rv.Kind()
				}
				fmt.Println("giving dest: ", saveName)
				dest := rv.MapIndex(reflect.ValueOf(saveName))
				if err := Unmarshal([]byte(builder4Nested.String()), &dest); err != nil {
					return err
				}
				continue
			}

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

	if kind == reflect.Pointer {
		rv = rv.Elem()
		kind = rv.Kind()
	}

	field := rv.Field(a.Pos)
	fieldKind := field.Kind()

	if fieldKind == reflect.Pointer {
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
		trimmedLine := strings.TrimSpace(scanner.Text())

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

func csvLike(scanner *bufio.Scanner, listLength int, orderList []string, sep string) ([]map[string]reflect.Value, error) {

	var maplists []map[string]reflect.Value
	for _ = range listLength {
		scanner.Scan()
		text := strings.TrimSpace(scanner.Text())

		splited := strings.Split(text, sep)

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
