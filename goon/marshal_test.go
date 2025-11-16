package goon_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/roboogg133/goon/goon"
)

type Test1 struct {
	ID     int     `toon:"id"`
	Name   *string `toon:"name"`
	Active bool    `toon:"active"`
	Email  string  `toon:"email"`
	Score  float32 `toon:"score"`
}

type Test2 struct {
	User struct {
		ID      int    `toon:"id"`
		Name    string `toon:"name"`
		Content struct {
			Email string `toon:"email"`
			Phone string `toon:"phone"`
		} `toon:"contact"`
		Settings struct {
			Theme         string `toon:"theme"`
			Notifications bool   `toon:"notifications"`
		} `toon:"settings"`
	} `toon:"user"`
}

type Test3 struct {
	Items []any `toon:"items"`
}

type Test4 struct {
	Tags    []string `toon:"tags"`
	Numbers []int    `toon:"numbers"`
	Empty   []string `toon:"empty"`
}

func TestMarshal(t *testing.T) {

	fd := "Ada Lovelace"
	test1 := Test1{
		ID:     123,
		Name:   &fd,
		Active: true,
		Email:  "ada@example.com",
		Score:  98.5,
	}

	test1map := make(map[string]any)

	test1map["id"] = 123
	test1map["name"] = "Ada Lovelace"
	test1map["active"] = true
	test1map["email"] = "ada@example.com"
	test1map["score"] = 98.5

	var test2 Test2

	test2.User.ID = 123
	test2.User.Name = "Ada Lovelace"

	test2.User.Content.Email = "ada@example.com"
	test2.User.Content.Phone = "+1-555-0100"

	test2.User.Settings.Theme = "dark"
	test2.User.Settings.Notifications = true

	var test3 Test3

	temp := make(map[string]string)
	temp["a"] = "hello"
	temp["b"] = "world"

	test3.Items = append(test3.Items, 1)
	test3.Items = append(test3.Items, temp)
	test3.Items = append(test3.Items, "text value")

	test4 := Test4{
		Tags:    []string{"admin", "ops", "dev"},
		Numbers: []int{1, 2, 3, 4, 5},
		Empty:   []string{},
	}

	test5 := make(map[string][]map[string]any)

	temp2 := make(map[string]any)
	temp2["name"] = "tamanho máximo"
	temp2["age"] = 555

	temp1 := make(map[string]any)

	temp1["name"] = "tamanlego   áximo"
	temp1["age"] = 12
	temp1["size"] = 500

	test5["users"] = []map[string]any{temp2, temp1}

	test6 := []string{"oi", "fsdafasdf", "ksç"}

	t.Run("objects", func(t *testing.T) {
		a, err := goon.Marshal(test1)
		if err != nil {
			t.Error(err)
		}
		os.WriteFile("object.toon", a, 0777)
		fmt.Println(string(a))
	})

	t.Run("objects-map", func(t *testing.T) {
		a, err := goon.Marshal(test1map)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(string(a))
	})

	t.Run("nested object", func(t *testing.T) {
		a, err := goon.Marshal(test2)
		if err != nil {
			t.Error(err)
		}
		os.WriteFile("nested-object.toon", a, 0777)
		fmt.Println(string(a))
	})

	t.Run("mixed array", func(t *testing.T) {
		a, err := goon.Marshal(test3)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(string(a))
	})

	t.Run("array", func(t *testing.T) {
		a, err := goon.Marshal(test4)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(string(a))
	})

	t.Run("slice of map/struct", func(t *testing.T) {
		a, err := goon.Marshal(test5)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(string(a))
	})

	t.Run("just an slice", func(t *testing.T) {
		a, err := goon.Marshal(test6)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(string(a))
	})

}
