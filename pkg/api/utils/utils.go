package utils

import (
	"reflect"
)

func UniqueStrings(v interface{}, field string) []string {
	hm := map[string]struct{}{}
	values := make([]string, 0)

	sl := reflect.ValueOf(v)
	for i := 0; i < sl.Len(); i++ {
		item := reflect.Indirect(sl.Index(i)).FieldByName(field).String()
		if _, ok := hm[item]; !ok {
			values = append(values, item)
			hm[item] = struct{}{}
		}
	}
	return values
}

func MapStringKeys(v interface{}) []string {
	m := reflect.ValueOf(v)
	ks := make([]string, 0, m.Len())
	iter := m.MapRange()
	for iter.Next() {
		ks = append(ks, iter.Key().String())
	}
	return ks
}
