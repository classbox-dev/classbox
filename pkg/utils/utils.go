package utils

import (
	"math/rand"
	"reflect"
	"strings"
	"time"
)

const (
	alphanum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func UniqueStringFields(v interface{}, field string) []string {
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

func RandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteByte(alphanum[rand.Intn(len(alphanum))])
	}
	return b.String()
}
