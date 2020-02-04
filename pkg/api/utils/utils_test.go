package utils_test

import (
	"github.com/mkuznets/classbox/pkg/api/utils"
	"reflect"
	"testing"
)

type Foo struct {
	TextField string
}

func TestUniqueStrings(t *testing.T) {
	data := []Foo{{"z"}, {"z"}, {"y"}, {"x"}, {"x"}}
	unique := utils.UniqueStrings(data, "TextField")
	expected := []string{"z", "y", "x"}
	if !reflect.DeepEqual(unique, expected) {
		t.Fatalf("expected %v, got %v", expected, unique)
	}
}
