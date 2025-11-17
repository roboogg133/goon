package goon_test

import (
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
}

func TestUnmarshal(t *testing.T) {

	data, _ := os.ReadFile("./object.toon")
	t.Run("Unmarshal", func(t *testing.T) {
		var obj Object
		err := goon.Unmarshal(data, &obj)
		if err != nil {
			t.Errorf("Unmarshal failed: %v", err)
		}
	})

	meumapa := make(map[string]any)

	t.Run("Unmarshal_map", func(t *testing.T) {
		err := goon.Unmarshal(data, &meumapa)
		if err != nil {
			t.Errorf("Unmarshal failed: %v", err)
		}
	})
}
