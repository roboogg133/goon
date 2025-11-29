package goon_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/roboogg133/goon/goon"
)

type Object struct {
	Id     int     `toon:"id"`
	Name   string  `toon:"name"`
	Active bool    `toon:"active"`
	Email  string  `toon:"email"`
	Score  float64 `toon:"score"`
	Names  []any   `toon:"names"`
}

type CsvToon struct {
	Users []Person `toon:"users"`
}
type Person struct {
	Name string `toon:"name"`
	Age  int    `toon:"age"`
	Size int    `toon:"size"`
}

func TestUnmarshal(t *testing.T) {

	data, _ := os.ReadFile("./object.toon")
	t.Run("Unmarshal", func(t *testing.T) {
		var obj Object
		err := goon.Unmarshal(data, &obj)
		if err != nil {
			t.Errorf("Unmarshal failed: %v", err)
		}
		returned := fmt.Sprintf("id : %d\nname : %s\nactive: %t\nemail : %s\nscore: %f\nnames: %v\n", obj.Id, obj.Name, obj.Active, obj.Email, obj.Score, obj.Names)
		if returned != `id : 123
name : Ada Lovelace
active: true
email : ada@example.com
score: 98.500000
names: [hello testing here]
` {
			t.Fail()
		}
	})

	meumapa := make(map[string]any)

	t.Run("Unmarshal_map", func(t *testing.T) {
		err := goon.Unmarshal(data, &meumapa)
		if err != nil {
			t.Errorf("Unmarshal failed: %v", err)
		}
		fmt.Println(meumapa)
	})

	t.Run("UnmarshalMixedList_Map", func(t *testing.T) {
		data, _ := os.ReadFile("./mixedList.toon")
		meumapa := make(map[string]any)
		err := goon.Unmarshal(data, &meumapa)
		if err != nil {
			t.Errorf("Unmarshal failed: %v", err)
		}
	})

	t.Run("slice of map", func(t *testing.T) {
		data, _ := os.ReadFile("./tooncsv.toon")

		sliceof := make(map[string]any)
		err := goon.Unmarshal(data, &sliceof)
		if err != nil {
			t.Errorf("Unmarshal failed: %v", err)
		}

		for i, v := range sliceof {
			fmt.Printf("%s :\n", i)
			v := v.([]map[string]any)

			for _, v2 := range v {
				fmt.Println("{")
				for j2, v3 := range v2 {
					fmt.Printf("  %v : %v\n", j2, v3)
				}
				fmt.Println("}")
			}

		}
	})

	t.Run("slice of struct", func(t *testing.T) {
		data, _ := os.ReadFile("./tooncsv.toon")

		var test CsvToon
		err := goon.Unmarshal(data, &test)
		if err != nil {
			t.Errorf("Unmarshal failed: %v", err)
		}

		fmt.Println(test)

	})

	t.Run("nested struct", func(t *testing.T) {
		data, _ := os.ReadFile("./nested-object.toon")

		test := make(map[string]any)
		err := goon.Unmarshal(data, &test)
		if err != nil {
			t.Errorf("Unmarshal failed: %v", err)
		}

		fmt.Println(test)

	})

}
