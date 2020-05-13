package utils_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/mkuznets/classbox/pkg/utils"
)

type Foo struct {
	TextField string
}

func TestUniqueStrings(t *testing.T) {
	data := []Foo{{"z"}, {"z"}, {"y"}, {"x"}, {"x"}}
	unique := utils.UniqueStringFields(data, "TextField")
	expected := []string{"z", "y", "x"}
	if !reflect.DeepEqual(unique, expected) {
		t.Fatalf("expected %v, got %v", expected, unique)
	}
}

func TestMapStringKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := utils.MapStringKeys(m)
	sort.StringSlice(keys).Sort()
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(keys, expected) {
		t.Fatalf("expected %v, got %v", expected, keys)
	}
}
